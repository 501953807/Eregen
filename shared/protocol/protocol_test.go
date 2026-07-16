package protocol

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestMessageTypeConstants(t *testing.T) {
	expectedTypes := []MessageType{
		TypeHeartbeat, TypeLocation, TypeHealth, TypeSOS,
		TypeFall, TypeMedStatus, TypeMedRule, TypeConfig,
		TypeTTS, TypeOTA, TypeDeviceInfo, TypeAlertForward,
	}
	for _, mt := range expectedTypes {
		if mt == "" {
			t.Error("message type constant is empty")
		}
	}
}

func TestDeviceTypeConstants(t *testing.T) {
	if DeviceBracelet != "bracelet" {
		t.Errorf("DeviceBracelet = %q, want %q", DeviceBracelet, "bracelet")
	}
	if DevicePillbox != "pillbox" {
		t.Errorf("DevicePillbox = %q, want %q", DevicePillbox, "pillbox")
	}
}

func TestAlertPriorityConstants(t *testing.T) {
	if PriorityP0 != "P0" {
		t.Errorf("PriorityP0 = %q, want %q", PriorityP0, "P0")
	}
	if PriorityP1 != "P1" {
		t.Errorf("PriorityP1 = %q, want %q", PriorityP1, "P1")
	}
	if PriorityP2 != "P2" {
		t.Errorf("PriorityP2 = %q, want %q", PriorityP2, "P2")
	}
}

func TestHeartbeatJSON(t *testing.T) {
	hb := Heartbeat{
		Type:    TypeHeartbeat,
		DevID:   "BR-0001",
		Battery: 85,
	}

	data, err := json.Marshal(hb)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Heartbeat
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.DevID != hb.DevID {
		t.Errorf("DevID = %q, want %q", decoded.DevID, hb.DevID)
	}
	if decoded.Battery != hb.Battery {
		t.Errorf("Battery = %d, want %d", decoded.Battery, hb.Battery)
	}
	if decoded.Type != hb.Type {
		t.Errorf("Type = %q, want %q", decoded.Type, hb.Type)
	}
}

func TestLocationJSON(t *testing.T) {
	now := time.Now()
	loc := Location{
		Type:  TypeLocation,
		DevID: "BR-0002",
		Lat:   31.2304,
		Lon:   121.4737,
		Acc:   5.0,
		Ts:    now,
	}

	data, err := json.Marshal(loc)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Location
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.DevID != loc.DevID {
		t.Errorf("DevID = %q, want %q", decoded.DevID, loc.DevID)
	}
	if decoded.Lat != loc.Lat {
		t.Errorf("Lat = %f, want %f", decoded.Lat, loc.Lat)
	}
	if decoded.Lon != loc.Lon {
		t.Errorf("Lon = %f, want %f", decoded.Lon, loc.Lon)
	}
}

func TestHealthJSON(t *testing.T) {
	now := time.Now()
	hr := 72.0
	spo2 := 98.0
	steps := int64(3456)

	h := Health{
		Type:  TypeHealth,
		DevID: "BR-0001",
		HR:    &hr,
		SPO2:  &spo2,
		Step:  &steps,
		Ts:    now,
	}

	data, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Health
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.HR == nil || *decoded.HR != hr {
		t.Errorf("HR mismatch")
	}
	if decoded.SPO2 == nil || *decoded.SPO2 != spo2 {
		t.Errorf("SPO2 mismatch")
	}
	if decoded.Step == nil || *decoded.Step != steps {
		t.Errorf("Step mismatch")
	}
}

func TestHealthOptionalFields(t *testing.T) {
	h := Health{
		Type: TypeHealth,
		DevID: "BR-0001",
	}
	data, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// Optional fields should not appear in JSON since all are nil
	str := string(data)
	for _, field := range []string{"\"hr\"", "\"spo2\"", "\"step\"", "\"sleep\"", "\"temp\""} {
		if strings.Contains(str, field) {
			t.Errorf("unexpected optional field %s in JSON: %s", field, str)
		}
	}
}

func TestSOSJSON(t *testing.T) {
	now := time.Now()
	sos := SOS{
		Type:  TypeSOS,
		DevID: "BR-0003",
		Lat:   31.2304,
		Lon:   121.4737,
		Ts:    now,
	}

	data, err := json.Marshal(sos)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded SOS
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.DevID != sos.DevID {
		t.Errorf("DevID = %q, want %q", decoded.DevID, sos.DevID)
	}
	if decoded.Type != sos.Type {
		t.Errorf("Type = %q, want %q", decoded.Type, sos.Type)
	}
}

func TestFallJSON(t *testing.T) {
	now := time.Now()
	fall := Fall{
		Type:  TypeFall,
		DevID: "BR-0001",
		Conf:  0.95,
		Lat:   31.2304,
		Lon:   121.4737,
		Ts:    now,
	}

	data, err := json.Marshal(fall)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Fall
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Conf != fall.Conf {
		t.Errorf("Conf = %f, want %f", decoded.Conf, fall.Conf)
	}
}

func TestMedStatusJSON(t *testing.T) {
	now := time.Now()
	ms := MedStatus{
		Type:        TypeMedStatus,
		DevID:       "PX-0001",
		Compartment: 3,
		Taken:       true,
		Timestamp:   now,
	}

	data, err := json.Marshal(ms)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded MedStatus
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Compartment != ms.Compartment {
		t.Errorf("Compartment = %d, want %d", decoded.Compartment, ms.Compartment)
	}
	if !decoded.Taken {
		t.Error("Taken should be true")
	}
}

func TestDeviceInfoJSON(t *testing.T) {
	now := time.Now()
	di := DeviceInfo{
		Type:       TypeDeviceInfo,
		DevID:      "BR-0001",
		Firmware:   "1.2.3",
		DeviceType: DeviceBracelet,
		Tier:       "plus",
		MAC:        "AA:BB:CC:DD:EE:FF",
		Battery:    72,
		Ts:         now,
	}

	data, err := json.Marshal(di)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded DeviceInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Firmware != di.Firmware {
		t.Errorf("Firmware = %q, want %q", decoded.Firmware, di.Firmware)
	}
	if decoded.DeviceType != di.DeviceType {
		t.Errorf("DeviceType = %q, want %q", decoded.DeviceType, di.DeviceType)
	}
}

func TestMedRuleJSON(t *testing.T) {
	rule := MedRule{
		Type:  TypeMedRule,
		DevID: "PX-0001",
		Rules: []MedSchedule{
			{Time: "08:00", Dose: 1, Type: "capsule", Name: "降压药"},
			{Time: "12:00", Dose: 1, Type: "tablet", Name: "维生素D"},
			{Time: "20:00", Dose: 2, Type: "capsule", Name: "钙片"},
		},
	}

	data, err := json.Marshal(rule)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded MedRule
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(decoded.Rules) != 3 {
		t.Errorf("Rules count = %d, want 3", len(decoded.Rules))
	}
	if decoded.Rules[0].Name != "降压药" {
		t.Errorf("Rule name = %q, want %q", decoded.Rules[0].Name, "降压药")
	}
}

func TestConfigJSON(t *testing.T) {
	cfg := Config{
		Type:  TypeConfig,
		DevID: "BR-0001",
		Settings: map[string]any{
			"interval": 30,
			"volume":   80,
			"bright":   100,
		},
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Config
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.DevID != cfg.DevID {
		t.Errorf("DevID = %q, want %q", decoded.DevID, cfg.DevID)
	}
}

func TestTTSJSON(t *testing.T) {
	tts := TTS{
		Type:  TypeTTS,
		DevID: "PX-0001",
		Text:  "爷爷，该吃降压药了",
	}

	data, err := json.Marshal(tts)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded TTS
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Text != tts.Text {
		t.Errorf("Text = %q, want %q", decoded.Text, tts.Text)
	}
}

func TestOTAJSON(t *testing.T) {
	ota := OTA{
		Type:  TypeOTA,
		DevID: "BR-0001",
		URL:   "https://cdn.eregen.dev/firmware/v1.3.0.bin",
		Hash:  "sha256:abc123def456",
		Size:  524288,
	}

	data, err := json.Marshal(ota)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded OTA
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.URL != ota.URL {
		t.Errorf("URL = %q, want %q", decoded.URL, ota.URL)
	}
	if decoded.Size != ota.Size {
		t.Errorf("Size = %d, want %d", decoded.Size, ota.Size)
	}
}

func TestAlertNotificationJSON(t *testing.T) {
	now := time.Now()
	loc := Location{
		Type:  TypeLocation,
		DevID: "BR-0001",
		Lat:   31.2304,
		Lon:   121.4737,
	}
	alert := AlertNotification{
		Type:      TypeAlertForward,
		AlertID:   "alert-001",
		ElderlyID: "elderly-123",
		Priority:  PriorityP0,
		Title:     "紧急告警",
		Message:   "检测到跌倒事件",
		Location:  &loc,
		CreatedAt: now,
		Read:      false,
	}

	data, err := json.Marshal(alert)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded AlertNotification
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Priority != PriorityP0 {
		t.Errorf("Priority = %q, want %q", decoded.Priority, PriorityP0)
	}
	if decoded.Read {
		t.Error("Read should be false")
	}
	if decoded.Location == nil {
		t.Fatal("Location should not be nil")
	}
	if decoded.Location.Lat != 31.2304 {
		t.Errorf("Location Lat = %f, want 31.2304", decoded.Location.Lat)
	}
}

func TestMQTTTopics(t *testing.T) {
	topics := []string{
		TopicUpHeartbeat,
		TopicUpLocation,
		TopicUpHealth,
		TopicUpSOS,
		TopicUpFall,
		TopicUpMedStatus,
		TopicUpDeviceInfo,
		TopicDownConfig,
		TopicDownTTS,
		TopicDownOTA,
		TopicDownMedRule,
		TopicAlertNotify,
	}
	for _, topic := range topics {
		if topic == "" {
			t.Error("MQTT topic constant is empty")
		}
		// Should contain format specifiers
		if len(topic) < 5 {
			t.Errorf("topic too short: %q", topic)
		}
	}
}

func TestNATSSubjects(t *testing.T) {
	subjects := []string{
		NATSHealthData, NATSLocData, NATSSOSAlert,
		NATSFallAlert, NATSMedStatus, NATSDeviceOnline,
		NATSDeviceOffline, NATSPushRequest,
	}
	for _, subj := range subjects {
		if subj == "" {
			t.Error("NATS subject constant is empty")
		}
		if subj != "eregen."+subj[7:] {
			t.Errorf("subject %q doesn't follow eregen.* pattern", subj)
		}
	}
}

func TestBraceletTierConstants(t *testing.T) {
	if TierStarter != "starter" {
		t.Errorf("TierStarter = %q, want %q", TierStarter, "starter")
	}
	if TierPlus != "plus" {
		t.Errorf("TierPlus = %q, want %q", TierPlus, "plus")
	}
	if TierPro != "pro" {
		t.Errorf("TierPro = %q, want %q", TierPro, "pro")
	}
}

func TestPillboxTierConstants(t *testing.T) {
	if TierBasic != "basic" {
		t.Errorf("TierBasic = %q, want %q", TierBasic, "basic")
	}
	if TierSmart != "smart" {
		t.Errorf("TierSmart = %q, want %q", TierSmart, "smart")
	}
	if TierAuto != "auto" {
		t.Errorf("TierAuto = %q, want %q", TierAuto, "auto")
	}
}

func TestHeartbeatRoundTrip(t *testing.T) {
	original := Heartbeat{Type: TypeHeartbeat, DevID: "BR-TEST", Battery: 42}

	data, _ := json.Marshal(original)
	var decoded Heartbeat
	json.Unmarshal(data, &decoded)

	if decoded.DevID != original.DevID || decoded.Battery != original.Battery || decoded.Type != original.Type {
		t.Error("heartbeat round-trip failed")
	}
}

func TestMedScheduleRoundTrip(t *testing.T) {
	original := MedSchedule{Time: "08:00", Dose: 1, Type: "tablet", Name: "阿司匹林"}

	data, _ := json.Marshal(original)
	var decoded MedSchedule
	json.Unmarshal(data, &decoded)

	if decoded.Time != original.Time || decoded.Dose != original.Dose || decoded.Name != original.Name {
		t.Error("med schedule round-trip failed")
	}
}
