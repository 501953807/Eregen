package model

import "time"

// Role represents a user role in the system.
type Role string

const (
	RoleFamily   Role = "family"
	RoleElderly  Role = "elderly"
	RoleInstitution Role = "institution"
)

// PlanTier represents a subscription plan tier.
type PlanTier string

const (
	PlanFree      PlanTier = "free"
	PlanPremium   PlanTier = "premium"
	PlanEnterprise PlanTier = "enterprise"
)

// AlertSeverity represents the priority of an alert.
type AlertSeverity string

const (
	AlertP0 AlertSeverity = "P0" // SOS, fall — immediate action
	AlertP1 AlertSeverity = "P1" // medication missed, geofence breach
	AlertP2 AlertSeverity = "P2" // device offline, low battery
)

// AlertStatus tracks whether an alert has been handled.
type AlertStatus string

const (
	AlertPending AlertStatus = "pending"
	AlertResolved AlertStatus = "resolved"
)

// DeviceStatus represents the online/offline state of a device.
type DeviceStatus string

const (
	DeviceOnline  DeviceStatus = "online"
	DeviceOffline DeviceStatus = "offline"
)

// User is a registered account holder.
type User struct {
	ID           string    `json:"id"`
	Email        *string   `json:"email,omitempty"`
	Phone        *string   `json:"phone,omitempty"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ElderlyProfile holds health and identity info for an elderly person.
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

// Device represents a connected hardware device.
type Device struct {
	ID          string          `json:"id"`
	DeviceID    string          `json:"device_id"` // BR-XXXX or PX-XXXX
	DeviceType  string          `json:"device_type"`
	Tier        string          `json:"tier"`       // starter/plus/pro
	OwnerUserID string          `json:"owner_user_id"`
	Status      DeviceStatus    `json:"status"`
	LastSeen    *time.Time      `json:"last_seen,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	Settings    map[string]any  `json:"settings,omitempty"`
}

// HealthRecord stores a point-in-time health reading.
type HealthRecord struct {
	ID            string    `json:"id"`
	ElderlyID     string    `json:"elderly_id"`
	Timestamp     time.Time `json:"timestamp"`
	HR            *int      `json:"hr,omitempty"`
	SPO2          *int      `json:"spo2,omitempty"`
	Steps         *int64    `json:"steps,omitempty"`
	SleepHours    *float64  `json:"sleep_hours,omitempty"`
	BPSystolic    *int      `json:"bp_systolic,omitempty"`
	BPDiastolic   *int      `json:"bp_diastolic,omitempty"`
}

// LocationRecord stores a GPS fix.
type LocationRecord struct {
	ID        string    `json:"id"`
	ElderlyID string    `json:"elderly_id"`
	Timestamp time.Time `json:"timestamp"`
	Lat       float64   `json:"lat"`
	Lon       float64   `json:"lon"`
	Accuracy  *float64  `json:"accuracy,omitempty"`
}

// MedicationRule defines when pills should be taken.
type MedicationRule struct {
	ID            string         `json:"id"`
	ElderlyID     string         `json:"elderly_id"`
	ScheduleTime  string         `json:"schedule_time"` // HH:MM format
	DoseCount     int            `json:"dose_count"`
	PillType      string         `json:"pill_type"`
	DaysOfWeek    []int          `json:"days_of_week"` // 1=Mon … 7=Sun
	Active        bool           `json:"active"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// MedStatusRecord tracks whether a dose was taken.
type MedStatusRecord struct {
	ID         string    `json:"id"`
	RuleID     string    `json:"rule_id"`
	ElderlyID  string    `json:"elderly_id"`
	TakenAt    *time.Time `json:"taken_at,omitempty"`
	Taken      bool      `json:"taken"`
	MissedAt   *time.Time `json:"missed_at,omitempty"`
}

// Alert is a notification-worthy event.
type Alert struct {
	ID         string        `json:"id"`
	ElderlyID  string        `json:"elderly_id"`
	AlertType  string        `json:"alert_type"` // sos, fall, med_missed, device_offline, geofence_breach
	Severity   AlertSeverity `json:"severity"`
	Status     AlertStatus   `json:"status"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
	ResolvedAt *time.Time    `json:"resolved_at,omitempty"`
}

// Subscription tracks a user's paid plan.
type Subscription struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	PlanTier  PlanTier  `json:"plan_tier"`
	Status    string    `json:"status"` // active, canceled, expired, trialing
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// OTPRequest holds a one-time-password verification payload.
type OTPRequest struct {
	Code string `json:"code"`
}

// LoginRequest contains credentials for authentication.
type LoginRequest struct {
	Identifier string `json:"identifier"` // phone or email
	Password   string `json:"password"`
}

// RegisterRequest for new account creation with OTP.
type RegisterRequest struct {
	Phone    *string `json:"phone,omitempty"`
	Email    *string `json:"email,omitempty"`
	Password string  `json:"password"`
	OTPCode  string  `json:"otp_code"`
	Name     string  `json:"name"`
}

// UpdateUserRequest for profile updates.
type UpdateUserRequest struct {
	Name  *string `json:"name,omitempty"`
	Phone *string `json:"phone,omitempty"`
	Email *string `json:"email,omitempty"`
}

// UpdateElderlyRequest for elderly profile updates.
type UpdateElderlyRequest struct {
	Name      string   `json:"name"`
	BirthDate *string  `json:"birth_date,omitempty"`
	AvatarURL *string  `json:"avatar_url,omitempty"`
	HealthTiers []string `json:"health_tiers"`
}

// BindDeviceRequest for linking a device to a user.
type BindDeviceRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
}

// DeviceSettingsRequest for updating device configuration.
type DeviceSettingsRequest struct {
	Settings map[string]any `json:"settings"`
}

// GeofenceRequest for setting up electronic fence.
type GeofenceRequest struct {
	Lat          float64 `json:"lat"`
	Lon          float64 `json:"lon"`
	RadiusMeters int     `json:"radius_meters"`
	Name         string  `json:"name"`
}

// Geofence is the persisted geofence entity.
type Geofence struct {
	ID           string    `json:"id"`
	ElderlyID    string    `json:"elderly_id"`
	Name         string    `json:"name"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	RadiusMeters int       `json:"radius_meters"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UpdateGeofenceRequest for modifying an existing geofence.
type UpdateGeofenceRequest struct {
	Name         string `json:"name"`
	Lat          float64 `json:"lat"`
	Lon          float64 `json:"lon"`
	RadiusMeters int    `json:"radius_meters"`
	Active       bool   `json:"active"`
}

// CreateMedicationRuleRequest for adding a new medication schedule.
type CreateMedicationRuleRequest struct {
	ScheduleTime string `json:"schedule_time"`
	DoseCount    int    `json:"dose_count"`
	PillType     string `json:"pill_type"`
	DaysOfWeek   []int  `json:"days_of_week"`
	Active       bool   `json:"active"`
}

// AlertFilter for querying alerts.
type AlertFilter struct {
	Severity *AlertSeverity `json:"severity,omitempty"`
	Status   *AlertStatus   `json:"status,omitempty"`
}
