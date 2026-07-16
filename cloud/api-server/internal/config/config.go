package config

import (
	"os"
	"strconv"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	JWTSecret     string
	TokenExpiry   int // seconds
	RefreshExpiry int // seconds

	DBURL         string
	RedisURL      string
	NATSURL       string
	InfluxDBURL   string
	InfluxDBOrg   string
	InfluxDBToken string
	InfluxDBBucket string

	FCMProjectID string
	FCMServerKey string

	SMSSignName  string
	SMPTemplateID string

	Port string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		JWTSecret:     getEnv("JWT_SECRET", "eregen-dev-secret-change-in-production"),
		TokenExpiry:   getEnvAsInt("TOKEN_EXPIRY", 3600),
		RefreshExpiry: getEnvAsInt("REFRESH_EXPIRY", 604800),

		DBURL:         getEnv("DB_URL", "postgres://postgres:postgres@localhost/eregen?sslmode=disable"),
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379/0"),
		NATSURL:       getEnv("NATS_URL", "nats://localhost:4222"),
		InfluxDBURL:   getEnv("INFLUXDB_URL", "http://localhost:8086"),
		InfluxDBOrg:   getEnv("INFLUXDB_ORG", "eregen"),
		InfluxDBToken: getEnv("INFLUXDB_TOKEN", "eregen-token"),
		InfluxDBBucket: getEnv("INFLUXDB_BUCKET", "health"),

		FCMProjectID: getEnv("FCM_PROJECT_ID", ""),
		FCMServerKey: getEnv("FCM_SERVER_KEY", ""),

		SMSSignName:  getEnv("SMS_SIGN_NAME", "颐贞"),
		SMPTemplateID: getEnv("SMS_TEMPLATE_ID", "SMS_XXXXXXXX"),

		Port: getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
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
