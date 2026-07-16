package channel

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// SMSClient wraps 阿里云SMS API for sending alert/reminder messages.
type SMSClient struct {
	accessKey    string
	accessSecret string
	signName     string

	mu         sync.Mutex
	lastSendAt time.Time
	dailyCount int
}

// NewSMSClient creates an SMS client. All fields must be set for production use.
func NewSMSClient(accessKey, accessSecret, signName string) *SMSClient {
	return &SMSClient{
		accessKey:    accessKey,
		accessSecret: accessSecret,
		signName:     signName,
	}
}

// SendAlert sends an alert SMS with rate limiting (max 10/day).
func (c *SMSClient) SendAlert(phone, message string) error {
	if !c.allowSend(10) {
		log.Printf("[sms] rate limit exceeded for %s today", phone)
		return fmt.Errorf("sms rate limit reached")
	}
	return c.send(phone, "SMS_HEALTH_ALERT", message)
}

// SendReminder sends a medication reminder SMS (max 50/day).
func (c *SMSClient) SendReminder(phone, message string) error {
	if !c.allowSend(50) {
		log.Printf("[sms] daily reminder limit reached for %s", phone)
		return fmt.Errorf("sms daily limit reached")
	}
	return c.send(phone, "SMS_MED_REMIND", message)
}

func (c *SMSClient) allowSend(maxPerDay int) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	if now.Day() != c.lastSendAt.Day() || now.Month() != c.lastSendAt.Month() {
		c.dailyCount = 0
	}
	return c.dailyCount < maxPerDay
}

func (c *SMSClient) send(phone, tmplID, message string) error {
	if c.accessKey == "" || c.accessSecret == "" {
		log.Printf("[sms] SKIP: not configured (phone=%s msg=%s)", phone, message)
		return nil
	}

	phoneNum := phone
	if !strings.HasPrefix(phoneNum, "+") {
		phoneNum = "+86" + phoneNum
	}

	params := map[string]string{"message": message}
	paramJSON, _ := json.Marshal(params)

	apiURL := fmt.Sprintf(
		"https://dysmsapi.aliyuncs.com/?PhoneNumbers=%s&SignName=%s&TemplateCode=%s&TemplateParam=%s&RegionId=cn-shanghai",
		url.QueryEscape(phoneNum),
		url.QueryEscape(c.signName),
		url.QueryEscape(tmplID),
		url.QueryEscape(string(paramJSON)),
	)

	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("sms request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Code      string `json:"Code"`
		Message   string `json:"Message"`
		RequestID string `json:"RequestId"`
	}
	json.Unmarshal(body, &result)

	if result.Code != "OK" {
		return fmt.Errorf("sms error: %s (%s)", result.Message, result.Code)
	}

	log.Printf("[sms] sent to %s: %s", phoneNum, message)
	return nil
}
