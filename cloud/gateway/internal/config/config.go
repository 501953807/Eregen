// © 2026 Eregen (颐贞). All rights reserved.

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
	LogLevel string        `yaml:"log_level"`
	MQTT     MQTTConfig    `yaml:"mqtt"`
	NATS     NATSConfig    `yaml:"nats"`
	Postgres PostgresConfig `yaml:"postgres"`
	Redis    RedisConfig   `yaml:"redis"`
	InfluxDB InfluxDBConfig `yaml:"influxdb"`
	Auth     AuthConfig    `yaml:"auth"`
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
	URL           string `yaml:"url"`
	JetStreamDomain string `yaml:"jetstream_domain"`
	StreamName    string `yaml:"stream_name"`
}

// PostgresConfig holds PostgreSQL connection settings.
type PostgresConfig struct {
	DSN string `yaml:"dsn"`
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// InfluxDBConfig holds InfluxDB v2 connection settings.
type InfluxDBConfig struct {
	URL      string `yaml:"url"`
	Token    string `yaml:"token"`
	Org      string `yaml:"org"`
	Bucket   string `yaml:"bucket"`
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
		return cfg
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(fmt.Sprintf("failed to parse config file: %v", err))
	}

	overrideString(&cfg.MQTT.Broker, "GATEWAY_MQTT_BROKER")
	overrideString(&cfg.NATS.URL, "GATEWAY_NATS_URL")
	overrideString(&cfg.Postgres.DSN, "GATEWAY_POSTGRES_DSN")
	overrideString(&cfg.Redis.Address, "GATEWAY_REDIS_ADDRESS")
	overrideString(&cfg.InfluxDB.URL, "GATEWAY_INFLUXDB_URL")
	overrideString(&cfg.Auth.SecretKey, "GATEWAY_AUTH_SECRET")

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
			URL:           "nats://localhost:4222",
			JetStreamDomain: "EREGEN",
			StreamName:    "DEVICE_EVENTS",
		},
		Postgres: PostgresConfig{
			DSN: "host=localhost port=5432 user=eregen password=eregen dbname=eregen sslmode=disable",
		},
		Redis: RedisConfig{
			Address: "localhost:6379",
			DB:      0,
		},
		InfluxDB: InfluxDBConfig{
			URL:    "http://localhost:8086",
			Token:  "eregen-token",
			Org:    "eregen",
			Bucket: "eregen-telemetry",
		},
		Auth: AuthConfig{
			SecretKey: "dev-secret-key-change-in-production",
			RateLimit: 100,
		},
	}
}
