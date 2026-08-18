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
	ExportStore
	RulesStore
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

// ExportStore handles health data export requests.
type ExportStore interface {
	CreateExport(ctx context.Context, export *model.ExportRequest) error
	GetExportByID(ctx context.Context, exportID string) (*model.ExportRequest, error)
	ListExportsByInstitution(ctx context.Context, instID string, elderlyID string) ([]model.ExportRequest, error)
}

// RulesStore handles hospital medication rules.
type RulesStore interface {
	CreateMedicationRule(ctx context.Context, rule *model.MedicationRuleV2) error
	GetMedicationRulesByPerson(ctx context.Context, personID string) ([]model.MedicationRuleV2, error)
	GetMedicationRulesByInstitution(ctx context.Context, instID string) ([]model.MedicationRuleV2, error)
}
