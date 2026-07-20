package model

import "testing"

func TestUpstreamMessageTypeConstants(t *testing.T) {
	tests := []struct {
		name string
		got  UpstreamMessageType
		want string
	}{
		{"heartbeat", TypeHeartbeat, "heartbeat"},
		{"location", TypeLocation, "location"},
		{"health", TypeHealth, "health"},
		{"sos", TypeSOS, "sos"},
		{"fall", TypeFall, "fall"},
		{"med_status", TypeMedStatus, "med_status"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestDeviceMessageJSON(t *testing.T) {
	msg := DeviceMessage{
		Type:      TypeHealth,
		DeviceID:  "BR-0001",
		Timestamp: 1720000000,
	}

	if msg.Type != "health" || msg.DeviceID != "BR-0001" || msg.Timestamp != 1720000000 {
		t.Error("DeviceMessage fields not set correctly")
	}
}

func TestHeartbeatPayloadDefaults(t *testing.T) {
	h := HeartbeatPayload{}
	if h.Battery != 0 || h.Model != "" || h.FWVer != "" {
		t.Error("zero-value HeartbeatPayload should have empty optional fields")
	}
}

func TestSOSPayloadTriggerValues(t *testing.T) {
	sosManual := SOSPayload{Lat: 31.23, Lon: 121.47, Trigger: "manual"}
	sosLong := SOSPayload{Lat: 31.23, Lon: 121.47, Trigger: "long_press"}

	if sosManual.Trigger != "manual" || sosLong.Trigger != "long_press" {
		t.Error("SOS trigger values not set correctly")
	}
}
