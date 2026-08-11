// Package config loads gateway configuration from YAML + env overrides.
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all gateway configuration.
type Config struct {
	LogLevel   string       `yaml:"log_level"`
	MQTT       MQTTConfig   `yaml:"mqtt"`
	NATS       NATSConfig   `yaml:"nats"`
	Storage    StorageConfig `yaml:"storage"`
	Auth       AuthConfig   `yaml:"auth"`
}

// StorageConfig selects between postgres and sqlite backend.
type StorageConfig struct {
	Type   string `yaml:"type"`   // "postgres" or "sqlite"
	DSN    string `yaml:"dsn"`
	SQLite string `yaml:"sqlite"` // path to .sqlite file
}

// MQTTConfig holds EMQX connection settings.
type MQTTConfig struct {
	Broker    string        `yaml:"broker"`
	ClientID  string        `yaml:"client_id"`
	Username  string        `yaml:"username"`
	Password  string        `yaml:"password"`
	TLS       TLSConfig     `yaml:"tls"`
	KeepAlive time.Duration `yaml:"keep_alive"`
}

// TLSConfig holds TLS certificate paths.
type TLSConfig struct {
	Enabled bool   `yaml:"enabled"`
	CACert  string `yaml:"ca_cert"`
	Cert    string `yaml:"cert"`
	Key     string `yaml:"key"`
}

// NATSConfig holds NATS JetStream settings.
type NATSConfig struct {
	URL             string `yaml:"url"`
	JetStreamDomain string `yaml:"jetstream_domain"`
	StreamName      string `yaml:"stream_name"`
}

// AuthConfig holds authentication settings.
type AuthConfig struct {
	SecretKey string `yaml:"secret_key"`
	RateLimit int    `yaml:"rate_limit"`
}

// Load reads configuration from YAML file and environment variable overrides.
func Load() Config {
	cfg := defaultConfig()

	data, err := os.ReadFile("./config/gateway.yaml")
	if err != nil {
		// Config file is optional; use defaults but will validate critical values below
	} else if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(fmt.Sprintf("failed to parse config file: %v", err))
	}

	// Override from environment variables (optional overrides)
	overrideString(&cfg.MQTT.Broker, "GATEWAY_MQTT_BROKER")
	overrideString(&cfg.NATS.URL, "GATEWAY_NATS_URL")
	overrideString(&cfg.Storage.Type, "DATABASE_TYPE")
	overrideString(&cfg.Storage.DSN, "POSTGRES_DSN")
	overrideString(&cfg.Storage.SQLite, "SQLITE_PATH")

	// CRITICAL: Auth secret MUST be set via environment variable — no default allowed
	authSecret := os.Getenv("GATEWAY_AUTH_SECRET")
	if authSecret == "" {
		panic("GATEWAY_AUTH_SECRET environment variable is required. JWT secret cannot be left at development default.")
	}
	cfg.Auth.SecretKey = authSecret

	// Final validation check
	if cfg.Auth.SecretKey == "dev-secret-key-change-in-production" || len(cfg.Auth.SecretKey) < 32 {
		panic("invalid or missing JWT secret key — must be a strong, randomly-generated value of at least 32 characters")
	}

	return cfg
}

func overrideString(s *string, env string) {
	if v := os.Getenv(env); v != "" {
		*s = v
	}
}

func defaultConfig() Config {
	return Config{
		LogLevel: "info",
		MQTT: MQTTConfig{
			Broker:    "tcp://localhost:1883",
			ClientID:  "gateway-1",
			Username:  "eregen",
			Password:  "eregen_password",
			KeepAlive: 60 * time.Second,
			TLS: TLSConfig{
				Enabled: false,
			},
		},
		NATS: NATSConfig{
			URL:             "nats://localhost:4222",
			JetStreamDomain: "EREGEN",
			StreamName:      "DEVICE_EVENTS",
		},
		Storage: StorageConfig{
			Type:   "sqlite",
			DSN:    "host=localhost port=5432 user=eregen password=eregen dbname=eregen sslmode=disable",
			SQLite: "./data/eregen.db",
		},
		Auth: AuthConfig{
			SecretKey: "dev-secret-key-change-in-production",
			RateLimit: 100,
		},
	}
}
