package channel

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

// signRequest builds the HMAC-SHA256 signature required by 阿里云SMS API.
func (c *SMSClient) signRequest(params url.Values, method, canonicalURI string) string {
	stringToSign := method + "\n" +
		"\n" + // Accept
		"\n" + // ContentType
		time.Now().UTC().Format(http.TimeFormat) + "\n" +
		canonicalURI + "?" + params.Encode()

	h := hmac.New(sha256.New, []byte(c.accessSecret))
	h.Write([]byte(stringToSign))
	return "HMAC-SHA256 Signature=" + hex.EncodeToString(h.Sum(nil))
}

func (c *SMSClient) send(phone, tmplID, message string) error {
	if c.accessKey == "" || c.accessSecret == "" {
		log.Printf("[sms] SKIP: not configured (phone=%s msg=%s)", phone, message)
		return nil
	}

	phoneNum := phone
	if len(phoneNum) > 0 && phoneNum[0] != '+' {
		phoneNum = "+86" + phoneNum
	}

	params := url.Values{}
	params.Set("PhoneNumbers", phoneNum)
	params.Set("RegionId", "cn-shanghai")
	params.Set("SignName", c.signName)
	params.Set("TemplateCode", tmplID)
	params.Set("TemplateParam", `{"message":"`+message+`"}`)

	canonicalURI := "/"
	signature := c.signRequest(params, "POST", canonicalURI)

	apiURL := "https://dysmsapi.aliyuncs.com/"

	req, _ := http.NewRequest("POST", apiURL, strings.NewReader(params.Encode()))
	req.Header.Set("Authorization", signature)
	req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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

	c.mu.Lock()
	c.lastSendAt = time.Now()
	c.dailyCount++
	c.mu.Unlock()

	log.Printf("[sms] sent to %s: %s", phoneNum, message)
	return nil
}
