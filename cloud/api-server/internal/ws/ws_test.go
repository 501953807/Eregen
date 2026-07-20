package ws

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()
	if hub == nil {
		t.Fatal("NewHub returned nil")
	}
	if hub.clients == nil {
		t.Error("hub.clients should be initialized")
	}
	if hub.alertChan == nil {
		t.Error("hub.alertChan should be initialized")
	}
	if hub.stop == nil {
		t.Error("hub.stop should be initialized")
	}
}

func TestAlertBroadcastJSON(t *testing.T) {
	msg := AlertBroadcast{
		ElderlyID: "elderly-1",
		Type:      "sos",
		Payload: map[string]interface{}{
			"lat": 31.23,
			"lon": 121.47,
		},
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("json marshal failed: %v", err)
	}

	var decoded AlertBroadcast
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}

	if decoded.ElderlyID != "elderly-1" {
		t.Errorf("ElderlyID = %q, want elderly-1", decoded.ElderlyID)
	}
	if decoded.Type != "sos" {
		t.Errorf("Type = %q, want sos", decoded.Type)
	}
	if decoded.Payload["lat"] != 31.23 {
		t.Errorf("Payload lat = %v, want 31.23", decoded.Payload["lat"])
	}
}

func TestClientSendChannelInitialized(t *testing.T) {
	alertChan := make(chan AlertBroadcast, 10)
	client := NewClient("user-1", alertChan)
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.send == nil {
		t.Error("client.send channel should be initialized")
	}
	if client.userID != "user-1" {
		t.Errorf("userID = %q, want user-1", client.userID)
	}
}

func TestClientCloseNoConn(t *testing.T) {
	alertChan := make(chan AlertBroadcast, 10)
	client := NewClient("user-1", alertChan)
	client.conn = nil
	// Should not panic when conn is nil
	client.Close()
}
