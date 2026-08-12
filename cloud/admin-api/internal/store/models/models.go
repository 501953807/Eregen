package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents an admin or family user.
type User struct {
	gorm.Model
	Name         string  `gorm:"size:100;not null;index"`
	Email        string  `gorm:"uniqueIndex;size:200"`
	Phone        string  `gorm:"uniqueIndex;size:20"`
	Role         string  `gorm:"size:50;default:'family'"`
	PasswordHash string  `gorm:"size:255;not null"`
	OpenID       string  `gorm:"size:255;index"`
}

func (User) TableName() string { return "users" }

// ElderlyProfile represents an elderly person's profile.
type ElderlyProfile struct {
	gorm.Model
	UserID      string   `gorm:"size:255;index"`
	Name        string   `gorm:"size:100;not null"`
	BirthDate   *time.Time
	AvatarURL   *string
	HealthTiers string   `gorm:"type:json"`
}

func (ElderlyProfile) TableName() string { return "elderly_profiles" }

// Device represents a wearable device (bracelet, pillbox, etc.).
type Device struct {
	gorm.Model
	DeviceID    string `gorm:"uniqueIndex;size:100;not null"`
	DeviceType  string `gorm:"size:50;not null;index"`
	Tier        string `gorm:"size:50;not null;index"`
	Status      string `gorm:"size:20;default:'offline';index"`
	LastSeen    *time.Time
	OwnerUserID string `gorm:"size:255;index"`
	Settings    string `gorm:"type:json"`
	OTAURL      string `gorm:"size:500"`
	OTAModel    string `gorm:"size:255"`
	OTAStatus   string `gorm:"size:50;default:''"`
}

func (Device) TableName() string { return "devices" }

// HealthRecord represents a health data point.
type HealthRecord struct {
	gorm.Model
	ElderlyID string `gorm:"size:255;index"`
	HR        *int
	SpO2      *int
	Steps     *int64
	SleepHours *float64
	Timestamp time.Time `gorm:"index"`
}

func (HealthRecord) TableName() string { return "health_records" }

// MedicationRule represents a medication schedule.
type MedicationRule struct {
	gorm.Model
	ElderlyID    string `gorm:"size:255;index"`
	ScheduleTime string `gorm:"size:50"`
	PillType     string `gorm:"size:100"`
	DoseCount    int
	DaysOfWeek   string `gorm:"type:json"`
	Active       bool   `gorm:"default:true"`
}

func (MedicationRule) TableName() string { return "medication_rules" }

// LocationHistory represents a GPS location point.
type LocationHistory struct {
	gorm.Model
	ElderlyID string  `gorm:"size:255;index"`
	Lat       float64
	Lng       float64
	Accuracy  *float64
	Timestamp time.Time `gorm:"index"`
}

func (LocationHistory) TableName() string { return "location_history" }

// Alert represents a safety alert.
type Alert struct {
	gorm.Model
	ElderlyID string `gorm:"size:255;index"`
	AlertType string `gorm:"size:50;not null;index"`
	Severity  string `gorm:"size:20;not null;index"`
	Status    string `gorm:"size:20;default:'pending';index"`
	DeviceID  string `gorm:"size:255"`
}

func (Alert) TableName() string { return "alerts" }

// Subscription represents a service subscription.
type Subscription struct {
	gorm.Model
	UserID    string  `gorm:"size:255;index"`
	Tier      string  `gorm:"size:50"`
	Status    string  `gorm:"size:20;default:'active'"`
	StartDate *time.Time
	EndDate   *time.Time
}

func (Subscription) TableName() string { return "subscriptions" }

// FirmwareRelease represents a firmware update release.
type FirmwareRelease struct {
	gorm.Model
	DeviceType  string `gorm:"size:50;index"`
	Tier        string `gorm:"size:50;index"`
	Version     string `gorm:"size:50;not null"`
	URL         string `gorm:"size:500"`
	Changelog   string
	ReleasedAt  time.Time
}

func (FirmwareRelease) TableName() string { return "firmware_releases" }

// Person represents a unified identity record.
type Person struct {
	gorm.Model
	Name    string  `gorm:"size:100;not null"`
	IDCard  string  `gorm:"uniqueIndex;size:50"`
	Gender  string  `gorm:"size:10"`
	BirthDate *time.Time
	Phone   string  `gorm:"size:20"`
}

func (Person) TableName() string { return "persons" }

// HospitalAdmission represents a hospital admission record.
type HospitalAdmission struct {
	gorm.Model
	PersonID     string  `gorm:"size:255;index"`
	HospitalID   string  `gorm:"size:255"`
	AdmissionNo  string  `gorm:"size:100"`
	Department   string  `gorm:"size:100"`
	BedNumber    string  `gorm:"size:50"`
	BloodType    string  `gorm:"size:10"`
	Status       string  `gorm:"size:20;default:'admitted'"`
	AdmittedAt   time.Time
	DischargedAt *time.Time
}

func (HospitalAdmission) TableName() string { return "hospital_admissions" }

// MedicalWristbandPatient represents a patient with medical wristband.
type MedicalWristbandPatient struct {
	gorm.Model
	PersonID   string `gorm:"size:255;index"`
	PatientID  string `gorm:"size:255"`
	Status     string `gorm:"size:20;default:'active'"`
}

func (MedicalWristbandPatient) TableName() string { return "medical_wristband_patients" }

// RegulatoryFenceConfig represents an electronic fence configuration.
type RegulatoryFenceConfig struct {
	gorm.Model
	HospitalID   string  `gorm:"size:255;index"`
	CenterLat    float64
	CenterLng    float64
	RadiusMeters int
	Enabled      bool `gorm:"default:true"`
}

func (RegulatoryFenceConfig) TableName() string { return "regulatory_fence_config" }

// AlertRule represents an alert generation rule.
type AlertRule struct {
	gorm.Model
	Name       string `gorm:"size:100;not null"`
	RuleType   string `gorm:"size:50"`
	Conditions string `gorm:"type:json"`
	Actions    string `gorm:"type:json"`
	Enabled    bool   `gorm:"default:true"`
}

func (AlertRule) TableName() string { return "alert_rules" }
