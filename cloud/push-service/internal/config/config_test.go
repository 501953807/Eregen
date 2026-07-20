package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_DefaultValues(t *testing.T) {
	// Clear relevant env vars
	for _, key := range []string{"NATS_URL", "EMQX_MQTT_HOST", "EMQX_MQTT_PORT", "POSTGRES_DSN", "REDIS_ADDR", "REDIS_PASSWORD", "FCM_KEY_PATH", "WECHAT_APP_ID", "WECHAT_APP_SECRET", "SMS_ACCESS_KEY", "SMS_ACCESS_SECRET", "SMS_SIGN_NAME", "SMS_HEALTH_ALERT_TPL", "SMS_MED_REMIND_TPL", "LOG_LEVEL", "PUSH_SERVICE_PORT"} {
		os.Unsetenv(key)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.NATSURL != "nats://nats:4222" {
		t.Errorf("NATSURL = %q, want nats://nats:4222", cfg.NATSURL)
	}
	if cfg.EMQXHost != "emqx" {
		t.Errorf("EMQXHost = %q, want emqx", cfg.EMQXHost)
	}
	if cfg.EMQXPort != 1883 {
		t.Errorf("EMQXPort = %d, want 1883", cfg.EMQXPort)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want info", cfg.LogLevel)
	}
	if cfg.Port != 8085 {
		t.Errorf("Port = %d, want 8085", cfg.Port)
	}
	if cfg.SSignName != "颐贞" {
		t.Errorf("SSignName = %q, want 颐贞", cfg.SSignName)
	}
}

func TestGetIntEnv_Valid(t *testing.T) {
	os.Setenv("TEST_INT", "8080")
	defer os.Unsetenv("TEST_INT")
	got := getIntEnv("TEST_INT", 9090)
	if got != 8080 {
		t.Errorf("getIntEnv() = %d, want 8080", got)
	}
}

func TestGetIntEnv_Invalid(t *testing.T) {
	os.Setenv("TEST_INT_BAD", "not-a-number")
	defer os.Unsetenv("TEST_INT_BAD")
	got := getIntEnv("TEST_INT_BAD", 9090)
	if got != 9090 {
		t.Errorf("getIntEnv() = %d, want fallback 9090", got)
	}
}

func TestGetDurationEnv_Valid(t *testing.T) {
	os.Setenv("TEST_DUR", "5m")
	defer os.Unsetenv("TEST_DUR")
	got := getDurationEnv("TEST_DUR", 10*time.Minute)
	if got != 5*time.Minute {
		t.Errorf("getDurationEnv() = %v, want 5m", got)
	}
}

func TestGetDurationEnv_Invalid(t *testing.T) {
	os.Setenv("TEST_DUR_BAD", "invalid")
	defer os.Unsetenv("TEST_DUR_BAD")
	got := getDurationEnv("TEST_DUR_BAD", 10*time.Minute)
	if got != 10*time.Minute {
		t.Errorf("getDurationEnv() = %v, want fallback 10m", got)
	}
}
