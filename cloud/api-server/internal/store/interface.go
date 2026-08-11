// Package store defines the unified data access interface for both SQLite and
// PostgreSQL backends in the api-server.
package store

import (
	"context"
	"time"

	"eregen.dev/api-server/internal/model"
)

// Store is the comprehensive data access interface for both SQLite and PostgreSQL backends.
// Both Postgres and SqliteStore implement this interface.
type Store interface {
	// Database interface
	Raw() RawDB
	Health(ctx context.Context) error

	// UserDomain
	CreateUser(ctx context.Context, u *model.User) error
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

	// DeviceDomain
	CreateDevice(ctx context.Context, d *model.Device) error
	ListDevices(ctx context.Context, ownerID string, deviceType *string, page, pageSize int) ([]model.Device, int, error)
	GetDeviceByDeviceID(ctx context.Context, deviceID string) (*model.Device, error)
	DeleteDevice(ctx context.Context, deviceID, ownerID string) error
	AdminDeleteDevice(ctx context.Context, deviceID string) error
	BindDevice(ctx context.Context, deviceID, ownerUserID, deviceType, tier string) (*model.Device, error)
	AdminDeviceList(ctx context.Context, deviceType, tier, status string, page, pageSize int) ([]model.Device, int, error)

	// HealthRecord
	CreateHealthRecord(ctx context.Context, r *model.HealthRecord) error
	GetHealthSummary(ctx context.Context, elderlyID string, day time.Time) (*model.HealthRecord, error)
	GetHealthHistory(ctx context.Context, elderlyID string, days int) ([]model.HealthRecord, error)
	GetHealthTrend(ctx context.Context, elderlyID, metric string, days int) ([]model.HealthRecord, error)
	GetHealthRecordsByElderlyID(ctx context.Context, elderlyID string, from, until time.Time) ([]model.HealthRecord, error)
	LatestHealthByElderlyID(ctx context.Context, elderlyID string, since time.Time) (*model.HealthRecord, error)
	HealthRecordsByElderlyIDs(ctx context.Context, elderIDs []string, days int) ([]model.HealthRecord, error)
	HealthTrendByElderlyID(ctx context.Context, elderlyID string, days int) (avgHR, avgSpO2, totalSteps int64, lastHR, lastSpO2 *int, err error)
	GetElderlyName(ctx context.Context, elderlyID string) (string, error)

	// LocationRecord
	CreateLocationRecord(ctx context.Context, r *model.LocationRecord) error
	GetLatestLocation(ctx context.Context, elderlyID string) (*model.LocationRecord, error)
	GetLocationHistory(ctx context.Context, elderlyID string, from, until time.Time) ([]model.LocationRecord, error)
	GetLocationHistoryByElderlyID(ctx context.Context, elderlyID string, from, until time.Time) ([]model.LocationRecord, error)

	// MedicationRule
	CreateMedicationRule(ctx context.Context, mr *model.MedicationRule) error
	ListMedicationRules(ctx context.Context, elderlyID string) ([]model.MedicationRule, error)
	GetMedicationRule(ctx context.Context, ruleID string) (*model.MedicationRule, error)
	UpdateMedicationRule(ctx context.Context, ruleID string, req *model.CreateMedicationRuleRequest) error
	DeleteMedicationRule(ctx context.Context, ruleID string) error
	GetMedicationRulesByElderlyID(ctx context.Context, elderlyID string) ([]model.MedicationRule, error)

	// MedStatusRecord
	CreateMedStatusRecord(ctx context.Context, r *model.MedStatusRecord) error
	GetTodayMedStatus(ctx context.Context, elderlyID string) ([]model.MedStatusRecord, error)
	GetMedicationHistory(ctx context.Context, elderlyID string, days int) ([]model.MedStatusRecord, error)
	CreateMedTakeRecord(ctx context.Context, ruleID, elderlyID string) error

	// AlertDomain
	CreateAlert(ctx context.Context, a *model.Alert) error
	ListAlerts(ctx context.Context, elderIDs []string, filter *model.AlertFilter, page, pageSize int) ([]model.Alert, int, error)
	GetAlert(ctx context.Context, alertID string) (*model.Alert, error)
	UpdateAlert(ctx context.Context, alertID string, status model.AlertStatus) error
	ResolveAlertByID(ctx context.Context, alertID string) error
	GetAlertsByElderlyID(ctx context.Context, elderlyID string, from, until time.Time) ([]model.Alert, error)
	GetAlertElderlyID(ctx context.Context, alertID string) (string, error)

	// Geofence
	CreateGeofence(ctx context.Context, gf *model.Geofence) error
	ListGeofences(ctx context.Context, elderlyID string) ([]model.Geofence, error)
	UpdateGeofence(ctx context.Context, id string, req *model.UpdateGeofenceRequest) error
	DeleteGeofence(ctx context.Context, id string) error

	// AdminDomain
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

	// Legacy Store interface (for backward compatibility)
	ListElderly(ctx context.Context, page, pageSize int) ([]model.ElderlyProfile, error)
	ListUsersAdmin(ctx context.Context, page, pageSize int, role string) ([]model.UserSummary, error)
	ListDevicesAdmin(ctx context.Context, status string) ([]model.DeviceSummary, error)
	GetActiveAlerts(ctx context.Context) ([]model.AlertSummary, error)
	ValidateToken(ctx context.Context, token string) (string, error)
	ListDailyTasks(ctx context.Context, elderlyID string, taskDate string) ([]model.ChronicDailyTask, error)
	UpdateDailyTaskComplete(ctx context.Context, taskID string) error
}

var _ Store = (*Postgres)(nil)
var _ Store = (*SqliteStore)(nil)
