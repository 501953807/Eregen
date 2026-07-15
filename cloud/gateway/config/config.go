// © 2026 Eregen (颐贞). All rights reserved.

// Package config provides configuration loading for the MQTT gateway.
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all gateway configuration.
type Config struct {
	ServerPort int       `yaml:"server_port"`
	MQTT       MQTTConfig `yaml:"mqtt"`
	NATS       NATSConfig `yaml:"nats"`
	Auth       AuthConfig `yaml:"auth"`
}

// MQTTConfig holds EMQX connection settings.
type MQTTConfig struct {
	Broker   string        `yaml:"broker"`
	ClientID string        `yaml:"client_id"`
	Username string        `yaml:"username"`
	Password string        `yaml:"password"`
	TLS      TLSConfig     `yaml:"tls"`
	KeepAlive time.Duration `yaml:"keep_alive"`
}

// TLSConfig holds TLS certificate paths.
type TLSConfig struct {
	Enabled bool   `yaml:"enabled"`
	CACert  string `yaml:"ca_cert"`
	Cert    string `yaml:"cert"`
	Key     string `yaml:"key"`
}

// NATSConfig holds NATS connection settings.
type NATSConfig struct {
	URL string `yaml:"url"`
}

// AuthConfig holds authentication settings.
type AuthConfig struct {
	SecretKey string `yaml:"secret_key"`
}

// Load reads configuration from YAML file and environment variable overrides.
func Load() Config {
	cfg := defaultConfig()

	data, err := os.ReadFile("./config/gateway.yaml")
	if err != nil {
		// No config file; use defaults.
		return cfg
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(fmt.Sprintf("failed to parse config file: %v", err))
	}

	// Environment variable overrides.
	if v := os.Getenv("GATEWAY_MQTT_BROKER"); v != "" {
		cfg.MQTT.Broker = v
	}
	if v := os.Getenv("GATEWAY_NATS_URL"); v != "" {
		cfg.NATS.URL = v
	}
	if v := os.Getenv("GATEWAY_AUTH_SECRET"); v != "" {
		cfg.Auth.SecretKey = v
	}

	return cfg
}

func defaultConfig() Config {
	return Config{
		ServerPort: 8080,
		MQTT: MQTTConfig{
			Broker:    "tls://localhost:8883",
			ClientID:  "gateway-dev",
			Username:  "eregen",
			Password:  "eregen_password",
			KeepAlive: 60 * time.Second,
			TLS: TLSConfig{
				Enabled: false, // self-signed cert in dev
			},
		},
		NATS: NATSConfig{
			URL: "nats://localhost:4222",
		},
		Auth: AuthConfig{
			SecretKey: "dev-secret-key-change-in-production",
		},
	}
}
