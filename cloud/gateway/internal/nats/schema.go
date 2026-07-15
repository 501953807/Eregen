// © 2026 Eregen (颐贞). All rights reserved.

// Package nats provides NATS message schema definitions for Eregen events.
package nats

// HeartbeatEvent represents a device heartbeat.
type HeartbeatEvent struct {
	Type     string `json:"type"`     // "heartbeat"
	DevID    string `json:"dev_id"`
	Battery  int    `json:"bat"`
	Timestamp int64 `json:"ts"`
}

// LocationEvent represents a GPS location update.
type LocationEvent struct {
	Type     string  `json:"type"`     // "location"
	DevID    string  `json:"dev_id"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Accuracy int     `json:"acc"`
	Timestamp int64  `json:"ts"`
}

// HealthEvent represents health metric readings.
type HealthEvent struct {
	Type      string `json:"type"`      // "health"
	DevID     string `json:"dev_id"`
	HeartRate int    `json:"hr"`
	SpO2      int    `json:"spo2"`
	Steps     int    `json:"step"`
	Timestamp int64  `json:"ts"`
}

// SosEvent represents an SOS emergency alert.
type SosEvent struct {
	Type      string  `json:"type"`      // "sos"
	DevID     string  `json:"dev_id"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Timestamp int64   `json:"ts"`
}

// FallEvent represents a detected fall.
type FallEvent struct {
	Type      string  `json:"type"`      // "fall"
	DevID     string  `json:"dev_id"`
	Confidence float64 `json:"conf"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Timestamp int64   `json:"ts"`
}

// MedStatusEvent represents a pillbox medication status update.
type MedStatusEvent struct {
	Type        string `json:"type"`        // "med_status"
	DevID       string `json:"dev_id"`
	Compartment int    `json:"compartment"`
	Taken       bool   `json:"taken"`
	Timestamp   int64  `json:"ts"`
}

// FenceAlertEvent represents an electronic fence breach alert.
type FenceAlertEvent struct {
	Type      string  `json:"type"`      // "fence_alert"
	DevID     string  `json:"dev_id"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Timestamp int64   `json:"ts"`
}

// InventoryWarningEvent represents a pillbox inventory warning.
type InventoryWarningEvent struct {
	Type      string `json:"type"`     // "inventory_warning"
	DevID     string `json:"dev_id"`
	Medicine  string `json:"medicine"`
	Remaining int    `json:"remaining"`
	Timestamp int64  `json:"ts"`
}
