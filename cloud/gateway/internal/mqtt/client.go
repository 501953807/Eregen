// © 2026 Eregen (颐贞). All rights reserved.

package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MessageHandler is the callback for received MQTT messages.
type MessageHandler func(topic string, payload []byte)

// Client wraps the paho MQTT client with gateway-specific logic.
type Client struct {
	mqtt   mqtt.Client
	config MQTTConfig
	mu     sync.Mutex
}

// MQTTConfig holds connection parameters for the MQTT client.
type MQTTConfig struct {
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
func NewClient(cfg MQTTConfig) *Client {
	rand.Seed(time.Now().UnixNano())
	timestamp := time.Now().Unix()
	random := rand.Intn(10000)
	clientID := fmt.Sprintf("gateway-%d-%04d", timestamp, random)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetClientID(clientID)
	opts.SetUsername(cfg.Username)
	opts.SetPassword(cfg.Password)
	opts.SetKeepAlive(cfg.KeepAlive)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(30 * time.Second)
	opts.SetCleanSession(true)

	// Last will message
	opts.SetWill("eregen/gateway/status", `{"type":"gateway_status","status":"offline"}`, 1, false)

	if cfg.TLS.Enabled {
		tlsConf, err := buildTLSConfig(cfg.TLS)
		if err != nil {
			log.Fatalf("failed to build TLS config: %v", err)
		}
		opts.SetTLSConfig(tlsConf)
	}

	return &Client{
		config: cfg,
	}
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
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
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

// Connect establishes the MQTT connection to EMQX.
func (c *Client) Connect() error {
	opts := c.createOptions()
	c.mqtt = mqtt.NewClient(opts)
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
	opts.SetOnConnectHandler(func(m mqtt.Client) {
		log.Printf("MQTT connected")
	})
	opts.SetConnectionLostHandler(func(m mqtt.Client, err error) {
		log.Printf("MQTT connection lost: %v", err)
	})

	if c.config.TLS.Enabled {
		tlsConf, _ := buildTLSConfig(c.config.TLS)
		opts.SetTLSConfig(tlsConf)
	}

	return opts
}

// Disconnect gracefully shuts down the MQTT client.
func (c *Client) Disconnect() {
	if c.mqtt != nil && c.mqtt.IsConnected() {
		c.mqtt.Disconnect(1000)
	}
}

// Subscribe registers a handler for an MQTT topic.
func (c *Client) Subscribe(topic string, handler MessageHandler) error {
	if c.mqtt == nil || !c.mqtt.IsConnected() {
		return fmt.Errorf("mqtt client not connected")
	}

	mqttCb := func(_ mqtt.Client, msg mqtt.Message) {
		handler(msg.Topic(), msg.Payload())
	}

	token := c.mqtt.Subscribe(topic, 1, mqttCb)
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

// Publish sends a message to an MQTT topic.
func (c *Client) Publish(topic string, qos byte, payload []byte) error {
	if c.mqtt == nil || !c.mqtt.IsConnected() {
		return fmt.Errorf("mqtt client not connected")
	}
	token := c.mqtt.Publish(topic, qos, false, payload)
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
