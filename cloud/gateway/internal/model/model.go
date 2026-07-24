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
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Accuracy int     `json:"acc"`
	Source   string  `json:"source,omitempty"` // "gps" | "base_station" | ""
	Speed    float64 `json:"speed,omitempty"`
	Heading  float64 `json:"heading,omitempty"`
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

// ========== Medical Wristband Upstream Message Types ==========

const (
	TypePatientRegister UpstreamMessageType = "patient_register"
	TypeVerificationScan UpstreamMessageType = "verification_scan"
	TypeDeviceStatus    UpstreamMessageType = "device_status"
	TypeAlertTag        UpstreamMessageType = "alert_tag"
)

// PatientRegisterPayload carries patient registration data sent by the medical wristband BLE scan.
type PatientRegisterPayload struct {
	PatientID   string `json:"patient_id"`
	AdmissionNo string `json:"admission_no"`
	Name        string `json:"name"`
	Gender      string `json:"gender,omitempty"`
	Age         int    `json:"age,omitempty"`
	Department  string `json:"department,omitempty"`
	BedNumber   string `json:"bed_number,omitempty"`
	BloodType   string `json:"blood_type,omitempty"`
	Allergies   string `json:"allergies,omitempty"`
	SpecialCond string `json:"special_conditions,omitempty"`
}

// VerificationScanPayload carries a nurse BLE verification scan result.
type VerificationScanPayload struct {
	PatientID     string  `json:"patient_id"`
	DeviceID      string  `json:"device_id"`
	ScanType      string  `json:"scan_type"` // "round", "medication", "test", "discharge"
	Result        string  `json:"result"`    // "matched", "unmatched", "not_found"
	VerifiedBy    string  `json:"verified_by,omitempty"`
	Lat           float64 `json:"lat,omitempty"`
	Lon           float64 `json:"lon,omitempty"`
	Notes         string  `json:"notes,omitempty"`
}

// DeviceStatusPayload carries medical wristband device status.
type DeviceStatusPayload struct {
	Battery       int    `json:"bat"`
	FirmwareVer   string `json:"fw_ver,omitempty"`
	Status        string `json:"status"` // "online", "offline", "error"
	LastBindTime  int64  `json:"last_bind_ts,omitempty"`
	BindCount     int    `json:"bind_count,omitempty"`
}

// AlertTagPayload carries an alert tag event from the medical wristband.
type AlertTagPayload struct {
	TagID   string  `json:"tag_id"`
	TagName string  `json:"tag_name"`
	Severity string `json:"severity"` // "info", "warning", "critical"
	PatientID string `json:"patient_id,omitempty"`
	Lat     float64 `json:"lat,omitempty"`
	Lon     float64 `json:"lon,omitempty"`
	Notes   string  `json:"notes,omitempty"`
}

// ========== Community Wristband Upstream Message Types ==========

const (
	TypeCommunitySignin UpstreamMessageType = "community_signin"
	TypeCommunityWelfareUpdate UpstreamMessageType = "community_welfare_update"
	TypeCommunityDispense UpstreamMessageType = "community_dispense"
)

// CommunitySigninPayload carries community elderly wristband check-in data.
type CommunitySigninPayload struct {
	ElderID       string   `json:"elder_id"`
	HospitalID    string   `json:"hospital_id"`
	Period        string   `json:"period"`
	ActivatedTags []string `json:"activated_tags,omitempty"`
	IsMedical     bool     `json:"is_medical_signin"`
	IsWelfare     bool     `json:"is_welfare_signin"`
	Lat           float64  `json:"lat,omitempty"`
	Lon           float64  `json:"lon,omitempty"`
}

// CommunityWelfareUpdatePayload carries welfare tag change events.
type CommunityWelfareUpdatePayload struct {
	ElderID string `json:"elder_id"`
	TagCode string `json:"tag_code"`
	Action  string `json:"action"` // "assign" | "revoke"
}

// CommunityDispensePayload carries pharmacy dispensing records.
type CommunityDispensePayload struct {
	ElderID      string   `json:"elder_id"`
	HospitalID   string   `json:"hospital_id"`
	Period       string   `json:"period"`
	Items        []string `json:"items"`
	TotalCost    float64  `json:"total_cost"`
	Insurance    float64  `json:"insurance_covered"`
	SelfPay      float64  `json:"self_pay"`
}
