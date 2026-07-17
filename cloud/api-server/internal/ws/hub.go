package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"
)

// Hub manages all connected WebSocket clients and broadcasts alert messages.
type Hub struct {
	clients    map[string]*Client // userID -> Client
	alertChan  chan AlertBroadcast
	mu         sync.RWMutex
	stop       chan struct{}
}

// AlertBroadcast is a message sent to all subscribed clients.
type AlertBroadcast struct {
	ElderlyID string                 `json:"elderly_id"`
	Type      string                 `json:"type"` // sos, fall, geofence_breach, high_risk_score
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		clients:   make(map[string]*Client),
		alertChan: make(chan AlertBroadcast, 256),
		stop:      make(chan struct{}),
	}
}

// Run starts the Hub's broadcast loop. Blocks until Stop is called.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			h.stopHub()
			return
		case msg := <-h.alertChan:
			h.broadcastAlert(msg)
		}
	}
}

// Stop closes the Hub and disconnects all clients.
func (h *Hub) stopHub() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, c := range h.clients {
		c.Close()
	}
	close(h.stop)
}

// broadcastAlert sends an alert to all clients subscribed to the affected elderly.
func (h *Hub) broadcastAlert(msg AlertBroadcast) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[ws] marshal alert: %v", err)
		return
	}
	for _, c := range h.clients {
		c.Send(data)
	}
}

// Subscribe registers a client for a specific elderly person's alerts.
func (h *Hub) Subscribe(userID string) *Client {
	h.mu.Lock()
	defer h.mu.Unlock()
	if existing, ok := h.clients[userID]; ok {
		existing.Close()
	}
	client := NewClient(userID, h.alertChan)
	h.clients[userID] = client
	go client.ReadPump()
	go client.WritePump()
	return client
}

// Unsubscribe removes a client.
func (h *Hub) Unsubscribe(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c, ok := h.clients[userID]; ok {
		c.Close()
		delete(h.clients, userID)
	}
}

// PublishAlert sends an alert into the Hub's broadcast channel.
func (h *Hub) PublishAlert(msg AlertBroadcast) {
	select {
	case h.alertChan <- msg:
	default:
		log.Println("[ws] alert channel full, dropping")
	}
}
