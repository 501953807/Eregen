package handler

import (
	"encoding/json"

	"eregen.dev/gateway/internal/model"
	"eregen.dev/gateway/internal/nats"
)

// validGPS checks if GPS coordinates are within valid ranges.
func validGPS(lat, lon float64) bool {
	return lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180
}

// validateHealth clamps health data to valid ranges.
func validateHealth(h *model.HealthPayload) {
	if h.HeartRate < 0 || h.HeartRate > 300 {
		h.HeartRate = 0
	}
	if h.SpO2 < 0 || h.SpO2 > 100 {
		h.SpO2 = 0
	}
	if h.Steps < 0 {
		h.Steps = 0
	}
}

// makeNATSEvent converts a device message and payload to a NATS event.
func makeNATSEvent(msg *model.DeviceMessage, payload any) *nats.Event {
	data, _ := json.Marshal(payload)
	ev := &nats.Event{
		Type:      string(msg.Type),
		DeviceID:  msg.DeviceID,
		Timestamp: msg.Timestamp,
		Payload:   data,
	}
	// Extract hospital_id from community signin payloads for cross-hospital tracking
	switch p := payload.(type) {
	case model.CommunitySigninPayload:
		if p.HospitalID != "" {
			ev.HospitalID = p.HospitalID
		}
	}
	return ev
}
