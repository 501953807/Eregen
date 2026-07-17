package model

import "time"

// Role represents an admin user role.
type Role string

const (
	RoleAdmin      Role = "admin"
	RoleOperator   Role = "operator"
	RoleSuperAdmin Role = "super_admin"
)

// DashboardStats aggregates the key metrics shown on the admin dashboard.
type DashboardStats struct {
	OnlineDevices       int          `json:"online_devices"`
	TotalDevices        int          `json:"total_devices"`
	ActiveAlerts        int          `json:"active_alerts"`
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
	ID         string   `json:"id"`
	UserID     string   `json:"user_id"`
	Name       string   `json:"name"`
	BirthDate  *string  `json:"birth_date,omitempty"`
	AvatarURL  string   `json:"avatar_url,omitempty"`
	HealthTiers []string `json:"health_tiers"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
}
