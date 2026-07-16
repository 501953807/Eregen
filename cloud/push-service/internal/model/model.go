package model

import "time"

// EventType identifies the kind of push event.
type EventType string

const (
	EventTypeAlert    EventType = "alert"
	EventTypeReminder EventType = "reminder"
	EventTypeVoiceCall EventType = "voice_call"
)

// Severity maps to alert priority levels.
type Severity string

const (
	SeverityP0 Severity = "P0" // SOS, fall — immediate
	SeverityP1 Severity = "P1" // Elevated risk
	SeverityP2 Severity = "P2" // Informational
)

// AlertPushEvent fires when an alert needs distribution.
type AlertPushEvent struct {
	AlertID    string    `json:"alert_id"`
	ElderlyID  string    `json:"elderly_id"`
	Severity   Severity  `json:"severity"`
	AlertType  string    `json:"alert_type"`
	Message    string    `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
	RawData    map[string]interface{} `json:"raw_data,omitempty"`
}

// ReminderPushEvent fires for medication or health reminders.
type ReminderPushEvent struct {
	ElderlyID string    `json:"elderly_id"`
	RuleID    string    `json:"rule_id"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// VoiceCallEvent fires for SOS voice call requests.
type VoiceCallEvent struct {
	ElderlyID string    `json:"elderly_id"`
	CallerID  string    `json:"caller_id"` // family member ID
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}
