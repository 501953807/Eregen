package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestClaimTypeConstants(t *testing.T) {
	tests := []struct {
		name string
		ct   ClaimType
		want string
	}{
		{"accident", ClaimAccident, "accident"},
		{"illness", ClaimIllness, "illness"},
		{"emergency", ClaimEmergency, "emergency"},
		{"health_check", ClaimHealthCheck, "health_check"},
		{"chronic_disease", ClaimChronicDisease, "chronic_disease"},
	}
	for _, tt := range tests {
		if string(tt.ct) != tt.want {
			t.Errorf("ClaimType %s = %q, want %q", tt.name, tt.ct, tt.want)
		}
	}
}

func TestClaimStatusConstants(t *testing.T) {
	tests := []struct {
		name string
		cs   ClaimStatus
		want string
	}{
		{"pending", ClaimPending, "pending"},
		{"submitted", ClaimSubmitted, "submitted"},
		{"under_review", ClaimUnderReview, "under_review"},
		{"approved", ClaimApproved, "approved"},
		{"rejected", ClaimRejected, "rejected"},
		{"paid", ClaimPaid, "paid"},
	}
	for _, tt := range tests {
		if string(tt.cs) != tt.want {
			t.Errorf("ClaimStatus %s = %q, want %q", tt.name, tt.cs, tt.want)
		}
	}
}

func TestInsuranceProviderJSON(t *testing.T) {
	prov := InsuranceProvider{
		ID:          "provider-001",
		Name:        "中国人保",
		Code:        "PICC",
		APIEndpoint: "https://api.picc.cn/v1",
		Active:      true,
	}

	data, err := json.Marshal(prov)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded InsuranceProvider
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Name != prov.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, prov.Name)
	}
	if decoded.Code != prov.Code {
		t.Errorf("Code = %q, want %q", decoded.Code, prov.Code)
	}
	if !decoded.Active {
		t.Error("Active should be true")
	}
}

func TestPolicyJSON(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	policy := Policy{
		ID:           "policy-001",
		ElderlyID:    "elderly-123",
		ProviderID:   "provider-001",
		PlanName:     "颐贞长寿医疗险",
		PlanCode:     "YZZX-2026",
		PolicyNumber: "POL-2026-001234",
		StartDate:    start,
		EndDate:      end,
		CoverageLimit: 500000,
		Premium:      3600,
		Status:       "active",
	}

	data, err := json.Marshal(policy)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Policy
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.PolicyNumber != policy.PolicyNumber {
		t.Errorf("PolicyNumber = %q, want %q", decoded.PolicyNumber, policy.PolicyNumber)
	}
	if decoded.CoverageLimit != policy.CoverageLimit {
		t.Errorf("CoverageLimit = %f, want %f", decoded.CoverageLimit, policy.CoverageLimit)
	}
}

func TestClaimJSON(t *testing.T) {
	incident := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	familyID := "family-user-001"
	claim := InsuranceClaim{
		ID:             "claim-001",
		ElderlyID:      "elderly-123",
		FamilyMemberID: familyID,
		ProviderID:     "provider-001",
		ClaimType:      ClaimIllness,
		Status:         ClaimPending,
		IncidentDate:   incident,
		ClaimAmount:    15000,
		CoverageLimit:  50000,
		Description:    "住院治疗费用报销",
	}

	data, err := json.Marshal(claim)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded InsuranceClaim
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.ClaimAmount != claim.ClaimAmount {
		t.Errorf("ClaimAmount = %f, want %f", decoded.ClaimAmount, claim.ClaimAmount)
	}
	if decoded.ClaimType != claim.ClaimType {
		t.Errorf("ClaimType = %q, want %q", decoded.ClaimType, claim.ClaimType)
	}
	if decoded.Status != claim.Status {
		t.Errorf("Status = %q, want %q", decoded.Status, claim.Status)
	}
}

func TestEvidenceFileJSON(t *testing.T) {
	ef := EvidenceFile{
		ID:        "file-001",
		ClaimID:   "claim-001",
		FileType:  "medical_report",
		FileName:  "hospital_receipt.pdf",
		FileURL:   "https://storage.eregen.dev/evidence/file-001.pdf",
	}

	data, err := json.Marshal(ef)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded EvidenceFile
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.FileName != ef.FileName {
		t.Errorf("FileName = %q, want %q", decoded.FileName, ef.FileName)
	}
	if decoded.FileType != ef.FileType {
		t.Errorf("FileType = %q, want %q", decoded.FileType, ef.FileType)
	}
}

func TestPremiumReminderJSON(t *testing.T) {
	remind := PremiumReminder{
		ID:         "reminder-001",
		PolicyID:   "policy-001",
		ElderlyID:  "elderly-123",
		FamilyID:   "family-user-001",
		RemindDate: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		Amount:     3600,
		Sent:       false,
	}

	data, err := json.Marshal(remind)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded PremiumReminder
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Amount != remind.Amount {
		t.Errorf("Amount = %f, want %f", decoded.Amount, remind.Amount)
	}
	if decoded.Sent {
		t.Error("Sent should be false")
	}
}

func TestClaimWithNilTimes(t *testing.T) {
	claim := InsuranceClaim{
		ElderlyID:     "elderly-123",
		ProviderID:    "provider-001",
		ClaimType:     ClaimAccident,
		Status:        ClaimPending,
		ClaimAmount:   5000,
		CoverageLimit: 50000,
	}

	data, err := json.Marshal(claim)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded InsuranceClaim
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.SubmittedAt != nil {
		t.Error("SubmittedAt should be nil for pending claims")
	}
	if decoded.ReviewedAt != nil {
		t.Error("ReviewedAt should be nil for pending claims")
	}
}

func TestClaimApproved(t *testing.T) {
	now := time.Now()
	claim := InsuranceClaim{
		ElderlyID:     "elderly-123",
		ProviderID:    "provider-001",
		ClaimType:     ClaimIllness,
		Status:        ClaimApproved,
		ClaimAmount:   25000,
		CoverageLimit: 50000,
		ReviewedAt:    &now,
		ReviewerNotes: "材料齐全，符合理赔条件",
	}

	data, err := json.Marshal(claim)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded InsuranceClaim
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Status != ClaimApproved {
		t.Errorf("Status = %q, want %q", decoded.Status, ClaimApproved)
	}
	if decoded.ReviewerNotes == "" {
		t.Error("ReviewerNotes should not be empty")
	}
}

func TestPolicyExpired(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	policy := Policy{
		PolicyNumber: "POL-2025-999999",
		PlanName:     "已过期保单",
		StartDate:    start,
		EndDate:      end,
		Status:       "expired",
	}

	data, err := json.Marshal(policy)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Policy
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Status != "expired" {
		t.Errorf("Status = %q, want %q", decoded.Status, "expired")
	}
}

func TestHealthDataExportJSON(t *testing.T) {
	claimID := "claim-001"
	export := HealthDataExport{
		ID:          "export-001",
		ElderlyID:   "elderly-123",
		ClaimID:     &claimID,
		ExportType:  "claim_support",
		PeriodStart: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		FileURL:     "https://storage.eregen.dev/exports/export-001.pdf",
		Status:      "ready",
	}

	data, err := json.Marshal(export)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded HealthDataExport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.ExportType != export.ExportType {
		t.Errorf("ExportType = %q, want %q", decoded.ExportType, export.ExportType)
	}
}

func TestHealthDataExportNoClaim(t *testing.T) {
	export := HealthDataExport{
		ElderlyID:   "elderly-123",
		ExportType:  "annual_report",
		Status:      "generating",
	}

	data, err := json.Marshal(export)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded HealthDataExport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.ClaimID != nil {
		t.Error("ClaimID should be nil for annual report exports")
	}
}

func TestInsuranceProviderInactive(t *testing.T) {
	prov := InsuranceProvider{
		Name:   "已停用保险公司",
		Code:   "OLD-INS",
		Active: false,
	}

	data, err := json.Marshal(prov)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded InsuranceProvider
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Active {
		t.Error("Active should be false")
	}
}

func TestClaimPaid(t *testing.T) {
	now := time.Now()
	claim := InsuranceClaim{
		ID:            "claim-paid",
		ElderlyID:     "elderly-123",
		ProviderID:    "provider-001",
		ClaimType:     ClaimEmergency,
		Status:        ClaimPaid,
		ClaimAmount:   30000,
		CoverageLimit: 100000,
		ReviewedAt:    &now,
	}

	data, err := json.Marshal(claim)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded InsuranceClaim
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Status != ClaimPaid {
		t.Errorf("Status = %q, want %q", decoded.Status, ClaimPaid)
	}
}
