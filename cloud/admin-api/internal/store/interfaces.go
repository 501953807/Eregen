package store

import (
	"context"
	"time"

	"eregen.dev/admin-api/internal/model"
)

// Domain interfaces split from Store for precise dependency injection.

type DashboardStore interface {
	GetDashboardStats(ctx context.Context) (*model.DashboardStats, error)
	GetSubscriptionStats(ctx context.Context) ([]model.SubscriptionStat, error)
	GetAlertTrend(ctx context.Context, days int) ([]model.AlertTrendPoint, error)
	GetAlertDistribution(ctx context.Context) ([]model.AlertDistributionItem, error)
	GetUserGrowth(ctx context.Context, months int) ([]model.UserGrowthPoint, error)
}

type DeviceStore interface {
	ListDevices(ctx context.Context, page, pageSize int, status, devType, tier string) ([]model.DeviceSummary, error)
	GetDeviceByID(ctx context.Context, id string) (*model.DeviceDetail, error)
	UpdateDeviceConfig(ctx context.Context, deviceID string, config map[string]interface{}) error
	TriggerOTA(ctx context.Context, deviceID, firmwareURL, sha256Hash string) error
	UnbindDevice(ctx context.Context, deviceID string) error
	BatchTriggerOTA(ctx context.Context, deviceIDs, firmwareURL, sha256Hash []string) error
}

type UserStore interface {
	ListUsers(ctx context.Context, page, pageSize int, role string) ([]model.UserSummary, error)
	CreateUser(ctx context.Context, name, email, phone, role, passwordHash string) (string, error)
	UpdateUser(ctx context.Context, id, name, email, phone, role string) error
	DeleteUser(ctx context.Context, id string) error
	SetUserRole(ctx context.Context, userID, role string) error
}

type AuthStore interface {
	GetUserByCredential(ctx context.Context, method, credential, secret string) (*model.UserLogin, error)
}

type AlertStore interface {
	ListAlerts(ctx context.Context, severity, status string, limit int) ([]model.AlertSummary, error)
	CreateAlert(ctx context.Context, alert *model.AlertSummary) error
	ResolveAlert(ctx context.Context, alertID string) error
	UpdateAlertStatus(ctx context.Context, alertID, status string) error
}

type ElderlyStore interface {
	ListElderly(ctx context.Context, page, pageSize int) ([]model.ElderlyProfile, error)
	GetElderly(ctx context.Context, id string) (*model.ElderlyProfile, error)
	CreateElderly(ctx context.Context, name, birthDate, userID string, healthTiers []string, avatarURL string) (*model.ElderlyProfile, error)
	UpdateElderly(ctx context.Context, id, name, birthDate, userID string, healthTiers []string, avatarURL string) (*model.ElderlyProfile, error)
	DeleteElderly(ctx context.Context, id string) error
	GetElderlyHealthStats(ctx context.Context, elderlyID string) (*model.HealthStats, error)
	CreateHealthRecord(ctx context.Context, r *model.HealthRecordRow) error
	GetElderlyHealthRecords(ctx context.Context, elderlyID string, limit int) ([]model.HealthRecordRow, error)
	GetElderlyMedicationRules(ctx context.Context, elderlyID string) ([]model.MedicationRuleRow, error)
	CreateMedicationRule(ctx context.Context, elderlyID string, rule *model.MedicationRuleRow) error
	UpdateMedicationRule(ctx context.Context, elderlyID, ruleID string, updates map[string]interface{}) error
	DeleteMedicationRule(ctx context.Context, elderlyID, ruleID string) error
	GetElderlyDevices(ctx context.Context, elderlyID string) ([]model.DeviceSummaryRow, error)
	CreateLocation(ctx context.Context, loc *model.LocationPoint) error
	GetElderlyLocationHistory(ctx context.Context, elderlyID string, limit int) ([]model.LocationPoint, error)
	GetElderlyAlertHistory(ctx context.Context, elderlyID string, limit int) ([]model.AlertSummaryRow, error)
}

type FirmwareStore interface {
	ListFirmwareVersions(ctx context.Context) ([]model.FirmwareVersion, error)
	CreateFirmwareVersion(ctx context.Context, v *model.FirmwareVersion) error
	DeleteFirmwareVersion(ctx context.Context, id string) error
	PushOTAJob(ctx context.Context, firmwareID string, deviceIDs []string) error
}

type SettingsStore interface {
	GetNotificationSettings(ctx context.Context) (map[string]any, error)
	UpdateNotificationSettings(ctx context.Context, data map[string]any) error
	ListAPIKeys(ctx context.Context) ([]model.APIKeySummary, error)
	CreateAPIKey(ctx context.Context, name, keyHash string, expiresAt *time.Time) (string, error)
	RevokeAPIKey(ctx context.Context, id string) error
	ChangeAdminPassword(ctx context.Context, userID, hash string) error
}

type PatientStore interface {
	CreatePatient(ctx context.Context, p *model.MedicalPatient) error
	GetPatient(ctx context.Context, id string) (*model.MedicalPatient, error)
	ListPatients(ctx context.Context, page, pageSize int, status string) ([]model.MedicalPatient, error)
	UpdatePatient(ctx context.Context, p *model.MedicalPatient) error
	DeletePatient(ctx context.Context, id string) error
	GetPatientByAdmissionNo(ctx context.Context, admissionNo string) (*model.MedicalPatient, error)
	BatchImportPatients(ctx context.Context, patients []model.MedicalPatient) error
	GetPatientHistory(ctx context.Context, patientID string) (*model.MedicalPatientHistory, error)
}

type WristbandStore interface {
	BindWristband(ctx context.Context, patientID, deviceID string) error
	UnbindWristband(ctx context.Context, bindingID string) error
	ClearWristband(ctx context.Context, deviceID string) error
	ListWristbands(ctx context.Context, page, pageSize int, status string) ([]model.MedicalWristbandDevice, error)
	GetWristbandFirmware(ctx context.Context, deviceID string) (string, error)
	WriteToWristband(ctx context.Context, deviceID, data string) error
}

type ClinicalStore interface {
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

type AdmissionStore interface {
	CreateAdmission(ctx context.Context, a *model.HospitalAdmission) error
	GetAdmission(ctx context.Context, id string) (*model.HospitalAdmission, error)
	ListAdmissions(ctx context.Context, page, pageSize int, department, status string) ([]model.HospitalAdmission, error)
	CompleteAdmission(ctx context.Context, id, dischargeType, notes, transferredTo string) error
	CreateWardRound(ctx context.Context, w *model.WardRoundEntry) error
	ListWardRounds(ctx context.Context, patientID string) ([]model.WardRoundEntry, error)
	EvaluateRegulatoryRules(ctx context.Context, event string, data map[string]string) ([]*model.RegulatoryRuleResult, error)
}

type RegulatoryStore interface {
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

type CommunityWBStore interface {
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
	GetSigninSummary(ctx context.Context, elderID, period string) (*model.CommunitySigninRecord, error)
	CreatePharmacyLog(ctx context.Context, p *model.CommunityPharmacyLog) error
	ListPharmacyLogs(ctx context.Context, elderID, period string, page, pageSize int) ([]model.CommunityPharmacyLog, error)
	CreateMinzhengSync(ctx context.Context, m *model.CommunityMinzhengSync) error
	ListMinzhengSync(ctx context.Context, page, pageSize int) ([]model.CommunityMinzhengSync, error)
	GetLatestMinzhengSync(ctx context.Context) (*model.CommunityMinzhengSync, error)
	CreateBatchPayment(ctx context.Context, p *model.CommunityBatchPayment) error
	BulkCreateBatchPayments(ctx context.Context, payments []model.CommunityBatchPayment) error
	ListBatchPayments(ctx context.Context, batchID string, page, pageSize int) ([]model.CommunityBatchPayment, error)
	UpdateBatchPaymentStatus(ctx context.Context, id, status string, failureReason string) error
	CountPendingPayments(ctx context.Context) (int64, error)
}

// RuleEngineStore aggregates the subset of stores needed by the compliance rule engine.
type RuleEngineStore interface {
	RegulatoryStore
	CommunityWBStore
	AdmissionStore
}

type SubscriptionStore interface {
	ListSubscriptions(ctx context.Context, page, pageSize int, status, planTier string) ([]model.SubscriptionItem, error)
	GetSubscription(ctx context.Context, id string) (*model.SubscriptionItem, error)
	CreateSubscription(ctx context.Context, s *model.SubscriptionItem) error
	UpdateSubscription(ctx context.Context, id string, updates map[string]any) error
	RenewSubscription(ctx context.Context, id string, endDate string) error
}

type InstitutionStore interface {
	ListInstitutions(ctx context.Context, page, pageSize int, name, typ, status string) ([]model.InstitutionSummary, error)
	GetInstitution(ctx context.Context, id string) (*model.InstitutionSummary, error)
	CreateInstitution(ctx context.Context, i *model.InstitutionSummary) error
	UpdateInstitution(ctx context.Context, id string, updates map[string]any) error
	DeleteInstitution(ctx context.Context, id string) error
	CreateInstitutionAPIKey(ctx context.Context, institutionID, name string) (string, error)
	RevokeInstitutionAPIKey(ctx context.Context, institutionID, keyID string) error
}

type AuditStore interface {
	CreateAuditLog(ctx context.Context, log *model.AuditLog) error
	ListAuditLogs(ctx context.Context, limit int) ([]model.AuditLog, error)
	ListAuditLogsByUser(ctx context.Context, userID string, limit int) ([]model.AuditLog, error)
	ListAuditLogsByAction(ctx context.Context, action string, limit int) ([]model.AuditLog, error)
}


type PersonStore interface {
	CreatePerson(ctx context.Context, p *model.Person) error
	GetPerson(ctx context.Context, id string) (*model.Person, error)
	GetPersonByIDCard(ctx context.Context, idCard string) (*model.Person, error)
	ListPersons(ctx context.Context, page, pageSize int, businessChain, status string) ([]model.Person, error)
	UpdatePerson(ctx context.Context, id string, updates map[string]any) error
	DeletePerson(ctx context.Context, id string) error
	CreateProfile(ctx context.Context, pp *model.PersonProfile) error
	GetProfile(ctx context.Context, personID string, chain model.BusinessChain) (*model.PersonProfile, error)
	ListProfiles(ctx context.Context, chain model.BusinessChain) ([]model.PersonProfile, error)
	UpdateProfile(ctx context.Context, pp *model.PersonProfile) error
	AssignPersonWelfareTag(ctx context.Context, wt *model.PersonWelfareTag) error
	RevokePersonWelfareTag(ctx context.Context, personID, tagCode string) error
	ListPersonWelfareTags(ctx context.Context, personID string) ([]model.PersonWelfareTag, error)
}

// LifecycleStore manages person lifecycle transitions across business chains.
type LifecycleStore interface {
	TransitionStatus(ctx context.Context, personID string, chain model.BusinessChain, newStatus, reason string) error
	GetPersonStatus(ctx context.Context, personID string, chain model.BusinessChain) (string, error)
	LinkPersons(ctx context.Context, personID1, personID2 string, chain1, chain2 model.BusinessChain) error
}

type MedicationRuleStore interface {
	ListMedicationRules(ctx context.Context, personID string, chain model.BusinessChain) ([]model.MedicationRuleRow, error)
	CreateMedicationRuleV2(ctx context.Context, r *model.MedicationRuleRow) error
	UpdateMedicationRuleV2(ctx context.Context, id string, updates map[string]any) error
	DeleteMedicationRuleV2(ctx context.Context, id string) error
	CreateMedicationExecution(ctx context.Context, e *model.MedicationExecution) error
	ListMedicationExecutions(ctx context.Context, personID string, chain model.BusinessChain, limit int) ([]model.MedicationExecution, error)
}

type PersonRoleStore interface {
	AssignRole(ctx context.Context, binding *model.PersonRoleBinding) error
	ListRoles(ctx context.Context, userID string) ([]model.PersonRoleBinding, error)
	ListRolesByChain(ctx context.Context, chain model.BusinessChain) ([]model.PersonRoleBinding, error)
	RevokeRole(ctx context.Context, bindingID string) error
	GetEffectiveRole(ctx context.Context, userID string, chain model.BusinessChain) (string, bool)
}

type AlertRuleStore interface {
	CreateAlertRule(ctx context.Context, r *model.AlertRule) error
	GetAlertRule(ctx context.Context, id string) (*model.AlertRule, error)
	ListAlertRules(ctx context.Context, chain model.BusinessChain) ([]model.AlertRule, error)
	UpdateAlertRule(ctx context.Context, id string, updates map[string]any) error
	DeleteAlertRule(ctx context.Context, id string) error
}

type HealthRecordStore interface {
	CreateHealthRecordV2(ctx context.Context, r *model.HealthRecordV2) error
	ListHealthRecordsV2(ctx context.Context, personID string, chain model.BusinessChain, recordType string, limit int) ([]model.HealthRecordV2, error)
	GetHealthSummaryV2(ctx context.Context, personID string, chain model.BusinessChain) (*model.PersonHealthSummary, error)
	UpdateHealthSummaryV2(ctx context.Context, s *model.PersonHealthSummary) error
}

type HealthGuidanceStore interface {
	CreateGuidanceRule(ctx context.Context, r *model.HealthGuidanceRule) error
	ListGuidanceRules(ctx context.Context, chain model.BusinessChain, enabledOnly bool) ([]model.HealthGuidanceRule, error)
	EvaluateGuidanceRules(ctx context.Context, personID string, chain model.BusinessChain, healthData map[string]any) ([]model.HealthGuidanceRule, error)
	CreateGuidanceDelivery(ctx context.Context, d *model.HealthGuidanceDelivery) error
	ListGuidanceDeliveries(ctx context.Context, personID string, chain model.BusinessChain, limit int) ([]model.HealthGuidanceDelivery, error)
}

type HealthReportStore interface {
	CreateReportTemplate(ctx context.Context, t *model.HealthReportTemplate) error
	ListReportTemplates(ctx context.Context, chain model.BusinessChain) ([]model.HealthReportTemplate, error)
	CreateReport(ctx context.Context, r *model.HealthReport) error
	ListReports(ctx context.Context, personID string, chain model.BusinessChain, limit int) ([]model.HealthReport, error)
}

type ComplianceStore interface {
	CreateComplianceRule(ctx context.Context, r *model.ComplianceRule) error
	ListComplianceRules(ctx context.Context, chain model.BusinessChain) ([]model.ComplianceRule, error)
	RunComplianceCheck(ctx context.Context, ruleCode string, personID string) (*model.ComplianceCheck, error)
	ListComplianceChecks(ctx context.Context, personID string, limit int) ([]model.ComplianceCheck, error)
	ReviewCheck(ctx context.Context, checkID string, reviewerID string, result string, notes string) error
}

type DeviceBindingStore interface {
	BindDevice(ctx context.Context, binding *model.DeviceBinding) error
	UnbindDevice(ctx context.Context, bindingID string) error
	ListDeviceBindings(ctx context.Context, personID string, chain model.BusinessChain) ([]model.DeviceBinding, error)
	ListDevicesByPerson(ctx context.Context, personID string) ([]model.DeviceSummary, error)
}

type NotificationStore interface {
	CreateNotificationTemplate(ctx context.Context, t *model.NotificationTemplate) error
	ListNotificationTemplates(ctx context.Context, chain model.BusinessChain) ([]model.NotificationTemplate, error)
	CreateNotificationLog(ctx context.Context, l *model.NotificationLog) error
	UpdateNotificationStatus(ctx context.Context, logID string, status string, sentAt, readAt *time.Time) error
	ListNotificationLogs(ctx context.Context, personID string, chain model.BusinessChain, limit int) ([]model.NotificationLog, error)
}
