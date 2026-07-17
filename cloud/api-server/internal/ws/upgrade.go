package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "" || isAllowedOrigin(origin)
	},
}

func isAllowedOrigin(origin string) bool {
	allowed := []string{
		"http://localhost:",
		"https://localhost:",
		"http://127.0.0.1:",
	}
	for _, a := range allowed {
		if len(origin) >= len(a) && origin[:len(a)] == a {
			return true
		}
	}
	return false
}

// UpgradeHandler upgrades HTTP to WebSocket and registers the client with the Hub.
func UpgradeHandler(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "upgrade failed", http.StatusBadRequest)
			return
		}

		// Extract userID from query parameter
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			conn.Close()
			http.Error(w, "user_id required", http.StatusBadRequest)
			return
		}

		client := hub.Subscribe(userID)
		client.conn = conn

		// Send connection confirmed
		client.Send([]byte(`{"type":"connected","message":"WebSocket connected"}`))
	}
}
