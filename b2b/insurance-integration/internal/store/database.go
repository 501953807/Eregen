package store

import "context"

// Store is the unified interface for all database backends.
type Store interface {
	ProviderStore
	PolicyStore
	ClaimStore
	EvidenceStore
	ExportStore
	ReminderStore
}

// ProviderStore handles insurance providers.
type ProviderStore interface {
	CreateProvider(ctx context.Context, p *InsuranceProvider) error
	UpdateProvider(ctx context.Context, p *InsuranceProvider) error
	GetProviderByID(ctx context.Context, id string) (*InsuranceProvider, error)
	ListProviders(ctx context.Context, page, pageSize int) ([]InsuranceProvider, int, error)
}

// PolicyStore handles insurance policies.
type PolicyStore interface {
	CreatePolicy(ctx context.Context, policy *Policy) error
	UpdatePolicy(ctx context.Context, policy *Policy) error
	GetPolicyByID(ctx context.Context, id string) (*Policy, error)
	GetPoliciesForElderly(ctx context.Context, elderlyID string) ([]Policy, error)
}

// ClaimStore handles insurance claims.
type ClaimStore interface {
	CreateClaim(ctx context.Context, claim *InsuranceClaim) error
	UpdateClaimStatus(ctx context.Context, claimID string, status ClaimStatus, notes string) error
	GetClaimByID(ctx context.Context, claimID string) (*InsuranceClaim, error)
	GetClaimsForElderly(ctx context.Context, elderlyID string) ([]InsuranceClaim, error)
	ListClaims(ctx context.Context, status ClaimStatus, page, pageSize int) ([]InsuranceClaim, int, error)
}

// EvidenceStore handles evidence files for claims.
type EvidenceStore interface {
	AddEvidenceFile(ctx context.Context, file *EvidenceFile) error
	GetEvidenceForClaim(ctx context.Context, claimID string) ([]EvidenceFile, error)
}

// ExportStore handles health data exports.
type ExportStore interface {
	CreateExport(ctx context.Context, export *HealthDataExport) error
	MarkExportReady(ctx context.Context, exportID string, fileURL string) error
	GetExportByID(ctx context.Context, id string) (*HealthDataExport, error)
}

// ReminderStore handles premium payment reminders.
type ReminderStore interface {
	CreateReminder(ctx context.Context, reminder *PremiumReminder) error
	GetUpcomingReminders(ctx context.Context, daysAhead int) ([]PremiumReminder, error)
}
