// Package store defines the data access interface for admin operations.
package store

import (
	"context"
	"time"

	"eregen.dev/admin-api/internal/model"
)

// Store is the interface that both PostgresStore and SqliteStore implement.
type Store interface {
	GetDashboardStats(ctx context.Context) (*model.DashboardStats, error)
	ListDevices(ctx context.Context, page, pageSize int, status, devType, tier string) ([]model.DeviceSummary, error)
	ListUsers(ctx context.Context, page, pageSize int, role string) ([]model.UserSummary, error)
	ListAlerts(ctx context.Context, severity, status string, limit int) ([]model.AlertSummary, error)
	SetUserRole(ctx context.Context, userID, role string) error
	UpdateDeviceConfig(ctx context.Context, deviceID string, config map[string]interface{}) error
	TriggerOTA(ctx context.Context, deviceID, firmwareURL, sha256Hash string) error
	ResolveAlert(ctx context.Context, alertID string) error
	GetSubscriptionStats(ctx context.Context) ([]model.SubscriptionStat, error)
	GetAlertTrend(ctx context.Context, days int) ([]model.AlertTrendPoint, error)
	GetAlertDistribution(ctx context.Context) ([]model.AlertDistributionItem, error)
	GetUserGrowth(ctx context.Context, months int) ([]model.UserGrowthPoint, error)
	GetDeviceByID(ctx context.Context, id string) (*model.DeviceDetail, error)
	UnbindDevice(ctx context.Context, deviceID string) error
	BatchTriggerOTA(ctx context.Context, deviceIDs, firmwareURL, sha256Hash []string) error
	CreateFirmwareVersion(ctx context.Context, v *model.FirmwareVersion) error
	ListFirmwareVersions(ctx context.Context) ([]model.FirmwareVersion, error)
	DeleteFirmwareVersion(ctx context.Context, id string) error
	PushOTAJob(ctx context.Context, firmwareID string, deviceIDs []string) error
	GetNotificationSettings(ctx context.Context) (map[string]any, error)
	UpdateNotificationSettings(ctx context.Context, data map[string]any) error
	ListAPIKeys(ctx context.Context) ([]model.APIKeySummary, error)
	CreateAPIKey(ctx context.Context, name, keyHash string, expiresAt *time.Time) (string, error)
	RevokeAPIKey(ctx context.Context, id string) error
	ChangeAdminPassword(ctx context.Context, userID, hash string) error

	// Elderly profile management
	ListElderly(ctx context.Context, page, pageSize int) ([]model.ElderlyProfile, error)
	GetElderly(ctx context.Context, id string) (*model.ElderlyProfile, error)
	CreateElderly(ctx context.Context, name, birthDate, userID string, healthTiers []string, avatarURL string) (*model.ElderlyProfile, error)
	UpdateElderly(ctx context.Context, id, name, birthDate, userID string, healthTiers []string, avatarURL string) (*model.ElderlyProfile, error)
	DeleteElderly(ctx context.Context, id string) error
	GetElderlyHealthStats(ctx context.Context, elderlyID string) (*model.HealthStats, error)
	GetElderlyHealthRecords(ctx context.Context, elderlyID string, limit int) ([]model.HealthRecordRow, error)
	GetElderlyMedicationRules(ctx context.Context, elderlyID string) ([]model.MedicationRuleRow, error)
	GetElderlyDevices(ctx context.Context, elderlyID string) ([]model.DeviceSummaryRow, error)
	GetElderlyLocationHistory(ctx context.Context, elderlyID string, limit int) ([]model.LocationPoint, error)
	GetElderlyAlertHistory(ctx context.Context, elderlyID string, limit int) ([]model.AlertSummaryRow, error)

	// Medical wristband management
	CreatePatient(ctx context.Context, p *model.MedicalPatient) error
	GetPatient(ctx context.Context, id string) (*model.MedicalPatient, error)
	ListPatients(ctx context.Context, page, pageSize int, status string) ([]model.MedicalPatient, error)
	UpdatePatient(ctx context.Context, p *model.MedicalPatient) error
	DeletePatient(ctx context.Context, id string) error
	BindWristband(ctx context.Context, patientID, deviceID string) error
	UnbindWristband(ctx context.Context, bindingID string) error
	ClearWristband(ctx context.Context, deviceID string) error
	ListWristbands(ctx context.Context, page, pageSize int, status string) ([]model.MedicalWristbandDevice, error)
	GetWristbandFirmware(ctx context.Context, deviceID string) (string, error)
	WriteToWristband(ctx context.Context, deviceID, data string) error
	CreateExpense(ctx context.Context, e *model.MedicalExpense) error
	ListExpenses(ctx context.Context, patientID string, page, pageSize int) ([]model.MedicalExpense, error)
	CreateMedication(ctx context.Context, m *model.MedicalMedication) error
	ListMedications(ctx context.Context, patientID string) ([]model.MedicalMedication, error)
	CreateTestResult(ctx context.Context, r *model.MedicalTestResult) error
	ListTestResults(ctx context.Context, patientID string) ([]model.MedicalTestResult, error)
	CreateDailyEntry(ctx context.Context, e *model.MedicalDailyEntry) error
	ListDailyEntries(ctx context.Context, patientID string, date string) ([]model.MedicalDailyEntry, error)
	CreateVerification(ctx context.Context, v *model.MedicalVerification) error
	ListVerifications(ctx context.Context, page, pageSize int) ([]model.MedicalVerification, error)
	UpdateVerificationStatus(ctx context.Context, id, status string) error
	GetTodayVerificationStats(ctx context.Context) (*model.MedicalVerificationStats, error)
	GetMedicalStatsOverview(ctx context.Context) (*model.MedicalStatsOverview, error)
	GetPatientByAdmissionNo(ctx context.Context, admissionNo string) (*model.MedicalPatient, error)
	BatchImportPatients(ctx context.Context, patients []model.MedicalPatient) error
	GetPatientHistory(ctx context.Context, patientID string) (*model.MedicalPatientHistory, error)
	CreateAlertTagConfig(ctx context.Context, c *model.MedicalAlertTagConfig) error
	ListAlertTagConfigs(ctx context.Context) ([]model.MedicalAlertTagConfig, error)

	// ===== Regulatory closure =====
	CreateFenceConfig(ctx context.Context, fc *model.RegulatoryFenceConfig) error
	GetFenceConfig(ctx context.Context, hospitalID string) (*model.RegulatoryFenceConfig, error)
	UpdateFenceConfig(ctx context.Context, fc *model.RegulatoryFenceConfig) error
	ListRegulatoryAlerts(ctx context.Context, ruleCode, level, status, department string, page, pageSize int) ([]model.RegulatoryAlert, error)
	GetRegulatoryAlert(ctx context.Context, alertID string) (*model.RegulatoryAlert, error)
	AcknowledgeAlert(ctx context.Context, alertID, userID string) error
	ResolveRegulatoryAlert(ctx context.Context, alertID, userID, notes string) error
	ListRegulatoryAlertsCountByRule(ctx context.Context, days int) ([]model.RuleAlertCount, error)
	SaveLocationLog(ctx context.Context, log *model.RegulatoryLocationLog) error
	ListLocationLogs(ctx context.Context, patientID string, limit int) ([]model.RegulatoryLocationLog, error)
	GetPatientFenceStatus(ctx context.Context, patientID string) (string, time.Time, int, error)
	GetRegulatoryOverview(ctx context.Context, department string) (*model.RegulatoryDashboardOverview, error)
	ListRegulatoryPatients(ctx context.Context, department string, page, pageSize int) ([]model.RegulatoryPatientRow, error)
	GetRegulatoryAuditTrail(ctx context.Context, patientID string) (*model.RegulatoryAuditTrail, error)
	ListRuleConfigs(ctx context.Context) ([]model.RegulatoryRuleConfig, error)
	UpdateRuleConfig(ctx context.Context, ruleCode string, configJSON string) error
	GetComplianceReport(ctx context.Context, hospitalID, startDate, endDate string) (*model.ComplianceReport, error)
	CreateDepartmentBinding(ctx context.Context, binding *model.DepartmentBinding) error
	ListDepartmentBindings(ctx context.Context, userID string) ([]model.DepartmentBinding, error)
	CreateRegulatoryAlert(ctx context.Context, a *model.RegulatoryAlert) error
	CountPendingAlertsByRule(ctx context.Context) ([]model.RuleAlertCount, error)
	CountAlertsByDept(ctx context.Context, startDate, endDate string) ([]model.DeptAlertCount, error)

	// ===== Community elderly wristband =====
	// Elder profiles
	CreateCommunityElder(ctx context.Context, e *model.CommunityElder) error
	GetCommunityElder(ctx context.Context, id string) (*model.CommunityElder, error)
	ListCommunityElders(ctx context.Context, page, pageSize int, status string) ([]model.CommunityElder, error)
	UpdateCommunityElder(ctx context.Context, e *model.CommunityElder) error
	DeleteCommunityElder(ctx context.Context, id string) error
	BulkUpsertCommunityElders(ctx context.Context, elders []model.CommunityElder) error
	GetCommunityElderStats(ctx context.Context) (*model.CommunityElderStats, error)
	// Device management
	CreateCommunityDevice(ctx context.Context, d *model.CommunityWristbandDevice) error
	GetCommunityDevice(ctx context.Context, deviceID string) (*model.CommunityWristbandDevice, error)
	ListCommunityDevices(ctx context.Context, page, pageSize int, status string) ([]model.CommunityWristbandDevice, error)
	UpdateCommunityDevice(ctx context.Context, d *model.CommunityWristbandDevice) error
	// Bindings
	BindCommunityElderDevice(ctx context.Context, elderID, deviceID string) error
	UnbindCommunityElderDevice(ctx context.Context, bindingID string) error
	// Welfare tags
	CreateWelfareTagConfig(ctx context.Context, c *model.CommunityWelfareTagConfig) error
	UpdateWelfareTagConfig(ctx context.Context, c *model.CommunityWelfareTagConfig) error
	ListWelfareTagConfigs(ctx context.Context) ([]model.CommunityWelfareTagConfig, error)
	GetWelfareTagConfig(ctx context.Context, tagCode string) (*model.CommunityWelfareTagConfig, error)
	AssignWelfareTag(ctx context.Context, a *model.CommunityElderWelfare) error
	RevokeWelfareTag(ctx context.Context, elderID, tagCode string) error
	ListElderWelfareTags(ctx context.Context, elderID string) ([]model.CommunityElderWelfare, error)
	// Sign-in
	CreateSigninRecord(ctx context.Context, s *model.CommunitySigninRecord) error
	ListSigninRecords(ctx context.Context, elderID, period, hospitalID string, page, pageSize int) ([]model.CommunitySigninRecord, error)
	GetSigninSummary(ctx context.Context, elderID, period string) (*model.CommunitySigninRecord, error)
	// Pharmacy
	CreatePharmacyLog(ctx context.Context, p *model.CommunityPharmacyLog) error
	ListPharmacyLogs(ctx context.Context, elderID, period string, page, pageSize int) ([]model.CommunityPharmacyLog, error)
	// Minzheng sync
	CreateMinzhengSync(ctx context.Context, m *model.CommunityMinzhengSync) error
	ListMinzhengSync(ctx context.Context, page, pageSize int) ([]model.CommunityMinzhengSync, error)
	GetLatestMinzhengSync(ctx context.Context) (*model.CommunityMinzhengSync, error)
	// Batch payments
	CreateBatchPayment(ctx context.Context, p *model.CommunityBatchPayment) error
	BulkCreateBatchPayments(ctx context.Context, payments []model.CommunityBatchPayment) error
	ListBatchPayments(ctx context.Context, batchID string, page, pageSize int) ([]model.CommunityBatchPayment, error)
	UpdateBatchPaymentStatus(ctx context.Context, id, status string, failureReason string) error
	CountPendingPayments(ctx context.Context) (int64, error)
}
