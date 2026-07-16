package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	NATSURL          string
	EMQXHost         string
	EMQXPort         int
	PostgresDSN      string
	RedisAddr        string
	RedisPassword    string
	FCMKeyPath       string
	WeChatAppID      string
	WeChatAppSecret  string
	SMSAccessKey     string
	SMSAccessSecret  string
	SSignName        string
	SMHealthAlertTmpl string
	SMSMedRemindTmpl  string
	LogLevel         string
	Port             int
}

func Load() (*Config, error) {
	c := &Config{
		NATSURL:         getEnv("NATS_URL", "nats://nats:4222"),
		EMQXHost:        getEnv("EMQX_MQTT_HOST", "emqx"),
		EMQXPort:        getIntEnv("EMQX_MQTT_PORT", 1883),
		PostgresDSN:     getEnv("POSTGRES_DSN", "postgres://eregen:eregen@postgres:5432/eregen?sslmode=disable"),
		RedisAddr:       getEnv("REDIS_ADDR", "redis:6379"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		FCMKeyPath:      getEnv("FCM_KEY_PATH", ""),
		WeChatAppID:     getEnv("WECHAT_APP_ID", ""),
		WeChatAppSecret: getEnv("WECHAT_APP_SECRET", ""),
		SMSAccessKey:    getEnv("SMS_ACCESS_KEY", ""),
		SMSAccessSecret: getEnv("SMS_ACCESS_SECRET", ""),
		SSignName:       getEnv("SMS_SIGN_NAME", "颐贞"),
		SMHealthAlertTmpl: getEnv("SMS_HEALTH_ALERT_TPL", ""),
		SMSMedRemindTmpl:  getEnv("SMS_MED_REMIND_TPL", ""),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		Port:            getIntEnv("PUSH_SERVICE_PORT", 8085),
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
