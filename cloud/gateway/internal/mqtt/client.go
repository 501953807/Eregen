// © 2026 Eregen (颐贞). All rights reserved.

// Package mqtt provides an EMQX connection with device topic routing.
package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var deviceIDRegex = regexp.MustCompile(`^(BR|PX)-[A-Za-z0-9]+$`)

// MessageHandler is the callback for received MQTT messages.
type MessageHandler func(topic string, payload []byte)

// Client wraps the paho MQTT client with gateway-specific logic.
type Client struct {
	mqtt   mqtt.Client
	config *Config
	mu     sync.Mutex
}

// Config holds connection parameters for the MQTT client.
type Config struct {
	Broker    string
	ClientID  string
	Username  string
	Password  string
	TLS       TLSConfig
	KeepAlive time.Duration
}

// TLSConfig holds TLS certificate paths.
type TLSConfig struct {
	Enabled bool
	CACert  string
	Cert    string
	Key     string
}

// NewClient creates a new MQTT client from configuration.
func NewClient(cfg *Config) *Client {
	rand.Seed(time.Now().UnixNano())
	timestamp := time.Now().Unix()
	random := rand.Intn(10000)
	clientID := fmt.Sprintf("gateway-%d-%04d", timestamp, random)

	return &Client{
		config: &Config{
			Broker:    cfg.Broker,
			ClientID:  clientID,
			Username:  cfg.Username,
			Password:  cfg.Password,
			TLS:       cfg.TLS,
			KeepAlive: cfg.KeepAlive,
		},
	}
}

// Connect establishes the MQTT connection to EMQX.
func (c *Client) Connect() error {
	opts := c.createOptions()
	c.mu.Lock()
	c.mqtt = mqtt.NewClient(opts)
	c.mu.Unlock()

	token := c.mqtt.Connect()
	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-token.Done():
			return token.Error()
		case <-timeout:
			return fmt.Errorf("mqtt connect timeout")
		}
	}
}

func (c *Client) createOptions() *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(c.config.Broker)
	opts.SetClientID(c.config.ClientID)
	opts.SetUsername(c.config.Username)
	opts.SetPassword(c.config.Password)
	opts.SetKeepAlive(c.config.KeepAlive)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(30 * time.Second)
	opts.SetCleanSession(true)

	// Last will message so downstream systems know when this gateway dies.
	opts.SetWill("eregen/gateway/status", `{"type":"gateway_status","status":"offline"}`, 1, false)

	if c.config.TLS.Enabled {
		tlsConf, err := buildTLSConfig(c.config.TLS)
		if err != nil {
			log.Fatalf("failed to build TLS config: %v", err)
		}
		opts.SetTLSConfig(tlsConf)
	}

	opts.SetOnConnectHandler(func(m mqtt.Client) {
		log.Printf("MQTT connected")
	})
	opts.SetConnectionLostHandler(func(m mqtt.Client, err error) {
		log.Printf("MQTT connection lost: %v", err)
	})

	return opts
}

// Disconnect gracefully shuts down the MQTT client.
func (c *Client) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.mqtt != nil && c.mqtt.IsConnected() {
		c.mqtt.Disconnect(1000)
	}
}

// Subscribe registers a handler for an MQTT topic.
func (c *Client) Subscribe(topic string, handler MessageHandler) error {
	c.mu.Lock()
	cl := c.mqtt
	c.mu.Unlock()

	if cl == nil || !cl.IsConnected() {
		return fmt.Errorf("mqtt client not connected")
	}

	mqttCb := func(_ mqtt.Client, msg mqtt.Message) {
		handler(msg.Topic(), msg.Payload())
	}

	token := cl.Subscribe(topic, 1, mqttCb)
	timeout := time.After(5 * time.Second)
	for {
		select {
		case <-token.Done():
			return token.Error()
		case <-timeout:
			return fmt.Errorf("subscribe timeout for topic: %s", topic)
		}
	}
}

// Publish sends a message to an MQTT topic (for downstream commands).
func (c *Client) Publish(topic string, qos byte, payload []byte) error {
	c.mu.Lock()
	cl := c.mqtt
	c.mu.Unlock()

	if cl == nil || !cl.IsConnected() {
		return fmt.Errorf("mqtt client not connected")
	}

	token := cl.Publish(topic, qos, false, payload)
	timeout := time.After(5 * time.Second)
	for {
		select {
		case <-token.Done():
			return token.Error()
		case <-timeout:
			return fmt.Errorf("publish timeout for topic: %s", topic)
		}
	}
}

// DeviceIDFromTopic extracts the device ID from an MQTT topic path.
// "eregen/device/bracelet/BR-1234/up" -> "BR-1234"
func DeviceIDFromTopic(topic string) string {
	parts := splitTopic(topic)
	if len(parts) >= 5 {
		return parts[3]
	}
	return ""
}

// ValidateDeviceID checks that a device ID matches the BR-/PX- prefix format.
func ValidateDeviceID(id string) bool {
	return deviceIDRegex.MatchString(id)
}

func buildTLSConfig(tc TLSConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // self-signed cert in dev
	}

	if tc.CACert != "" {
		caCert, err := os.ReadFile(tc.CACert)
		if err != nil {
			return nil, fmt.Errorf("read CA cert: %w", err)
		}
		caPool := x509.NewCertPool()
		caPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caPool
	}

	if tc.Cert != "" && tc.Key != "" {
		cert, err := tls.LoadX509KeyPair(tc.Cert, tc.Key)
		if err != nil {
			return nil, fmt.Errorf("load client cert: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}

func splitTopic(s string) []string {
	out := make([]string, 0, 5)
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}
