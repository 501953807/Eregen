package store

import (
	"context"

	"eregen.dev/b2b-hospital-api/internal/model"
)

// Database is the unified interface for all database backends.
type Database interface {
	InstitutionStore
	VitalStore
	DiagnosisStore
	MedicationStore
	PatientLinkStore
	ElderlyLinkStore
}

// InstitutionStore handles institution and API key operations.
type InstitutionStore interface {
	CreateInstitution(ctx context.Context, inst *model.Institution) error
	GetInstitutionByID(ctx context.Context, id string) (*model.Institution, error)
	GetInstitutionByCode(ctx context.Context, code string) (*model.Institution, error)
	ListInstitutions(ctx context.Context, page, pageSize int) ([]model.Institution, int, error)
	UpdateInstitution(ctx context.Context, id string, inst *model.Institution) error
	CreateAPIKey(ctx context.Context, key *model.InstitutionAPIKey) error
	GetInstitutionByAPIKey(ctx context.Context, keyHash string) (*model.Institution, error)
}

// VitalStore handles vital sign records.
type VitalStore interface {
	StoreVitals(ctx context.Context, v *model.VitalSignRecord) error
	BulkStoreVitals(ctx context.Context, vitals []*model.VitalSignRecord) error
	GetVitalsForElderly(ctx context.Context, elderlyID string, days int) ([]model.VitalSignRecord, error)
}

// DiagnosisStore handles diagnosis records.
type DiagnosisStore interface {
	StoreDiagnoses(ctx context.Context, records []*model.DiagnosisRecord) error
	GetDiagnosesForElderly(ctx context.Context, elderlyID string, days int) ([]model.DiagnosisRecord, error)
}

// MedicationStore handles medication records.
type MedicationStore interface {
	StoreMedications(ctx context.Context, records []*model.MedicationRecord) error
	GetMedicationsForElderly(ctx context.Context, elderlyID string) ([]model.MedicationRecord, error)
}

// PatientLinkStore handles patient linking.
type PatientLinkStore interface {
	LinkElderlyToExternalPatient(ctx context.Context, elderlyID, patientID, eregenID string) error
	FindElderlyByExternalPatient(ctx context.Context, patientID string) (string, error)
}

// ElderlyLinkStore handles elderly-institution links.
type ElderlyLinkStore interface {
	LinkElderlyToInstitution(ctx context.Context, link *model.ElderlyInstitutionLink) error
	GetActiveLinksForInstitution(ctx context.Context, instID string) ([]model.ElderlyInstitutionLink, error)
	GetActiveLinksForElderly(ctx context.Context, elderlyID string) ([]model.ElderlyInstitutionLink, error)
}
