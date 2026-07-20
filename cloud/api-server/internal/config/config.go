package config

import (
	"os"
	"strconv"
	"strings"
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
	FCMKeyPath   string // path to service account JSON for FCM OAuth2

	SMSAccessKey   string
	SMSAccessSecret string
	SMSSignName     string
	SMPTemplateID   string

	Port          string
	CORSOrigins   []string // comma-separated allowed origins
	BodyLimitMB   int      // max request body in MB
	DeviceSecret  string   // HMAC key for device tokens
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	corsStr := getEnv("CORS_ORIGINS", "")
	var corsOrigins []string
	if corsStr != "" {
		for _, o := range strings.Split(corsStr, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				corsOrigins = append(corsOrigins, o)
			}
		}
	}

	return &Config{
		// JWT_SECRET must be set in production — no fallback allowed
		JWTSecret:     getEnv("JWT_SECRET", ""),
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
		FCMKeyPath:   getEnv("FCM_KEY_PATH", ""),

		SMSAccessKey:    getEnv("SMS_ACCESS_KEY", ""),
		SMSAccessSecret: getEnv("SMS_ACCESS_SECRET", ""),
		SMSSignName:     getEnv("SMS_SIGN_NAME", "颐贞"),
		SMPTemplateID:   getEnv("SMS_TEMPLATE_ID", "SMS_XXXXXXXX"),

		Port:          getEnv("PORT", "8080"),
		CORSOrigins:   corsOrigins,
		BodyLimitMB:   getEnvAsInt("BODY_LIMIT_MB", 1),
		DeviceSecret:  getEnv("DEVICE_SECRET", ""),
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
