package store

import (
	"context"
	"time"

	"eregen.dev/admin-api/internal/model"
)

// MedicationRuleStore stubs
func (s *SqliteStore) ListMedicationRules(ctx context.Context, personID string, chain model.BusinessChain) ([]model.MedicationRuleRow, error) {
	return nil, nil
}
func (s *SqliteStore) CreateMedicationRuleV2(ctx context.Context, r *model.MedicationRuleRow) error {
	return nil
}
func (s *SqliteStore) UpdateMedicationRuleV2(ctx context.Context, id string, updates map[string]any) error {
	return nil
}
func (s *SqliteStore) DeleteMedicationRuleV2(ctx context.Context, id string) error {
	return nil
}
func (s *SqliteStore) CreateMedicationExecution(ctx context.Context, e *model.MedicationExecution) error {
	return nil
}
func (s *SqliteStore) ListMedicationExecutions(ctx context.Context, personID string, chain model.BusinessChain, limit int) ([]model.MedicationExecution, error) {
	return nil, nil
}

// PersonRoleStore stubs
func (s *SqliteStore) AssignRole(ctx context.Context, binding *model.PersonRoleBinding) error {
	return nil
}
func (s *SqliteStore) ListRoles(ctx context.Context, userID string) ([]model.PersonRoleBinding, error) {
	return nil, nil
}
func (s *SqliteStore) ListRolesByChain(ctx context.Context, chain model.BusinessChain) ([]model.PersonRoleBinding, error) {
	return nil, nil
}
func (s *SqliteStore) RevokeRole(ctx context.Context, bindingID string) error {
	return nil
}
func (s *SqliteStore) GetEffectiveRole(ctx context.Context, userID string, chain model.BusinessChain) (string, bool) {
	return "", false
}

// AlertRuleStore stubs
func (s *SqliteStore) CreateAlertRule(ctx context.Context, r *model.AlertRule) error {
	return nil
}
func (s *SqliteStore) GetAlertRule(ctx context.Context, id string) (*model.AlertRule, error) {
	return nil, nil
}
func (s *SqliteStore) ListAlertRules(ctx context.Context, chain model.BusinessChain) ([]model.AlertRule, error) {
	return nil, nil
}
func (s *SqliteStore) UpdateAlertRule(ctx context.Context, id string, updates map[string]any) error {
	return nil
}
func (s *SqliteStore) DeleteAlertRule(ctx context.Context, id string) error {
	return nil
}

// HealthRecordStore stubs
func (s *SqliteStore) CreateHealthRecordV2(ctx context.Context, r *model.HealthRecordV2) error {
	return nil
}
func (s *SqliteStore) ListHealthRecordsV2(ctx context.Context, personID string, chain model.BusinessChain, recordType string, limit int) ([]model.HealthRecordV2, error) {
	return nil, nil
}
func (s *SqliteStore) GetHealthSummaryV2(ctx context.Context, personID string, chain model.BusinessChain) (*model.PersonHealthSummary, error) {
	return nil, nil
}
func (s *SqliteStore) UpdateHealthSummaryV2(ctx context.Context, s2 *model.PersonHealthSummary) error {
	return nil
}

// HealthGuidanceStore stubs
func (s *SqliteStore) CreateGuidanceRule(ctx context.Context, r *model.HealthGuidanceRule) error {
	return nil
}
func (s *SqliteStore) ListGuidanceRules(ctx context.Context, chain model.BusinessChain, enabledOnly bool) ([]model.HealthGuidanceRule, error) {
	return nil, nil
}
func (s *SqliteStore) EvaluateGuidanceRules(ctx context.Context, personID string, chain model.BusinessChain, healthData map[string]any) ([]model.HealthGuidanceRule, error) {
	return nil, nil
}
func (s *SqliteStore) CreateGuidanceDelivery(ctx context.Context, d *model.HealthGuidanceDelivery) error {
	return nil
}
func (s *SqliteStore) ListGuidanceDeliveries(ctx context.Context, personID string, chain model.BusinessChain, limit int) ([]model.HealthGuidanceDelivery, error) {
	return nil, nil
}

// HealthReportStore stubs
func (s *SqliteStore) CreateReportTemplate(ctx context.Context, t *model.HealthReportTemplate) error {
	return nil
}
func (s *SqliteStore) ListReportTemplates(ctx context.Context, chain model.BusinessChain) ([]model.HealthReportTemplate, error) {
	return nil, nil
}
func (s *SqliteStore) CreateReport(ctx context.Context, r *model.HealthReport) error {
	return nil
}
func (s *SqliteStore) ListReports(ctx context.Context, personID string, chain model.BusinessChain, limit int) ([]model.HealthReport, error) {
	return nil, nil
}

// ComplianceStore stubs
func (s *SqliteStore) CreateComplianceRule(ctx context.Context, r *model.ComplianceRule) error {
	return nil
}
func (s *SqliteStore) ListComplianceRules(ctx context.Context, chain model.BusinessChain) ([]model.ComplianceRule, error) {
	return nil, nil
}
func (s *SqliteStore) RunComplianceCheck(ctx context.Context, ruleCode string, personID string) (*model.ComplianceCheck, error) {
	return nil, nil
}
func (s *SqliteStore) ListComplianceChecks(ctx context.Context, personID string, limit int) ([]model.ComplianceCheck, error) {
	return nil, nil
}
func (s *SqliteStore) ReviewCheck(ctx context.Context, checkID string, reviewerID string, result string, notes string) error {
	return nil
}

// DeviceBindingStore stubs
func (s *SqliteStore) BindDevice(ctx context.Context, binding *model.DeviceBinding) error {
	return nil
}
func (s *SqliteStore) ListDeviceBindings(ctx context.Context, personID string, chain model.BusinessChain) ([]model.DeviceBinding, error) {
	return nil, nil
}
func (s *SqliteStore) ListDevicesByPerson(ctx context.Context, personID string) ([]model.DeviceSummary, error) {
	return nil, nil
}

// NotificationStore stubs
func (s *SqliteStore) CreateNotificationTemplate(ctx context.Context, t *model.NotificationTemplate) error {
	return nil
}
func (s *SqliteStore) ListNotificationTemplates(ctx context.Context, chain model.BusinessChain) ([]model.NotificationTemplate, error) {
	return nil, nil
}
func (s *SqliteStore) CreateNotificationLog(ctx context.Context, l *model.NotificationLog) error {
	return nil
}
func (s *SqliteStore) UpdateNotificationStatus(ctx context.Context, logID string, status string, sentAt, readAt *time.Time) error {
	return nil
}
func (s *SqliteStore) ListNotificationLogs(ctx context.Context, personID string, chain model.BusinessChain, limit int) ([]model.NotificationLog, error) {
	return nil, nil
}
