package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel is the base model with string ID and timestamps.
type BaseModel struct {
	ID        string         `gorm:"primaryKey;size:255" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// User represents an admin or family user.
type User struct {
	BaseModel
	Name         string `gorm:"size:100;not null;index"`
	Email        string `gorm:"uniqueIndex;size:200"`
	Phone        string `gorm:"uniqueIndex;size:20"`
	Role         string `gorm:"size:50;default:'family'"`
	PasswordHash string `gorm:"size:255;not null"`
	OpenID       string `gorm:"size:255;index"`
}

func (User) TableName() string { return "users" }

// ElderlyProfile represents an elderly person's profile.
type ElderlyProfile struct {
	BaseModel
	UserID      string   `gorm:"size:255;index"`
	Name        string   `gorm:"size:100;not null"`
	BirthDate   *time.Time
	AvatarURL   *string
	HealthTiers string   `gorm:"type:json"`
}

func (ElderlyProfile) TableName() string { return "elderly_profiles" }

// Device represents a wearable device.
type Device struct {
	BaseModel
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
	BaseModel
	ElderlyID  string  `gorm:"size:255;index"`
	HR         *int
	SpO2       *int
	Steps      *int64
	SleepHours *float64
	Timestamp  time.Time `gorm:"index"`
}

func (HealthRecord) TableName() string { return "health_records" }

// MedicationRule represents a medication schedule.
type MedicationRule struct {
	BaseModel
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
	BaseModel
	ElderlyID string  `gorm:"size:255;index"`
	Lat       float64
	Lng       float64
	Accuracy  *float64
	Timestamp time.Time `gorm:"index"`
}

func (LocationHistory) TableName() string { return "location_history" }

// Alert represents a safety alert.
type Alert struct {
	BaseModel
	ElderlyID string `gorm:"size:255;index"`
	AlertType string `gorm:"size:50;not null;index"`
	Severity  string `gorm:"size:20;not null;index"`
	Status    string `gorm:"size:20;default:'pending';index"`
	DeviceID  string `gorm:"size:255"`
}

func (Alert) TableName() string { return "alerts" }

// Subscription represents a service subscription.
type Subscription struct {
	BaseModel
	UserID               string     `gorm:"size:255;index"`
	PlanTier             string     `gorm:"size:50;not null"`
	Status               string     `gorm:"size:20;default:'active'"`
	BillingCycle         string     `gorm:"size:20;default:'monthly'"`
	StartsAt             *time.Time
	ExpiresAt            *time.Time
	CancellationReason   string
	TotalSpent           float64
}

func (Subscription) TableName() string { return "subscriptions" }

// FirmwareRelease represents a firmware update release.
type FirmwareRelease struct {
	BaseModel
	DeviceType string `gorm:"size:50;index"`
	Tier       string `gorm:"size:50;index"`
	Version    string `gorm:"size:50;not null"`
	URL        string `gorm:"size:500"`
	Changelog  string
	ReleasedAt time.Time
}

func (FirmwareRelease) TableName() string { return "firmware_releases" }

// Person represents a unified identity record.
type Person struct {
	BaseModel
	IDCard           string  `gorm:"uniqueIndex;size:50"`
	Name             string  `gorm:"size:100;not null"`
	Gender           int
	BirthDate        *time.Time
	Phone            *string
	EmergencyContact string
	Address          string
	AvatarURL        *string
	Status           string `gorm:"size:20;default:'active'"`
}

func (Person) TableName() string { return "persons" }

// HospitalAdmission represents a hospital admission record.
type HospitalAdmission struct {
	BaseModel
	PatientID           string  `gorm:"size:255;index"`
	AdmissionNo         string  `gorm:"size:100"`
	BedNo               string  `gorm:"size:50"`
	Department          string  `gorm:"size:100"`
	Diagnosis           string
	EmergencyContact    string
	Allergies           string
	AdmittedAt          time.Time
	ExpectedDischargeAt *time.Time
	DischargedAt        *time.Time
	DischargeType       string
	TransferredTo       string
	Notes               string
}

func (HospitalAdmission) TableName() string { return "hospital_admissions" }

// MedicalWristbandPatient represents a patient with medical wristband.
type MedicalWristbandPatient struct {
	BaseModel
	PersonID  string `gorm:"size:255;index"`
	PatientID string `gorm:"size:255"`
	Status    string `gorm:"size:20;default:'active'"`
}

func (MedicalWristbandPatient) TableName() string { return "medical_wristband_patients" }

// RegulatoryFenceConfig represents an electronic fence configuration.
type RegulatoryFenceConfig struct {
	BaseModel
	HospitalID   string  `gorm:"size:255;index"`
	CenterLat    float64
	CenterLng    float64
	RadiusMeters int
	Enabled      bool `gorm:"default:true"`
}

func (RegulatoryFenceConfig) TableName() string { return "regulatory_fence_config" }

// AlertRule represents an alert generation rule.
type AlertRule struct {
	BaseModel
	Name       string `gorm:"size:100;not null"`
	RuleType   string `gorm:"size:50"`
	Conditions string `gorm:"type:json"`
	Actions    string `gorm:"type:json"`
	Enabled    bool   `gorm:"default:true"`
}

func (AlertRule) TableName() string { return "alert_rules" }

// APIKey represents a B2B API key.
type APIKey struct {
	BaseModel
	InstitutionID string `gorm:"size:255"`
	Name          string `gorm:"size:100;not null"`
	KeyHash       string `gorm:"size:255;not null"`
	KeyPrefix     string `gorm:"size:50"`
	ExpiresAt     *time.Time
	Active        bool `gorm:"default:true"`
}

func (APIKey) TableName() string { return "b2b_api_keys" }

// SystemSetting represents a system configuration key-value.
type SystemSetting struct {
	Key          string `gorm:"primaryKey;size:100"`
	SettingValue string `gorm:"type:text"`
}

func (SystemSetting) TableName() string { return "system_settings" }

// OTAJob represents an OTA push job.
type OTAJob struct {
	BaseModel
	FirmwareID    string `gorm:"size:255;index"`
	TargetDevices string `gorm:"type:text"`
	Progress      string `gorm:"type:text"`
}

func (OTAJob) TableName() string { return "ota_jobs" }

// WardRoundEntry represents a nurse ward round record.
type WardRoundEntry struct {
	BaseModel
	PatientID     string  `gorm:"size:255;index"`
	NurseID       string  `gorm:"size:255"`
	BloodPressure string
	HeartRate     *int
	SpO2          *int
	Temperature   *float64
	Weight        *float64
	Notes         string
	Observations  string
	CompletedAt   time.Time
}

func (WardRoundEntry) TableName() string { return "ward_rounds" }

// MedicalWristbandDevice represents a wristband device.
type MedicalWristbandDevice struct {
	BaseModel
	DeviceID          string     `gorm:"uniqueIndex;size:100"`
	FirmwareVersion   string     `gorm:"size:50;default:''"`
	Status            string     `gorm:"size:20;default:'idle'"`
	BoundPatientID    *string
}

func (MedicalWristbandDevice) TableName() string { return "medical_wristband_devices" }

// MedicalBinding represents the association between a patient and a wristband.
type MedicalBinding struct {
	BaseModel
	PatientID string     `gorm:"size:255;index"`
	DeviceID  string     `gorm:"size:255;index"`
	BoundAt   time.Time
	UnboundAt *time.Time
}

func (MedicalBinding) TableName() string { return "medical_bindings" }

// AuditLog represents an audit trail entry.
type AuditLog struct {
	BaseModel
	UserID       string `gorm:"size:255;index"`
	Action       string `gorm:"size:100"`
	Resource     string `gorm:"size:100"`
	ResourceID   string `gorm:"size:255"`
	Details      string `gorm:"type:text"`
	IPAddress    string `gorm:"size:50"`
	UserAgent    string `gorm:"size:255"`
}

func (AuditLog) TableName() string { return "audit_log" }

// MedicalPatient represents a hospital patient for wristband tracking.
type MedicalPatient struct {
	BaseModel
	AdmissionNo       string `gorm:"uniqueIndex;size:100"`
	Name              string `gorm:"size:100;not null"`
	Gender            string `gorm:"size:10"`
	Age               *int
	Department        string `gorm:"size:100"`
	BedNumber         string `gorm:"size:50"`
	BloodType         string `gorm:"size:10"`
	Allergies         string
	SpecialConditions string
	TagIDs            string `gorm:"type:text"`
	Status            string `gorm:"size:20;default:'admitted'"`
}

func (MedicalPatient) TableName() string { return "medical_wristband_patients" }

// MedicalDailyEntry represents a daily nursing entry.
type MedicalDailyEntry struct {
	BaseModel
	PatientID string `gorm:"size:255;index"`
	EntryDate string `gorm:"size:20"`
	EntryType string `gorm:"size:50"`
	Content   string
	NurseID   string `gorm:"size:255"`
}

func (MedicalDailyEntry) TableName() string { return "medical_daily_entries" }

// PersonProfile represents a business-chain-specific profile extension.
type PersonProfile struct {
	PersonID            string  `gorm:"primaryKey;size:255"`
	BusinessChain       string  `gorm:"size:50"`
	SubscriptionTier    string  `gorm:"size:50"`
	SubscriptionStatus  string  `gorm:"size:50"`
	SubscriptionStart   *time.Time
	SubscriptionEnd     *time.Time
	HealthRiskLevel     string  `gorm:"size:50"`
	AdmissionNo         string  `gorm:"size:100"`
	Department          string  `gorm:"size:100"`
	BedNumber           string  `gorm:"size:50"`
	BloodType           string  `gorm:"size:10"`
	AttendingDoctor     string  `gorm:"size:100"`
	Diagnosis           string
	AdmissionDate       *time.Time
	ExpectedDischarge   *time.Time
	DischargeDate       *time.Time
	DischargeType       string  `gorm:"size:50"`
	HospitalID          string  `gorm:"size:255"`
	HospitalIDCommunity string  `gorm:"size:255"`
	MinzhengCertified   string  `gorm:"size:255"`
	SubsidyType         string  `gorm:"size:100"`
	CertificationDate   *time.Time
	CertificationDoc    string
	NextReviewDate      *time.Time
	LinkedPersonID      string  `gorm:"size:255"`
}

func (PersonProfile) TableName() string { return "person_profiles" }

// PersonWelfareTag represents a welfare tag assignment.
type PersonWelfareTag struct {
	PersonID string `gorm:"size:255"`
	TagCode  string `gorm:"size:100"`
	ValidFrom time.Time
	ValidTo   time.Time
}

func (PersonWelfareTag) TableName() string { return "person_welfare_tags" }

// MedicalExpense represents a hospital expense.
type MedicalExpense struct {
	BaseModel
	PatientID string  `gorm:"size:255;index"`
	ItemName  string  `gorm:"size:200"`
	Category  string  `gorm:"size:100"`
	Amount    float64
	Quantity  int
	UnitPrice float64
	Notes     string
}

func (MedicalExpense) TableName() string { return "medical_expenses" }

// MedicalMedication represents a medication record.
type MedicalMedication struct {
	BaseModel
	PatientID string `gorm:"size:255;index"`
	Name      string `gorm:"size:200"`
	Dosage    string
	Frequency string
	Duration  string
	Route     string
	Notes     string
}

func (MedicalMedication) TableName() string { return "medical_medications" }

// MedicalTestResult represents a lab test result.
type MedicalTestResult struct {
	BaseModel
	PatientID      string  `gorm:"size:255;index"`
	TestName       string  `gorm:"size:200"`
	Result         string
	ReferenceRange string
	Unit           string
	CollectedAt    *time.Time
	ReportedAt     *time.Time
	Notes          string
}

func (MedicalTestResult) TableName() string { return "medical_test_results" }

// MedicalVerification represents a verification record.
type MedicalVerification struct {
	BaseModel
	DeviceID         string  `gorm:"size:255"`
	PatientID        *string `gorm:"size:255"`
	VerificationType string  `gorm:"size:100"`
	Result           string
	Matched          bool
	VerifiedBy       string
	VerifiedAt       *time.Time
	Notes            string
}

func (MedicalVerification) TableName() string { return "medical_verifications" }

// MedicalAlertTagConfig represents an alert tag configuration.
type MedicalAlertTagConfig struct {
	BaseModel
	TagName string `gorm:"size:100"`
	TagColor string `gorm:"size:20"`
	TagIcon string `gorm:"size:50"`
	Enabled bool   `gorm:"default:true"`
}

func (MedicalAlertTagConfig) TableName() string { return "medical_alert_tag_configs" }

// MedicationRuleV2 represents a medication rule (v2).
type MedicationRuleV2 struct {
	BaseModel
	PersonID       string `gorm:"size:255;index"`
	BusinessChain  string `gorm:"size:50"`
	SourceType     string `gorm:"size:50"`
	SourceID       string
	DrugName       string `gorm:"size:200"`
	GenericName    string `gorm:"size:200"`
	DrugCategory   string `gorm:"size:100"`
	Dosage         string
	Frequency      string
	Route          string
	ScheduleTime1  string
	ScheduleTime2  string
	ScheduleTime3  string
	DaysOfWeek     string
	Duration       string
	PreMeal        bool
	PostMeal       bool
	SpecialInstructions string
	PrescribedBy   string
	PrescribedAt   string
	Active         bool `gorm:"default:true"`
}

func (MedicationRuleV2) TableName() string { return "medication_rules_v2" }

// MedicationExecution represents a medication execution record.
type MedicationExecution struct {
	BaseModel
	PersonID          string `gorm:"size:255;index"`
	BusinessChain     string `gorm:"size:50"`
	RuleID            string `gorm:"size:255"`
	ScheduledTime     string
	ActualTime        string
	Status            string `gorm:"size:50"`
	TakenBy           string
	DeviceID          string
	VerificationMethod string
	Notes             string
}

func (MedicationExecution) TableName() string { return "medication_executions" }

// UserRoleBinding represents a role binding.
type UserRoleBinding struct {
	BaseModel
	UserID          string  `gorm:"size:255;index"`
	BusinessChain   string  `gorm:"size:50"`
	Role            string  `gorm:"size:50"`
	InstitutionID   string  `gorm:"size:255"`
	GrantedBy       string
	ExpiresAt       *time.Time
	Active          bool    `gorm:"default:true"`
}

func (UserRoleBinding) TableName() string { return "user_role_bindings" }

// AlertRuleGorm represents an alert rule.
type AlertRuleGorm struct {
	BaseModel
	Name                 string `gorm:"size:100"`
	BusinessChain        string `gorm:"size:50"`
	AlertType            string `gorm:"size:50"`
	Severity             string `gorm:"size:20"`
	ConditionField       string
	ConditionOperator    string
	ConditionThreshold   *int
	ConditionDurationMin *int
	NotifyRoles          string
	NotifyChannels       string
	EscalationTimeoutMin int
	Enabled              bool `gorm:"default:true"`
}

func (AlertRuleGorm) TableName() string { return "alert_rules" }

// HealthRecordV2 represents a health record (v2).
type HealthRecordV2 struct {
	BaseModel
	PersonID        string  `gorm:"size:255;index"`
	BusinessChain   string  `gorm:"size:50"`
	RecordType      string  `gorm:"size:50"`
	Source          string  `gorm:"size:50"`
	DeviceID        string
	RecordedAt      time.Time
	HeartRate       *int
	BloodPressureSys *int
	BloodPressureDia *int
	SpO2            *int
	Temperature     *float64
	RespiratoryRate *int
	PulseRate       *int
	GlucoseFasting  *float64
	UricAcid        *float64
	Steps           *int64
	SleepHours      *float64
	Content         string
}

func (HealthRecordV2) TableName() string { return "health_records_v2" }

// PersonHealthSummary represents a health summary.
type PersonHealthSummary struct {
	PersonID            string  `gorm:"primaryKey;size:255"`
	BusinessChain       string  `gorm:"size:50"`
	LatestHR            *int
	LatestSpO2          *int
	LatestBPSys         *int
	LatestBPDia         *int
	LatestGlucoseFasting *float64
	LatestUricAcid      *float64
	LatestSteps         *int64
	LatestSleepHours    *float64
	RiskScore           *float64
	TrendDirection      string
	Recommendation      string
}

func (PersonHealthSummary) TableName() string { return "person_health_summaries" }

// HealthGuidanceRule represents a guidance rule.
type HealthGuidanceRule struct {
	BaseModel
	Name           string  `gorm:"size:200"`
	BusinessChain  string  `gorm:"size:50"`
	TriggerCondition string
	ConditionField string
	ConditionOp    string
	ConditionThresh *float64
	GuidanceType   string
	Title          string
	Content        string
	Priority       int
	Enabled        bool    `gorm:"default:true"`
}

func (HealthGuidanceRule) TableName() string { return "health_guidance_rules" }

// HealthGuidanceDelivery represents a guidance delivery record.
type HealthGuidanceDelivery struct {
	BaseModel
	PersonID      string `gorm:"size:255;index"`
	BusinessChain string `gorm:"size:50"`
	RuleID        string `gorm:"size:255"`
	GuidanceType  string
	Title         string
	Content       string
	Channel       string
	DeliveredAt   time.Time
	ReadStatus    int
	Feedback      string
}

func (HealthGuidanceDelivery) TableName() string { return "health_guidance_deliveries" }

// HealthReportTemplate represents a report template.
type HealthReportTemplate struct {
	BaseModel
	Name          string `gorm:"size:200"`
	BusinessChain string `gorm:"size:50"`
	Frequency     string
	TemplateType  string
}

func (HealthReportTemplate) TableName() string { return "health_report_templates" }

// HealthReport represents a health report.
type HealthReport struct {
	BaseModel
	PersonID     string `gorm:"size:255;index"`
	BusinessChain string `gorm:"size:50"`
	TemplateID   string
	ReportPeriod string
	Content      string
}

func (HealthReport) TableName() string { return "health_reports" }

// ComplianceRule represents a compliance rule.
type ComplianceRule struct {
	BaseModel
	RuleCode    string `gorm:"uniqueIndex;size:100"`
	Name        string `gorm:"size:200"`
	Description string
	BusinessChain string `gorm:"size:50"`
	Condition   string
	Action      string
	Enabled     bool `gorm:"default:true"`
}

func (ComplianceRule) TableName() string { return "compliance_rules" }

// ComplianceCheck represents a compliance check result.
type ComplianceCheck struct {
	BaseModel
	RuleID    string `gorm:"size:255"`
	PersonID  string `gorm:"size:255;index"`
	Violated  bool
	Result    string
	Notes     string
	ReviewerID string
	ReviewedAt *time.Time
}

func (ComplianceCheck) TableName() string { return "compliance_checks" }

// DeviceBinding represents a device binding.
type DeviceBinding struct {
	BaseModel
	DeviceID      string `gorm:"size:255;index"`
	PersonID      string `gorm:"size:255;index"`
	BusinessChain string `gorm:"size:50"`
}

func (DeviceBinding) TableName() string { return "device_bindings" }

// NotificationTemplate represents a notification template.
type NotificationTemplate struct {
	BaseModel
	Name          string `gorm:"size:200"`
	BusinessChain string `gorm:"size:50"`
	Channel       string `gorm:"size:50"`
	Subject       string
	Content       string
	Enabled       bool `gorm:"default:true"`
}

func (NotificationTemplate) TableName() string { return "notification_templates" }

// NotificationLog represents a notification log.
type NotificationLog struct {
	BaseModel
	PersonID     string `gorm:"size:255;index"`
	BusinessChain string `gorm:"size:50"`
	TemplateID   string
	Channel      string
	Status       string `gorm:"size:50"`
	SentAt       *time.Time
	ReadAt       *time.Time
	Content      string
}

func (NotificationLog) TableName() string { return "notification_logs" }
