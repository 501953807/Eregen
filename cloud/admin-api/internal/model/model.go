package model

import "time"

// Role represents an admin user role.
type Role string

const (
	RoleAdmin      Role = "admin"
	RoleOperator   Role = "operator"
	RoleSuperAdmin Role = "super_admin"
	RoleNurse      Role = "nurse"
	RoleRegulator  Role = "regulator"
)

// DashboardStats aggregates the key metrics shown on the admin dashboard.
type DashboardStats struct {
	OnlineDevices       int          `json:"online_devices"`
	TotalDevices        int          `json:"total_devices"`
	ActiveAlerts        int          `json:"active_alerts"`
	P0Alerts            int          `json:"p0"`
	P1Alerts            int          `json:"p1"`
	P2Alerts            int          `json:"p2"`
	TotalUsers          int          `json:"total_users"`
	ActiveSubscriptions int          `json:"active_subscriptions"`
	AlertTrend          []TrendPoint `json:"alert_trend,omitempty"`
}

// TrendPoint is a single data point in a time-series chart.
type TrendPoint struct {
	Date  string `json:"date"`
	Value int    `json:"value"`
}

// DeviceSummary is a lightweight row returned by the device list endpoint.
type DeviceSummary struct {
	ID          string    `json:"id"`
	DeviceID    string    `json:"device_id"`
	Type        string    `json:"type"`
	Tier        string    `json:"tier"`
	Status      string    `json:"status"`
	LastSeen    time.Time `json:"last_seen"`
	OwnerName   string    `json:"owner_name"`
	FirmwareVer string    `json:"firmware_version"`
}

// UserSummary is a lightweight row returned by the user list endpoint.
type UserSummary struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	Devices   int       `json:"devices"`
}

// UserLogin is returned by GetUserByCredential.
type UserLogin struct {
	ID   string
	Name string
	Role string
}

// AlertSummary is a lightweight row returned by the alert list endpoint.
type AlertSummary struct {
	ID         string    `json:"id"`
	ElderlyID  string    `json:"elderly_id"`
	AlertType  string    `json:"alert_type"`
	Severity   string    `json:"severity"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	DeviceID   string    `json:"device_id"`
}

// SubscriptionStat holds a plan-tier breakdown.
type SubscriptionStat struct {
	Tier  string  `json:"tier"`
	Count int     `json:"count"`
	Pct   float64 `json:"pct"`
}

// ElderlyProfile is a lightweight row returned by the elderly list endpoint.
type ElderlyProfile struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Name        string     `json:"name"`
	BirthDate   *time.Time `json:"birth_date,omitempty"`
	AvatarURL   *string    `json:"avatar_url,omitempty"`
	HealthTiers []string   `json:"health_tiers"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// AlertTrendPoint is a single data point in the alert trend chart.
type AlertTrendPoint struct {
	Date           string `json:"date"`
	BraceletCount  int    `json:"bracelet_count"`
	PillboxCount   int    `json:"pillbox_count"`
}

// AlertDistributionItem holds an alert type and its count.
type AlertDistributionItem struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
	Color string `json:"color"`
}

// UserGrowthPoint is a monthly new-user count.
type UserGrowthPoint struct {
	Month    string `json:"month"`
	NewUsers int    `json:"new_users"`
}

// DeviceDetail includes settings JSONB for the device detail view.
type DeviceDetail struct {
	ID           string         `json:"id"`
	DeviceID     string         `json:"device_id"`
	Type         string         `json:"type"`
	Tier         string         `json:"tier"`
	Status       string         `json:"status"`
	LastSeen     time.Time      `json:"last_seen"`
	OwnerName    string         `json:"owner_name"`
	FirmwareVer  string         `json:"firmware_version"`
	SettingsJSON map[string]any  `json:"settings,omitempty"`
	ElderlyName  string         `json:"elderly_name,omitempty"`
}

// FirmwareVersion represents a firmware release for OTA tracking.
type FirmwareVersion struct {
	ID            string    `json:"id"`
	DeviceType    string    `json:"device_type"`
	Tier          string    `json:"tier"`
	Version       string    `json:"version"`
	DownloadURL   string    `json:"download_url"`
	Sha256Hash    string    `json:"sha256_hash"`
	Changelog     string    `json:"changelog"`
	MinAppVersion string    `json:"min_app_version,omitempty"`
	ForceUpdate   bool      `json:"force_update"`
	IsActive      bool      `json:"is_active"`
	ReleaseDate   time.Time `json:"release_date"`
}

// APIKeySummary is a lightweight row for B2B API key listing.
type APIKeySummary struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	KeyPrefix string     `json:"key_prefix"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Active    bool       `json:"active"`
	CreatedAt time.Time  `json:"created_at"`
}

// ========== Medical Wristband Models ==========

// MedicalPatient represents a hospital patient registered via wristband system.
type MedicalPatient struct {
	ID                 string    `json:"id"`
	AdmissionNo        string    `json:"admission_no"`
	Name               string    `json:"name"`
	Gender             string    `json:"gender,omitempty"`
	Age                int       `json:"age,omitempty"`
	Department         string    `json:"department,omitempty"`
	BedNumber          string    `json:"bed_number,omitempty"`
	BloodType          string    `json:"blood_type,omitempty"`
	Allergies          string    `json:"allergies,omitempty"`
	SpecialConditions  string    `json:"special_conditions,omitempty"`
	TagIDs             []string  `json:"tag_ids,omitempty"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// MedicalWristbandDevice represents an electronic wristband device.
type MedicalWristbandDevice struct {
	ID             string    `json:"id"`
	DeviceID       string    `json:"device_id"`
	FirmwareVersion string   `json:"firmware_version"`
	Status         string    `json:"status"` // idle, bound, cleared
	BoundPatientID string    `json:"bound_patient_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// MedicalBinding represents the association between a patient and a wristband device.
type MedicalBinding struct {
	ID        string    `json:"id"`
	PatientID string    `json:"patient_id"`
	DeviceID  string    `json:"device_id"`
	BoundAt   time.Time `json:"bound_at"`
	UnboundAt *time.Time `json:"unbound_at,omitempty"`
}

// MedicalExpense represents a hospital expense item.
type MedicalExpense struct {
	ID         string    `json:"id"`
	PatientID  string    `json:"patient_id"`
	ItemName   string    `json:"item_name"`
	Category   string    `json:"category,omitempty"`
	Amount     float64   `json:"amount"`
	Quantity   int       `json:"quantity"`
	UnitPrice  float64   `json:"unit_price"`
	Notes      string    `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// MedicalMedication represents a medication order.
type MedicalMedication struct {
	ID         string    `json:"id"`
	PatientID  string    `json:"patient_id"`
	Name       string    `json:"name"`
	Dosage     string    `json:"dosage,omitempty"`
	Frequency  string    `json:"frequency,omitempty"`
	Duration   string    `json:"duration,omitempty"`
	Route      string    `json:"route,omitempty"`
	Notes      string    `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// MedicalTestResult represents a lab/test result.
type MedicalTestResult struct {
	ID             string    `json:"id"`
	PatientID      string    `json:"patient_id"`
	TestName       string    `json:"test_name"`
	Result         string    `json:"result,omitempty"`
	ReferenceRange string    `json:"reference_range,omitempty"`
	Unit           string    `json:"unit,omitempty"`
	CollectedAt    *time.Time `json:"collected_at,omitempty"`
	ReportedAt     *time.Time `json:"reported_at,omitempty"`
	Notes          string    `json:"notes,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// MedicalDailyEntry represents a daily nursing/doctor entry.
type MedicalDailyEntry struct {
	ID         string    `json:"id"`
	PatientID  string    `json:"patient_id"`
	EntryDate  string    `json:"entry_date"`
	EntryType  string    `json:"entry_type"`
	Content    string    `json:"content"`
	NurseID    string    `json:"nurse_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// MedicalVerification represents a nurse NFC verification record.
type MedicalVerification struct {
	ID               string    `json:"id"`
	DeviceID         string    `json:"device_id"`
	PatientID        *string   `json:"patient_id,omitempty"`
	VerificationType string    `json:"verification_type"`
	Result           string    `json:"result,omitempty"`
	Matched          bool      `json:"matched"`
	VerifiedBy       string    `json:"verified_by,omitempty"`
	VerifiedAt       *time.Time `json:"verified_at,omitempty"`
	Notes            string    `json:"notes,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// MedicalAlertTagConfig represents an alert tag configuration.
type MedicalAlertTagConfig struct {
	ID        string    `json:"id"`
	TagName   string    `json:"tag_name"`
	TagColor  string    `json:"tag_color"`
	TagIcon   string    `json:"tag_icon"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ========== Regulatory Closure Models ==========

// RegulatoryFenceConfig defines a geofence for a hospital/department.
type RegulatoryFenceConfig struct {
	ID           string    `json:"id"`
	HospitalID   string    `json:"hospital_id"`
	HospitalName string    `json:"hospital_name"`
	CenterLat    float64   `json:"center_lat"`
	CenterLng    float64   `json:"center_lng"`
	RadiusMeters int       `json:"radius_meters"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RegulatoryLocationLog records a patient's location for fence checking.
type RegulatoryLocationLog struct {
	ID            string    `json:"id"`
	PatientID     string    `json:"patient_id"`
	DeviceID      string    `json:"device_id"`
	Lat           float64   `json:"lat"`
	Lng           float64   `json:"lng"`
	Accuracy      *float64  `json:"accuracy,omitempty"`
	LocationSource string   `json:"location_source,omitempty"` // "gps" | "base_station"
	InsideFence   bool      `json:"inside_fence"`
	RecordedAt    time.Time `json:"recorded_at"`
}

// RegulatoryAlert represents a rule-engine triggered alert.
type RegulatoryAlert struct {
	ID            string     `json:"id"`
	RuleCode      string     `json:"rule_code"`
	PatientID     *string    `json:"patient_id,omitempty"`
	HospitalID    string     `json:"hospital_id"`
	Department    string     `json:"department"`
	Severity      string     `json:"severity"` // low, medium, high
	AlertType     string     `json:"alert_type"`
	Detail        string     `json:"detail"`
	Status        string     `json:"status"` // pending, acknowledged, resolved, false_positive
	TriggeredAt   time.Time  `json:"triggered_at"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
	AcknowledgedBy string    `json:"acknowledged_by,omitempty"`
	ResolvedAt    *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy    string     `json:"resolved_by,omitempty"`
	Notes         string     `json:"notes,omitempty"`
}

// RegulatoryDashboardOverview holds summary stats for the regulatory dashboard.
type RegulatoryDashboardOverview struct {
	TotalAdmitted       int                    `json:"total_admitted"`
	TodayAdmit          int                    `json:"today_admit"`
	TodayDischarge      int                    `json:"today_discharge"`
	ByDepartment        []RegulatoryDeptStat   `json:"by_department"`
	FenceViolationsToday int                   `json:"fence_violations_today"`
	NoVerify24h         int                    `json:"no_verify_24h"`
}

// RegulatoryDeptStat is a department row in the dashboard overview.
type RegulatoryDeptStat struct {
	Name       string `json:"name"`
	Count      int    `json:"count"`
	AlertCount int    `json:"alert_count"`
}

// RegulatoryPatientRow is a patient row in the regulatory dashboard list.
type RegulatoryPatientRow struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	AdmissionNo        string    `json:"admission_no"`
	Department         string    `json:"department"`
	BedNumber          string    `json:"bed_number"`
	BoundAt            time.Time `json:"bound_at"`
	LastVerify         *time.Time `json:"last_verify,omitempty"`
	VerifyGapHours     int       `json:"verify_gap_hours"`
	FenceStatus        string    `json:"fence_status"`
	FenceExitDurationSec int     `json:"fence_exit_duration_sec"`
	AlertTags          []string  `json:"alert_tags"`
	AlertsTriggered    []string  `json:"alerts_triggered"`
}

// RegulatoryRuleConfig holds a configurable detection rule.
type RegulatoryRuleConfig struct {
	Code   string                 `json:"code"`
	Name   string                 `json:"name"`
	Enabled bool                  `json:"enabled"`
	Config map[string]interface{} `json:"config"`
}

// ComplianceReport holds periodic compliance statistics.
type ComplianceReport struct {
	Summary          ComplianceSummary          `json:"summary"`
	DepartmentBreakdown []ComplianceDeptBreakdown `json:"department_breakdown"`
}

// ComplianceSummary is the top-level summary of a compliance report.
type ComplianceSummary struct {
	TotalPatientsPeriod  int     `json:"total_patients_period"`
	AvgStayDays          float64 `json:"avg_stay_days"`
	FenceViolations      int     `json:"fence_violations"`
	NoVerifyAlerts       int     `json:"no_verify_alerts"`
	ExpenseAnomalies     int     `json:"expense_anomalies"`
	MedVerifyMismatch    int     `json:"med_verify_mismatch"`
	ComplianceRate       float64 `json:"compliance_rate"`
}

// ComplianceDeptBreakdown is a department row in the compliance report.
type ComplianceDeptBreakdown struct {
	Name          string  `json:"name"`
	TotalPatients int     `json:"total_patients"`
	Alerts        int     `json:"alerts"`
	ComplianceRate float64 `json:"compliance_rate"`
}

// DepartmentBinding links a user to a hospital department.
type DepartmentBinding struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Department string    `json:"department"`
	BoundAt    time.Time `json:"bound_at"`
}

// ========== Community Elderly Models ==========

// CommunityElder represents a community elderly person profile.
type CommunityElder struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	IDCard          string     `json:"id_card"`
	Gender          int        `json:"gender"`
	Age             int        `json:"age,omitempty"`
	Address         string     `json:"address,omitempty"`
	EmergencyContact string    `json:"emergency_contact,omitempty"`
	BankAccount     string     `json:"bank_account,omitempty"`
	HospitalID      string     `json:"hospital_id,omitempty"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeactivatedAt   *time.Time `json:"deactivated_at,omitempty"`
	DeactivatedReason string   `json:"deactivated_reason,omitempty"`
}

// CommunityWristbandDevice represents a community-mode wristband device.
type CommunityWristbandDevice struct {
	ID            string    `json:"id"`
	DeviceID      string    `json:"device_id"`
	FirmwareVersion string  `json:"firmware_version"`
	Mode          string    `json:"mode"` // hospital, community
	Status        string    `json:"status"`
	LastSeen      *time.Time `json:"last_seen,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CommunityElderBinding links an elder to a wristband device.
type CommunityElderBinding struct {
	ID        string    `json:"id"`
	ElderID   string    `json:"elder_id"`
	DeviceID  string    `json:"device_id"`
	BoundAt   time.Time `json:"bound_at"`
	UnboundAt *time.Time `json:"unbound_at,omitempty"`
}

// CommunityWelfareTagConfig is a configurable welfare tag definition.
type CommunityWelfareTagConfig struct {
	ID                string    `json:"id"`
	TagCode           string    `json:"tag_code"`
	TagName           string    `json:"tag_name"`
	Issuer            string    `json:"issuer"`
	RenewalPeriodDays int       `json:"renewal_period_days,omitempty"`
	BenefitAmount     float64   `json:"benefit_amount,omitempty"`
	Enabled           bool      `json:"enabled"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// CommunityElderWelfare is a welfare tag assignment to an elder.
type CommunityElderWelfare struct {
	ID                string     `json:"id"`
	ElderID           string     `json:"elder_id"`
	TagCode           string     `json:"tag_code"`
	ValidFrom         string     `json:"valid_from"`
	ValidTo           string     `json:"valid_to"`
	CertifiedBy       string     `json:"certified_by,omitempty"`
	CertificationDoc  string     `json:"certification_doc,omitempty"`
	EffectiveAt       time.Time  `json:"effective_at"`
	RevokedAt         *time.Time `json:"revoked_at,omitempty"`
}

// CommunitySigninRecord represents a sign-in activation event.
type CommunitySigninRecord struct {
	ID              string    `json:"id"`
	ElderID         string    `json:"elder_id"`
	DeviceID        string    `json:"device_id"`
	HospitalID      string    `json:"hospital_id"`
	PharmacistID    string    `json:"pharmacist_id,omitempty"`
	SigninTime      time.Time `json:"signin_time"`
	Period          string    `json:"period"` // YYYY-MM
	IDCard          string    `json:"id_card,omitempty"`
	ActivatedTags   string    `json:"activated_tags"` // JSON array
	IsMedicalSignin bool      `json:"is_medical_signin"`
	IsWelfareSignin bool      `json:"is_welfare_signin"`
	Notes           string    `json:"notes,omitempty"`
}

// CommunityPharmacyLog represents a drug dispensing record.
type CommunityPharmacyLog struct {
	ID             string    `json:"id"`
	ElderID        string    `json:"elder_id"`
	DeviceID       string    `json:"device_id,omitempty"`
	HospitalID     string    `json:"hospital_id"`
	PharmacistID   string    `json:"pharmacist_id,omitempty"`
	DispenseTime   time.Time `json:"dispense_time"`
	Period         string    `json:"period"`
	Items          string    `json:"items"` // JSON array of dispensed items
	TotalCost      float64   `json:"total_cost"`
	InsuranceCovered float64 `json:"insurance_covered"`
	SelfPay        float64   `json:"self_pay"`
	Notes          string    `json:"notes,omitempty"`
}

// CommunityMinzhengSync tracks a batch import task from government data.
type CommunityMinzhengSync struct {
	ID                string    `json:"id"`
	Source            string    `json:"source"`
	Filename          string    `json:"filename,omitempty"`
	ImportedCount     int       `json:"imported_count"`
	MatchedCount      int       `json:"matched_count"`
	PendingReviewCount int      `json:"pending_review_count"`
	ErrorCount        int       `json:"error_count"`
	Status            string    `json:"status"` // processing, completed, failed
	CreatedAt         time.Time `json:"created_at"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
}

// CommunityBatchPayment represents a batch subsidy disbursement record.
type CommunityBatchPayment struct {
	ID            string     `json:"id"`
	BatchID       string     `json:"batch_id"`
	Period        string     `json:"period"`
	PayType       string     `json:"pay_type"`
	ElderID       string     `json:"elder_id"`
	Amount        float64    `json:"amount"`
	BankAccount   string     `json:"bank_account,omitempty"`
	Status        string     `json:"status"` // pending, success, failed, retrying
	FailureReason string     `json:"failure_reason,omitempty"`
	ExecutedAt    *time.Time `json:"executed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// CommunityElderStats holds overview stats for the community dashboard.
type CommunityElderStats struct {
	TotalElders      int `json:"total_elders"`
	ActiveElders     int `json:"active_elders"`
	TodaySignins     int `json:"today_signins"`
	TodayDispenses   int `json:"today_dispenses"`
	PendingSignins   int `json:"pending_signins"`
	ActiveWelfareTags int `json:"active_welfare_tags"`
}

// MedicalVerificationStats holds today's verification statistics.
type MedicalVerificationStats struct {
	Total    int `json:"total"`
	Matched  int `json:"matched"`
	Unmatched int `json:"unmatched"`
}

// MedicalStatsOverview holds overall medical statistics for the dashboard.
type MedicalStatsOverview struct {
	ActivePatients   int `json:"active_patients"`
	TodayAdmitted    int `json:"today_admitted"`
	TodayDischarged  int `json:"today_discharged"`
	BoundDevices     int `json:"bound_devices"`
	TotalDevices     int `json:"total_devices"`
}

// ========== Elderly Detail Types ==========

// HealthStats aggregates recent health metrics for an elderly person.
type HealthStats struct {
	ElderlyID  string    `json:"elderly_id"`
	AvgHR      *float64  `json:"avg_hr,omitempty"`
	MaxHR      *int      `json:"max_hr,omitempty"`
	AvgSpO2    *float64  `json:"avg_spo2,omitempty"`
	TotalSteps *int64    `json:"total_steps,omitempty"`
	LastSeen   time.Time `json:"last_seen"`
}

// HealthRecordRow represents a single health record from the database.
type HealthRecordRow struct {
	ID         string    `json:"id"`
	ElderlyID  string    `json:"elderly_id"`
	Timestamp  time.Time `json:"timestamp"`
	HR         *int      `json:"hr,omitempty"`
	SpO2       *int      `json:"spo2,omitempty"`
	Steps      *int64    `json:"steps,omitempty"`
	SleepHours *float64  `json:"sleep_hours,omitempty"`
}

// MedicationRuleRow represents a medication rule.
type MedicationRuleRow struct {
	ID                   string    `json:"id"`
	PersonID             string    `json:"person_id"`
	BusinessChain        string    `json:"business_chain"`
	SourceType           string    `json:"source_type"`
	SourceID             string    `json:"source_id"`
	ElderlyID            string    `json:"elderly_id"`
	Dosage               string    `json:"dosage"`
	Frequency            string    `json:"frequency"`
	Route                string    `json:"route"`
	ScheduleTime         string    `json:"schedule_time"`          // legacy alias
	ScheduleTime1        string    `json:"schedule_time1"`
	ScheduleTime2        string    `json:"schedule_time2"`
	ScheduleTime3        string    `json:"schedule_time3"`
	DaysOfWeek           string    `json:"days_of_week"`
	Duration             string    `json:"duration"`
	PreMeal              int       `json:"pre_meal"`
	PostMeal             int       `json:"post_meal"`
	SpecialInstructions  string    `json:"special_instructions"`
	PrescribedBy         string    `json:"prescribed_by"`
	PrescribedAt         string    `json:"prescribed_at"`
	DrugName             string    `json:"drug_name"`
	GenericName          string    `json:"generic_name"`
	DrugCategory         string    `json:"drug_category"`
	DoseCount            int       `json:"dose_count"`
	PillType             string    `json:"pill_type"`
	Active               bool      `json:"active"`
	CreatedAt            string    `json:"created_at"`
	UpdatedAt            string    `json:"updated_at"`
}

// DeviceSummaryRow is a device linked to an elderly person.
type DeviceSummaryRow struct {
	ID          string    `json:"id"`
	DeviceID    string    `json:"device_id"`
	Type        string    `json:"type"`
	Tier        string    `json:"tier"`
	Status      string    `json:"status"`
	FirmwareVer string    `json:"firmware_version"`
	LastSeen    time.Time `json:"last_seen"`
}

// LocationPoint represents a location record.
type LocationPoint struct {
	ID        string    `json:"id"`
	ElderlyID string    `json:"elderly_id"`
	Lat       float64   `json:"lat"`
	Lon       float64   `json:"lon"`
	Accuracy  *float64  `json:"accuracy,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// AlertSummaryRow represents an alert.
type AlertSummaryRow struct {
	ID        string    `json:"id"`
	ElderlyID string    `json:"elderly_id"`
	AlertType string    `json:"alert_type"`
	Severity  string    `json:"severity"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// MedicalPatientHistory represents a patient with their daily entries.
type MedicalPatientHistory struct {
	Patient     *MedicalPatient     `json:"patient"`
	DailyEntries []MedicalDailyEntry `json:"daily_entries"`
}

// RuleAlertCount holds alert count grouped by rule code.
type RuleAlertCount struct {
	RuleCode string `json:"rule_code"`
	Count    int    `json:"count"`
}

// DeptAlertCount holds alert count grouped by department.
type DeptAlertCount struct {
	Department string `json:"department"`
	Count      int    `json:"count"`
}

// ========== Clinical Workflow Models ==========

// HospitalAdmission represents a patient's hospital stay record.
type HospitalAdmission struct {
	ID                  string     `json:"id"`
	PatientID           string     `json:"patient_id"`
	AdmissionNo         string     `json:"admission_no"`
	BedNo               string     `json:"bed_no"`
	Department          string     `json:"department"`
	Diagnosis           string     `json:"diagnosis,omitempty"`
	EmergencyContact    string     `json:"emergency_contact,omitempty"`
	Allergies           string     `json:"allergies,omitempty"`
	AdmittedAt          time.Time  `json:"admitted_at"`
	ExpectedDischargeAt *time.Time `json:"expected_discharge_at,omitempty"`
	DischargedAt        *time.Time `json:"discharged_at,omitempty"`
	DischargeType       string     `json:"discharge_type,omitempty"` // "discharged", "transferred", "deceased"
	TransferredTo       string     `json:"transferred_to,omitempty"`
	Notes               string     `json:"notes,omitempty"`
}

// WardRoundEntry represents a nursing round visit with vitals.
type WardRoundEntry struct {
	ID             string    `json:"id"`
	PatientID      string    `json:"patient_id"`
	NurseID        string    `json:"nurse_id"`
	BloodPressure  string    `json:"blood_pressure,omitempty"` // e.g. "120/80"
	HeartRate      int       `json:"heart_rate,omitempty"`
	SpO2           int       `json:"spo2,omitempty"`
	Temperature    float64   `json:"temperature,omitempty"`
	Weight         float64   `json:"weight,omitempty"`
	Notes          string    `json:"notes,omitempty"`
	Observations   string    `json:"observations,omitempty"` // JSON array of checkboxes
	CompletedAt    time.Time `json:"completed_at"`
}

// RegulatoryRuleResult holds the outcome of a single rule evaluation.
type RegulatoryRuleResult struct {
	RuleCode    string    `json:"rule_code"` // "R01"..."R08"
	Severity    string    `json:"severity"`  // "P0", "P1", "P2"
	PatientID   string    `json:"patient_id,omitempty"`
	Message     string    `json:"message"`
	TriggeredAt time.Time `json:"triggered_at"`
	Resolved    bool      `json:"resolved"`
}

// RegulatoryAuditTrail is the full audit trail for a single patient.
type RegulatoryAuditTrail struct {
	Patient         *MedicalPatient       `json:"patient"`
	Binding         *MedicalBinding       `json:"binding"`
	Verifications   []MedicalVerification `json:"verifications"`
	Medications     []MedicalMedication   `json:"medications"`
	Expenses        []MedicalExpense      `json:"expenses"`
	DailyEntries    []MedicalDailyEntry   `json:"daily_entries"`
	FenceLogs       []RegulatoryLocationLog `json:"fence_logs"`
	AlertsGenerated []RegulatoryAlert     `json:"alerts_generated"`
}

// RegulatoryRuleConfigDB is the DB model for rule configuration.
type RegulatoryRuleConfigDB struct {
	RuleCode   string `json:"code"`
	RuleName   string `json:"name"`
	Enabled    bool   `json:"enabled"`
	ConfigJSON string `json:"config_json"`
	UpdatedAt  time.Time `json:"updated_at"`
}


// SubscriptionItem represents a subscription record for listing.
type SubscriptionItem struct {
	ID                   string    `json:"id"`
	UserID               string    `json:"user_id"`
	UserName             string    `json:"user_name,omitempty"`
	UserPhone            string    `json:"user_phone,omitempty"`
	PlanTier             string    `json:"plan_tier"`
	Status               string    `json:"status"`
	BillingCycle         string    `json:"billing_cycle"`
	StartDate            string    `json:"start_date"`
	EndDate              string    `json:"end_date"`
	CancellationReason   string    `json:"cancellation_reason,omitempty"`
	TotalSpent           float64   `json:"total_spent,omitempty"`
	Devices              []string  `json:"devices,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
}

// InstitutionSummary is a lightweight row for institution listing.
type InstitutionSummary struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Code         string    `json:"code"`
	ContactName  string    `json:"contact_name,omitempty"`
	ContactPhone string    `json:"contact_phone,omitempty"`
	AccessLevel  string    `json:"access_level"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	APIKeyCount  int       `json:"api_key_count"`
}

// AuditLog represents an audit log entry.
type AuditLog struct {
	ID         string                 `json:"id"`
	UserID     string                 `json:"user_id"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	ResourceID string                 `json:"resource_id,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	IP         string                 `json:"ip"`
	UserAgent  string                 `json:"user_agent"`
	CreatedAt  time.Time              `json:"created_at"`
}

// Person represents a unified identity across all business chains.
type Person struct {
	ID               string     `json:"id"`
	IDCard           string     `json:"id_card"`
	Name             string     `json:"name"`
	Gender           int        `json:"gender"`
	BirthDate        *time.Time `json:"birth_date,omitempty"`
	Phone            string     `json:"phone"`
	EmergencyContact string     `json:"emergency_contact"`
	Address          string     `json:"address"`
	AvatarURL        *string    `json:"avatar_url,omitempty"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// PersonProfile holds business-chain-specific extension fields for a person.
type PersonProfile struct {
	PersonID            string     `json:"person_id"`
	BusinessChain       string     `json:"business_chain"`
	SubscriptionTier    string     `json:"subscription_tier,omitempty"`
	SubscriptionStatus  string     `json:"subscription_status,omitempty"`
	SubscriptionStart   *time.Time `json:"subscription_start,omitempty"`
	SubscriptionEnd     *time.Time `json:"subscription_end,omitempty"`
	HealthRiskLevel     string     `json:"health_risk_level,omitempty"`
	AdmissionNo         string     `json:"admission_no,omitempty"`
	Department          string     `json:"department,omitempty"`
	BedNumber           string     `json:"bed_number,omitempty"`
	BloodType           string     `json:"blood_type,omitempty"`
	AttendingDoctor     string     `json:"attending_doctor,omitempty"`
	Diagnosis           string     `json:"diagnosis,omitempty"`
	AdmissionDate       *time.Time `json:"admission_date,omitempty"`
	ExpectedDischarge   *time.Time `json:"expected_discharge_date,omitempty"`
	DischargeDate       *time.Time `json:"discharge_date,omitempty"`
	DischargeType       string     `json:"discharge_type,omitempty"`
	HospitalID          string     `json:"hospital_id,omitempty"`
	HospitalIDCommunity string     `json:"hospital_id_community,omitempty"`
	MinzhengCertified   int        `json:"minzheng_certified"`
	SubsidyType         string     `json:"subsidy_type,omitempty"`
	CertificationDate   *time.Time `json:"certification_date,omitempty"`
	CertificationDoc    string     `json:"certification_doc,omitempty"`
	NextReviewDate      *time.Time `json:"next_review_date,omitempty"`
	LinkedPersonID      string     `json:"linked_person_id,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// PersonWelfareTag links a person to a welfare tag.
type PersonWelfareTag struct {
	PersonID string    `json:"person_id"`
	TagCode  string    `json:"tag_code"`
	ValidFrom time.Time `json:"valid_from"`
	ValidTo  time.Time `json:"valid_to"`
}

// PersonRoleBinding represents a user's role assignment to a business chain.
type PersonRoleBinding struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	BusinessChain string     `json:"business_chain"`
	Role          string     `json:"role"`
	InstitutionID string     `json:"institution_id,omitempty"`
	GrantedBy     string     `json:"granted_by,omitempty"`
	GrantedAt     time.Time  `json:"granted_at"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	Active        int        `json:"active"`
	CreatedAt     time.Time  `json:"created_at"`
}

// AlertRule defines an alert configuration for a business chain.
type AlertRule struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	BusinessChain       string    `json:"business_chain"`
	AlertType           string    `json:"alert_type"`
	Severity            string    `json:"severity"`
	ConditionField      string    `json:"condition_field"`
	ConditionOperator   string    `json:"condition_operator"`
	ConditionThreshold  *int      `json:"condition_threshold,omitempty"`
	ConditionDurationMin *int     `json:"condition_duration_min,omitempty"`
	NotifyRoles         string    `json:"notify_roles"`
	NotifyChannels      string    `json:"notify_channels"`
	NotifyInstitutionIDs string   `json:"notify_institution_ids"`
	EscalationTimeoutMin int     `json:"escalation_timeout_min"`
	EscalationRoles     string    `json:"escalation_roles"`
	AutoAction          string    `json:"auto_action"`
	Active              int       `json:"active"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// HealthRecordV2 is the extended health record supporting all chains.
type HealthRecordV2 struct {
	ID              string     `json:"id"`
	PersonID        string     `json:"person_id"`
	BusinessChain   string     `json:"business_chain"`
	RecordType      string     `json:"record_type"`
	Source          string     `json:"source"`
	DeviceID        string     `json:"device_id,omitempty"`
	RecordedAt      time.Time  `json:"recorded_at"`
	HeartRate       *int       `json:"heart_rate,omitempty"`
	BloodPressureSys *int      `json:"blood_pressure_sys,omitempty"`
	BloodPressureDia *int      `json:"blood_pressure_dia,omitempty"`
	SpO2            *int       `json:"spo2,omitempty"`
	Temperature     *float64   `json:"temperature,omitempty"`
	RespiratoryRate *int       `json:"respiratory_rate,omitempty"`
	PulseRate       *int       `json:"pulse_rate,omitempty"`
	GlucoseFasting  *float64   `json:"blood_glucose_fasting,omitempty"`
	GlucosePost     *float64   `json:"blood_glucose_postprandial,omitempty"`
	UricAcid        *float64   `json:"uric_acid,omitempty"`
	Creatinine      *float64   `json:"creatinine,omitempty"`
	HbA1c           *float64   `json:"hemoglobin_a1c,omitempty"`
	Weight          *float64   `json:"weight,omitempty"`
	Height          *float64   `json:"height,omitempty"`
	BMI             *float64   `json:"bmi,omitempty"`
	Steps           *int64     `json:"steps,omitempty"`
	SleepHours      *float64   `json:"sleep_hours,omitempty"`
	ExerciseMinutes *int       `json:"exercise_minutes,omitempty"`
	Notes           string     `json:"notes,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// PersonHealthSummary is the pre-computed health summary.
type PersonHealthSummary struct {
	PersonID          string    `json:"person_id"`
	BusinessChain     string    `json:"business_chain"`
	LatestHR          *int      `json:"latest_hr,omitempty"`
	LatestSpO2        *int      `json:"latest_spo2,omitempty"`
	LatestBPSys       *int      `json:"latest_bp_sys,omitempty"`
	LatestBPDia       *int      `json:"latest_bp_dia,omitempty"`
	LatestGlucoseFasting *float64 `json:"latest_glucose_fasting,omitempty"`
	LatestUricAcid    *float64  `json:"latest_uric_acid,omitempty"`
	LatestSteps       *int64    `json:"latest_steps,omitempty"`
	LatestSleepHours  *float64  `json:"latest_sleep_hours,omitempty"`
	RiskScore         *float64  `json:"risk_score,omitempty"`
	TrendDirection    string    `json:"trend_direction,omitempty"`
	LastUpdated       time.Time `json:"last_updated"`
	ARecommendation   string    `json:"ai_recommendation,omitempty"`
}

// HealthGuidanceRule defines a health guidance rule.
type HealthGuidanceRule struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	BusinessChain   string    `json:"business_chain"`
	TriggerCondition string   `json:"trigger_condition"`
	ConditionField  string    `json:"condition_field,omitempty"`
	ConditionOp     string    `json:"condition_operator,omitempty"`
	ConditionThresh *float64  `json:"condition_threshold,omitempty"`
	GuidanceType    string    `json:"guidance_type"`
	Title           string    `json:"title"`
	Content         string    `json:"content"`
	Priority        int       `json:"priority"`
	Enabled         int       `json:"enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// HealthGuidanceDelivery records a guidance push.
type HealthGuidanceDelivery struct {
	ID           string    `json:"id"`
	PersonID     string    `json:"person_id"`
	BusinessChain string   `json:"business_chain"`
	RuleID       string    `json:"rule_id"`
	GuidanceType string    `json:"guidance_type"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Channel      string    `json:"channel"`
	DeliveredAt  time.Time `json:"delivered_at"`
	ReadStatus   int       `json:"read_status"`
	Feedback     string    `json:"feedback,omitempty"`
}

// HealthReportTemplate defines a report template.
type HealthReportTemplate struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	BusinessChain  string    `json:"business_chain"`
	Frequency      string    `json:"frequency"`
	TemplateType   string    `json:"template_type"`
	IncludeSections string   `json:"include_sections"`
	Enabled        int       `json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
}

// HealthReport records a generated report.
type HealthReport struct {
	ID                 string     `json:"id"`
	PersonID           string     `json:"person_id"`
	BusinessChain      string     `json:"business_chain"`
	TemplateID         string     `json:"template_id,omitempty"`
	ReportPeriodStart  time.Time  `json:"report_period_start"`
	ReportPeriodEnd    time.Time  `json:"report_period_end"`
	GeneratedAt        time.Time  `json:"generated_at"`
	ReportData         string     `json:"report_data,omitempty"`
	DeliveredChannels  string     `json:"delivered_channels,omitempty"`
	Status             string     `json:"status"`
	CreatedAt          time.Time  `json:"created_at"`
}

// ComplianceRule defines a compliance check rule.
type ComplianceRule struct {
	ID            string    `json:"id"`
	RuleCode      string    `json:"rule_code"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	BusinessChain string    `json:"business_chain"`
	RuleType      string    `json:"rule_type"`
	ConditionSQL  string    `json:"condition_sql"`
	Severity      string    `json:"severity"`
	ActionRequired string   `json:"action_required,omitempty"`
	Enabled       int       `json:"enabled"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ComplianceCheck records a compliance check result.
type ComplianceCheck struct {
	ID                string     `json:"id"`
	RuleID            string     `json:"rule_id"`
	PersonID          string     `json:"person_id"`
	CheckTime         time.Time  `json:"check_time"`
	Violated          int        `json:"violated"`
	ViolationDetails  string     `json:"violation_details,omitempty"`
	ReviewedBy        string     `json:"reviewed_by,omitempty"`
	ReviewedAt        *time.Time `json:"reviewed_at,omitempty"`
	ReviewResult      string     `json:"review_result,omitempty"`
	ActionTaken       string     `json:"action_taken,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

// DeviceBinding links a device to a person on a business chain.
type DeviceBinding struct {
	ID            string     `json:"id"`
	DeviceID      string     `json:"device_id"`
	PersonID      string     `json:"person_id"`
	BusinessChain string     `json:"business_chain"`
	BoundAt       time.Time  `json:"bound_at"`
	UnboundAt     *time.Time `json:"unbound_at,omitempty"`
	BindingType   string     `json:"binding_type"`
	CreatedAt     time.Time  `json:"created_at"`
}

// NotificationTemplate defines a notification template.
type NotificationTemplate struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	BusinessChain string   `json:"business_chain"`
	Channel      string    `json:"channel"`
	Subject      string    `json:"subject,omitempty"`
	BodyTemplate string    `json:"body_template"`
	Enabled      int       `json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
}

// NotificationLog records a notification send.
type NotificationLog struct {
	ID             string     `json:"id"`
	PersonID       string     `json:"person_id"`
	BusinessChain  string     `json:"business_chain"`
	TemplateID     string     `json:"template_id,omitempty"`
	RecipientRole  string     `json:"recipient_role,omitempty"`
	RecipientID    string     `json:"recipient_id,omitempty"`
	Channel        string     `json:"channel"`
	Status         string     `json:"status"`
	SentAt         *time.Time `json:"sent_at,omitempty"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}


type MedicationExecution struct {
	ID                 string `json:"id"`
	PersonID           string `json:"person_id"`
	BusinessChain      string `json:"business_chain"`
	RuleID             string `json:"rule_id"`
	ScheduledTime      string `json:"scheduled_time"`
	ActualTime         string `json:"actual_time,omitempty"`
	Status             string `json:"status"`
	TakenBy            string `json:"taken_by"`
	DeviceID           string `json:"device_id"`
	VerificationMethod string `json:"verification_method"`
	Notes              string `json:"notes"`
	CreatedAt          string `json:"created_at"`
}
