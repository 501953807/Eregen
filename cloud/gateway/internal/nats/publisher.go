// © 2026 Eregen (颐贞). All rights reserved.

// Package nats provides NATS JetStream publishing for device events.
package nats

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	natss "github.com/nats-io/nats.go"
)

const (
	maxRetries    = 3
	retryDelay    = 100 * time.Millisecond
	subjectPrefix = "eregen.event."
)

// Client wraps a NATS JetStream connection.
type Client struct {
	conn      *natss.Conn
	js        natss.JetStreamContext
	gatewayID string
	stream    string
	url       string
	domain    string
	mu        sync.Mutex
}

// Config holds NATS connection parameters.
type Config struct {
	URL             string
	JetStreamDomain string
	StreamName      string
	GatewayID       string
}

// Event is the canonical envelope published to NATS JetStream.
type Event struct {
	Type        string          `json:"type"`
	DeviceID    string          `json:"dev_id"`
	Timestamp   int64           `json:"ts"`
	Payload     json.RawMessage `json:"payload"`
	Gateway     string          `json:"_gateway"`
	PublishedAt string          `json:"_published_at"`
}

// NewClient creates a NATS JetStream client.
func NewClient(cfg Config) *Client {
	if cfg.GatewayID == "" {
		cfg.GatewayID = fmt.Sprintf("gateway-%d", time.Now().Unix())
	}
	if cfg.StreamName == "" {
		cfg.StreamName = "DEVICE_EVENTS"
	}
	return &Client{
		gatewayID: cfg.GatewayID,
		stream:    cfg.StreamName,
		url:       cfg.URL,
		domain:    cfg.JetStreamDomain,
	}
}

// Connect establishes the NATS connection and ensures the JetStream stream exists.
func (c *Client) Connect() error {
	conn, err := natss.Connect(c.url)
	if err != nil {
		return fmt.Errorf("nats connect: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	var jsOpts []natss.JSOpt
	if c.domain != "" {
		jsOpts = append(jsOpts, natss.Domain(c.domain))
	}
	c.js, err = conn.JetStream(jsOpts...)
	c.mu.Unlock()
	if err != nil {
		conn.Close()
		return fmt.Errorf("nats jetstream: %w", err)
	}

	// Ensure the stream exists; create it if missing.
	_, err = c.js.StreamInfo(c.stream)
	if err == natss.ErrStreamNotFound {
		_, err = c.js.AddStream(&natss.StreamConfig{
			Name:      c.stream,
			Subjects:  []string{subjectPrefix + "*"},
			Storage:   natss.FileStorage,
			Retention: natss.LimitsPolicy,
			MaxMsgs:   -1,
			Replicas:  1,
		})
		if err != nil {
			conn.Close()
			return fmt.Errorf("nats create stream: %w", err)
		}
		log.Printf("Created JetStream stream %q", c.stream)
	} else if err != nil {
		conn.Close()
		return fmt.Errorf("nats stream info: %w", err)
	}

	log.Println("Connected to NATS JetStream")
	return nil
}

// Close shuts down the NATS connection.
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		c.conn.Close()
	}
}

// Publish sends a device event to NATS JetStream with retry logic.
func (c *Client) Publish(ev *Event) error {
	data, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		_, err = c.js.Publish(subjectPrefix+ev.Type, data)
		if err == nil {
			return nil
		}
		lastErr = err
		if attempt < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}
	log.Printf("ERROR: failed to publish %s for %s after %d retries: %v",
		ev.Type, ev.DeviceID, maxRetries, lastErr)
	return fmt.Errorf("publish after retries: %w", lastErr)
}
