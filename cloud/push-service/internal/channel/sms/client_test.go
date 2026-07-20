package channel

import (
	"net/url"
	"testing"
	"time"
)

func TestSMSClient_AllowSendInitial(t *testing.T) {
	c := NewSMSClient("key", "secret", "sign")
	if !c.allowSend(10) {
		t.Error("first send should be allowed")
	}
}

func TestSMSClient_RateLimitExceeded(t *testing.T) {
	c := NewSMSClient("key", "secret", "sign")
	// Pre-fill counter beyond limit
	c.mu.Lock()
	c.dailyCount = 10
	c.lastSendAt = time.Now()
	c.mu.Unlock()

	if c.allowSend(10) {
		t.Error("should reject when daily limit reached")
	}
}

func TestSMSClient_DailyReset(t *testing.T) {
	c := NewSMSClient("key", "secret", "sign")
	c.mu.Lock()
	c.dailyCount = 10
	c.lastSendAt = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	c.mu.Unlock()

	if !c.allowSend(10) {
		t.Error("should allow after day rollover")
	}
}

func TestSMSClient_SendWithoutConfigSkips(t *testing.T) {
	c := NewSMSClient("", "", "")
	err := c.send("+8613800138000", "TEST_TPL", "hello")
	if err != nil {
		t.Fatalf("send without config should skip, got: %v", err)
	}
}

func TestSMSClient_SignRequestFormat(t *testing.T) {
	c := NewSMSClient("key", "secret", "sign")
	params := map[string]string{
		"PhoneNumbers": "+8613800138000",
		"RegionId":     "cn-shanghai",
	}
	sig := c.signRequest(paramsToValues(params), "POST", "/")
	if sig == "" {
		t.Error("signature should not be empty")
	}
	if len(sig) < 20 {
		t.Errorf("signature too short: %s", sig)
	}
}

func paramsToValues(m map[string]string) url.Values {
	v := make(url.Values)
	for k, val := range m {
		v.Set(k, val)
	}
	return v
}
