package model

import "time"

// InsuranceProvider represents an insurance company connected to the platform.
type InsuranceProvider struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"` // internal provider code
	APIEndpoint string    `json:"api_endpoint"`
	APIKey      string    `json:"-"` // never expose in responses
	Secret      string    `json:"-"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
}

// ClaimType categorizes the type of insurance claim.
type ClaimType string

const (
	ClaimAccident   ClaimType = "accident"
	ClaimIllness    ClaimType = "illness"
	ClaimEmergency  ClaimType = "emergency"
	ClaimHealthCheck ClaimType = "health_check"
	ClaimChronicDisease ClaimType = "chronic_disease"
)

// ClaimStatus tracks the lifecycle of an insurance claim.
type ClaimStatus string

const (
	ClaimPending    ClaimStatus = "pending"
	ClaimSubmitted  ClaimStatus = "submitted"
	ClaimUnderReview ClaimStatus = "under_review"
	ClaimApproved   ClaimStatus = "approved"
	ClaimRejected   ClaimStatus = "rejected"
	ClaimPaid       ClaimStatus = "paid"
)

// InsuranceClaim is the main claim entity.
type InsuranceClaim struct {
	ID             string       `json:"id"`
	ElderlyID      string       `json:"elderly_id"`
	FamilyMemberID string       `json:"family_member_id"`
	ProviderID     string       `json:"provider_id"`
	ClaimType      ClaimType    `json:"claim_type"`
	Status         ClaimStatus  `json:"status"`
	IncidentDate   time.Time    `json:"incident_date"`
	ClaimAmount    float64      `json:"claim_amount"`
	CoverageLimit  float64      `json:"coverage_limit"`
	Description    string       `json:"description"`
	EvidenceFiles  []EvidenceFile `json:"evidence_files,omitempty"`
	SubmittedAt    *time.Time   `json:"submitted_at,omitempty"`
	ReviewedAt     *time.Time   `json:"reviewed_at,omitempty"`
	ReviewerNotes  string       `json:"reviewer_notes,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

// EvidenceFile stores uploaded evidence for a claim.
type EvidenceFile struct {
	ID        string    `json:"id"`
	ClaimID   string    `json:"claim_id"`
	FileType  string    `json:"file_type"` // medical_report, photo, video, lab_result
	FileName  string    `json:"file_name"`
	FileURL   string    `json:"file_url"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// Policy represents an elderly person's insurance policy.
type Policy struct {
	ID           string     `json:"id"`
	ElderlyID    string     `json:"elderly_id"`
	ProviderID   string     `json:"provider_id"`
	PlanName     string     `json:"plan_name"`
	PlanCode     string     `json:"plan_code"`
	PolicyNumber string     `json:"policy_number"`
	StartDate    time.Time  `json:"start_date"`
	EndDate      time.Time  `json:"end_date"`
	CoverageLimit float64   `json:"coverage_limit"`
	Premium      float64    `json:"premium"`
	Status       string     `json:"status"` // active, expired, cancelled
	CreatedAt    time.Time  `json:"created_at"`
}

// HealthDataExport is a standardized health report exported for insurance purposes.
type HealthDataExport struct {
	ID            string    `json:"id"`
	ElderlyID     string    `json:"elderly_id"`
	ClaimID       *string   `json:"claim_id,omitempty"`
	ExportType    string    `json:"export_type"` // claim_support, annual_report, risk_assessment
	PeriodStart   time.Time `json:"period_start"`
	PeriodEnd     time.Time `json:"period_end"`
	FileURL       string    `json:"file_url"`
	GeneratedAt   time.Time `json:"generated_at"`
	Status        string    `json:"status"` // generating, ready, expired
}

// PremiumReminder is a scheduled notification for policy renewal.
type PremiumReminder struct {
	ID         string    `json:"id"`
	PolicyID   string    `json:"policy_id"`
	ElderlyID  string    `json:"elderly_id"`
	FamilyID   string    `json:"family_id"`
	RemindDate time.Time `json:"remind_date"`
	Amount     float64   `json:"amount"`
	Sent       bool      `json:"sent"`
	CreatedAt  time.Time `json:"created_at"`
}
