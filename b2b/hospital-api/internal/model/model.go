package model

import (
	"encoding/json"
	"time"
)

// InstitutionType classifies the kind of healthcare institution.
type InstitutionType string

const (
	InstitutionHospital    InstitutionType = "hospital"
	InstitutionNursingHome InstitutionType = "nursing_home"
	InstitutionCommunity   InstitutionType = "community_center"
	InstitutionClinic      InstitutionType = "clinic"
)

// AccessLevel controls what an institution can see.
type AccessLevel string

const (
	AccessRead       AccessLevel = "read"
	AccessReadWrite  AccessLevel = "read_write"
	AccessEmergency  AccessLevel = "emergency_only"
)

// Institution represents a connected hospital, nursing home, or clinic.
type Institution struct {
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	Type          InstitutionType  `json:"type"`
	Code          string           `json:"code"` // hospital registration code
	ContactName   string           `json:"contact_name"`
	ContactPhone  string           `json:"contact_phone"`
	AccessLevel   AccessLevel      `json:"access_level"`
	Status        string           `json:"status"` // active, suspended, pending
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

// InstitutionAPIKey is used for machine-to-machine authentication.
type InstitutionAPIKey struct {
	ID         string    `json:"id"`
	InstitutionID string `json:"institution_id"`
	KeyHash    string    `json:"-"`
	Name       string    `json:"name"`
	ExpiresAt  time.Time `json:"expires_at"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
}

// HealthDataRequest is the standard inbound health data from hospital HIS.
type HealthDataRequest struct {
	PatientID   string                  `json:"patient_id"`   // HIS patient ID
	EregenID    *string                 `json:"eregen_id,omitempty"` // linked eregen elderly profile
	Timestamp   time.Time               `json:"timestamp"`
	Vitals      []VitalSign             `json:"vitals"`
	Diagnoses   []Diagnosis             `json:"diagnoses,omitempty"`
	Medications []HospitalMedication    `json:"medications,omitempty"`
	Notes       *string                 `json:"notes,omitempty"`
}

// VitalSign represents a single vital sign reading from hospital systems.
type VitalSign struct {
	Type     string  `json:"type"` // hr, spo2, bp_systolic, bp_diastolic, temp, weight, height
	Value    float64 `json:"value"`
	Unit     string  `json:"unit,omitempty"`
	Normal   *bool   `json:"normal,omitempty"` // true=normal, false=abnormal
}

// Diagnosis follows ICD-10 coding convention.
type Diagnosis struct {
	Code    string `json:"code"`    // ICD-10 code
	Name    string `json:"name"`    // diagnosis name
	Severity string `json:"severity,omitempty"` // mild, moderate, severe
}

// HospitalMedication represents a prescribed medication from hospital.
type HospitalMedication struct {
	Name     string  `json:"name"`
	Dose     string  `json:"dose"`
	Freq     string  `json:"freq"`     // frequency e.g. "bid", "tid"
	Route    string  `json:"route"`    // oral, iv, etc.
	Duration string  `json:"duration"` // e.g. "7 days"
}

// ElderlyInstitutionLink connects an elderly profile to an institution.
type ElderlyInstitutionLink struct {
	ID            string     `json:"id"`
	ElderlyID     string     `json:"elderly_id"`
	InstitutionID string     `json:"institution_id"`
	AdmittedAt    *time.Time `json:"admitted_at,omitempty"`
	DischargedAt  *time.Time `json:"discharged_at,omitempty"`
	PrimaryDoc    *string   `json:"primary_doc,omitempty"`
	Notes         json.RawMessage `json:"notes,omitempty"`
	Active        bool      `json:"active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// HealthReport is a standardized output for institutions to consume.
type HealthReport struct {
	ElderlyID    string               `json:"elderly_id"`
	ReportDate   time.Time            `json:"report_date"`
	PeriodStart  time.Time            `json:"period_start"`
	PeriodEnd    time.Time            `json:"period_end"`
	Summary      ReportSummary        `json:"summary"`
	VitalsTrend  []VitalTrend         `json:"vitals_trend"`
	Diagnoses    []DiagnosisRecord    `json:"diagnoses,omitempty"`
	Medications  []MedicationRecord   `json:"medications,omitempty"`
	AlertCount   int                  `json:"alert_count"`
	MedAdherence float64              `json:"med_adherence"` // percentage
}

// ReportSummary is a high-level health snapshot.
type ReportSummary struct {
	AvgHR       *float64 `json:"avg_hr,omitempty"`
	AvgSPO2     *float64 `json:"avg_spo2,omitempty"`
	AvgSteps    *float64 `json:"avg_steps,omitempty"`
	AvgSleepHrs *float64 `json:"avg_sleep_hrs,omitempty"`
	RiskLevel   string   `json:"risk_level"` // low, medium, high, critical
}

// VitalTrend tracks a vital sign over time.
type VitalTrend struct {
	Type   string    `json:"type"`
	Values []TrendPoint `json:"values"`
}

// TrendPoint is a single point on a trend line.
type TrendPoint struct {
	Time  time.Time `json:"time"`
	Value float64   `json:"value"`
}

// AlertForwardRequest is sent from Eregen to an institution when P0/P1 alerts fire.
type AlertForwardRequest struct {
	ElderlyID  string    `json:"elderly_id"`
	InstitutionID string `json:"institution_id"`
	AlertType  string    `json:"alert_type"`
	Severity   string    `json:"severity"`
	Message    string    `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
	LatLon     *struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"lat_lon,omitempty"`
}

// VitalSignRecord is the persisted vital sign entry in b2b_vital_signs table.
type VitalSignRecord struct {
	ID           string    `json:"id"`
	ElderlyID    string    `json:"elderly_id"`
	InstitutionID string   `json:"institution_id"`
	PatientID    string    `json:"patient_id"`
	HeartRate    *int      `json:"heart_rate,omitempty"`
	SPO2         *int      `json:"spo2,omitempty"`
	SystolicBP   *int      `json:"systolic_bp,omitempty"`
	DiastolicBP  *int      `json:"diastolic_bp,omitempty"`
	Temperature  *float64  `json:"temperature,omitempty"`
	Steps        *int64    `json:"steps,omitempty"`
	RecordedAt   time.Time `json:"recorded_at"`
}

// DiagnosisRecord is a persisted diagnosis from hospital HIS.
type DiagnosisRecord struct {
	ID            string     `json:"id"`
	ElderlyID     string     `json:"elderly_id"`
	InstitutionID string     `json:"institution_id"`
	PatientID     string     `json:"patient_id"`
	DiagnosisCode string     `json:"diagnosis_code"`
	DiagnosisName string     `json:"diagnosis_name"`
	Severity      string     `json:"severity"`
	DiagnosedAt   time.Time  `json:"diagnosed_at"`
}

// MedicationRecord is a persisted medication prescription from hospital HIS.
type MedicationRecord struct {
	ID             string    `json:"id"`
	ElderlyID      string    `json:"elderly_id"`
	InstitutionID  string    `json:"institution_id"`
	PatientID      string    `json:"patient_id"`
	MedicationName string    `json:"medication_name"`
	Dose           string    `json:"dose"`
	Frequency      string    `json:"frequency"`
	Route          string    `json:"route"`
	Duration       string    `json:"duration"`
	PrescribedAt   time.Time `json:"prescribed_at"`
}
