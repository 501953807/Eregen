package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestServiceTypeConstants(t *testing.T) {
	tests := []struct {
		name string
		st   ServiceType
		want string
	}{
		{"home_visit", ServiceHomeVisit, "home_visit"},
		{"health_check", ServiceHealthCheck, "health_check"},
		{"med_delivery", ServiceMedDelivery, "med_delivery"},
		{"emergency", ServiceEmergency, "emergency_response"},
		{"rehab", ServiceRehab, "rehabilitation"},
	}
	for _, tt := range tests {
		if string(tt.st) != tt.want {
			t.Errorf("ServiceType %s = %q, want %q", tt.name, tt.st, tt.want)
		}
	}
}

func TestCommunityRoleConstants(t *testing.T) {
	tests := []struct {
		name string
		cr   CommunityRole
		want string
	}{
		{"admin", RoleAdmin, "admin"},
		{"caregiver", RoleCaregiver, "caregiver"},
		{"volunteer", RoleVolunteer, "volunteer"},
		{"elderly", RoleElderly, "elderly"},
	}
	for _, tt := range tests {
		if string(tt.cr) != tt.want {
			t.Errorf("CommunityRole %s = %q, want %q", tt.name, tt.cr, tt.want)
		}
	}
}

func TestEventJSON(t *testing.T) {
	start := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 1, 17, 0, 0, 0, time.UTC)
	event := CommunityEvent{
		ID:             "event-001",
		Name:           "夏季健康讲座",
		Description:    "高血压预防与控制",
		ServiceType:    ServiceHealthCheck,
		Location:       "社区活动中心一楼",
		StartTime:      start,
		EndTime:        end,
		MaxParticipants: 50,
		Status:         "scheduled",
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded CommunityEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Name != event.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, event.Name)
	}
	if decoded.ServiceType != event.ServiceType {
		t.Errorf("ServiceType = %q, want %q", decoded.ServiceType, event.ServiceType)
	}
	if decoded.MaxParticipants != event.MaxParticipants {
		t.Errorf("MaxParticipants = %d, want %d", decoded.MaxParticipants, event.MaxParticipants)
	}
}

func TestEventRegistrationJSON(t *testing.T) {
	caregiverID := "caregiver-001"
	reg := EventRegistration{
		EventID:       "event-001",
		ElderlyID:     "elderly-123",
		CaregiverID:   &caregiverID,
		Status:        "confirmed",
		RegisteredAt:  time.Now(),
	}

	data, err := json.Marshal(reg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded EventRegistration
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.ElderlyID != reg.ElderlyID {
		t.Errorf("ElderlyID = %q, want %q", decoded.ElderlyID, reg.ElderlyID)
	}
	if decoded.Status != reg.Status {
		t.Errorf("Status = %q, want %q", decoded.Status, reg.Status)
	}
	if decoded.CaregiverID == nil || *decoded.CaregiverID != caregiverID {
		t.Error("CaregiverID should be set")
	}
}

func TestEventRegistrationNoCaregiver(t *testing.T) {
	reg := EventRegistration{
		EventID:    "event-001",
		ElderlyID:  "elderly-456",
		Status:     "attended",
		RegisteredAt: time.Now(),
	}

	data, err := json.Marshal(reg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded EventRegistration
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.CaregiverID != nil {
		t.Error("CaregiverID should be nil")
	}
}

func TestHealthCheckRecordJSON(t *testing.T) {
	bpSys := 128.0
	bpDia := 82.0
	hr := 72.0
	glucose := 5.4
	record := HealthCheckRecord{
		ID:            "check-001",
		ElderlyID:     "elderly-123",
		CheckDate:     time.Now(),
		BP_Systolic:   &bpSys,
		BP_Diastolic:  &bpDia,
		HR:            &hr,
		Glucose:       &glucose,
		Weight:        nil,
		Height:        nil,
		Notes:         "血压略高，建议低盐饮食",
		CheckedBy:     "nurse-wang",
	}

	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded HealthCheckRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.BP_Systolic == nil || *decoded.BP_Systolic != bpSys {
		t.Errorf("BP_Systolic mismatch")
	}
	if decoded.HR == nil || *decoded.HR != hr {
		t.Errorf("HR mismatch")
	}
	if decoded.Weight != nil {
		t.Error("Weight should be nil")
	}
}

func TestCarePlanJSON(t *testing.T) {
	now := time.Now()
	end := now.Add(90 * 24 * time.Hour)
	plan := CarePlan{
		ID:            "cp-001",
		ElderlyID:     "elderly-123",
		Title:         "高血压管理计划",
		Description:   "日常运动与监测计划",
		Tasks: []CareTask{
			{ID: "task-1", Title: "晨间散步", Type: ServiceHomeVisit, Schedule: "daily", DueTime: ptrTime(now.Add(7 * time.Hour)), Completed: false},
			{ID: "task-2", Title: "血压测量", Type: ServiceHealthCheck, Schedule: "daily", DueTime: ptrTime(now.Add(9 * time.Hour)), Completed: true, CompletedAt: &now},
		},
		AssignedTo: "caregiver-001",
		Status:     "active",
		StartDate:  now,
		EndDate:    &end,
	}

	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded CarePlan
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Title != plan.Title {
		t.Errorf("Title = %q, want %q", decoded.Title, plan.Title)
	}
	if len(decoded.Tasks) != 2 {
		t.Errorf("Tasks count = %d, want 2", len(decoded.Tasks))
	}
	if !decoded.Tasks[1].Completed {
		t.Error("Second task should be completed")
	}
}

func TestCareTaskJSON(t *testing.T) {
	ct := CareTask{
		ID:        "task-001",
		Title:     "康复训练",
		Type:      ServiceRehab,
		Schedule:  "mon_wed_fri",
		Completed: false,
	}

	data, err := json.Marshal(ct)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded CareTask
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Title != ct.Title {
		t.Errorf("Title = %q, want %q", decoded.Title, ct.Title)
	}
	if decoded.Completed {
		t.Error("Completed should be false")
	}
}

func TestEventWithZeroTimes(t *testing.T) {
	event := CommunityEvent{
		ID:          "event-002",
		Name:        "无时间限制活动",
		ServiceType: ServiceHomeVisit,
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded CommunityEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.ID != event.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, event.ID)
	}
}

func TestCarePlanPaused(t *testing.T) {
	plan := CarePlan{
		ID:          "cp-002",
		ElderlyID:   "elderly-456",
		Title:       "已暂停计划",
		Status:      "paused",
		AssignedTo:  "caregiver-002",
	}

	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded CarePlan
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Status != "paused" {
		t.Errorf("Status = %q, want %q", decoded.Status, "paused")
	}
}

func ptrTime(t time.Time) *time.Time { return &t }
