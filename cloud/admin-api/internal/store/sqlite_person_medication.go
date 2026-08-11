package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"eregen.dev/admin-api/internal/model"
	"github.com/google/uuid"
)

// MedicationRuleStore implementation

func (s *SqliteStore) ListMedicationRules(ctx context.Context, personID string, chain model.BusinessChain) ([]model.MedicationRuleRow, error) {
	query := `SELECT id, person_id, business_chain, source_type, source_id, drug_name, generic_name,
				drug_category, dosage, frequency, route, schedule_time1, schedule_time2, schedule_time3,
				days_of_week, duration, pre_meal, post_meal, special_instructions, prescribed_by,
				prescribed_at, active, created_at
			  FROM medication_rules_v2 WHERE person_id = ?`
	args := []any{personID}
	if chain != "" {
		query += ` AND business_chain = ?`
		args = append(args, chain)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list medication rules: %w", err)
	}
	defer rows.Close()

	var rules []model.MedicationRuleRow
	for rows.Next() {
		var r model.MedicationRuleRow
		if err := rows.Scan(&r.ID, &r.PersonID, &r.BusinessChain, &r.SourceType, &r.SourceID,
			&r.DrugName, &r.GenericName, &r.DrugCategory, &r.Dosage, &r.Frequency, &r.Route,
			&r.ScheduleTime1, &r.ScheduleTime2, &r.ScheduleTime3, &r.DaysOfWeek, &r.Duration,
			&r.PreMeal, &r.PostMeal, &r.SpecialInstructions, &r.PrescribedBy, &r.PrescribedAt,
			&r.Active, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan medication rule: %w", err)
		}
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

func (s *SqliteStore) CreateMedicationRuleV2(ctx context.Context, r *model.MedicationRuleRow) error {
	r.ID = uuid.New().String()
	r.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
	r.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medication_rules_v2 (id, person_id, business_chain, source_type, source_id,
		 drug_name, generic_name, drug_category, dosage, frequency, route, schedule_time1,
		 schedule_time2, schedule_time3, days_of_week, duration, pre_meal, post_meal,
		 special_instructions, prescribed_by, prescribed_at, active)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.PersonID, r.BusinessChain, r.SourceType, r.SourceID, r.DrugName, r.GenericName,
		r.DrugCategory, r.Dosage, r.Frequency, r.Route, r.ScheduleTime1, r.ScheduleTime2,
		r.ScheduleTime3, r.DaysOfWeek, r.Duration, r.PreMeal, r.PostMeal, r.SpecialInstructions,
		r.PrescribedBy, r.PrescribedAt, 1)
	return err
}

func (s *SqliteStore) UpdateMedicationRuleV2(ctx context.Context, id string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	setClauses := make([]string, 0, len(updates))
	args := []any{}
	for k, v := range updates {
		setClauses = append(setClauses, k+" = ?")
		args = append(args, v)
	}
	args = append(args, id)
	_, err := s.db.ExecContext(ctx,
		fmt.Sprintf("UPDATE medication_rules_v2 SET %s, updated_at = datetime('now') WHERE id = ?",
			strings.Join(setClauses, ", ")), args...)
	return err
}

func (s *SqliteStore) DeleteMedicationRuleV2(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM medication_rules_v2 WHERE id = ?`, id)
	return err
}

func (s *SqliteStore) CreateMedicationExecution(ctx context.Context, e *model.MedicationExecution) error {
	e.ID = uuid.New().String()
	e.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medication_executions (id, person_id, business_chain, rule_id, scheduled_time,
		 actual_time, status, taken_by, device_id, verification_method, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.PersonID, e.BusinessChain, e.RuleID, e.ScheduledTime, e.ActualTime,
		e.Status, e.TakenBy, e.DeviceID, e.VerificationMethod, e.Notes)
	return err
}

func (s *SqliteStore) ListMedicationExecutions(ctx context.Context, personID string, chain model.BusinessChain, limit int) ([]model.MedicationExecution, error) {
	query := `SELECT id, person_id, business_chain, rule_id, scheduled_time, actual_time,
				status, taken_by, device_id, verification_method, notes, created_at
			  FROM medication_executions WHERE person_id = ?`
	args := []any{personID}
	if chain != "" {
		query += ` AND business_chain = ?`
		args = append(args, chain)
	}
	query += ` ORDER BY scheduled_time DESC LIMIT ?`
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list medication executions: %w", err)
	}
	defer rows.Close()

	var executions []model.MedicationExecution
	for rows.Next() {
		var e model.MedicationExecution
		if err := rows.Scan(&e.ID, &e.PersonID, &e.BusinessChain, &e.RuleID, &e.ScheduledTime,
			&e.ActualTime, &e.Status, &e.TakenBy, &e.DeviceID, &e.VerificationMethod,
			&e.Notes, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan medication execution: %w", err)
		}
		executions = append(executions, e)
	}
	return executions, rows.Err()
}

// PersonRoleStore implementation

func (s *SqliteStore) AssignRole(ctx context.Context, binding *model.PersonRoleBinding) error {
	binding.ID = uuid.New().String()
	binding.CreatedAt = time.Now()
	expiresAt := binding.ExpiresAt
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO user_role_bindings (id, user_id, business_chain, role, institution_id,
		 granted_by, expires_at, active)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		binding.ID, binding.UserID, binding.BusinessChain, binding.Role,
		binding.InstitutionID, binding.GrantedBy, expiresAt,
		binding.Active)
	return err
}

func (s *SqliteStore) ListRoles(ctx context.Context, userID string) ([]model.PersonRoleBinding, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, business_chain, role, institution_id, granted_by, granted_at,
		 expires_at, active, created_at
		 FROM user_role_bindings WHERE user_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	defer rows.Close()

	var bindings []model.PersonRoleBinding
	for rows.Next() {
		var b model.PersonRoleBinding
		var expiresRaw sql.NullString
		if err := rows.Scan(&b.ID, &b.UserID, &b.BusinessChain, &b.Role, &b.InstitutionID,
			&b.GrantedBy, &b.GrantedAt, &expiresRaw, &b.Active, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan role binding: %w", err)
		}
		if expiresRaw.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", expiresRaw.String)
			b.ExpiresAt = &t
		}
		bindings = append(bindings, b)
	}
	return bindings, rows.Err()
}

func (s *SqliteStore) ListRolesByChain(ctx context.Context, chain model.BusinessChain) ([]model.PersonRoleBinding, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, business_chain, role, institution_id, granted_by, granted_at,
		 expires_at, active, created_at
		 FROM user_role_bindings WHERE business_chain = ? AND active = 1 ORDER BY created_at DESC`, chain)
	if err != nil {
		return nil, fmt.Errorf("list roles by chain: %w", err)
	}
	defer rows.Close()

	var bindings []model.PersonRoleBinding
	for rows.Next() {
		var b model.PersonRoleBinding
		var expiresRaw sql.NullString
		if err := rows.Scan(&b.ID, &b.UserID, &b.BusinessChain, &b.Role, &b.InstitutionID,
			&b.GrantedBy, &b.GrantedAt, &expiresRaw, &b.Active, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan role binding: %w", err)
		}
		if expiresRaw.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", expiresRaw.String)
			b.ExpiresAt = &t
		}
		bindings = append(bindings, b)
	}
	return bindings, rows.Err()
}

func (s *SqliteStore) RevokeRole(ctx context.Context, bindingID string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE user_role_bindings SET active = 0 WHERE id = ?`, bindingID)
	return err
}

func (s *SqliteStore) GetEffectiveRole(ctx context.Context, userID string, chain model.BusinessChain) (string, bool) {
	var role string
	err := s.db.QueryRowContext(ctx,
		`SELECT role FROM user_role_bindings WHERE user_id = ? AND business_chain = ? AND active = 1
		 AND (expires_at IS NULL OR expires_at > datetime('now'))
		 ORDER BY granted_at DESC LIMIT 1`, userID, chain).Scan(&role)
	if err == sql.ErrNoRows {
		return "", false
	}
	if err != nil {
		return "", false
	}
	return role, true
}

// AlertRuleStore implementation

func (s *SqliteStore) CreateAlertRule(ctx context.Context, r *model.AlertRule) error {
	r.ID = uuid.New().String()
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO alert_rules (id, name, business_chain, alert_type, severity, condition_field,
		 condition_operator, condition_threshold, condition_duration_min, notify_roles, notify_channels,
		 notify_institution_ids, escalation_timeout_min, escalation_roles, auto_action, active)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.Name, r.BusinessChain, r.AlertType, r.Severity, r.ConditionField,
		r.ConditionOperator, nullableInt(r.ConditionThreshold), nullableInt(r.ConditionDurationMin),
		r.NotifyRoles, r.NotifyChannels, r.NotifyInstitutionIDs, r.EscalationTimeoutMin,
		r.EscalationRoles, r.AutoAction, r.Active)
	return err
}

func nullableInt(i *int) interface{} {
	if i == nil {
		return nil
	}
	return *i
}

func (s *SqliteStore) GetAlertRule(ctx context.Context, id string) (*model.AlertRule, error) {
	var r model.AlertRule
	var threshold, durationMin *int
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, business_chain, alert_type, severity, condition_field, condition_operator,
		 condition_threshold, condition_duration_min, notify_roles, notify_channels,
		 notify_institution_ids, escalation_timeout_min, escalation_roles, auto_action, active,
		 created_at, updated_at
		 FROM alert_rules WHERE id = ?`, id).Scan(
		&r.ID, &r.Name, &r.BusinessChain, &r.AlertType, &r.Severity, &r.ConditionField,
		&r.ConditionOperator, &threshold, &durationMin, &r.NotifyRoles, &r.NotifyChannels,
		&r.NotifyInstitutionIDs, &r.EscalationTimeoutMin, &r.EscalationRoles, &r.AutoAction,
		&r.Active, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get alert rule: %w", err)
	}
	r.ConditionThreshold = threshold
	r.ConditionDurationMin = durationMin
	return &r, nil
}

func (s *SqliteStore) ListAlertRules(ctx context.Context, chain model.BusinessChain) ([]model.AlertRule, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, business_chain, alert_type, severity, condition_field, condition_operator,
		 condition_threshold, condition_duration_min, notify_roles, notify_channels,
		 notify_institution_ids, escalation_timeout_min, escalation_roles, auto_action, active,
		 created_at, updated_at
		 FROM alert_rules WHERE business_chain = ? ORDER BY created_at DESC`, chain)
	if err != nil {
		return nil, fmt.Errorf("list alert rules: %w", err)
	}
	defer rows.Close()

	var rules []model.AlertRule
	for rows.Next() {
		var r model.AlertRule
		var threshold, durationMin *int
		if err := rows.Scan(&r.ID, &r.Name, &r.BusinessChain, &r.AlertType, &r.Severity,
			&r.ConditionField, &r.ConditionOperator, &threshold, &durationMin, &r.NotifyRoles,
			&r.NotifyChannels, &r.NotifyInstitutionIDs, &r.EscalationTimeoutMin, &r.EscalationRoles,
			&r.AutoAction, &r.Active, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan alert rule: %w", err)
		}
		r.ConditionThreshold = threshold
		r.ConditionDurationMin = durationMin
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

func (s *SqliteStore) UpdateAlertRule(ctx context.Context, id string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	setClauses := make([]string, 0, len(updates))
	args := []any{}
	for k, v := range updates {
		setClauses = append(setClauses, k+" = ?")
		args = append(args, v)
	}
	args = append(args, id)
	_, err := s.db.ExecContext(ctx,
		fmt.Sprintf("UPDATE alert_rules SET %s, updated_at = datetime('now') WHERE id = ?",
			fmt.Sprintf("%s", setClauses[0])), args...)
	return err
}

func (s *SqliteStore) DeleteAlertRule(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM alert_rules WHERE id = ?`, id)
	return err
}

// HealthRecordStore implementation

func (s *SqliteStore) CreateHealthRecordV2(ctx context.Context, r *model.HealthRecordV2) error {
	r.ID = uuid.New().String()
	r.CreatedAt = time.Now()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO health_records_v2 (id, person_id, business_chain, record_type, source,
		 device_id, recorded_at, heart_rate, blood_pressure_sys, blood_pressure_dia, spo2,
		 temperature, respiratory_rate, pulse_rate, blood_glucose_fasting, blood_glucose_postprandial,
		 uric_acid, creatinine, hemoglobin_a1c, weight, height, bmi, steps, sleep_hours,
		 exercise_minutes, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.PersonID, r.BusinessChain, r.RecordType, r.Source, r.DeviceID, r.RecordedAt,
		nullableIntPtr(r.HeartRate), nullableIntPtr(r.BloodPressureSys), nullableIntPtr(r.BloodPressureDia),
		nullableIntPtr(r.SpO2), nullableFloatPtr(r.Temperature), nullableIntPtr(r.RespiratoryRate),
		nullableIntPtr(r.PulseRate), nullableFloatPtr(r.GlucoseFasting), nullableFloatPtr(r.GlucosePost),
		nullableFloatPtr(r.UricAcid), nullableFloatPtr(r.Creatinine), nullableFloatPtr(r.HbA1c),
		nullableFloatPtr(r.Weight), nullableFloatPtr(r.Height), nullableFloatPtr(r.BMI),
		nullableInt64Ptr(r.Steps), nullableFloatPtr(r.SleepHours), nullableIntPtr(r.ExerciseMinutes),
		r.Notes)
	return err
}

func nullableIntPtr(i *int) interface{} {
	if i == nil {
		return nil
	}
	return *i
}
func nullableFloatPtr(f *float64) interface{} {
	if f == nil {
		return nil
	}
	return *f
}
func nullableInt64Ptr(i *int64) interface{} {
	if i == nil {
		return nil
	}
	return *i
}

func (s *SqliteStore) ListHealthRecordsV2(ctx context.Context, personID string, chain model.BusinessChain, recordType string, limit int) ([]model.HealthRecordV2, error) {
	query := `SELECT id, person_id, business_chain, record_type, source, device_id, recorded_at,
				heart_rate, blood_pressure_sys, blood_pressure_dia, spo2, temperature,
				respiratory_rate, pulse_rate, blood_glucose_fasting, blood_glucose_postprandial,
				uric_acid, creatinine, hemoglobin_a1c, weight, height, bmi, steps, sleep_hours,
				exercise_minutes, notes, created_at
			  FROM health_records_v2 WHERE person_id = ?`
	args := []any{personID}
	if chain != "" {
		query += ` AND business_chain = ?`
		args = append(args, chain)
	}
	if recordType != "" {
		query += ` AND record_type = ?`
		args = append(args, recordType)
	}
	query += ` ORDER BY recorded_at DESC LIMIT ?`
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list health records: %w", err)
	}
	defer rows.Close()

	var records []model.HealthRecordV2
	for rows.Next() {
		var r model.HealthRecordV2
		if err := rows.Scan(&r.ID, &r.PersonID, &r.BusinessChain, &r.RecordType, &r.Source,
			&r.DeviceID, &r.RecordedAt, &r.HeartRate, &r.BloodPressureSys, &r.BloodPressureDia,
			&r.SpO2, &r.Temperature, &r.RespiratoryRate, &r.PulseRate, &r.GlucoseFasting,
			&r.GlucosePost, &r.UricAcid, &r.Creatinine, &r.HbA1c, &r.Weight, &r.Height,
			&r.BMI, &r.Steps, &r.SleepHours, &r.ExerciseMinutes, &r.Notes, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan health record: %w", err)
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

func (s *SqliteStore) GetHealthSummaryV2(ctx context.Context, personID string, chain model.BusinessChain) (*model.PersonHealthSummary, error) {
	var s2 model.PersonHealthSummary
	var hr, spo2, bpSys, bpDia *int
	var glucose, uricAcid *float64
	var steps *int64
	var sleepHours *float64
	var riskScore *float64
	var latestUpdated string

	err := s.db.QueryRowContext(ctx,
		`SELECT person_id, business_chain, latest_hr, latest_spo2, latest_bp_sys, latest_bp_dia,
		 latest_glucose_fasting, latest_uric_acid, latest_steps, latest_sleep_hours,
		 risk_score, trend_direction, last_updated, ai_recommendation
		 FROM person_health_summaries WHERE person_id = ? AND business_chain = ?`,
		personID, chain).Scan(
		&s2.PersonID, &s2.BusinessChain, &hr, &spo2, &bpSys, &bpDia,
		&glucose, &uricAcid, &steps, &sleepHours, &riskScore, &s2.TrendDirection,
		&latestUpdated, &s2.ARecommendation)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get health summary: %w", err)
	}
	s2.LatestHR = hr
	s2.LatestSpO2 = spo2
	s2.LatestBPSys = bpSys
	s2.LatestBPDia = bpDia
	s2.LatestGlucoseFasting = glucose
	s2.LatestUricAcid = uricAcid
	s2.LatestSteps = steps
	s2.LatestSleepHours = sleepHours
	s2.RiskScore = riskScore
	if latestUpdated != "" {
		t, _ := time.Parse("2006-01-02 15:04:05", latestUpdated)
		s2.LastUpdated = t
	}
	return &s2, nil
}

func (s *SqliteStore) UpdateHealthSummaryV2(ctx context.Context, s2 *model.PersonHealthSummary) error {
	s2.LastUpdated = time.Now()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO person_health_summaries (person_id, business_chain, latest_hr, latest_spo2,
		 latest_bp_sys, latest_bp_dia, latest_glucose_fasting, latest_uric_acid, latest_steps,
		 latest_sleep_hours, risk_score, trend_direction, last_updated, ai_recommendation)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(person_id, business_chain) DO UPDATE SET
		 latest_hr=excluded.latest_hr, latest_spo2=excluded.latest_spo2,
		 latest_bp_sys=excluded.latest_bp_sys, latest_bp_dia=excluded.latest_bp_dia,
		 latest_glucose_fasting=excluded.latest_glucose_fasting, latest_uric_acid=excluded.latest_uric_acid,
		 latest_steps=excluded.latest_steps, latest_sleep_hours=excluded.latest_sleep_hours,
		 risk_score=excluded.risk_score, trend_direction=excluded.trend_direction,
		 last_updated=excluded.last_updated, ai_recommendation=excluded.ai_recommendation`,
		s2.PersonID, s2.BusinessChain, nullableIntPtr(s2.LatestHR), nullableIntPtr(s2.LatestSpO2),
		nullableIntPtr(s2.LatestBPSys), nullableIntPtr(s2.LatestBPDia),
		nullableFloatPtr(s2.LatestGlucoseFasting), nullableFloatPtr(s2.LatestUricAcid),
		nullableInt64Ptr(s2.LatestSteps), nullableFloatPtr(s2.LatestSleepHours),
		nullableFloatPtr(s2.RiskScore), s2.TrendDirection, s2.LastUpdated.Format("2006-01-02 15:04:05"),
		s2.ARecommendation)
	return err
}

// HealthGuidanceStore implementation

func (s *SqliteStore) CreateGuidanceRule(ctx context.Context, r *model.HealthGuidanceRule) error {
	r.ID = uuid.New().String()
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO health_guidance_rules (id, name, business_chain, trigger_condition,
		 condition_field, condition_operator, condition_threshold, guidance_type, title,
		 content, priority, enabled)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.Name, r.BusinessChain, r.TriggerCondition, r.ConditionField, r.ConditionOp,
		r.ConditionThresh, r.GuidanceType, r.Title, r.Content, r.Priority, r.Enabled)
	return err
}

func (s *SqliteStore) ListGuidanceRules(ctx context.Context, chain model.BusinessChain, enabledOnly bool) ([]model.HealthGuidanceRule, error) {
	query := `SELECT id, name, business_chain, trigger_condition, condition_field, condition_operator,
			 condition_threshold, guidance_type, title, content, priority, enabled, created_at, updated_at
			 FROM health_guidance_rules WHERE business_chain = ?`
	args := []any{chain}
	if enabledOnly {
		query += ` AND enabled = 1`
	}
	query += ` ORDER BY priority DESC, created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list guidance rules: %w", err)
	}
	defer rows.Close()

	var rules []model.HealthGuidanceRule
	for rows.Next() {
		var r model.HealthGuidanceRule
		if err := rows.Scan(&r.ID, &r.Name, &r.BusinessChain, &r.TriggerCondition, &r.ConditionField,
			&r.ConditionOp, &r.ConditionThresh, &r.GuidanceType, &r.Title, &r.Content,
			&r.Priority, &r.Enabled, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan guidance rule: %w", err)
		}
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

func (s *SqliteStore) EvaluateGuidanceRules(ctx context.Context, personID string, chain model.BusinessChain, healthData map[string]any) ([]model.HealthGuidanceRule, error) {
	rules, err := s.ListGuidanceRules(ctx, chain, true)
	if err != nil {
		return nil, err
	}
	// Simple evaluation: return all enabled rules for now
	// In production, this would check trigger conditions against healthData
	return rules, nil
}

func (s *SqliteStore) CreateGuidanceDelivery(ctx context.Context, d *model.HealthGuidanceDelivery) error {
	d.ID = uuid.New().String()
	d.DeliveredAt = time.Now()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO health_guidance_deliveries (id, person_id, business_chain, rule_id,
		 guidance_type, title, content, channel, delivered_at, read_status, feedback)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		d.ID, d.PersonID, d.BusinessChain, d.RuleID, d.GuidanceType, d.Title, d.Content,
		d.Channel, d.DeliveredAt, d.ReadStatus, d.Feedback)
	return err
}

func (s *SqliteStore) ListGuidanceDeliveries(ctx context.Context, personID string, chain model.BusinessChain, limit int) ([]model.HealthGuidanceDelivery, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, person_id, business_chain, rule_id, guidance_type, title, content,
		 channel, delivered_at, read_status, feedback
		 FROM health_guidance_deliveries WHERE person_id = ? AND business_chain = ?
		 ORDER BY delivered_at DESC LIMIT ?`, personID, chain, limit)
	if err != nil {
		return nil, fmt.Errorf("list guidance deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []model.HealthGuidanceDelivery
	for rows.Next() {
		var d model.HealthGuidanceDelivery
		if err := rows.Scan(&d.ID, &d.PersonID, &d.BusinessChain, &d.RuleID, &d.GuidanceType,
			&d.Title, &d.Content, &d.Channel, &d.DeliveredAt, &d.ReadStatus, &d.Feedback); err != nil {
			return nil, fmt.Errorf("scan guidance delivery: %w", err)
		}
		deliveries = append(deliveries, d)
	}
	return deliveries, rows.Err()
}

// HealthReportStore implementation

func (s *SqliteStore) CreateReportTemplate(ctx context.Context, t *model.HealthReportTemplate) error {
	t.ID = uuid.New().String()
	t.CreatedAt = time.Now()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO health_report_templates (id, name, business_chain, frequency, template_type,
		 include_sections, enabled)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.Name, t.BusinessChain, t.Frequency, t.TemplateType, t.IncludeSections, t.Enabled)
	return err
}

func (s *SqliteStore) ListReportTemplates(ctx context.Context, chain model.BusinessChain) ([]model.HealthReportTemplate, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, business_chain, frequency, template_type, include_sections, enabled, created_at
		 FROM health_report_templates WHERE business_chain = ? ORDER BY created_at DESC`, chain)
	if err != nil {
		return nil, fmt.Errorf("list report templates: %w", err)
	}
	defer rows.Close()

	var templates []model.HealthReportTemplate
	for rows.Next() {
		var t model.HealthReportTemplate
		if err := rows.Scan(&t.ID, &t.Name, &t.BusinessChain, &t.Frequency, &t.TemplateType,
			&t.IncludeSections, &t.Enabled, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan report template: %w", err)
		}
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

func (s *SqliteStore) CreateReport(ctx context.Context, r *model.HealthReport) error {
	r.ID = uuid.New().String()
	r.GeneratedAt = time.Now()
	r.CreatedAt = time.Now()
	if r.Status == "" {
		r.Status = "generated"
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO health_reports (id, person_id, business_chain, template_id,
		 report_period_start, report_period_end, generated_at, report_data,
		 delivered_channels, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.PersonID, r.BusinessChain, r.TemplateID,
		r.ReportPeriodStart.Format("2006-01-02"), r.ReportPeriodEnd.Format("2006-01-02"),
		r.GeneratedAt, r.ReportData, r.DeliveredChannels, r.Status)
	return err
}

func (s *SqliteStore) ListReports(ctx context.Context, personID string, chain model.BusinessChain, limit int) ([]model.HealthReport, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, person_id, business_chain, template_id, report_period_start, report_period_end,
		 generated_at, report_data, delivered_channels, status, created_at
		 FROM health_reports WHERE person_id = ? AND business_chain = ?
		 ORDER BY generated_at DESC LIMIT ?`, personID, chain, limit)
	if err != nil {
		return nil, fmt.Errorf("list reports: %w", err)
	}
	defer rows.Close()

	var reports []model.HealthReport
	for rows.Next() {
		var r model.HealthReport
		var startRaw, endRaw, genRaw string
		if err := rows.Scan(&r.ID, &r.PersonID, &r.BusinessChain, &r.TemplateID, &startRaw,
			&endRaw, &genRaw, &r.ReportData, &r.DeliveredChannels, &r.Status, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan report: %w", err)
		}
		r.ReportPeriodStart, _ = time.Parse("2006-01-02", startRaw)
		r.ReportPeriodEnd, _ = time.Parse("2006-01-02", endRaw)
		r.GeneratedAt, _ = time.Parse("2006-01-02 15:04:05", genRaw)
		reports = append(reports, r)
	}
	return reports, rows.Err()
}

// ComplianceStore implementation

func (s *SqliteStore) CreateComplianceRule(ctx context.Context, r *model.ComplianceRule) error {
	r.ID = uuid.New().String()
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO compliance_rules (id, rule_code, name, description, business_chain,
		 rule_type, condition_sql, severity, action_required, enabled)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.RuleCode, r.Name, r.Description, r.BusinessChain, r.RuleType,
		r.ConditionSQL, r.Severity, r.ActionRequired, r.Enabled)
	return err
}

func (s *SqliteStore) ListComplianceRules(ctx context.Context, chain model.BusinessChain) ([]model.ComplianceRule, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, rule_code, name, description, business_chain, rule_type, condition_sql,
		 severity, action_required, enabled, created_at, updated_at
		 FROM compliance_rules WHERE business_chain = ? ORDER BY created_at DESC`, chain)
	if err != nil {
		return nil, fmt.Errorf("list compliance rules: %w", err)
	}
	defer rows.Close()

	var rules []model.ComplianceRule
	for rows.Next() {
		var r model.ComplianceRule
		if err := rows.Scan(&r.ID, &r.RuleCode, &r.Name, &r.Description, &r.BusinessChain,
			&r.RuleType, &r.ConditionSQL, &r.Severity, &r.ActionRequired, &r.Enabled,
			&r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan compliance rule: %w", err)
		}
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

func (s *SqliteStore) RunComplianceCheck(ctx context.Context, ruleCode string, personID string) (*model.ComplianceCheck, error) {
	check := &model.ComplianceCheck{
		ID:         uuid.New().String(),
		PersonID:   personID,
		CheckTime:  time.Now(),
		CreatedAt:  time.Now(),
		Violated:   0,
	}
	// In production, this would execute the rule's condition_sql
	// For now, insert a placeholder check record
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO compliance_checks (id, rule_id, person_id, check_time, violated, violation_details, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		check.ID, ruleCode, personID, check.CheckTime, check.Violated, check.ViolationDetails, check.CreatedAt)
	return check, err
}

func (s *SqliteStore) ListComplianceChecks(ctx context.Context, personID string, limit int) ([]model.ComplianceCheck, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, rule_id, person_id, check_time, violated, violation_details, reviewed_by,
		 reviewed_at, review_result, action_taken, created_at
		 FROM compliance_checks WHERE person_id = ? ORDER BY check_time DESC LIMIT ?`,
		personID, limit)
	if err != nil {
		return nil, fmt.Errorf("list compliance checks: %w", err)
	}
	defer rows.Close()

	var checks []model.ComplianceCheck
	for rows.Next() {
		var c model.ComplianceCheck
		var reviewedAtRaw sql.NullString
		if err := rows.Scan(&c.ID, &c.RuleID, &c.PersonID, &c.CheckTime, &c.Violated,
			&c.ViolationDetails, &c.ReviewedBy, &reviewedAtRaw, &c.ReviewResult,
			&c.ActionTaken, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan compliance check: %w", err)
		}
		if reviewedAtRaw.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", reviewedAtRaw.String)
			c.ReviewedAt = &t
		}
		checks = append(checks, c)
	}
	return checks, rows.Err()
}

func (s *SqliteStore) ReviewCheck(ctx context.Context, checkID string, reviewerID string, result string, notes string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE compliance_checks SET reviewed_by = ?, reviewed_at = datetime('now'),
		 review_result = ?, action_taken = ? WHERE id = ?`,
		reviewerID, result, notes, checkID)
	return err
}

// DeviceBindingStore implementation

func (s *SqliteStore) BindDevice(ctx context.Context, binding *model.DeviceBinding) error {
	binding.ID = uuid.New().String()
	binding.BoundAt = time.Now()
	binding.CreatedAt = time.Now()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO device_bindings (id, device_id, person_id, business_chain, bound_at,
		 binding_type, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		binding.ID, binding.DeviceID, binding.PersonID, binding.BusinessChain,
		binding.BoundAt, binding.BindingType, binding.CreatedAt)
	return err
}

func (s *SqliteStore) ListDeviceBindings(ctx context.Context, personID string, chain model.BusinessChain) ([]model.DeviceBinding, error) {
	query := `SELECT id, device_id, person_id, business_chain, bound_at, unbound_at,
			 binding_type, created_at
			 FROM device_bindings WHERE person_id = ?`
	args := []any{personID}
	if chain != "" {
		query += ` AND business_chain = ?`
		args = append(args, chain)
	}
	query += ` ORDER BY bound_at DESC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list device bindings: %w", err)
	}
	defer rows.Close()

	var bindings []model.DeviceBinding
	for rows.Next() {
		var b model.DeviceBinding
		var unboundRaw sql.NullString
		if err := rows.Scan(&b.ID, &b.DeviceID, &b.PersonID, &b.BusinessChain, &b.BoundAt,
			&unboundRaw, &b.BindingType, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan device binding: %w", err)
		}
		if unboundRaw.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", unboundRaw.String)
			b.UnboundAt = &t
		}
		bindings = append(bindings, b)
	}
	return bindings, rows.Err()
}

func (s *SqliteStore) ListDevicesByPerson(ctx context.Context, personID string) ([]model.DeviceSummary, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT d.id, d.device_id, d.device_type, d.tier, d.status, d.last_seen,
				d.firmware_version, e.name as owner_name
			 FROM devices d
			 JOIN device_bindings db ON d.id = db.device_id
			 LEFT JOIN elderly_profiles e ON db.person_id = e.id
			 WHERE db.person_id = ? ORDER BY d.created_at DESC`, personID)
	if err != nil {
		return nil, fmt.Errorf("list devices by person: %w", err)
	}
	defer rows.Close()

	var devices []model.DeviceSummary
	for rows.Next() {
		var d model.DeviceSummary
		if err := rows.Scan(&d.ID, &d.DeviceID, &d.Type, &d.Tier, &d.Status, &d.LastSeen,
			&d.FirmwareVer, &d.OwnerName); err != nil {
			return nil, fmt.Errorf("scan device: %w", err)
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

// NotificationStore implementation

func (s *SqliteStore) CreateNotificationTemplate(ctx context.Context, t *model.NotificationTemplate) error {
	t.ID = uuid.New().String()
	t.CreatedAt = time.Now()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO notification_templates (id, name, business_chain, channel, subject,
		 body_template, enabled)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.Name, t.BusinessChain, t.Channel, t.Subject, t.BodyTemplate, t.Enabled)
	return err
}

func (s *SqliteStore) ListNotificationTemplates(ctx context.Context, chain model.BusinessChain) ([]model.NotificationTemplate, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, business_chain, channel, subject, body_template, enabled, created_at
		 FROM notification_templates WHERE business_chain = ? ORDER BY created_at DESC`, chain)
	if err != nil {
		return nil, fmt.Errorf("list notification templates: %w", err)
	}
	defer rows.Close()

	var templates []model.NotificationTemplate
	for rows.Next() {
		var t model.NotificationTemplate
		if err := rows.Scan(&t.ID, &t.Name, &t.BusinessChain, &t.Channel, &t.Subject,
			&t.BodyTemplate, &t.Enabled, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan notification template: %w", err)
		}
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

func (s *SqliteStore) CreateNotificationLog(ctx context.Context, l *model.NotificationLog) error {
	l.ID = uuid.New().String()
	l.CreatedAt = time.Now()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO notification_logs (id, person_id, business_chain, template_id,
		 recipient_role, recipient_id, channel, status, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		l.ID, l.PersonID, l.BusinessChain, l.TemplateID, l.RecipientRole, l.RecipientID,
		l.Channel, l.Status, l.CreatedAt)
	return err
}

func (s *SqliteStore) UpdateNotificationStatus(ctx context.Context, logID string, status string, sentAt, readAt *time.Time) error {
	setClauses := []string{"status = ?"}
	args := []any{status}
	if sentAt != nil {
		setClauses = append(setClauses, "sent_at = ?")
		args = append(args, sentAt.Format("2006-01-02 15:04:05"))
	}
	if readAt != nil {
		setClauses = append(setClauses, "read_at = ?")
		args = append(args, readAt.Format("2006-01-02 15:04:05"))
	}
	args = append(args, logID)
	_, err := s.db.ExecContext(ctx,
		fmt.Sprintf("UPDATE notification_logs SET %s WHERE id = ?", fmt.Sprintf("%s", setClauses[0])), args...)
	return err
}

func (s *SqliteStore) ListNotificationLogs(ctx context.Context, personID string, chain model.BusinessChain, limit int) ([]model.NotificationLog, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, person_id, business_chain, template_id, recipient_role, recipient_id,
		 channel, status, sent_at, read_at, error_message, created_at
		 FROM notification_logs WHERE person_id = ? AND business_chain = ?
		 ORDER BY created_at DESC LIMIT ?`, personID, chain, limit)
	if err != nil {
		return nil, fmt.Errorf("list notification logs: %w", err)
	}
	defer rows.Close()

	var logs []model.NotificationLog
	for rows.Next() {
		var l model.NotificationLog
		var sentRaw, readRaw sql.NullString
		if err := rows.Scan(&l.ID, &l.PersonID, &l.BusinessChain, &l.TemplateID, &l.RecipientRole,
			&l.RecipientID, &l.Channel, &l.Status, &sentRaw, &readRaw, &l.ErrorMessage, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan notification log: %w", err)
		}
		if sentRaw.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", sentRaw.String)
			l.SentAt = &t
		}
		if readRaw.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", readRaw.String)
			l.ReadAt = &t
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
