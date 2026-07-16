package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	NATSURL          string
	PostgresDSN      string
	InfluxDBURL      string
	InfluxDBToken    string
	InfluxDBOrg      string
	InfluxDBBucket   string
	LogLevel         string
	Port             int
	BaselineDays     int
	RiskVitalsWeight float64
	RiskMedWeight    float64
	RiskActivityWt   float64
	RiskSleepWt      float64
}

func Load() (*Config, error) {
	c := &Config{
		NATSURL:          getEnv("NATS_URL", "nats://nats:4222"),
		PostgresDSN:      getEnv("POSTGRES_DSN", "postgres://eregen:eregen@postgres:5432/eregen?sslmode=disable"),
		InfluxDBURL:      getEnv("INFLUXDB_URL", "http://influxdb:8086"),
		InfluxDBToken:    getEnv("INFLUXDB_TOKEN", "eregen_token"),
		InfluxDBOrg:      getEnv("INFLUXDB_ORG", "eregen"),
		InfluxDBBucket:   getEnv("INFLUXDB_BUCKET", "health_data"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		Port:             getIntEnv("PIPELINE_PORT", 8086),
		BaselineDays:     getIntEnv("BASELINE_DAYS", 7),
		RiskVitalsWeight: parseFloatEnv("RISK_VITALS_WEIGHT", 0.40),
		RiskMedWeight:    parseFloatEnv("RISK_MED_WEIGHT", 0.30),
		RiskActivityWt:   parseFloatEnv("RISK_ACTIVITY_WT", 0.20),
		RiskSleepWt:      parseFloatEnv("RISK_SLEEP_WT", 0.10),
	}
	return c, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func parseFloatEnv(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return f
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
