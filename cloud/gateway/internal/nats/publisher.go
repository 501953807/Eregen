// © 2026 Eregen (颐贞). All rights reserved.

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

// Client wraps the NATS connection.
type Client struct {
	conn      *natss.Conn
	gatewayID string
	mu        sync.Mutex
	url       string
}

// Config holds NATS connection parameters.
type Config struct {
	URL       string
	GatewayID string
}

// NewClient creates a NATS client.
func NewClient(cfg Config) *Client {
	if cfg.GatewayID == "" {
		cfg.GatewayID = fmt.Sprintf("gateway-%d", time.Now().Unix())
	}
	return &Client{
		gatewayID: cfg.GatewayID,
		url:       cfg.URL,
	}
}

// Connect establishes the NATS connection.
func (c *Client) Connect() error {
	conn, err := natss.Connect(c.url,
		natss.DisconnectErrHandler(func(_ *natss.Conn, err error) {
			log.Printf("NATS disconnected: %v", err)
		}),
		natss.ReconnectHandler(func(_ *natss.Conn) {
			log.Println("NATS reconnected")
		}),
		natss.ClosedHandler(func(_ *natss.Conn) {
			log.Println("NATS connection closed")
		}),
	)
	if err != nil {
		return fmt.Errorf("nats connect: %w", err)
	}
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	log.Println("Connected to NATS")
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

// Publish sends an event to NATS with retry logic.
// eventType maps to subject: eregen.event.{event_type}
func (c *Client) Publish(eventType string, payload []byte) error {
	subject := subjectPrefix + eventType

	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil || conn.IsClosed() {
		return fmt.Errorf("nats client not connected")
	}

	data := enrichWithMetadata(payload, c.gatewayID)

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := conn.Publish(subject, data)
		if err == nil {
			return nil
		}
		lastErr = err
		if attempt < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	log.Printf("ERROR: failed to publish to %s after %d retries: %v",
		subject, maxRetries, lastErr)
	return fmt.Errorf("publish after retries: %w", lastErr)
}

// enrichWithMetadata adds gateway metadata to the payload.
func enrichWithMetadata(payload []byte, gatewayID string) []byte {
	var fields map[string]interface{}
	if err := json.Unmarshal(payload, &fields); err != nil {
		return []byte(fmt.Sprintf(`{"_error":"invalid_payload","gateway":"%s","ts":%d}`,
			gatewayID, time.Now().Unix()))
	}

	fields["_gateway"] = gatewayID
	fields["_published_at"] = time.Now().UTC().Format(time.RFC3339)

	enriched, err := json.Marshal(fields)
	if err != nil {
		return payload
	}
	return enriched
}
