package channel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// WeChatClient sends subscription messages via WeChat API.
type WeChatClient struct {
	appID     string
	appSecret string
	mu        sync.Mutex
	token     string
	expireAt  time.Time
}

// NewWeChatClient creates a WeChat client. Requires app ID and secret.
func NewWeChatClient(appID, appSecret string) *WeChatClient {
	return &WeChatClient{appID: appID, appSecret: appSecret}
}

// GetAccessToken fetches or returns cached access_token.
func (c *WeChatClient) GetAccessToken() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Now().Before(c.expireAt) {
		return c.token, nil
	}

	url := fmt.Sprintf(
		"https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		c.appID, c.appSecret)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("wechat token request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		AccessToken string  `json:"access_token"`
		ExpiresIn   float64 `json:"expires_in"`
		ErrCode     float64 `json:"errcode"`
		ErrMsg      string  `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("wechat token parse: %w", err)
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("wechat token error: %.0f %s", result.ErrCode, result.ErrMsg)
	}

	c.token = result.AccessToken
	c.expireAt = time.Now().Add(time.Duration(result.ExpiresIn-600) * time.Second)
	return c.token, nil
}

// WeChatData represents a single template field value.
type WeChatData struct {
	Value string `json:"value"`
}

// SendTemplateMessage sends a WeChat subscription message.
func (c *WeChatClient) SendTemplateMessage(openID, templateID string, data map[string]WeChatData) error {
	token, err := c.GetAccessToken()
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"touser":      openID,
		"template_id": templateID,
		"data":        data,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal wechat msg: %w", err)
	}

	url := "https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=" + token
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("wechat send: %w", err)
	}
	defer resp.Body.Close()

	result, _ := io.ReadAll(resp.Body)
	var resultJSON struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	json.Unmarshal(result, &resultJSON)

	if resultJSON.ErrCode != 0 {
		log.Printf("[wechat] send error: %.0f %s", resultJSON.ErrCode, resultJSON.ErrMsg)
		return fmt.Errorf("wechat send error: %s", resultJSON.ErrMsg)
	}

	log.Printf("[wechat] sent to %s: %s", openID, string(result))
	return nil
}
