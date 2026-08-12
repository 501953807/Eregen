package store

import "context"

// Database is the unified interface for all database backends.
type Database interface {
	InstitutionStore
	VitalStore
	DiagnosisStore
	MedicationStore
	PatientLinkStore
}

// InstitutionStore handles institution and API key operations.
type InstitutionStore interface {
	CreateInstitution(ctx context.Context, inst *Institution) error
	GetInstitutionByID(ctx context.Context, id string) (*Institution, error)
	GetInstitutionByCode(ctx context.Context, code string) (*Institution, error)
	ListInstitutions(ctx context.Context, page, pageSize int) ([]Institution, int, error)
	UpdateInstitution(ctx context.Context, id string, inst *Institution) error
	CreateAPIKey(ctx context.Context, key *InstitutionAPIKey) error
	GetInstitutionByAPIKey(ctx context.Context, keyHash string) (*Institution, error)
}

// VitalStore handles vital sign records.
type VitalStore interface {
	StoreVitals(ctx context.Context, v *VitalSignRecord) error
	BulkStoreVitals(ctx context.Context, vitals []*VitalSignRecord) error
	GetVitalsForElderly(ctx context.Context, elderlyID string, days int) ([]VitalSignRecord, error)
}

// DiagnosisStore handles diagnosis records.
type DiagnosisStore interface {
	StoreDiagnoses(ctx context.Context, records []*DiagnosisRecord) error
	GetDiagnosesForElderly(ctx context.Context, elderlyID string, days int) ([]DiagnosisRecord, error)
}

// MedicationStore handles medication records.
type MedicationStore interface {
	StoreMedications(ctx context.Context, records []*MedicationRecord) error
	GetMedicationsForElderly(ctx context.Context, elderlyID string) ([]MedicationRecord, error)
}

// PatientLinkStore handles patient linking.
type PatientLinkStore interface {
	LinkElderlyToExternalPatient(ctx context.Context, elderlyID, patientID, eregenID string) error
	FindElderlyByExternalPatient(ctx context.Context, patientID string) (string, error)
}

// ElderlyLinkStore handles elderly-institution links.
type ElderlyLinkStore interface {
	LinkElderlyToInstitution(ctx context.Context, link *ElderlyInstitutionLink) error
	GetActiveLinksForInstitution(ctx context.Context, instID string) ([]ElderlyInstitutionLink, error)
	GetActiveLinksForElderly(ctx context.Context, elderlyID string) ([]ElderlyInstitutionLink, error)
}
