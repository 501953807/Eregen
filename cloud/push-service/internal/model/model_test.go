package model

import "testing"

func TestEventTypes(t *testing.T) {
	tests := []struct {
		name string
		got  EventType
		want string
	}{
		{"alert", EventTypeAlert, "alert"},
		{"reminder", EventTypeReminder, "reminder"},
		{"voice_call", EventTypeVoiceCall, "voice_call"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Errorf("EventType = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestSeverities(t *testing.T) {
	tests := []struct {
		name string
		got  Severity
		want string
	}{
		{"P0", SeverityP0, "P0"},
		{"P1", SeverityP1, "P1"},
		{"P2", SeverityP2, "P2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Errorf("Severity = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestAlertPushEventFields(t *testing.T) {
	ev := AlertPushEvent{
		AlertID:   "a-1",
		ElderlyID: "elderly-1",
		Severity:  SeverityP0,
		AlertType: "sos",
		Message:   "SOS triggered",
		RawData:   map[string]interface{}{"lat": 31.23},
	}
	if ev.AlertID != "a-1" || ev.ElderlyID != "elderly-1" || ev.Severity != SeverityP0 {
		t.Error("AlertPushEvent fields not set correctly")
	}
	if ev.RawData["lat"] != 31.23 {
		t.Error("RawData not preserved")
	}
}

func TestReminderPushEventFields(t *testing.T) {
	ev := ReminderPushEvent{
		ElderlyID: "elderly-1",
		RuleID:    "rule-1",
		Message:   "Take your medicine",
	}
	if ev.ElderlyID != "elderly-1" || ev.RuleID != "rule-1" {
		t.Error("ReminderPushEvent fields not set correctly")
	}
}

func TestVoiceCallEventFields(t *testing.T) {
	ev := VoiceCallEvent{
		ElderlyID: "elderly-1",
		CallerID:  "user-family-1",
		Message:   "Call back immediately",
	}
	if ev.CallerID != "user-family-1" {
		t.Error("VoiceCallEvent caller ID not set correctly")
	}
}
