package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"os"

	"eregen.dev/admin-api/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// alertEvent is the JSON payload pushed to SSE clients.
type alertEvent struct {
	Type  string                 `json:"type"`
	Alert *model.AlertSummary    `json:"alert,omitempty"`
	Alerts []model.AlertSummary  `json:"alerts,omitempty"`
	Stats *model.DashboardStats  `json:"stats,omitempty"`
}

// AlertsHub broadcasts new alerts to all connected SSE clients.
type AlertsHub struct {
	mu      sync.RWMutex
	clients map[http.ResponseWriter]struct{}
	pending chan alertEvent
	logger  *zap.Logger
}

// Global hub instance.
var AlertsHubInstance *AlertsHub

func init() {
	AlertsHubInstance = NewAlertsHub()
}

func NewAlertsHub() *AlertsHub {
	h := &AlertsHub{
		clients: make(map[http.ResponseWriter]struct{}),
		pending: make(chan alertEvent, 256),
		logger:  zap.L(),
	}
	go h.process()
	return h
}

func (h *AlertsHub) process() {
	for evt := range h.pending {
		h.broadcast(evt)
	}
}

func (h *AlertsHub) broadcast(evt alertEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for w := range h.clients {
		data, err := json.Marshal(evt)
		if err != nil {
			log.Printf("sse marshal error: %v", err)
			continue
		}
		fmt.Fprintf(w, "event: alert\n")
		fmt.Fprintf(w, "data: %s\n\n", data)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

// Push sends a single alert to all connected clients.
func (h *AlertsHub) Push(evtType string, alert model.AlertSummary) {
	h.pending <- alertEvent{Type: evtType, Alert: &alert}
}

// PushBatch sends a batch of alerts (e.g. on connect for catch-up).
func (h *AlertsHub) PushBatch(evtType string, alerts []model.AlertSummary) {
	h.pending <- alertEvent{Type: evtType, Alerts: alerts}
}

// PushStats sends a dashboard stats update.
func (h *AlertsHub) PushStats(stats model.DashboardStats) {
	h.pending <- alertEvent{Type: "stats", Stats: &stats}
}

// Subscribe registers a client writer for SSE.
func (h *AlertsHub) Subscribe(w http.ResponseWriter) {
	h.mu.Lock()
	h.clients[w] = struct{}{}
	h.mu.Unlock()
}

// Unsubscribe removes a client.
func (h *AlertsHub) Unsubscribe(w http.ResponseWriter) {
	h.mu.Lock()
	delete(h.clients, w)
	h.mu.Unlock()
}

// validateTokenFromQuery checks the JWT token passed via ?token= query param.
// EventSource API cannot send custom headers, so tokens must be passed as query params.
func validateTokenFromQuery(r *http.Request) bool {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		return false
	}
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return false
	}
	_, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	return err == nil
}

// Handler returns an HTTP handler for GET /api/v1/admin/stream/alerts.
// Expects ?token=<jwt> query parameter for authentication (EventSource limitation).
func (h *AlertsHub) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		if !validateTokenFromQuery(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		h.Subscribe(w)
		h.logger.Info("SSE client connected", zap.String("remote", r.RemoteAddr))

		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		done := r.Context().Done()
		for {
			select {
			case <-ticker.C:
				fmt.Fprint(w, ": ping\n\n")
				flusher.Flush()
			case <-done:
				h.Unsubscribe(w)
				h.logger.Info("SSE client disconnected", zap.String("remote", r.RemoteAddr))
				return
			}
		}
	}
}
