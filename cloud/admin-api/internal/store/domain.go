// Package store defines domain-driven interfaces for the admin-api.
// These interfaces align with api-server domain names for consistency
// across the platform while maintaining admin-specific functionality.
package store

import (
	"context"
	"time"

	"eregen.dev/admin-api/internal/model"
)

// UserDomain encapsulates user and elderly profile operations (aligned with api-server).
type UserDomain interface {
	ListUsers(ctx context.Context, page, pageSize int, role string) ([]model.UserSummary, error)
	CreateUser(ctx context.Context, name, email, phone, role, passwordHash string) (string, error)
	UpdateUser(ctx context.Context, id, name, email, phone, role string) error
	DeleteUser(ctx context.Context, id string) error
	SetUserRole(ctx context.Context, userID, role string) error
	ListElderly(ctx context.Context, page, pageSize int) ([]model.ElderlyProfile, error)
	GetElderly(ctx context.Context, id string) (*model.ElderlyProfile, error)
	CreateElderly(ctx context.Context, name, birthDate, userID string, healthTiers []string, avatarURL string) (*model.ElderlyProfile, error)
	UpdateElderly(ctx context.Context, id, name, birthDate, userID string, healthTiers []string, avatarURL string) (*model.ElderlyProfile, error)
	DeleteElderly(ctx context.Context, id string) error
	GetElderlyHealthStats(ctx context.Context, elderlyID string) (*model.HealthStats, error)
}

// DeviceDomain encapsulates device management operations (aligned with api-server).
type DeviceDomain interface {
	ListDevices(ctx context.Context, page, pageSize int, status, devType, tier string) ([]model.DeviceSummary, error)
	GetDeviceByID(ctx context.Context, id string) (*model.DeviceDetail, error)
	UpdateDeviceConfig(ctx context.Context, deviceID string, config map[string]interface{}) error
	TriggerOTA(ctx context.Context, deviceID, firmwareURL, sha256Hash string) error
	UnbindDevice(ctx context.Context, deviceID string) error
	BatchTriggerOTA(ctx context.Context, deviceIDs, firmwareURL, sha256Hash []string) error
}

// AlertDomain encapsulates alert operations (aligned with api-server).
type AlertDomain interface {
	ListAlerts(ctx context.Context, severity, status string, limit int) ([]model.AlertSummary, error)
	CreateAlert(ctx context.Context, alert *model.AlertSummary) error
	ResolveAlert(ctx context.Context, alertID string) error
	UpdateAlertStatus(ctx context.Context, alertID, status string) error
}

// SessionDomain encapsulates authentication and session operations (aligned with api-server).
type SessionDomain interface {
	GetUserByCredential(ctx context.Context, method, credential, secret string) (*model.UserLogin, error)
}

// AdminDomain encapsulates admin-only operations (stats, firmware, settings).
type AdminDomain interface {
	GetDashboardStats(ctx context.Context) (*model.DashboardStats, error)
	GetSubscriptionStats(ctx context.Context) ([]model.SubscriptionStat, error)
	GetAlertTrend(ctx context.Context, days int) ([]model.AlertTrendPoint, error)
	GetAlertDistribution(ctx context.Context) ([]model.AlertDistributionItem, error)
	GetUserGrowth(ctx context.Context, months int) ([]model.UserGrowthPoint, error)
	CreateFirmwareVersion(ctx context.Context, v *model.FirmwareVersion) error
	ListFirmwareVersions(ctx context.Context) ([]model.FirmwareVersion, error)
	DeleteFirmwareVersion(ctx context.Context, id string) error
	PushOTAJob(ctx context.Context, firmwareID string, deviceIDs []string) error
	GetNotificationSettings(ctx context.Context) (map[string]any, error)
	UpdateNotificationSettings(ctx context.Context, data map[string]any) error
	ListAPIKeys(ctx context.Context) ([]model.APIKeySummary, error)
	CreateAPIKey(ctx context.Context, name, keyHash string, expiresAt *time.Time) (string, error)
	RevokeAPIKey(ctx context.Context, id string) error
}

// ClinicalDomain encapsulates clinical/medical operations.
type ClinicalDomain interface {
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
	CreateAlertTagConfig(ctx context.Context, c *model.MedicalAlertTagConfig) error
	ListAlertTagConfigs(ctx context.Context) ([]model.MedicalAlertTagConfig, error)
}

// PatientDomain encapsulates patient management operations.
type PatientDomain interface {
	CreatePatient(ctx context.Context, p *model.MedicalPatient) error
	GetPatient(ctx context.Context, id string) (*model.MedicalPatient, error)
	ListPatients(ctx context.Context, page, pageSize int, status string) ([]model.MedicalPatient, error)
	UpdatePatient(ctx context.Context, p *model.MedicalPatient) error
	DeletePatient(ctx context.Context, id string) error
	GetPatientByAdmissionNo(ctx context.Context, admissionNo string) (*model.MedicalPatient, error)
	BatchImportPatients(ctx context.Context, patients []model.MedicalPatient) error
	GetPatientHistory(ctx context.Context, patientID string) (*model.MedicalPatientHistory, error)
}

// CommunityDomain encapsulates community wristband operations.
type CommunityDomain interface {
	CreateCommunityElder(ctx context.Context, e *model.CommunityElder) error
	GetCommunityElder(ctx context.Context, id string) (*model.CommunityElder, error)
	ListCommunityElders(ctx context.Context, page, pageSize int, status string) ([]model.CommunityElder, error)
	UpdateCommunityElder(ctx context.Context, e *model.CommunityElder) error
	DeleteCommunityElder(ctx context.Context, id string) error
	BulkUpsertCommunityElders(ctx context.Context, elders []model.CommunityElder) error
	GetCommunityElderStats(ctx context.Context) (*model.CommunityElderStats, error)
	CreateCommunityDevice(ctx context.Context, d *model.CommunityWristbandDevice) error
	GetCommunityDevice(ctx context.Context, deviceID string) (*model.CommunityWristbandDevice, error)
	ListCommunityDevices(ctx context.Context, page, pageSize int, status string) ([]model.CommunityWristbandDevice, error)
	UpdateCommunityDevice(ctx context.Context, d *model.CommunityWristbandDevice) error
	BindCommunityElderDevice(ctx context.Context, elderID, deviceID string) error
	UnbindCommunityElderDevice(ctx context.Context, bindingID string) error
	CreateWelfareTagConfig(ctx context.Context, c *model.CommunityWelfareTagConfig) error
	UpdateWelfareTagConfig(ctx context.Context, c *model.CommunityWelfareTagConfig) error
	ListWelfareTagConfigs(ctx context.Context) ([]model.CommunityWelfareTagConfig, error)
	GetWelfareTagConfig(ctx context.Context, tagCode string) (*model.CommunityWelfareTagConfig, error)
	AssignWelfareTag(ctx context.Context, a *model.CommunityElderWelfare) error
	RevokeWelfareTag(ctx context.Context, elderID, tagCode string) error
	ListElderWelfareTags(ctx context.Context, elderID string) ([]model.CommunityElderWelfare, error)
	CreateSigninRecord(ctx context.Context, s *model.CommunitySigninRecord) error
	ListSigninRecords(ctx context.Context, elderID, period, hospitalID string, page, pageSize int) ([]model.CommunitySigninRecord, error)
	GetSigninSummary(ctx context.Context, elderlyID, period string) (*model.CommunitySigninRecord, error)
	CreatePharmacyLog(ctx context.Context, p *model.CommunityPharmacyLog) error
	ListPharmacyLogs(ctx context.Context, elderlyID, period string, page, pageSize int) ([]model.CommunityPharmacyLog, error)
}

// RegulatoryDomain encapsulates regulatory compliance operations.
type RegulatoryDomain interface {
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
	CreateRegulatoryAlert(ctx context.Context, alert *model.RegulatoryAlert) error
	CountPendingAlertsByRule(ctx context.Context) ([]model.RuleAlertCount, error)
	CountAlertsByDept(ctx context.Context, startDate, endDate string) ([]model.DeptAlertCount, error)
}

// InstitutionDomain encapsulates B2B institution management.
type InstitutionDomain interface {
	ListInstitutions(ctx context.Context, page, pageSize int, name, typ, status string) ([]model.InstitutionSummary, error)
	GetInstitution(ctx context.Context, id string) (*model.InstitutionSummary, error)
	CreateInstitution(ctx context.Context, i *model.InstitutionSummary) error
	UpdateInstitution(ctx context.Context, id string, updates map[string]any) error
	DeleteInstitution(ctx context.Context, id string) error
	CreateInstitutionAPIKey(ctx context.Context, institutionID, name string) (string, error)
	RevokeInstitutionAPIKey(ctx context.Context, institutionID, keyID string) error
}

// SubscriptionDomain encapsulates subscription management.
type SubscriptionDomain interface {
	ListSubscriptions(ctx context.Context, page, pageSize int, status, planTier string) ([]model.SubscriptionItem, error)
	GetSubscription(ctx context.Context, id string) (*model.SubscriptionItem, error)
	CreateSubscription(ctx context.Context, s *model.SubscriptionItem) error
	UpdateSubscription(ctx context.Context, id string, updates map[string]any) error
	RenewSubscription(ctx context.Context, id string, endDate string) error
}

// AuditDomain encapsulates audit logging.
type AuditDomain interface {
	CreateAuditLog(ctx context.Context, log *model.AuditLog) error
	ListAuditLogs(ctx context.Context, limit int) ([]model.AuditLog, error)
	ListAuditLogsByUser(ctx context.Context, userID string, limit int) ([]model.AuditLog, error)
	ListAuditLogsByAction(ctx context.Context, action string, limit int) ([]model.AuditLog, error)
}

// WristbandDomain encapsulates medical wristband device management.
type WristbandDomain interface {
	BindWristband(ctx context.Context, patientID, deviceID string) error
	UnbindWristband(ctx context.Context, bindingID string) error
	ClearWristband(ctx context.Context, deviceID string) error
	ListWristbands(ctx context.Context, page, pageSize int, status string) ([]model.MedicalWristbandDevice, error)
	GetWristbandFirmware(ctx context.Context, deviceID string) (string, error)
	WriteToWristband(ctx context.Context, deviceID, data string) error
}

// AdmissionDomain encapsulates hospital admission management.
type AdmissionDomain interface {
	CreateAdmission(ctx context.Context, a *model.HospitalAdmission) error
	GetAdmission(ctx context.Context, id string) (*model.HospitalAdmission, error)
	ListAdmissions(ctx context.Context, page, pageSize int, department, status string) ([]model.HospitalAdmission, error)
	CompleteAdmission(ctx context.Context, id, dischargeType, notes, transferredTo string) error
	CreateWardRound(ctx context.Context, w *model.WardRoundEntry) error
	ListWardRounds(ctx context.Context, patientID string) ([]model.WardRoundEntry, error)
}
