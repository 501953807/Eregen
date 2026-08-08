package model

import "time"

// RiskLevel indicates severity of a health anomaly.
type RiskLevel string

const (
	RiskNormal   RiskLevel = "normal"
	RiskElevated RiskLevel = "elevated"
	RiskCritical RiskLevel = "critical"
)

// AnalysisResult records a single metric anomaly detection.
type AnalysisResult struct {
	ID          uint64    `json:"id"`
	ElderlyID   string    `json:"elderly_id"`
	Metric      string    `json:"metric"`       // "hr", "spo2", "steps", "bp_systolic", etc.
	Value       float64   `json:"value"`
	Baseline    float64   `json:"baseline"`     // 7-day average
	Deviation   float64   `json:"deviation"`    // percentage deviation from baseline
	RiskLevel   RiskLevel `json:"risk_level"`
	Message     string    `json:"message,omitempty"` // human-readable description of the anomaly
	Timestamp   time.Time `json:"timestamp"`
	CreatedAt   time.Time `json:"created_at"`
}

// MedicationAdherence tracks medication compliance over a period.
type MedicationAdherence struct {
	ID              uint64    `json:"id"`
	ElderlyID       string    `json:"elderly_id"`
	PeriodStart     time.Time `json:"period_start"`
	PeriodEnd       time.Time `json:"period_end"`
	ScheduledDoses  int       `json:"scheduled_doses"`
	TakenDoses      int       `json:"taken_doses"`
	AdherenceRate   float64   `json:"adherence_rate"` // 0-100
	MissedMedications []MissedMed `json:"missed_medications,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// MissedMed represents a single missed dose.
type MissedMed struct {
	RuleID    string    `json:"rule_id"`
	Scheduled time.Time `json:"scheduled"`
	MissedBy  time.Duration `json:"missed_by"` // how long past scheduled time
}

// RiskScore is the composite health risk score (0-100).
type RiskScore struct {
	ID             uint64    `json:"id"`
	ElderlyID      string    `json:"elderly_id"`
	CompositeScore int       `json:"composite_score"` // 0-100
	VitalsDeviation float64  `json:"vitals_deviation"`
	MedicationAdherence float64 `json:"medication_adherence"`
	ActivityLevel  float64   `json:"activity_level"`
	SleepQuality   float64   `json:"sleep_quality"`
	RecordedAt     time.Time `json:"recorded_at"`
}

// Geofence defines a circular geographic boundary for an elderly person.
type Geofence struct {
	ID           string  `json:"id"`
	ElderlyID    string  `json:"elderly_id"`
	Name         string  `json:"name"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	RadiusMeters int     `json:"radius_meters"`
	Active       bool    `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
}

// Recommendation holds a single chronic-care actionable item.
type Recommendation struct {
	ID           uint64    `json:"id"`
	ElderlyID    string    `json:"elderly_id"`
	Category     string    `json:"category"`     // "glucose", "uric_acid", "blood_pressure"
	Severity     string    `json:"severity"`     // "P0", "P1", "P2"
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Action       string    `json:"action"`        // suggested next action
	SourceMetric string    `json:"source_metric"` // which metric triggered this
	Value        float64   `json:"value"`         // the actual measurement value
	CreatedAt    time.Time `json:"created_at"`
}

// HealthRecord represents a single device health data point.
type HealthRecord struct {
	ID           uint64     `json:"id"`
	ElderlyID    string     `json:"elderly_id"`
	HeartRate    *float64   `json:"heart_rate,omitempty"`
	SpO2         *float64   `json:"spo2,omitempty"`
	Steps        *int       `json:"steps,omitempty"`
	SystolicBP   *int       `json:"systolic_bp,omitempty"`
	Temperature  *float64   `json:"temperature,omitempty"`
	RecordedAt   time.Time  `json:"recorded_at"`
}
