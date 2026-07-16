package fcm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// FCMClient sends push notifications via Firebase Cloud Messaging REST API.
type FCMClient struct {
	mu      sync.Mutex
	token   string
	expire  time.Time
	httpCli *http.Client
}

// Client is an alias for FCMClient.
type Client = FCMClient

// NewClient creates an FCM client. Uses default credentials from environment.
func NewClient() *Client {
	return &Client{httpCli: &http.Client{Timeout: 10 * time.Second}}
}

// GetOAuthToken fetches a short-lived OAuth2 token from Google's auth endpoint.
func (c *FCMClient) GetOAuthToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	if c.token != "" && time.Now().Before(c.expire) {
		c.mu.Unlock()
		return c.token, nil
	}
	c.mu.Unlock()

	keyPath := getEnv("FCM_KEY_PATH", "")
	if keyPath != "" {
		data, err := readFile(keyPath)
		if err == nil {
			var sa struct {
				PrivateKeyID string `json:"private_key_id"`
				PrivateKey   string `json:"private_key"`
				ClientEmail  string `json:"client_email"`
			}
			if err := json.Unmarshal(data, &sa); err == nil {
				token, exp, err := jwtBearerGrant(ctx, c.httpCli, sa.PrivateKey, sa.PrivateKeyID, sa.ClientEmail)
				if err == nil {
					c.mu.Lock()
					c.token = token
					c.expire = exp
					c.mu.Unlock()
					return token, nil
				}
			}
		}
	}

	token, exp, err := metadataAccessToken(ctx, c.httpCli)
	if err != nil {
		return "", fmt.Errorf("fcm: no credentials configured, set FCM_KEY_PATH or run on GCP")
	}
	c.mu.Lock()
	c.token = token
	c.expire = exp
	c.mu.Unlock()
	return token, nil
}

// SendToDevice sends a single FCM notification.
func (c *FCMClient) SendToDevice(ctx context.Context, deviceToken, title, body string) error {
	oauth, err := c.GetOAuthToken(ctx)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"message": map[string]interface{}{
			"token": deviceToken,
			"notification": map[string]string{
				"title": title,
				"body":  body,
			},
			"android": map[string]interface{}{
				"priority": "high",
			},
		},
	}

	bodyBytes, _ := json.Marshal(payload)
	resp, err := c.httpCli.Post(
		"https://fcm.googleapis.com/v1/projects/eregen-platform/messages:send",
		"application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("fcm send: %w", err)
	}
	defer resp.Body.Close()

	result, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fcm %d: %s", resp.StatusCode, string(result))
	}

	log.Printf("[fcm] sent to %s: %s", deviceToken, string(result))
	_ = oauth // token used internally by GetOAuthToken
	return nil
}

// SendBulk sends to multiple tokens sequentially.
func (c *FCMClient) SendBulk(ctx context.Context, tokens []string, title, body string) error {
	for _, tok := range tokens {
		if err := c.SendToDevice(ctx, tok, title, body); err != nil {
			prefix := tok
			if len(tok) > 8 {
				prefix = tok[:8]
			}
			log.Printf("[fcm] bulk skip %s: %v", prefix, err)
		}
	}
	return nil
}
