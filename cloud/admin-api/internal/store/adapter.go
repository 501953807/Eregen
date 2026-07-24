// Package store provides the Store interface adapter for *sql.DB.
package store

import (
	"context"
	"database/sql"
	"eregen.dev/admin-api/internal/model"
	"time"
)

// StoreAdapter adapts *sql.DB to the Store interface by selecting the appropriate backend.
type StoreAdapter struct {
	db     *sql.DB
	dbType string // "postgres" or "sqlite"
}

func NewStore(db *sql.DB, dbType string) *StoreAdapter {
	return &StoreAdapter{db: db, dbType: dbType}
}

func (a *StoreAdapter) GetDashboardStats(ctx context.Context) (*model.DashboardStats, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).GetDashboardStats(ctx)
	}
	return (&SqliteStore{db: a.db}).GetDashboardStats(ctx)
}

func (a *StoreAdapter) ListDevices(ctx context.Context, page, pageSize int, status, devType, tier string) ([]model.DeviceSummary, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).ListDevices(ctx, page, pageSize, status, devType, tier)
	}
	return (&SqliteStore{db: a.db}).ListDevices(ctx, page, pageSize, status, devType, tier)
}

func (a *StoreAdapter) ListUsers(ctx context.Context, page, pageSize int, role string) ([]model.UserSummary, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).ListUsers(ctx, page, pageSize, role)
	}
	return (&SqliteStore{db: a.db}).ListUsers(ctx, page, pageSize, role)
}

func (a *StoreAdapter) ListAlerts(ctx context.Context, severity, status string, limit int) ([]model.AlertSummary, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).ListAlerts(ctx, severity, status, limit)
	}
	return (&SqliteStore{db: a.db}).ListAlerts(ctx, severity, status, limit)
}

func (a *StoreAdapter) SetUserRole(ctx context.Context, userID, role string) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).SetUserRole(ctx, userID, role)
	}
	return (&SqliteStore{db: a.db}).SetUserRole(ctx, userID, role)
}

func (a *StoreAdapter) UpdateDeviceConfig(ctx context.Context, deviceID string, config map[string]interface{}) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).UpdateDeviceConfig(ctx, deviceID, config)
	}
	return (&SqliteStore{db: a.db}).UpdateDeviceConfig(ctx, deviceID, config)
}

func (a *StoreAdapter) TriggerOTA(ctx context.Context, deviceID, firmwareURL, sha256Hash string) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).TriggerOTA(ctx, deviceID, firmwareURL, sha256Hash)
	}
	return (&SqliteStore{db: a.db}).TriggerOTA(ctx, deviceID, firmwareURL, sha256Hash)
}

func (a *StoreAdapter) ResolveAlert(ctx context.Context, alertID string) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).ResolveAlert(ctx, alertID)
	}
	return (&SqliteStore{db: a.db}).ResolveAlert(ctx, alertID)
}

func (a *StoreAdapter) GetSubscriptionStats(ctx context.Context) ([]model.SubscriptionStat, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).GetSubscriptionStats(ctx)
	}
	return (&SqliteStore{db: a.db}).GetSubscriptionStats(ctx)
}

func (a *StoreAdapter) GetAlertTrend(ctx context.Context, days int) ([]model.AlertTrendPoint, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).GetAlertTrend(ctx, days)
	}
	return (&SqliteStore{db: a.db}).GetAlertTrend(ctx, days)
}

func (a *StoreAdapter) GetAlertDistribution(ctx context.Context) ([]model.AlertDistributionItem, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).GetAlertDistribution(ctx)
	}
	return (&SqliteStore{db: a.db}).GetAlertDistribution(ctx)
}

func (a *StoreAdapter) GetUserGrowth(ctx context.Context, months int) ([]model.UserGrowthPoint, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).GetUserGrowth(ctx, months)
	}
	return (&SqliteStore{db: a.db}).GetUserGrowth(ctx, months)
}

func (a *StoreAdapter) GetDeviceByID(ctx context.Context, id string) (*model.DeviceDetail, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).GetDeviceByID(ctx, id)
	}
	return (&SqliteStore{db: a.db}).GetDeviceByID(ctx, id)
}

func (a *StoreAdapter) UnbindDevice(ctx context.Context, deviceID string) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).UnbindDevice(ctx, deviceID)
	}
	return (&SqliteStore{db: a.db}).UnbindDevice(ctx, deviceID)
}

func (a *StoreAdapter) BatchTriggerOTA(ctx context.Context, deviceIDs, firmwareURL, sha256Hash []string) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).BatchTriggerOTA(ctx, deviceIDs, firmwareURL, sha256Hash)
	}
	return (&SqliteStore{db: a.db}).BatchTriggerOTA(ctx, deviceIDs, firmwareURL, sha256Hash)
}

func (a *StoreAdapter) CreateFirmwareVersion(ctx context.Context, v *model.FirmwareVersion) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).CreateFirmwareVersion(ctx, v)
	}
	return (&SqliteStore{db: a.db}).CreateFirmwareVersion(ctx, v)
}

func (a *StoreAdapter) ListFirmwareVersions(ctx context.Context) ([]model.FirmwareVersion, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).ListFirmwareVersions(ctx)
	}
	return (&SqliteStore{db: a.db}).ListFirmwareVersions(ctx)
}

func (a *StoreAdapter) DeleteFirmwareVersion(ctx context.Context, id string) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).DeleteFirmwareVersion(ctx, id)
	}
	return (&SqliteStore{db: a.db}).DeleteFirmwareVersion(ctx, id)
}

func (a *StoreAdapter) PushOTAJob(ctx context.Context, firmwareID string, deviceIDs []string) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).PushOTAJob(ctx, firmwareID, deviceIDs)
	}
	return (&SqliteStore{db: a.db}).PushOTAJob(ctx, firmwareID, deviceIDs)
}

func (a *StoreAdapter) GetNotificationSettings(ctx context.Context) (map[string]any, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).GetNotificationSettings(ctx)
	}
	return (&SqliteStore{db: a.db}).GetNotificationSettings(ctx)
}

func (a *StoreAdapter) UpdateNotificationSettings(ctx context.Context, data map[string]any) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).UpdateNotificationSettings(ctx, data)
	}
	return (&SqliteStore{db: a.db}).UpdateNotificationSettings(ctx, data)
}

func (a *StoreAdapter) ListAPIKeys(ctx context.Context) ([]model.APIKeySummary, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).ListAPIKeys(ctx)
	}
	return (&SqliteStore{db: a.db}).ListAPIKeys(ctx)
}

func (a *StoreAdapter) CreateAPIKey(ctx context.Context, name, keyHash string, expiresAt *time.Time) (string, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).CreateAPIKey(ctx, name, keyHash, expiresAt)
	}
	return (&SqliteStore{db: a.db}).CreateAPIKey(ctx, name, keyHash, expiresAt)
}

func (a *StoreAdapter) RevokeAPIKey(ctx context.Context, id string) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).RevokeAPIKey(ctx, id)
	}
	return (&SqliteStore{db: a.db}).RevokeAPIKey(ctx, id)
}

func (a *StoreAdapter) ChangeAdminPassword(ctx context.Context, userID, hash string) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).ChangeAdminPassword(ctx, userID, hash)
	}
	return (&SqliteStore{db: a.db}).ChangeAdminPassword(ctx, userID, hash)
}

// ========== Elderly Profile Management ==========

func (a *StoreAdapter) ListElderly(ctx context.Context, page, pageSize int) ([]model.ElderlyProfile, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).ListElderly(ctx, page, pageSize)
	}
	return (&SqliteStore{db: a.db}).ListElderly(ctx, page, pageSize)
}

func (a *StoreAdapter) GetElderly(ctx context.Context, id string) (*model.ElderlyProfile, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).GetElderly(ctx, id)
	}
	return (&SqliteStore{db: a.db}).GetElderly(ctx, id)
}

func (a *StoreAdapter) CreateElderly(ctx context.Context, name, birthDate, userID string, healthTiers []string, avatarURL string) (*model.ElderlyProfile, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).CreateElderly(ctx, name, birthDate, userID, healthTiers, avatarURL)
	}
	return (&SqliteStore{db: a.db}).CreateElderly(ctx, name, birthDate, userID, healthTiers, avatarURL)
}

func (a *StoreAdapter) UpdateElderly(ctx context.Context, id, name, birthDate, userID string, healthTiers []string, avatarURL string) (*model.ElderlyProfile, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).UpdateElderly(ctx, id, name, birthDate, userID, healthTiers, avatarURL)
	}
	return (&SqliteStore{db: a.db}).UpdateElderly(ctx, id, name, birthDate, userID, healthTiers, avatarURL)
}

func (a *StoreAdapter) DeleteElderly(ctx context.Context, id string) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).DeleteElderly(ctx, id)
	}
	return (&SqliteStore{db: a.db}).DeleteElderly(ctx, id)
}

func (a *StoreAdapter) GetElderlyHealthStats(ctx context.Context, elderlyID string) (*model.HealthStats, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).GetElderlyHealthStats(ctx, elderlyID)
	}
	return (&SqliteStore{db: a.db}).GetElderlyHealthStats(ctx, elderlyID)
}

func (a *StoreAdapter) GetElderlyHealthRecords(ctx context.Context, elderlyID string, limit int) ([]model.HealthRecordRow, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).GetElderlyHealthRecords(ctx, elderlyID, limit)
	}
	return (&SqliteStore{db: a.db}).GetElderlyHealthRecords(ctx, elderlyID, limit)
}

func (a *StoreAdapter) GetElderlyMedicationRules(ctx context.Context, elderlyID string) ([]model.MedicationRuleRow, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).GetElderlyMedicationRules(ctx, elderlyID)
	}
	return (&SqliteStore{db: a.db}).GetElderlyMedicationRules(ctx, elderlyID)
}

func (a *StoreAdapter) GetElderlyDevices(ctx context.Context, elderlyID string) ([]model.DeviceSummaryRow, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).GetElderlyDevices(ctx, elderlyID)
	}
	return (&SqliteStore{db: a.db}).GetElderlyDevices(ctx, elderlyID)
}

func (a *StoreAdapter) GetElderlyLocationHistory(ctx context.Context, elderlyID string, limit int) ([]model.LocationPoint, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).GetElderlyLocationHistory(ctx, elderlyID, limit)
	}
	return (&SqliteStore{db: a.db}).GetElderlyLocationHistory(ctx, elderlyID, limit)
}

func (a *StoreAdapter) GetElderlyAlertHistory(ctx context.Context, elderlyID string, limit int) ([]model.AlertSummaryRow, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).GetElderlyAlertHistory(ctx, elderlyID, limit)
	}
	return (&SqliteStore{db: a.db}).GetElderlyAlertHistory(ctx, elderlyID, limit)
}

// ========== Medical Wristband Management ==========

func (a *StoreAdapter) CreatePatient(ctx context.Context, p *model.MedicalPatient) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).CreatePatient(ctx, p)
	}
	return (&SqliteStore{db: a.db}).CreatePatient(ctx, p)
}

func (a *StoreAdapter) GetPatient(ctx context.Context, id string) (*model.MedicalPatient, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).GetPatient(ctx, id)
	}
	return (&SqliteStore{db: a.db}).GetPatient(ctx, id)
}

func (a *StoreAdapter) ListPatients(ctx context.Context, page, pageSize int, status string) ([]model.MedicalPatient, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).ListPatients(ctx, page, pageSize, status)
	}
	return (&SqliteStore{db: a.db}).ListPatients(ctx, page, pageSize, status)
}

func (a *StoreAdapter) UpdatePatient(ctx context.Context, p *model.MedicalPatient) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).UpdatePatient(ctx, p)
	}
	return (&SqliteStore{db: a.db}).UpdatePatient(ctx, p)
}

func (a *StoreAdapter) DeletePatient(ctx context.Context, id string) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).DeletePatient(ctx, id)
	}
	return (&SqliteStore{db: a.db}).DeletePatient(ctx, id)
}

func (a *StoreAdapter) BindWristband(ctx context.Context, patientID, deviceID string) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).BindWristband(ctx, patientID, deviceID)
	}
	return (&SqliteStore{db: a.db}).BindWristband(ctx, patientID, deviceID)
}

func (a *StoreAdapter) UnbindWristband(ctx context.Context, bindingID string) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).UnbindWristband(ctx, bindingID)
	}
	return (&SqliteStore{db: a.db}).UnbindWristband(ctx, bindingID)
}

func (a *StoreAdapter) ClearWristband(ctx context.Context, deviceID string) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).ClearWristband(ctx, deviceID)
	}
	return (&SqliteStore{db: a.db}).ClearWristband(ctx, deviceID)
}

func (a *StoreAdapter) ListWristbands(ctx context.Context, page, pageSize int, status string) ([]model.MedicalWristbandDevice, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).ListWristbands(ctx, page, pageSize, status)
	}
	return (&SqliteStore{db: a.db}).ListWristbands(ctx, page, pageSize, status)
}

func (a *StoreAdapter) GetWristbandFirmware(ctx context.Context, deviceID string) (string, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).GetWristbandFirmware(ctx, deviceID)
	}
	return (&SqliteStore{db: a.db}).GetWristbandFirmware(ctx, deviceID)
}

func (a *StoreAdapter) WriteToWristband(ctx context.Context, deviceID, data string) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).WriteToWristband(ctx, deviceID, data)
	}
	return (&SqliteStore{db: a.db}).WriteToWristband(ctx, deviceID, data)
}

func (a *StoreAdapter) CreateExpense(ctx context.Context, e *model.MedicalExpense) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).CreateExpense(ctx, e)
	}
	return (&SqliteStore{db: a.db}).CreateExpense(ctx, e)
}

func (a *StoreAdapter) ListExpenses(ctx context.Context, patientID string, page, pageSize int) ([]model.MedicalExpense, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).ListExpenses(ctx, patientID, page, pageSize)
	}
	return (&SqliteStore{db: a.db}).ListExpenses(ctx, patientID, page, pageSize)
}

func (a *StoreAdapter) CreateMedication(ctx context.Context, m *model.MedicalMedication) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).CreateMedication(ctx, m)
	}
	return (&SqliteStore{db: a.db}).CreateMedication(ctx, m)
}

func (a *StoreAdapter) ListMedications(ctx context.Context, patientID string) ([]model.MedicalMedication, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).ListMedications(ctx, patientID)
	}
	return (&SqliteStore{db: a.db}).ListMedications(ctx, patientID)
}

func (a *StoreAdapter) CreateTestResult(ctx context.Context, r *model.MedicalTestResult) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).CreateTestResult(ctx, r)
	}
	return (&SqliteStore{db: a.db}).CreateTestResult(ctx, r)
}

func (a *StoreAdapter) ListTestResults(ctx context.Context, patientID string) ([]model.MedicalTestResult, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).ListTestResults(ctx, patientID)
	}
	return (&SqliteStore{db: a.db}).ListTestResults(ctx, patientID)
}

func (a *StoreAdapter) CreateDailyEntry(ctx context.Context, e *model.MedicalDailyEntry) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).CreateDailyEntry(ctx, e)
	}
	return (&SqliteStore{db: a.db}).CreateDailyEntry(ctx, e)
}

func (a *StoreAdapter) ListDailyEntries(ctx context.Context, patientID string, date string) ([]model.MedicalDailyEntry, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).ListDailyEntries(ctx, patientID, date)
	}
	return (&SqliteStore{db: a.db}).ListDailyEntries(ctx, patientID, date)
}

func (a *StoreAdapter) CreateVerification(ctx context.Context, v *model.MedicalVerification) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).CreateVerification(ctx, v)
	}
	return (&SqliteStore{db: a.db}).CreateVerification(ctx, v)
}

func (a *StoreAdapter) ListVerifications(ctx context.Context, page, pageSize int) ([]model.MedicalVerification, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).ListVerifications(ctx, page, pageSize)
	}
	return (&SqliteStore{db: a.db}).ListVerifications(ctx, page, pageSize)
}

func (a *StoreAdapter) UpdateVerificationStatus(ctx context.Context, id, status string) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).UpdateVerificationStatus(ctx, id, status)
	}
	return (&SqliteStore{db: a.db}).UpdateVerificationStatus(ctx, id, status)
}

func (a *StoreAdapter) GetTodayVerificationStats(ctx context.Context) (*model.MedicalVerificationStats, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).GetTodayVerificationStats(ctx)
	}
	return (&SqliteStore{db: a.db}).GetTodayVerificationStats(ctx)
}

func (a *StoreAdapter) GetMedicalStatsOverview(ctx context.Context) (*model.MedicalStatsOverview, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).GetMedicalStatsOverview(ctx)
	}
	return (&SqliteStore{db: a.db}).GetMedicalStatsOverview(ctx)
}

func (a *StoreAdapter) GetPatientByAdmissionNo(ctx context.Context, admissionNo string) (*model.MedicalPatient, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).GetPatientByAdmissionNo(ctx, admissionNo)
	}
	return (&SqliteStore{db: a.db}).GetPatientByAdmissionNo(ctx, admissionNo)
}

func (a *StoreAdapter) BatchImportPatients(ctx context.Context, patients []model.MedicalPatient) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).BatchImportPatients(ctx, patients)
	}
	return (&SqliteStore{db: a.db}).BatchImportPatients(ctx, patients)
}

func (a *StoreAdapter) GetPatientHistory(ctx context.Context, patientID string) (*model.MedicalPatientHistory, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).GetPatientHistory(ctx, patientID)
	}
	return (&SqliteStore{db: a.db}).GetPatientHistory(ctx, patientID)
}

func (a *StoreAdapter) CreateAlertTagConfig(ctx context.Context, c *model.MedicalAlertTagConfig) error {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).CreateAlertTagConfig(ctx, c)
	}
	return (&SqliteStore{db: a.db}).CreateAlertTagConfig(ctx, c)
}

func (a *StoreAdapter) ListAlertTagConfigs(ctx context.Context) ([]model.MedicalAlertTagConfig, error) {
	if a.dbType == "postgres" {
		return (&PostgresStore{db: a.db}).ListAlertTagConfigs(ctx)
	}
	return (&SqliteStore{db: a.db}).ListAlertTagConfigs(ctx)
}

// Regulatory dispatch
func (a *StoreAdapter) CreateFenceConfig(ctx context.Context, fc *model.RegulatoryFenceConfig) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).CreateFenceConfig(ctx, fc) }
	return (&SqliteStore{db: a.db}).CreateFenceConfig(ctx, fc)
}
func (a *StoreAdapter) GetFenceConfig(ctx context.Context, hospitalID string) (*model.RegulatoryFenceConfig, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).GetFenceConfig(ctx, hospitalID) }
	return (&SqliteStore{db: a.db}).GetFenceConfig(ctx, hospitalID)
}
func (a *StoreAdapter) UpdateFenceConfig(ctx context.Context, fc *model.RegulatoryFenceConfig) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).UpdateFenceConfig(ctx, fc) }
	return (&SqliteStore{db: a.db}).UpdateFenceConfig(ctx, fc)
}
func (a *StoreAdapter) ListRegulatoryAlerts(ctx context.Context, ruleCode, level, status, department string, page, pageSize int) ([]model.RegulatoryAlert, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).ListRegulatoryAlerts(ctx, ruleCode, level, status, department, page, pageSize) }
	return (&SqliteStore{db: a.db}).ListRegulatoryAlerts(ctx, ruleCode, level, status, department, page, pageSize)
}
func (a *StoreAdapter) GetRegulatoryAlert(ctx context.Context, alertID string) (*model.RegulatoryAlert, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).GetRegulatoryAlert(ctx, alertID) }
	return (&SqliteStore{db: a.db}).GetRegulatoryAlert(ctx, alertID)
}
func (a *StoreAdapter) AcknowledgeAlert(ctx context.Context, alertID, userID string) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).AcknowledgeAlert(ctx, alertID, userID) }
	return (&SqliteStore{db: a.db}).AcknowledgeAlert(ctx, alertID, userID)
}
func (a *StoreAdapter) ResolveRegulatoryAlert(ctx context.Context, alertID, userID, notes string) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).ResolveRegulatoryAlert(ctx, alertID, userID, notes) }
	return (&SqliteStore{db: a.db}).ResolveRegulatoryAlert(ctx, alertID, userID, notes)
}
func (a *StoreAdapter) ListRegulatoryAlertsCountByRule(ctx context.Context, days int) ([]model.RuleAlertCount, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).ListRegulatoryAlertsCountByRule(ctx, days) }
	return (&SqliteStore{db: a.db}).ListRegulatoryAlertsCountByRule(ctx, days)
}
func (a *StoreAdapter) SaveLocationLog(ctx context.Context, log *model.RegulatoryLocationLog) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).SaveLocationLog(ctx, log) }
	return (&SqliteStore{db: a.db}).SaveLocationLog(ctx, log)
}
func (a *StoreAdapter) ListLocationLogs(ctx context.Context, patientID string, limit int) ([]model.RegulatoryLocationLog, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).ListLocationLogs(ctx, patientID, limit) }
	return (&SqliteStore{db: a.db}).ListLocationLogs(ctx, patientID, limit)
}
func (a *StoreAdapter) GetPatientFenceStatus(ctx context.Context, patientID string) (string, time.Time, int, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).GetPatientFenceStatus(ctx, patientID) }
	return (&SqliteStore{db: a.db}).GetPatientFenceStatus(ctx, patientID)
}
func (a *StoreAdapter) GetRegulatoryOverview(ctx context.Context, department string) (*model.RegulatoryDashboardOverview, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).GetRegulatoryOverview(ctx, department) }
	return (&SqliteStore{db: a.db}).GetRegulatoryOverview(ctx, department)
}
func (a *StoreAdapter) ListRegulatoryPatients(ctx context.Context, department string, page, pageSize int) ([]model.RegulatoryPatientRow, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).ListRegulatoryPatients(ctx, department, page, pageSize) }
	return (&SqliteStore{db: a.db}).ListRegulatoryPatients(ctx, department, page, pageSize)
}
func (a *StoreAdapter) GetRegulatoryAuditTrail(ctx context.Context, patientID string) (*model.RegulatoryAuditTrail, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).GetRegulatoryAuditTrail(ctx, patientID) }
	return (&SqliteStore{db: a.db}).GetRegulatoryAuditTrail(ctx, patientID)
}
func (a *StoreAdapter) ListRuleConfigs(ctx context.Context) ([]model.RegulatoryRuleConfig, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).ListRuleConfigs(ctx) }
	return (&SqliteStore{db: a.db}).ListRuleConfigs(ctx)
}
func (a *StoreAdapter) UpdateRuleConfig(ctx context.Context, ruleCode string, configJSON string) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).UpdateRuleConfig(ctx, ruleCode, configJSON) }
	return (&SqliteStore{db: a.db}).UpdateRuleConfig(ctx, ruleCode, configJSON)
}
func (a *StoreAdapter) GetComplianceReport(ctx context.Context, hospitalID, startDate, endDate string) (*model.ComplianceReport, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).GetComplianceReport(ctx, hospitalID, startDate, endDate) }
	return (&SqliteStore{db: a.db}).GetComplianceReport(ctx, hospitalID, startDate, endDate)
}
func (a *StoreAdapter) CreateDepartmentBinding(ctx context.Context, binding *model.DepartmentBinding) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).CreateDepartmentBinding(ctx, binding) }
	return (&SqliteStore{db: a.db}).CreateDepartmentBinding(ctx, binding)
}
func (a *StoreAdapter) ListDepartmentBindings(ctx context.Context, userID string) ([]model.DepartmentBinding, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).ListDepartmentBindings(ctx, userID) }
	return (&SqliteStore{db: a.db}).ListDepartmentBindings(ctx, userID)
}
func (a *StoreAdapter) CreateRegulatoryAlert(ctx context.Context, alert *model.RegulatoryAlert) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).CreateRegulatoryAlert(ctx, alert) }
	return (&SqliteStore{db: a.db}).CreateRegulatoryAlert(ctx, alert)
}
func (a *StoreAdapter) CountPendingAlertsByRule(ctx context.Context) ([]model.RuleAlertCount, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).CountPendingAlertsByRule(ctx) }
	return (&SqliteStore{db: a.db}).CountPendingAlertsByRule(ctx)
}
func (a *StoreAdapter) CountAlertsByDept(ctx context.Context, startDate, endDate string) ([]model.DeptAlertCount, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).CountAlertsByDept(ctx, startDate, endDate) }
	return (&SqliteStore{db: a.db}).CountAlertsByDept(ctx, startDate, endDate)
}

// Community dispatch
func (a *StoreAdapter) CreateCommunityElder(ctx context.Context, e *model.CommunityElder) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).CreateCommunityElder(ctx, e) }
	return (&SqliteStore{db: a.db}).CreateCommunityElder(ctx, e)
}
func (a *StoreAdapter) GetCommunityElder(ctx context.Context, id string) (*model.CommunityElder, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).GetCommunityElder(ctx, id) }
	return (&SqliteStore{db: a.db}).GetCommunityElder(ctx, id)
}
func (a *StoreAdapter) ListCommunityElders(ctx context.Context, page, pageSize int, status string) ([]model.CommunityElder, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).ListCommunityElders(ctx, page, pageSize, status) }
	return (&SqliteStore{db: a.db}).ListCommunityElders(ctx, page, pageSize, status)
}
func (a *StoreAdapter) UpdateCommunityElder(ctx context.Context, e *model.CommunityElder) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).UpdateCommunityElder(ctx, e) }
	return (&SqliteStore{db: a.db}).UpdateCommunityElder(ctx, e)
}
func (a *StoreAdapter) DeleteCommunityElder(ctx context.Context, id string) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).DeleteCommunityElder(ctx, id) }
	return (&SqliteStore{db: a.db}).DeleteCommunityElder(ctx, id)
}
func (a *StoreAdapter) BulkUpsertCommunityElders(ctx context.Context, elders []model.CommunityElder) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).BulkUpsertCommunityElders(ctx, elders) }
	return (&SqliteStore{db: a.db}).BulkUpsertCommunityElders(ctx, elders)
}
func (a *StoreAdapter) GetCommunityElderStats(ctx context.Context) (*model.CommunityElderStats, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).GetCommunityElderStats(ctx) }
	return (&SqliteStore{db: a.db}).GetCommunityElderStats(ctx)
}
func (a *StoreAdapter) CreateCommunityDevice(ctx context.Context, d *model.CommunityWristbandDevice) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).CreateCommunityDevice(ctx, d) }
	return (&SqliteStore{db: a.db}).CreateCommunityDevice(ctx, d)
}
func (a *StoreAdapter) GetCommunityDevice(ctx context.Context, deviceID string) (*model.CommunityWristbandDevice, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).GetCommunityDevice(ctx, deviceID) }
	return (&SqliteStore{db: a.db}).GetCommunityDevice(ctx, deviceID)
}
func (a *StoreAdapter) ListCommunityDevices(ctx context.Context, page, pageSize int, status string) ([]model.CommunityWristbandDevice, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).ListCommunityDevices(ctx, page, pageSize, status) }
	return (&SqliteStore{db: a.db}).ListCommunityDevices(ctx, page, pageSize, status)
}
func (a *StoreAdapter) UpdateCommunityDevice(ctx context.Context, d *model.CommunityWristbandDevice) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).UpdateCommunityDevice(ctx, d) }
	return (&SqliteStore{db: a.db}).UpdateCommunityDevice(ctx, d)
}
func (a *StoreAdapter) BindCommunityElderDevice(ctx context.Context, elderID, deviceID string) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).BindCommunityElderDevice(ctx, elderID, deviceID) }
	return (&SqliteStore{db: a.db}).BindCommunityElderDevice(ctx, elderID, deviceID)
}
func (a *StoreAdapter) UnbindCommunityElderDevice(ctx context.Context, bindingID string) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).UnbindCommunityElderDevice(ctx, bindingID) }
	return (&SqliteStore{db: a.db}).UnbindCommunityElderDevice(ctx, bindingID)
}
func (a *StoreAdapter) CreateWelfareTagConfig(ctx context.Context, c *model.CommunityWelfareTagConfig) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).CreateWelfareTagConfig(ctx, c) }
	return (&SqliteStore{db: a.db}).CreateWelfareTagConfig(ctx, c)
}
func (a *StoreAdapter) UpdateWelfareTagConfig(ctx context.Context, c *model.CommunityWelfareTagConfig) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).UpdateWelfareTagConfig(ctx, c) }
	return (&SqliteStore{db: a.db}).UpdateWelfareTagConfig(ctx, c)
}
func (a *StoreAdapter) ListWelfareTagConfigs(ctx context.Context) ([]model.CommunityWelfareTagConfig, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).ListWelfareTagConfigs(ctx) }
	return (&SqliteStore{db: a.db}).ListWelfareTagConfigs(ctx)
}
func (a *StoreAdapter) GetWelfareTagConfig(ctx context.Context, tagCode string) (*model.CommunityWelfareTagConfig, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).GetWelfareTagConfig(ctx, tagCode) }
	return (&SqliteStore{db: a.db}).GetWelfareTagConfig(ctx, tagCode)
}
func (a *StoreAdapter) AssignWelfareTag(ctx context.Context, welfare *model.CommunityElderWelfare) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).AssignWelfareTag(ctx, welfare) }
	return (&SqliteStore{db: a.db}).AssignWelfareTag(ctx, welfare)
}
func (a *StoreAdapter) RevokeWelfareTag(ctx context.Context, elderID, tagCode string) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).RevokeWelfareTag(ctx, elderID, tagCode) }
	return (&SqliteStore{db: a.db}).RevokeWelfareTag(ctx, elderID, tagCode)
}
func (a *StoreAdapter) ListElderWelfareTags(ctx context.Context, elderID string) ([]model.CommunityElderWelfare, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).ListElderWelfareTags(ctx, elderID) }
	return (&SqliteStore{db: a.db}).ListElderWelfareTags(ctx, elderID)
}
func (a *StoreAdapter) CreateSigninRecord(ctx context.Context, sRec *model.CommunitySigninRecord) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).CreateSigninRecord(ctx, sRec) }
	return (&SqliteStore{db: a.db}).CreateSigninRecord(ctx, sRec)
}
func (a *StoreAdapter) ListSigninRecords(ctx context.Context, elderID, period, hospitalID string, page, pageSize int) ([]model.CommunitySigninRecord, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).ListSigninRecords(ctx, elderID, period, hospitalID, page, pageSize) }
	return (&SqliteStore{db: a.db}).ListSigninRecords(ctx, elderID, period, hospitalID, page, pageSize)
}
func (a *StoreAdapter) GetSigninSummary(ctx context.Context, elderID, period string) (*model.CommunitySigninRecord, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).GetSigninSummary(ctx, elderID, period) }
	return (&SqliteStore{db: a.db}).GetSigninSummary(ctx, elderID, period)
}
func (a *StoreAdapter) CreatePharmacyLog(ctx context.Context, p *model.CommunityPharmacyLog) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).CreatePharmacyLog(ctx, p) }
	return (&SqliteStore{db: a.db}).CreatePharmacyLog(ctx, p)
}
func (a *StoreAdapter) ListPharmacyLogs(ctx context.Context, elderID, period string, page, pageSize int) ([]model.CommunityPharmacyLog, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).ListPharmacyLogs(ctx, elderID, period, page, pageSize) }
	return (&SqliteStore{db: a.db}).ListPharmacyLogs(ctx, elderID, period, page, pageSize)
}
func (a *StoreAdapter) CreateMinzhengSync(ctx context.Context, m *model.CommunityMinzhengSync) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).CreateMinzhengSync(ctx, m) }
	return (&SqliteStore{db: a.db}).CreateMinzhengSync(ctx, m)
}
func (a *StoreAdapter) ListMinzhengSync(ctx context.Context, page, pageSize int) ([]model.CommunityMinzhengSync, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).ListMinzhengSync(ctx, page, pageSize) }
	return (&SqliteStore{db: a.db}).ListMinzhengSync(ctx, page, pageSize)
}
func (a *StoreAdapter) GetLatestMinzhengSync(ctx context.Context) (*model.CommunityMinzhengSync, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).GetLatestMinzhengSync(ctx) }
	return (&SqliteStore{db: a.db}).GetLatestMinzhengSync(ctx)
}
func (a *StoreAdapter) CreateBatchPayment(ctx context.Context, p *model.CommunityBatchPayment) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).CreateBatchPayment(ctx, p) }
	return (&SqliteStore{db: a.db}).CreateBatchPayment(ctx, p)
}
func (a *StoreAdapter) BulkCreateBatchPayments(ctx context.Context, payments []model.CommunityBatchPayment) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).BulkCreateBatchPayments(ctx, payments) }
	return (&SqliteStore{db: a.db}).BulkCreateBatchPayments(ctx, payments)
}
func (a *StoreAdapter) ListBatchPayments(ctx context.Context, batchID string, page, pageSize int) ([]model.CommunityBatchPayment, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).ListBatchPayments(ctx, batchID, page, pageSize) }
	return (&SqliteStore{db: a.db}).ListBatchPayments(ctx, batchID, page, pageSize)
}
func (a *StoreAdapter) UpdateBatchPaymentStatus(ctx context.Context, id, status string, failureReason string) error {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).UpdateBatchPaymentStatus(ctx, id, status, failureReason) }
	return (&SqliteStore{db: a.db}).UpdateBatchPaymentStatus(ctx, id, status, failureReason)
}
func (a *StoreAdapter) CountPendingPayments(ctx context.Context) (int64, error) {
	if a.dbType == "postgres" { return (&PostgresStore{db: a.db}).CountPendingPayments(ctx) }
	return (&SqliteStore{db: a.db}).CountPendingPayments(ctx)
}

var _ Store = (*StoreAdapter)(nil)
