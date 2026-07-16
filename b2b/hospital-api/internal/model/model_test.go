package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestInstitutionTypeConstants(t *testing.T) {
	tests := []struct {
		name string
		it   InstitutionType
		want string
	}{
		{"hospital", InstitutionHospital, "hospital"},
		{"nursing_home", InstitutionNursingHome, "nursing_home"},
		{"community_center", InstitutionCommunity, "community_center"},
		{"clinic", InstitutionClinic, "clinic"},
	}
	for _, tt := range tests {
		if string(tt.it) != tt.want {
			t.Errorf("InstitutionType %s = %q, want %q", tt.name, tt.it, tt.want)
		}
	}
}

func TestAccessLevelConstants(t *testing.T) {
	tests := []struct {
		name string
		al   AccessLevel
		want string
	}{
		{"read", AccessRead, "read"},
		{"read_write", AccessReadWrite, "read_write"},
		{"emergency_only", AccessEmergency, "emergency_only"},
	}
	for _, tt := range tests {
		if string(tt.al) != tt.want {
			t.Errorf("AccessLevel %s = %q, want %q", tt.name, tt.al, tt.want)
		}
	}
}

func TestHealthDataRequestJSON(t *testing.T) {
	now := time.Now()
	eregenID := "elderly-456"
	vitals := []VitalSign{
		{Type: "heart_rate", Value: 72, Unit: "bpm", Normal: nil},
		{Type: "bp_systolic", Value: 120, Unit: "mmHg", Normal: nil},
	}
	diagnoses := []Diagnosis{
		{Code: "I10", Name: "Essential hypertension", Severity: "mild"},
	}
	meds := []HospitalMedication{
		{Name: "氨氯地平", Dose: "5mg", Freq: "qd", Route: "oral", Duration: "30 days"},
	}
	req := HealthDataRequest{
		PatientID: "HIS-00123",
		EregenID:  &eregenID,
		Timestamp: now,
		Vitals:    vitals,
		Diagnoses: diagnoses,
		Medications: meds,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded HealthDataRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.PatientID != req.PatientID {
		t.Errorf("PatientID = %q, want %q", decoded.PatientID, req.PatientID)
	}
	if decoded.EregenID == nil || *decoded.EregenID != eregenID {
		t.Errorf("EregenID = %v, want %q", decoded.EregenID, eregenID)
	}
	if len(decoded.Vitals) != len(vitals) {
		t.Errorf("Vitals count = %d, want %d", len(decoded.Vitals), len(vitals))
	}
	if len(decoded.Medications) != len(meds) {
		t.Errorf("Medications count = %d, want %d", len(decoded.Medications), len(meds))
	}
}

func TestVitalSignNormalTrue(t *testing.T) {
	ok := true
	vs := VitalSign{Type: "spo2", Value: 98, Unit: "%", Normal: &ok}

	data, err := json.Marshal(vs)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded VitalSign
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Normal == nil || !*decoded.Normal {
		t.Error("Normal should be true")
	}
}

func TestVitalSignNormalFalse(t *testing.T) {
	fail := false
	vs := VitalSign{Type: "temperature", Value: 39.5, Unit: "°C", Normal: &fail}

	data, err := json.Marshal(vs)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded VitalSign
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Normal == nil || *decoded.Normal {
		t.Error("Normal should be false for high temperature")
	}
}

func TestDiagnosisJSON(t *testing.T) {
	d := Diagnosis{Code: "E11", Name: "Type 2 diabetes mellitus", Severity: "moderate"}

	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Diagnosis
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Code != d.Code {
		t.Errorf("Code = %q, want %q", decoded.Code, d.Code)
	}
	if decoded.Severity != d.Severity {
		t.Errorf("Severity = %q, want %q", decoded.Severity, d.Severity)
	}
}

func TestHospitalMedicationJSON(t *testing.T) {
	med := HospitalMedication{Name: "阿司匹林", Dose: "100mg", Freq: "qd", Route: "oral", Duration: "30 days"}

	data, err := json.Marshal(med)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded HospitalMedication
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Freq != med.Freq {
		t.Errorf("Freq = %q, want %q", decoded.Freq, med.Freq)
	}
}

func TestElderlyInstitutionLinkJSON(t *testing.T) {
	now := time.Now()
	link := ElderlyInstitutionLink{
		ElderlyID:     "elderly-789",
		InstitutionID: "inst-001",
		AdmittedAt:    &now,
		PrimaryDoc:    strPtr("Dr. Zhang"),
		Active:        true,
	}

	data, err := json.Marshal(link)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded ElderlyInstitutionLink
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.ElderlyID != link.ElderlyID {
		t.Errorf("ElderlyID = %q, want %q", decoded.ElderlyID, link.ElderlyID)
	}
	if !decoded.Active {
		t.Error("Active should be true")
	}
	if decoded.DischargedAt != nil {
		t.Error("DischargedAt should be nil")
	}
}

func TestHealthReportJSON(t *testing.T) {
	now := time.Now()
	report := HealthReport{
		ElderlyID:   "elderly-123",
		ReportDate:  now,
		PeriodStart: now.AddDate(0, 0, -30),
		PeriodEnd:   now,
		Summary: ReportSummary{
			AvgHR:     floatPtr(72.0),
			AvgSPO2:   floatPtr(98.0),
			AvgSteps:  floatPtr(6500.0),
			RiskLevel: "low",
		},
		AlertCount:   2,
		MedAdherence: 95.5,
	}

	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded HealthReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.MedAdherence != report.MedAdherence {
		t.Errorf("MedAdherence = %f, want %f", decoded.MedAdherence, report.MedAdherence)
	}
	if decoded.Summary.RiskLevel != "low" {
		t.Errorf("RiskLevel = %q, want %q", decoded.Summary.RiskLevel, "low")
	}
}

func TestInstitutionJSON(t *testing.T) {
	now := time.Now()
	inst := Institution{
		Name:         "上海市第一人民医院",
		Type:         InstitutionHospital,
		Code:         "HOS-SH-001",
		ContactName:  "李医生",
		ContactPhone: "021-12345678",
		AccessLevel:  AccessReadWrite,
		Status:       "active",
		CreatedAt:    now,
	}

	data, err := json.Marshal(inst)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Institution
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Name != inst.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, inst.Name)
	}
	if decoded.Type != inst.Type {
		t.Errorf("Type = %q, want %q", decoded.Type, inst.Type)
	}
	if decoded.AccessLevel != inst.AccessLevel {
		t.Errorf("AccessLevel = %q, want %q", decoded.AccessLevel, inst.AccessLevel)
	}
}

func TestAPIKeyJSON(t *testing.T) {
	now := time.Now()
	key := InstitutionAPIKey{
		InstitutionID: "inst-001",
		KeyHash:       "sha256:hashed_value",
		Name:          "Production Key",
		ExpiresAt:     now.Add(24 * 365 * 24 * time.Hour),
		Active:        true,
	}

	data, err := json.Marshal(key)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded InstitutionAPIKey
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Name != key.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, key.Name)
	}
}

func TestAlertForwardRequestJSON(t *testing.T) {
	now := time.Now()
	req := AlertForwardRequest{
		ElderlyID:     "elderly-123",
		InstitutionID: "inst-001",
		AlertType:     "sos",
		Severity:      "P0",
		Message:       "紧急告警：老人按下SOS按钮",
		Timestamp:     now,
		LatLon: &struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		}{Lat: 31.2304, Lon: 121.4737},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded AlertForwardRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.AlertType != req.AlertType {
		t.Errorf("AlertType = %q, want %q", decoded.AlertType, req.AlertType)
	}
	if decoded.LatLon == nil {
		t.Fatal("LatLon should not be nil")
	}
}

func TestEmptyEregenID(t *testing.T) {
	req := HealthDataRequest{
		PatientID: "HIS-001",
		Timestamp: time.Now(),
		Vitals:    []VitalSign{},
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded HealthDataRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.EregenID != nil {
		t.Error("EregenID should be nil when not set")
	}
}

func TestTrendPointJSON(t *testing.T) {
	tp := TrendPoint{Time: time.Date(2026, 7, 16, 10, 0, 0, 0, time.UTC), Value: 72.5}

	data, err := json.Marshal(tp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded TrendPoint
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Value != tp.Value {
		t.Errorf("Value = %f, want %f", decoded.Value, tp.Value)
	}
}

func strPtr(s string) *string { return &s }
func floatPtr(f float64) *float64 { return &f }
