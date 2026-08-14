package store

import (
	"context"

	"eregen.dev/b2b-insurance-integration/internal/model"
)

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
	CreateProvider(ctx context.Context, p *model.InsuranceProvider) error
	UpdateProvider(ctx context.Context, p *model.InsuranceProvider) error
	GetProviderByID(ctx context.Context, id string) (*model.InsuranceProvider, error)
	ListProviders(ctx context.Context, page, pageSize int) ([]model.InsuranceProvider, int, error)
}

// PolicyStore handles insurance policies.
type PolicyStore interface {
	CreatePolicy(ctx context.Context, policy *model.Policy) error
	UpdatePolicy(ctx context.Context, policy *model.Policy) error
	GetPolicyByID(ctx context.Context, id string) (*model.Policy, error)
	GetPoliciesForElderly(ctx context.Context, elderlyID string) ([]model.Policy, error)
}

// ClaimStore handles insurance claims.
type ClaimStore interface {
	CreateClaim(ctx context.Context, claim *model.InsuranceClaim) error
	UpdateClaimStatus(ctx context.Context, claimID string, status model.ClaimStatus, notes string) error
	GetClaimByID(ctx context.Context, claimID string) (*model.InsuranceClaim, error)
	GetClaimsForElderly(ctx context.Context, elderlyID string) ([]model.InsuranceClaim, error)
	ListClaims(ctx context.Context, status model.ClaimStatus, page, pageSize int) ([]model.InsuranceClaim, int, error)
}

// EvidenceStore handles evidence files for claims.
type EvidenceStore interface {
	AddEvidenceFile(ctx context.Context, file *model.EvidenceFile) error
	GetEvidenceForClaim(ctx context.Context, claimID string) ([]model.EvidenceFile, error)
}

// ExportStore handles health data exports.
type ExportStore interface {
	CreateExport(ctx context.Context, export *model.HealthDataExport) error
	MarkExportReady(ctx context.Context, exportID string, fileURL string) error
	GetExportByID(ctx context.Context, id string) (*model.HealthDataExport, error)
}

// ReminderStore handles premium payment reminders.
type ReminderStore interface {
	CreateReminder(ctx context.Context, reminder *model.PremiumReminder) error
	GetUpcomingReminders(ctx context.Context, daysAhead int) ([]model.PremiumReminder, error)
}
