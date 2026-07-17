package ws

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client represents a single WebSocket connection.
type Client struct {
	userID    string
	conn      *websocket.Conn
	send      chan []byte
	mu        sync.Mutex
	alertChan chan<- AlertBroadcast
}

// NewClient creates a new Client bound to a user.
func NewClient(userID string, alertChan chan AlertBroadcast) *Client {
	return &Client{
		userID:    userID,
		send:      make(chan []byte, 256),
		alertChan: alertChan,
	}
}

// Send writes raw JSON bytes to the WebSocket.
func (c *Client) Send(data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	select {
	case c.send <- data:
	default:
		log.Printf("[ws] send queue full for %s, dropping", c.userID)
	}
}

// Close shuts down the client connection.
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		c.conn.Close()
	}
}

// ReadPump reads from the WebSocket and handles incoming messages.
func (c *Client) ReadPump() {
	defer c.Close()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway, websocket.CloseNormalClosure,
				websocket.CloseAbnormalClosure) {
				log.Printf("[ws] read error for %s: %v", c.userID, err)
			}
			break
		}
		c.handleIncoming(message)
	}
}

// WritePump writes to the WebSocket from the send channel.
func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	defer c.Close()
	for {
		select {
		case msg := <-c.send:
			c.mu.Lock()
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				c.mu.Unlock()
				break
			}
			c.mu.Unlock()
		case <-ticker.C:
			c.mu.Lock()
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.mu.Unlock()
				break
			}
			c.mu.Unlock()
		}
	}
}

// handleIncoming processes client messages.
func (c *Client) handleIncoming(msg []byte) {
	var req struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(msg, &req); err != nil {
		return // ignore malformed
	}
	// Currently we only support push notifications from server to client.
	// Future: could support subscribe/unsubscribe commands.
}
