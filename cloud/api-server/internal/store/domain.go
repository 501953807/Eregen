// Package store defines domain-driven interfaces for the api-server.
// Each interface represents a bounded context that a handler depends on.
package store

import (
	"context"
	"time"

	"eregen.dev/api-server/internal/model"
)

// UserDomain encapsulates all user and elderly profile operations.
type UserDomain interface {
	CreateUser(ctx context.Context, u *model.User) error
	CreateDevice(ctx context.Context, d *model.Device) error
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByOpenID(ctx context.Context, openID string) (*model.User, error)
	UpdateUser(ctx context.Context, id string, name, phone, email *string) error
	DeleteUser(ctx context.Context, userID string) error
	ListUsers(ctx context.Context, page, pageSize int) ([]model.User, int, error)
	UpdateUserRole(ctx context.Context, userID, role string) error
	GetElderlyIDsByUserID(ctx context.Context, userID string) ([]string, error)
	GetElderlyProfilesByUserID(ctx context.Context, userID string) ([]model.ElderlyProfile, error)
	CreateElderlyProfile(ctx context.Context, ep *model.ElderlyProfile) error
	GetElderlyProfile(ctx context.Context, elderlyID string) (*model.ElderlyProfile, error)
	UpdateElderlyProfile(ctx context.Context, elderlyID string, req *model.UpdateElderlyRequest) error
	ListElderlyProfiles(ctx context.Context, userID string, page, pageSize int) ([]model.ElderlyProfile, int, error)
	GetDeviceByElderlyID(ctx context.Context, elderlyID string) (string, error)
	CheckElderlyAccess(ctx context.Context, elderlyID, userID string) (bool, error)
	MedRuleElderlyID(ctx context.Context, ruleID string) (string, error)
	GetDevice(ctx context.Context, deviceID string) (*model.Device, error)
	UpdateDeviceSettings(ctx context.Context, deviceID string, settings map[string]any) error
}

// DeviceDomain encapsulates all device, health, location, and medication operations.
type DeviceDomain interface {
	CreateDevice(ctx context.Context, d *model.Device) error
	ListDevices(ctx context.Context, ownerID string, deviceType *string, page, pageSize int) ([]model.Device, int, error)
	GetDevice(ctx context.Context, deviceID string) (*model.Device, error)
	GetDeviceByDeviceID(ctx context.Context, deviceID string) (*model.Device, error)
	UpdateDeviceSettings(ctx context.Context, deviceID string, settings map[string]any) error
	DeleteDevice(ctx context.Context, deviceID, ownerID string) error
	AdminDeleteDevice(ctx context.Context, deviceID string) error
	BindDevice(ctx context.Context, deviceID, ownerUserID, deviceType, tier string) (*model.Device, error)
	AdminDeviceList(ctx context.Context, deviceType, tier, status string, page, pageSize int) ([]model.Device, int, error)
	CreateHealthRecord(ctx context.Context, r *model.HealthRecord) error
	GetHealthSummary(ctx context.Context, elderlyID string, day time.Time) (*model.HealthRecord, error)
	GetHealthHistory(ctx context.Context, elderlyID string, days int) ([]model.HealthRecord, error)
	GetHealthTrend(ctx context.Context, elderlyID, metric string, days int) ([]model.HealthRecord, error)
	GetHealthRecordsByElderlyID(ctx context.Context, elderlyID string, from, until time.Time) ([]model.HealthRecord, error)
	CreateLocationRecord(ctx context.Context, r *model.LocationRecord) error
	GetLatestLocation(ctx context.Context, elderlyID string) (*model.LocationRecord, error)
	GetLocationHistory(ctx context.Context, elderlyID string, from, until time.Time) ([]model.LocationRecord, error)
	GetLocationHistoryByElderlyID(ctx context.Context, elderlyID string, from, until time.Time) ([]model.LocationRecord, error)
	CreateMedStatusRecord(ctx context.Context, r *model.MedStatusRecord) error
	GetTodayMedStatus(ctx context.Context, elderlyID string) ([]model.MedStatusRecord, error)
	GetMedicationHistory(ctx context.Context, elderlyID string, days int) ([]model.MedStatusRecord, error)
	CreateMedicationRule(ctx context.Context, mr *model.MedicationRule) error
	ListMedicationRules(ctx context.Context, elderlyID string) ([]model.MedicationRule, error)
	GetMedicationRule(ctx context.Context, ruleID string) (*model.MedicationRule, error)
	UpdateMedicationRule(ctx context.Context, ruleID string, req *model.CreateMedicationRuleRequest) error
	DeleteMedicationRule(ctx context.Context, ruleID string) error
	GetMedicationRulesByElderlyID(ctx context.Context, elderlyID string) ([]model.MedicationRule, error)
	CreateMedTakeRecord(ctx context.Context, ruleID, elderlyID string) error
	LatestHealthByElderlyID(ctx context.Context, elderlyID string, since time.Time) (*model.HealthRecord, error)
	HealthRecordsByElderlyIDs(ctx context.Context, elderIDs []string, days int) ([]model.HealthRecord, error)
	HealthTrendByElderlyID(ctx context.Context, elderlyID string, days int) (avgHR, avgSpO2, totalSteps int64, lastHR, lastSpO2 *int, err error)
	GetElderlyName(ctx context.Context, elderlyID string) (string, error)
}

// AlertDomain encapsulates all alert and geofence operations.
type AlertDomain interface {
	CreateAlert(ctx context.Context, a *model.Alert) error
	ListAlerts(ctx context.Context, elderIDs []string, filter *model.AlertFilter, page, pageSize int) ([]model.Alert, int, error)
	GetAlert(ctx context.Context, alertID string) (*model.Alert, error)
	UpdateAlert(ctx context.Context, alertID string, status model.AlertStatus) error
	ResolveAlertByID(ctx context.Context, alertID string) error
	GetAlertsByElderlyID(ctx context.Context, elderlyID string, from, until time.Time) ([]model.Alert, error)
	CreateGeofence(ctx context.Context, gf *model.Geofence) error
	ListGeofences(ctx context.Context, elderlyID string) ([]model.Geofence, error)
	UpdateGeofence(ctx context.Context, id string, req *model.UpdateGeofenceRequest) error
	DeleteGeofence(ctx context.Context, id string) error
	GetAlertElderlyID(ctx context.Context, alertID string) (string, error)
}

// SessionDomain encapsulates all session, token, and cache operations.
type SessionDomain interface {
	SetRefreshToken(ctx context.Context, token string, userID string, ttl time.Duration) error
	ValidateRefreshToken(ctx context.Context, token string) (string, error)
	InvalidateRefreshToken(ctx context.Context, token string) error
	SetOTP(ctx context.Context, phoneOrEmail, code string, ttl time.Duration) error
	VerifyOTP(ctx context.Context, phoneOrEmail, code string) error
	DelByPattern(ctx context.Context, pattern string) error
	SetResetToken(ctx context.Context, token, userID string, ttl time.Duration) error
	Store(ctx context.Context, key, value string, ttl time.Duration) error
	Verify(ctx context.Context, key, value string) error
	SetDeviceOnline(ctx context.Context, deviceID string) error
	IsDeviceOnline(ctx context.Context, deviceID string) (bool, error)
	InvalidateDevice(ctx context.Context, deviceID string) error
}

// AdminDomain encapsulates admin-only operations (stats, subscriptions, firmware).
type AdminDomain interface {
	AdminStatsOverview(ctx context.Context) (*StatsOverview, error)
	AdminStatsAlertTrend(ctx context.Context, days int) ([]AlertTrendPoint, error)
	AdminStatsAlertDistribution(ctx context.Context) ([]AlertDistributionItem, error)
	AdminStatsUserGrowth(ctx context.Context, months int) ([]UserGrowthPoint, error)
	CreateSubscription(ctx context.Context, s *model.Subscription) error
	GetSubscription(ctx context.Context, userID string) (*model.Subscription, error)
	ListSubscriptions(ctx context.Context, userID string, page, pageSize int) ([]model.Subscription, int, error)
	SubscriptionTierStats(ctx context.Context) ([]struct{ Tier string; Count int }, error)
	CreateFirmwareRelease(ctx context.Context, r *model.FirmwareRelease) error
	ListFirmwareReleases(ctx context.Context, deviceType, tier string) ([]model.FirmwareRelease, error)
	GetFirmwareRelease(ctx context.Context, id string) (*model.FirmwareRelease, error)
	CreateOTAJob(ctx context.Context, j *model.OTAJob) error
	GetOTAJob(ctx context.Context, id string) (*model.OTAJob, error)
	UpdateOTAJobProgress(ctx context.Context, jobID string, fn UpdateOTAJobProgressFn) error
}

// DeviceLister is a minimal interface for listing devices (used by OTA service).
type DeviceLister interface {
	ListDevices(ctx context.Context, ownerID string, deviceType *string, page, pageSize int) ([]model.Device, int, error)
}
