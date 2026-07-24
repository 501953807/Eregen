// Package protocol defines the BLE GATT protocol for medical wristband communication.
package protocol

import "time"

// BLE Service UUIDs for medical wristband
const (
	ServiceUUID = "0000xxxx-0000-1000-8000-00805f9b34fb"
)

// Characteristic UUIDs
const (
	CharPairingCode    = "0000xxx1-0000-1000-8000-00805f9b34fb" // Pairing code input/output
	CharPatientInfo    = "0000xxx2-0000-1000-8000-00805f9b34fb" // Patient information
	CharMedicalSummary = "0000xxx3-0000-1000-8000-00805f9b34fb" // Medical summary
	CharVerification   = "0000xxx4-0000-1000-8000-00805f9b34fb" // Verification request/response
	CharStatus         = "0000xxx5-0000-1000-8000-00805f9b34fb" // Device status
	CharCommand        = "0000xxx6-0000-1000-8000-00805f9b34fb" // Command channel
)

// BLEMessageType identifies the kind of BLE message
type BLEMessageType uint8

const (
	TypePairingRequest      BLEMessageType = 0x01
	TypePairingResponse     BLEMessageType = 0x02
	TypePatientInfoReq      BLEMessageType = 0x10
	TypePatientInfoResp     BLEMessageType = 0x11
	TypeMedicalSummaryReq   BLEMessageType = 0x20
	TypeMedicalSummaryResp  BLEMessageType = 0x21
	TypeVerificationReq     BLEMessageType = 0x30
	TypeVerificationResp    BLEMessageType = 0x31
	TypeStatusReport        BLEMessageType = 0x40
	TypeCommandAck          BLEMessageType = 0x50
)

// BLEPairingCode represents a 4-digit pairing code
type BLEPairingCode struct {
	Code string `json:"code"` // 4 digits
}

// BLEPatientInfo contains patient identification data
type BLEPatientInfo struct {
	PatientID   string    `json:"patient_id"`
	AdmissionNo string    `json:"admission_no"`
	Name        string    `json:"name,omitempty"`
	BoundAt     time.Time `json:"bound_at"`
}

// BLEMedicalSummary contains encrypted medical summary data
type BLEMedicalSummary struct {
	Version       string `json:"version"`
	EncryptedData []byte `json:"encrypted_data"`
	IV            []byte `json:"iv"`
	HMACSignature []byte `json:"hmac_signature"`
}

// BLEVerificationRequest is sent from nurse device to wristband
type BLEVerificationRequest struct {
	RequestID    string `json:"request_id"`
	Challenge    []byte `json:"challenge"`
	Timestamp    int64  `json:"timestamp"`
	VerificationType string `json:"verification_type"` // "medication", "treatment", "discharge"
}

// BLEVerificationResponse is returned by wristband
type BLEVerificationResponse struct {
	RequestID string `json:"request_id"`
	Result    string `json:"result"` // "matched", "unmatched", "error"
	SignedChallenge []byte `json:"signed_challenge"`
	Timestamp int64  `json:"timestamp"`
}

// BLEDeviceStatus reports wristband state
type BLEDeviceStatus struct {
	Status      string `json:"status"` // "idle", "bound", "pairing", "error"
	Battery     int    `json:"battery"` // 0-100
	FirmwareVer string `json:"firmware_version"`
	DeviceID    string `json:"device_id"`
	Timestamp   int64  `json:"timestamp"`
}

// BLECommand represents commands from nurse device to wristband
type BLECommand struct {
	CommandType string `json:"command_type"` // "clear_binding", "read_status", "reboot"
	CommandID   string `json:"command_id"`
	Timestamp   int64  `json:"timestamp"`
}

// BLECommandAck acknowledges command execution
type BLECommandAck struct {
	CommandID string `json:"command_id"`
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// ========== NFC Authentication Protocol ==========

const (
	NFCServiceUUID = "0000fff0-0000-1000-8000-00805f9b34fb"
)

// NFCRecordType defines NDEF record type labels.
type NFCRecordType string

const (
	NFCRecordTypeText  NFCRecordType = "text/utf-8"
	NFCRecordTypeAuth  NFCRecordType = "application/vnd.eregen.auth"
)

// NFCAuthPayload represents the data structure for NFC-based authentication.
type NFCAuthPayload struct {
	ElderID     string `json:"elder_id"`
	SerialNo    string `json:"serial_number,omitempty"`
	Timestamp   int64  `json:"timestamp"`
	Challenge   []byte `json:"challenge,omitempty"`
	Response    []byte `json:"response,omitempty"`
}
