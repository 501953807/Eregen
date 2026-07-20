package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_DefaultValues(t *testing.T) {
	for _, key := range []string{"NATS_URL", "POSTGRES_DSN", "INFLUXDB_URL", "INFLUXDB_TOKEN", "INFLUXDB_ORG", "INFLUXDB_BUCKET", "LOG_LEVEL", "PIPELINE_PORT", "BASELINE_DAYS", "RISK_VITALS_WEIGHT", "RISK_MED_WEIGHT", "RISK_ACTIVITY_WT", "RISK_SLEEP_WT"} {
		os.Unsetenv(key)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.NATSURL != "nats://nats:4222" {
		t.Errorf("NATSURL = %q, want nats://nats:4222", cfg.NATSURL)
	}
	if cfg.PostgresDSN == "" {
		t.Error("PostgresDSN should not be empty")
	}
	if cfg.InfluxDBBucket != "health_data" {
		t.Errorf("InfluxDBBucket = %q, want health_data", cfg.InfluxDBBucket)
	}
	if cfg.BaselineDays != 7 {
		t.Errorf("BaselineDays = %d, want 7", cfg.BaselineDays)
	}
	if cfg.RiskVitalsWeight != 0.40 {
		t.Errorf("RiskVitalsWeight = %f, want 0.40", cfg.RiskVitalsWeight)
	}
	if cfg.Port != 8086 {
		t.Errorf("Port = %d, want 8086", cfg.Port)
	}
}

func TestGetIntEnv_Valid(t *testing.T) {
	os.Setenv("TEST_INT", "30")
	defer os.Unsetenv("TEST_INT")
	got := getIntEnv("TEST_INT", 7)
	if got != 30 {
		t.Errorf("getIntEnv() = %d, want 30", got)
	}
}

func TestGetFloatEnv_Valid(t *testing.T) {
	os.Setenv("TEST_FLOAT", "0.5")
	defer os.Unsetenv("TEST_FLOAT")
	got := parseFloatEnv("TEST_FLOAT", 0.3)
	if got != 0.5 {
		t.Errorf("parseFloatEnv() = %f, want 0.5", got)
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
