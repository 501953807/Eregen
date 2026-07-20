package mqtt

import (
	"testing"
)

func TestParseAndValidate_ValidHeartbeat(t *testing.T) {
	payload := []byte(`{"type":"heartbeat","dev_id":"BR-0001","ts":1720000000,"bat":85}`)
	msg, err := ParseAndValidate(payload)
	if err != nil {
		t.Fatalf("ParseAndValidate failed: %v", err)
	}
	if msg.Type != "heartbeat" {
		t.Errorf("Type = %q, want heartbeat", msg.Type)
	}
	if msg.DeviceID != "BR-0001" {
		t.Errorf("DeviceID = %q, want BR-0001", msg.DeviceID)
	}
	if msg.Timestamp != 1720000000 {
		t.Errorf("Timestamp = %d, want 1720000000", msg.Timestamp)
	}
}

func TestParseAndValidate_ValidSOS(t *testing.T) {
	payload := []byte(`{"type":"sos","dev_id":"PX-0001","ts":1720000000,"lat":31.23,"lon":121.47}`)
	msg, err := ParseAndValidate(payload)
	if err != nil {
		t.Fatalf("ParseAndValidate failed: %v", err)
	}
	if msg.Type != "sos" {
		t.Errorf("Type = %q, want sos", msg.Type)
	}
}

func TestParseAndValidate_MissingType(t *testing.T) {
	payload := []byte(`{"dev_id":"BR-0001","ts":1720000000}`)
	_, err := ParseAndValidate(payload)
	if err == nil {
		t.Error("expected error for missing type")
	}
}

func TestParseAndValidate_EmptyType(t *testing.T) {
	payload := []byte(`{"type":"","dev_id":"BR-0001","ts":1720000000}`)
	_, err := ParseAndValidate(payload)
	if err == nil {
		t.Error("expected error for empty type")
	}
}

func TestParseAndValidate_MissingDevID(t *testing.T) {
	payload := []byte(`{"type":"heartbeat","ts":1720000000}`)
	_, err := ParseAndValidate(payload)
	if err == nil {
		t.Error("expected error for missing dev_id")
	}
}

func TestParseAndValidate_InvalidDeviceID(t *testing.T) {
	payload := []byte(`{"type":"heartbeat","dev_id":"!!!invalid","ts":1720000000}`)
	_, err := ParseAndValidate(payload)
	if err == nil {
		t.Error("expected error for invalid device ID format")
	}
}

func TestParseAndValidate_MissingTimestamp(t *testing.T) {
	payload := []byte(`{"type":"heartbeat","dev_id":"BR-0001"}`)
	_, err := ParseAndValidate(payload)
	if err == nil {
		t.Error("expected error for missing ts")
	}
}

func TestParseAndValidate_InvalidJSON(t *testing.T) {
	payload := []byte(`not json at all`)
	_, err := ParseAndValidate(payload)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestToInt64(t *testing.T) {
	tests := []struct {
		input interface{}
		want  int64
		ok    bool
	}{
		{float64(1720000000), 1720000000, true},
		{int(42), 42, true},
		{int64(999), 999, true},
		{"string", 0, false},
		{nil, 0, false},
	}
	for _, tt := range tests {
		got, ok := toInt64(tt.input)
		if ok != tt.ok || got != tt.want {
			t.Errorf("toInt64(%v) = (%d, %v), want (%d, %v)", tt.input, got, ok, tt.want, tt.ok)
		}
	}
}

func TestExtractDeviceID(t *testing.T) {
	tests := []struct {
		topic string
		want  string
	}{
		{"eregen/device/bracelet/BR-0001/up", "BR-0001"},
		{"eregen/device/pillbox/PX-0001/up", "PX-0001"},
		{"short/topic", ""},
		{"eregen/device/bracelet//up", ""},
	}
	for _, tt := range tests {
		got := extractDeviceID(tt.topic)
		if got != tt.want {
			t.Errorf("extractDeviceID(%q) = %q, want %q", tt.topic, got, tt.want)
		}
	}
}

func TestParsedMessageToDeviceMessage(t *testing.T) {
	pm := &ParsedMessage{
		Type:      "health",
		DeviceID:  "BR-0001",
		Timestamp: 1720000000,
		Raw:       []byte(`{"hr":72}`),
	}
	dm := pm.ToDeviceMessage()
	if dm.Type != "health" || dm.DeviceID != "BR-0001" || dm.Timestamp != 1720000000 {
		t.Error("ToDeviceMessage conversion failed")
	}
}
