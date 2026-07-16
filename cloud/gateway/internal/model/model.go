// © 2026 Eregen (颐贞). All rights reserved.

// Package model defines device message types per the Eregen protocol spec.
package model

import (
	"encoding/json"
)

// UpstreamMessageType identifies the kind of device event.
type UpstreamMessageType string

const (
	TypeHeartbeat    UpstreamMessageType = "heartbeat"
	TypeLocation     UpstreamMessageType = "location"
	TypeHealth       UpstreamMessageType = "health"
	TypeSOS          UpstreamMessageType = "sos"
	TypeFall         UpstreamMessageType = "fall"
	TypeMedStatus    UpstreamMessageType = "med_status"
)

// DeviceMessage is the envelope for every upstream MQTT payload.
type DeviceMessage struct {
	Type      UpstreamMessageType `json:"type"`
	DeviceID  string              `json:"dev_id"`
	Timestamp int64               `json:"ts"`
	Raw       json.RawMessage     `json:"-"`
}

// HeartbeatPayload carries device battery and status info.
type HeartbeatPayload struct {
	Battery int    `json:"bat"`
	Model   string `json:"model,omitempty"`
	FWVer   string `json:"fw_ver,omitempty"`
}

// LocationPayload carries GPS coordinates from a bracelet device.
type LocationPayload struct {
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Accuracy int    `json:"acc"`
	Speed   float64 `json:"speed,omitempty"`
	Heading float64 `json:"heading,omitempty"`
}

// HealthPayload carries biometric readings from a bracelet device.
type HealthPayload struct {
	HeartRate int `json:"hr"`
	SpO2      int `json:"spo2"`
	Steps     int `json:"step"`
	Sleep     int `json:"sleep,omitempty"` // minutes
	BPM       int `json:"bpm,omitempty"`   // body temp * 100
}

// SOSPayload carries emergency alert data.
type SOSPayload struct {
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Trigger  string  `json:"trigger,omitempty"` // "manual" | "long_press"
}

// FallPayload carries fall detection results.
type FallPayload struct {
	Confidence float64 `json:"conf"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
}

// MedStatusPayload carries pillbox medication compartment status.
type MedStatusPayload struct {
	Compartment int  `json:"compartment"`
	Taken       bool `json:"taken"`
}
