// Package protocol defines the device-cloud communication protocol for Eregen devices.
// All message types use JSON encoding over MQTT with NATS as the internal transport.
package protocol

import "time"

// MessageType identifies the kind of MQTT message payload.
type MessageType string

const (
	TypeHeartbeat    MessageType = "heartbeat"
	TypeLocation     MessageType = "location"
	TypeHealth       MessageType = "health"
	TypeSOS          MessageType = "sos"
	TypeFall         MessageType = "fall"
	TypeMedStatus    MessageType = "med_status"
	TypeMedRule      MessageType = "med_rule"
	TypeConfig       MessageType = "config"
	TypeTTS          MessageType = "tts"
	TypeOTA          MessageType = "ota"
	TypeDeviceInfo   MessageType = "device_info"
	TypeAlertForward MessageType = "alert_forward" // cloud → institution
)

// DeviceType identifies the hardware family.
type DeviceType string

const (
	DeviceBracelet    DeviceType = "bracelet"
	DevicePillbox     DeviceType = "pillbox"
)

// BraceletTier identifies the bracelet model tier.
type BraceletTier string

const (
	TierStarter BraceletTier = "starter"
	TierPlus    BraceletTier = "plus"
	TierPro     BraceletTier = "pro"
)

// PillboxTier identifies the pillbox model tier.
type PillboxTier string

const (
	TierBasic   PillboxTier = "basic"
	TierSmart   PillboxTier = "smart"
	TierAuto    PillboxTier = "auto"
)

// AlertPriority indicates urgency level.
type AlertPriority string

const (
	PriorityP0 AlertPriority = "P0" // SOS, fall detection
	PriorityP1 AlertPriority = "P1" // medication missed, geofence breach
	PriorityP2 AlertPriority = "P2" // device offline, low battery
)

// ---------- Upstream Messages (Device → Cloud) ----------

// Heartbeat is sent by devices every heartbeat_interval seconds.
type Heartbeat struct {
	Type    MessageType `json:"type"`
	DevID   string      `json:"dev_id"`
	Battery int         `json:"bat"` // 0-100
}

// Location contains GPS coordinates from a bracelet.
type Location struct {
	Type  MessageType  `json:"type"`
	DevID string       `json:"dev_id"`
	Lat   float64      `json:"lat"`
	Lon   float64      `json:"lon"`
	Acc   float64      `json:"acc"`  // accuracy in meters
	Ts    time.Time    `json:"ts"`
}

// Health contains aggregated health data from bracelet sensors.
type Health struct {
	Type  MessageType  `json:"type"`
	DevID string       `json:"dev_id"`
	HR    *float64     `json:"hr,omitempty"`     // heart rate bpm
	SPO2  *float64     `json:"spo2,omitempty"`   // blood oxygen %
	Step  *int64       `json:"step,omitempty"`   // step count
	Sleep *float64     `json:"sleep,omitempty"`  // sleep hours
	Temp  *float64     `json:"temp,omitempty"`   // body temperature °C
	Ts    time.Time    `json:"ts"`
}

// SOS is triggered when the user presses the SOS button.
type SOS struct {
	Type  MessageType  `json:"type"`
	DevID string       `json:"dev_id"`
	Lat   float64      `json:"lat"`
	Lon   float64      `json:"lon"`
	Ts    time.Time    `json:"ts"`
}

// Fall indicates detected fall event with confidence score.
type Fall struct {
	Type  MessageType  `json:"type"`
	DevID string       `json:"dev_id"`
	Conf  float64      `json:"conf"` // 0.0-1.0 confidence
	Lat   float64      `json:"lat"`
	Lon   float64      `json:"lon"`
	Ts    time.Time    `json:"ts"`
}

// MedStatus reports pillbox compartment status.
type MedStatus struct {
	Type        MessageType `json:"type"`
	DevID       string      `json:"dev_id"`
	Compartment int         `json:"compartment"` // 0-based index
	Taken       bool        `json:"taken"`
	Timestamp   time.Time   `json:"ts"`
}

// DeviceInfo is sent on first boot and after OTA updates.
type DeviceInfo struct {
	Type       MessageType  `json:"type"`
	DevID      string       `json:"dev_id"`
	Firmware   string       `json:"firmware"`
	DeviceType DeviceType   `json:"device_type"`
	Tier       string       `json:"tier"`
	MAC        string       `json:"mac"`
	Battery    int          `json:"bat"`
	Ts         time.Time    `json:"ts"`
}

// ---------- Downstream Messages (Cloud → Device) ----------

// MedRule configures medication schedule on a pillbox.
type MedRule struct {
	Type  MessageType   `json:"type"`
	DevID string        `json:"dev_id"`
	Rules []MedSchedule `json:"rules"`
}

// MedSchedule represents a single medication time slot.
type MedSchedule struct {
	Time   string   `json:"time"`  // HH:MM format
	Dose   int      `json:"dose"`  // number of pills
	Type   string   `json:"type"`  // capsule, tablet, liquid
	Name   string   `json:"name"`  // medication name
}

// ConfigUpdate pushes settings to a bracelet.
type Config struct {
	Type     MessageType    `json:"type"`
	DevID    string         `json:"dev_id"`
	Settings map[string]any `json:"settings"`
}

// TTS triggers text-to-speech playback on a pillbox.
type TTS struct {
	Type  MessageType `json:"type"`
	DevID string      `json:"dev_id"`
	Text  string      `json:"text"`
}

// OTA provides firmware update URL and verification hash.
type OTA struct {
	Type  MessageType `json:"type"`
	DevID string      `json:"dev_id"`
	URL   string      `json:"url"`
	Hash  string      `json:"hash"` // sha256
	Size  int64       `json:"size"` // bytes
}

// AlertNotification is sent to family members' apps when an alert fires.
type AlertNotification struct {
	Type      MessageType   `json:"type"`
	AlertID   string        `json:"alert_id"`
	ElderlyID string        `json:"elderly_id"`
	Priority  AlertPriority `json:"priority"`
	Title     string        `json:"title"`
	Message   string        `json:"message"`
	Location  *Location     `json:"location,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	Read      bool          `json:"read"`
}

// ---------- MQTT Topic Layout ----------

// Topics follow the pattern: eregen/{direction}/{device_type}/{device_id}
const (
	TopicUpHeartbeat   = "eregen/up/%s/%s/heartbeat"
	TopicUpLocation    = "eregen/up/%s/%s/location"
	TopicUpHealth      = "eregen/up/%s/%s/health"
	TopicUpSOS         = "eregen/up/%s/%s/sos"
	TopicUpFall        = "eregen/up/%s/%s/fall"
	TopicUpMedStatus   = "eregen/up/%s/%s/med_status"
	TopicUpDeviceInfo  = "eregen/up/%s/%s/device_info"
	TopicDownConfig    = "eregen/down/%s/%s/config"
	TopicDownTTS       = "eregen/down/%s/%s/tts"
	TopicDownOTA       = "eregen/down/%s/%s/ota"
	TopicDownMedRule   = "eregen/down/%s/%s/med_rule"
	TopicAlertNotify   = "eregen/alert/%s/%s"
)

// ---------- NATS Subject Mapping ----------

// NATS subjects used internally between microservices.
const (
	NATSHealthData   = "eregen.health.data"
	NATSLocData      = "eregen.location.data"
	NATSSOSAlert     = "eregen.alert.sos"
	NATSFallAlert    = "eregen.alert.fall"
	NATSMedStatus    = "eregen.med.status"
	NATSDeviceOnline = "eregen.device.online"
	NATSDeviceOffline = "eregen.device.offline"
	NATSPushRequest  = "eregen.push.request"
)
