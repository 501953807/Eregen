package channel

import (
	"testing"
	"time"
)

func TestWeChatClient_GetAccessTokenCache(t *testing.T) {
	c := NewWeChatClient("app-id", "app-secret")
	// Pre-populate cache with valid token
	c.mu.Lock()
	c.token = "cached-token"
	c.expireAt = time.Now().Add(time.Hour)
	c.mu.Unlock()

	token, err := c.GetAccessToken()
	if err != nil {
		t.Fatalf("GetAccessToken failed: %v", err)
	}
	if token != "cached-token" {
		t.Errorf("token = %q, want cached-token", token)
	}
}

func TestWeChatDataStructure(t *testing.T) {
	data := WeChatData{Value: "test-value"}
	if data.Value != "test-value" {
		t.Error("WeChatData value not set correctly")
	}
}
