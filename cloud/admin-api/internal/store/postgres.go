package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"eregen.dev/admin-api/internal/model"
)

// PostgresStore wraps database access for admin operations.
var ErrNotImplemented = errors.New("feature not yet implemented — backend schema required")

type PostgresStore struct {
	db *sql.DB
}

// NewPostgres opens a connection pool to PostgreSQL.
func NewPostgres(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		panic(err)
	}
	return db
}

// NewPostgresStore creates a new PostgresStore from an existing db connection.
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// PersonStore stub implementations for PostgresStore
func (s *PostgresStore) CreatePerson(ctx context.Context, p *model.Person) error { return ErrNotImplemented }
func (s *PostgresStore) GetPerson(ctx context.Context, id string) (*model.Person, error) { return nil, ErrNotImplemented }
func (s *PostgresStore) GetPersonByIDCard(ctx context.Context, idCard string) (*model.Person, error) { return nil, ErrNotImplemented }
func (s *PostgresStore) ListPersons(ctx context.Context, page, pageSize int, businessChain, status string) ([]model.Person, error) { return nil, ErrNotImplemented }
func (s *PostgresStore) UpdatePerson(ctx context.Context, id string, updates map[string]any) error { return ErrNotImplemented }
func (s *PostgresStore) DeletePerson(ctx context.Context, id string) error { return ErrNotImplemented }
func (s *PostgresStore) CreateProfile(ctx context.Context, pp *model.PersonProfile) error { return ErrNotImplemented }
func (s *PostgresStore) GetProfile(ctx context.Context, personID string, chain model.BusinessChain) (*model.PersonProfile, error) { return nil, ErrNotImplemented }
func (s *PostgresStore) ListProfiles(ctx context.Context, chain model.BusinessChain) ([]model.PersonProfile, error) { return nil, ErrNotImplemented }
func (s *PostgresStore) UpdateProfile(ctx context.Context, pp *model.PersonProfile) error { return ErrNotImplemented }
func (s *PostgresStore) AssignPersonWelfareTag(ctx context.Context, wt *model.PersonWelfareTag) error { return ErrNotImplemented }
func (s *PostgresStore) RevokePersonWelfareTag(ctx context.Context, personID, tagCode string) error { return ErrNotImplemented }
func (s *PostgresStore) ListPersonWelfareTags(ctx context.Context, personID string) ([]model.PersonWelfareTag, error) { return nil, ErrNotImplemented }

// MedicationRuleStore stub implementations
func (s *PostgresStore) ListMedicationRules(ctx context.Context, personID string, chain model.BusinessChain) ([]model.MedicationRuleRow, error) { return nil, ErrNotImplemented }
func (s *PostgresStore) CreateMedicationRuleV2(ctx context.Context, r *model.MedicationRuleRow) error { return ErrNotImplemented }
func (s *PostgresStore) UpdateMedicationRuleV2(ctx context.Context, id string, updates map[string]any) error { return ErrNotImplemented }
func (s *PostgresStore) DeleteMedicationRuleV2(ctx context.Context, id string) error { return ErrNotImplemented }
func (s *PostgresStore) CreateMedicationExecution(ctx context.Context, e *model.MedicationExecution) error { return ErrNotImplemented }
func (s *PostgresStore) ListMedicationExecutions(ctx context.Context, personID string, chain model.BusinessChain, limit int) ([]model.MedicationExecution, error) { return nil, ErrNotImplemented }

// PersonRoleStore stub implementations
func (s *PostgresStore) AssignRole(ctx context.Context, binding *model.PersonRoleBinding) error { return ErrNotImplemented }
func (s *PostgresStore) ListRoles(ctx context.Context, userID string) ([]model.PersonRoleBinding, error) { return nil, ErrNotImplemented }
func (s *PostgresStore) ListRolesByChain(ctx context.Context, chain model.BusinessChain) ([]model.PersonRoleBinding, error) { return nil, ErrNotImplemented }
func (s *PostgresStore) RevokeRole(ctx context.Context, bindingID string) error { return ErrNotImplemented }
func (s *PostgresStore) GetEffectiveRole(ctx context.Context, userID string, chain model.BusinessChain) (string, bool) { return "", false }

// AlertRuleStore stub implementations
func (s *PostgresStore) CreateAlertRule(ctx context.Context, r *model.AlertRule) error { return ErrNotImplemented }
func (s *PostgresStore) GetAlertRule(ctx context.Context, id string) (*model.AlertRule, error) { return nil, ErrNotImplemented }
func (s *PostgresStore) ListAlertRules(ctx context.Context, chain model.BusinessChain) ([]model.AlertRule, error) { return nil, ErrNotImplemented }
func (s *PostgresStore) UpdateAlertRule(ctx context.Context, id string, updates map[string]any) error { return ErrNotImplemented }
func (s *PostgresStore) DeleteAlertRule(ctx context.Context, id string) error { return ErrNotImplemented }

// HealthRecordStore stub implementations
func (s *PostgresStore) CreateHealthRecordV2(ctx context.Context, r *model.HealthRecordV2) error { return ErrNotImplemented }
func (s *PostgresStore) ListHealthRecordsV2(ctx context.Context, personID string, chain model.BusinessChain, recordType string, limit int) ([]model.HealthRecordV2, error) { return nil, ErrNotImplemented }
func (s *PostgresStore) GetHealthSummaryV2(ctx context.Context, personID string, chain model.BusinessChain) (*model.PersonHealthSummary, error) { return nil, ErrNotImplemented }
func (s *PostgresStore) UpdateHealthSummaryV2(ctx context.Context, s2 *model.PersonHealthSummary) error { return ErrNotImplemented }

// HealthGuidanceStore stub implementations
func (s *PostgresStore) CreateGuidanceRule(ctx context.Context, r *model.HealthGuidanceRule) error { return ErrNotImplemented }
func (s *PostgresStore) ListGuidanceRules(ctx context.Context, chain model.BusinessChain, enabledOnly bool) ([]model.HealthGuidanceRule, error) { return nil, ErrNotImplemented }
func (s *PostgresStore) EvaluateGuidanceRules(ctx context.Context, personID string, chain model.BusinessChain, healthData map[string]any) ([]model.HealthGuidanceRule, error) { return nil, ErrNotImplemented }
func (s *PostgresStore) CreateGuidanceDelivery(ctx context.Context, d *model.HealthGuidanceDelivery) error { return ErrNotImplemented }
func (s *PostgresStore) ListGuidanceDeliveries(ctx context.Context, personID string, chain model.BusinessChain, limit int) ([]model.HealthGuidanceDelivery, error) { return nil, ErrNotImplemented }

// HealthReportStore stub implementations
func (s *PostgresStore) CreateReportTemplate(ctx context.Context, t *model.HealthReportTemplate) error { return ErrNotImplemented }
func (s *PostgresStore) ListReportTemplates(ctx context.Context, chain model.BusinessChain) ([]model.HealthReportTemplate, error) { return nil, ErrNotImplemented }
func (s *PostgresStore) CreateReport(ctx context.Context, r *model.HealthReport) error { return ErrNotImplemented }
func (s *PostgresStore) ListReports(ctx context.Context, personID string, chain model.BusinessChain, limit int) ([]model.HealthReport, error) { return nil, ErrNotImplemented }

// ComplianceStore stub implementations
func (s *PostgresStore) CreateComplianceRule(ctx context.Context, r *model.ComplianceRule) error { return ErrNotImplemented }
func (s *PostgresStore) ListComplianceRules(ctx context.Context, chain model.BusinessChain) ([]model.ComplianceRule, error) { return nil, ErrNotImplemented }
func (s *PostgresStore) RunComplianceCheck(ctx context.Context, ruleCode string, personID string) (*model.ComplianceCheck, error) { return nil, ErrNotImplemented }
func (s *PostgresStore) ListComplianceChecks(ctx context.Context, personID string, limit int) ([]model.ComplianceCheck, error) { return nil, ErrNotImplemented }
func (s *PostgresStore) ReviewCheck(ctx context.Context, checkID string, reviewerID string, result string, notes string) error { return ErrNotImplemented }

// DeviceBindingStore stub implementations
func (s *PostgresStore) BindDevice(ctx context.Context, binding *model.DeviceBinding) error { return ErrNotImplemented }
func (s *PostgresStore) ListDeviceBindings(ctx context.Context, personID string, chain model.BusinessChain) ([]model.DeviceBinding, error) { return nil, ErrNotImplemented }
func (s *PostgresStore) ListDevicesByPerson(ctx context.Context, personID string) ([]model.DeviceSummary, error) { return nil, ErrNotImplemented }

// NotificationStore stub implementations
func (s *PostgresStore) CreateNotificationTemplate(ctx context.Context, t *model.NotificationTemplate) error { return ErrNotImplemented }
func (s *PostgresStore) ListNotificationTemplates(ctx context.Context, chain model.BusinessChain) ([]model.NotificationTemplate, error) { return nil, ErrNotImplemented }
func (s *PostgresStore) CreateNotificationLog(ctx context.Context, l *model.NotificationLog) error { return ErrNotImplemented }
func (s *PostgresStore) UpdateNotificationStatus(ctx context.Context, logID string, status string, sentAt, readAt *time.Time) error { return ErrNotImplemented }
func (s *PostgresStore) ListNotificationLogs(ctx context.Context, personID string, chain model.BusinessChain, limit int) ([]model.NotificationLog, error) { return nil, ErrNotImplemented }

// LifecycleStore stub implementations for PostgresStore
func (s *PostgresStore) TransitionStatus(ctx context.Context, personID string, chain model.BusinessChain, newStatus, reason string) error { return ErrNotImplemented }
func (s *PostgresStore) GetPersonStatus(ctx context.Context, personID string, chain model.BusinessChain) (string, error) { return "", ErrNotImplemented }
func (s *PostgresStore) LinkPersons(ctx context.Context, personID1, personID2 string, chain1, chain2 model.BusinessChain) error { return ErrNotImplemented }
