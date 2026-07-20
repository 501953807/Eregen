package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear relevant env vars to test defaults
	envVars := []string{
		"PORT", "JWT_SECRET", "CORS_ORIGINS", "DB_URL", "REDIS_URL",
		"NATS_URL", "INFLUXDB_URL", "INFLUXDB_ORG", "INFLUXDB_TOKEN",
		"INFLUXDB_BUCKET", "FCM_PROJECT_ID", "FCM_KEY_PATH",
		"SMS_ACCESS_KEY", "SMS_ACCESS_SECRET", "SMS_SIGN_NAME",
		"SMS_TEMPLATE_ID", "BODY_LIMIT_MB", "DEVICE_SECRET",
	}
	for _, key := range envVars {
		os.Unsetenv(key)
	}

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.Port)
	}
	if cfg.TokenExpiry != 3600 {
		t.Errorf("TokenExpiry = %d, want 3600", cfg.TokenExpiry)
	}
	if cfg.RefreshExpiry != 604800 {
		t.Errorf("RefreshExpiry = %d, want 604800", cfg.RefreshExpiry)
	}
	if cfg.BodyLimitMB != 1 {
		t.Errorf("BodyLimitMB = %d, want 1", cfg.BodyLimitMB)
	}
	if cfg.SMSSignName != "颐贞" {
		t.Errorf("SMSSignName = %q, want 颐贞", cfg.SMSSignName)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("BODY_LIMIT_MB", "5")
	os.Setenv("CORS_ORIGINS", "http://localhost:3000,http://example.com")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("BODY_LIMIT_MB")
		os.Unsetenv("CORS_ORIGINS")
	}()

	cfg := Load()

	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want 9090", cfg.Port)
	}
	if cfg.BodyLimitMB != 5 {
		t.Errorf("BodyLimitMB = %d, want 5", cfg.BodyLimitMB)
	}
	if len(cfg.CORSOrigins) != 2 {
		t.Errorf("CORSOrigins length = %d, want 2", len(cfg.CORSOrigins))
	}
}
