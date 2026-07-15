// © 2026 Eregen (颐贞). All rights reserved.

package nats

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSubjectFormatting(t *testing.T) {
	tests := []struct {
		eventType string
		want      string
	}{
		{"heartbeat", "eregen.event.heartbeat"},
		{"location", "eregen.event.location"},
		{"health", "eregen.event.health"},
		{"sos", "eregen.event.sos"},
		{"fall", "eregen.event.fall"},
		{"med_status", "eregen.event.med_status"},
		{"fence_alert", "eregen.event.fence_alert"},
		{"inventory_warning", "eregen.event.inventory_warning"},
	}

	for _, tt := range tests {
		t.Run(tt.eventType, func(t *testing.T) {
			subject := subjectPrefix + tt.eventType
			if subject != tt.want {
				t.Errorf("subject = %q, want %q", subject, tt.want)
			}
		})
	}
}

func TestEnrichWithMetadata(t *testing.T) {
	payload := []byte(`{"type":"heartbeat","dev_id":"BR-1234","bat":85,"ts":1720000000}`)
	enriched := enrichWithMetadata(payload, "gw-test")

	var result map[string]interface{}
	if err := json.Unmarshal(enriched, &result); err != nil {
		t.Fatalf("enriched payload is not valid JSON: %v", err)
	}

	// Check original fields preserved.
	if result["type"] != "heartbeat" {
		t.Errorf("missing type field: %v", result)
	}
	if result["dev_id"] != "BR-1234" {
		t.Errorf("missing dev_id field: %v", result)
	}
	if result["bat"] != float64(85) {
		t.Errorf("missing bat field: %v", result)
	}

	// Check metadata added.
	if _, ok := result["_gateway"]; !ok {
		t.Error("missing _gateway metadata")
	}
	if _, ok := result["_published_at"]; !ok {
		t.Error("missing _published_at metadata")
	}
}

func TestEnrichWithInvalidJSON(t *testing.T) {
	payload := []byte(`not json at all`)
	enriched := enrichWithMetadata(payload, "gw-test")

	// Should return a wrapped error payload.
	if !strings.Contains(string(enriched), "_error") {
		t.Errorf("expected error wrapper, got: %s", string(enriched))
	}
}

func TestMessageSerialization(t *testing.T) {
	// Verify schema types serialize correctly.
	hb := HeartbeatEvent{
		Type:     "heartbeat",
		DevID:    "BR-1234",
		Battery:  85,
		Timestamp: 1720000000,
	}
	data, err := json.Marshal(hb)
	if err != nil {
		t.Fatalf("failed to marshal HeartbeatEvent: %v", err)
	}

	var decoded HeartbeatEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal HeartbeatEvent: %v", err)
	}
	if decoded.DevID != "BR-1234" || decoded.Battery != 85 {
		t.Errorf("round-trip failed: %+v", decoded)
	}

	loc := LocationEvent{
		Type:      "location",
		DevID:     "PX-5678",
		Lat:       31.2304,
		Lon:       121.4737,
		Accuracy:  5,
		Timestamp: 1720000001,
	}
	data, err = json.Marshal(loc)
	if err != nil {
		t.Fatalf("failed to marshal LocationEvent: %v", err)
	}

	var decodedLoc LocationEvent
	if err := json.Unmarshal(data, &decodedLoc); err != nil {
		t.Fatalf("failed to unmarshal LocationEvent: %v", err)
	}
	if decodedLoc.Lat != 31.2304 || decodedLoc.Accuracy != 5 {
		t.Errorf("round-trip failed: %+v", decodedLoc)
	}

	sos := SosEvent{
		Type:      "sos",
		DevID:     "BR-9999",
		Lat:       31.0,
		Lon:       121.0,
		Timestamp: 1720000002,
	}
	data, err = json.Marshal(sos)
	if err != nil {
		t.Fatalf("failed to marshal SosEvent: %v", err)
	}

	fall := FallEvent{
		Type:       "fall",
		DevID:      "BR-8888",
		Confidence: 0.95,
		Lat:        31.0,
		Lon:        121.0,
		Timestamp:  1720000003,
	}
	data, err = json.Marshal(fall)
	if err != nil {
		t.Fatalf("failed to marshal FallEvent: %v", err)
	}

	med := MedStatusEvent{
		Type:        "med_status",
		DevID:       "PX-7777",
		Compartment: 3,
		Taken:       true,
		Timestamp:   1720000004,
	}
	data, err = json.Marshal(med)
	if err != nil {
		t.Fatalf("failed to marshal MedStatusEvent: %v", err)
	}

	inv := InventoryWarningEvent{
		Type:      "inventory_warning",
		DevID:     "PX-6666",
		Medicine:  "aspirin",
		Remaining: 2,
		Timestamp: 1720000005,
	}
	data, err = json.Marshal(inv)
	if err != nil {
		t.Fatalf("failed to marshal InventoryWarningEvent: %v", err)
	}
}
