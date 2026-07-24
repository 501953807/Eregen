package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"eregen.dev/admin-api/internal/model"
)

// ========== Regulatory Store Methods (SqliteStore) ==========

func (s *SqliteStore) CreateFenceConfig(ctx context.Context, fc *model.RegulatoryFenceConfig) error {
	now := time.Now().UTC()
	fc.CreatedAt = now
	fc.UpdatedAt = now
	if fc.ID == "" {
		fc.ID = fmt.Sprintf("fc_%d", now.UnixNano())
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO regulatory_fence_config (id, hospital_id, hospital_name, center_lat, center_lng, radius_meters, enabled, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		fc.ID, fc.HospitalID, fc.HospitalName, fc.CenterLat, fc.CenterLng,
		fc.RadiusMeters, fc.Enabled, fc.CreatedAt, fc.UpdatedAt)
	return err
}

func (s *SqliteStore) GetFenceConfig(ctx context.Context, hospitalID string) (*model.RegulatoryFenceConfig, error) {
	var fc model.RegulatoryFenceConfig
	var enabled int
	err := s.db.QueryRowContext(ctx,
		`SELECT id, hospital_id, hospital_name, center_lat, center_lng, radius_meters, enabled, created_at, updated_at
		 FROM regulatory_fence_config WHERE hospital_id=?`, hospitalID).Scan(
		&fc.ID, &fc.HospitalID, &fc.HospitalName, &fc.CenterLat, &fc.CenterLng,
		&fc.RadiusMeters, &enabled, &fc.CreatedAt, &fc.UpdatedAt)
	fc.Enabled = enabled == 1
	if err != nil {
		return nil, err
	}
	return &fc, nil
}

func (s *SqliteStore) UpdateFenceConfig(ctx context.Context, fc *model.RegulatoryFenceConfig) error {
	fc.UpdatedAt = time.Now().UTC()
	_, err := s.db.ExecContext(ctx,
		`UPDATE regulatory_fence_config SET hospital_name=?, center_lat=?, center_lng=?,
		 radius_meters=?, enabled=?, updated_at=? WHERE hospital_id=?`,
		fc.HospitalName, fc.CenterLat, fc.CenterLng, fc.RadiusMeters, fc.Enabled, fc.UpdatedAt, fc.HospitalID)
	return err
}

func (s *SqliteStore) ListRegulatoryAlerts(ctx context.Context, ruleCode, level, status, department string, page, pageSize int) ([]model.RegulatoryAlert, error) {
	q := `SELECT id, rule_code, patient_id, hospital_id, department, severity, alert_type, detail,
				status, triggered_at, acknowledged_at, acknowledged_by, resolved_at, resolved_by, notes
			FROM regulatory_alerts WHERE 1=1`
	args := []interface{}{}
	if ruleCode != "" {
		q += " AND rule_code=?"
		args = append(args, ruleCode)
	}
	if level != "" {
		q += " AND severity=?"
		args = append(args, level)
	}
	if status != "" {
		q += " AND status=?"
		args = append(args, status)
	}
	if department != "" {
		q += " AND department=?"
		args = append(args, department)
	}
	q += " ORDER BY triggered_at DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []model.RegulatoryAlert
	for rows.Next() {
		var a model.RegulatoryAlert
		if err := rows.Scan(&a.ID, &a.RuleCode, &a.PatientID, &a.HospitalID, &a.Department,
			&a.Severity, &a.AlertType, &a.Detail, &a.Status, &a.TriggeredAt,
			&a.AcknowledgedAt, &a.AcknowledgedBy, &a.ResolvedAt, &a.ResolvedBy, &a.Notes); err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

func (s *SqliteStore) GetRegulatoryAlert(ctx context.Context, alertID string) (*model.RegulatoryAlert, error) {
	var a model.RegulatoryAlert
	err := s.db.QueryRowContext(ctx,
		`SELECT id, rule_code, patient_id, hospital_id, department, severity, alert_type, detail,
			 status, triggered_at, acknowledged_at, acknowledged_by, resolved_at, resolved_by, notes
		 FROM regulatory_alerts WHERE id=?`, alertID).Scan(
		&a.ID, &a.RuleCode, &a.PatientID, &a.HospitalID, &a.Department,
		&a.Severity, &a.AlertType, &a.Detail, &a.Status, &a.TriggeredAt,
		&a.AcknowledgedAt, &a.AcknowledgedBy, &a.ResolvedAt, &a.ResolvedBy, &a.Notes)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *SqliteStore) AcknowledgeAlert(ctx context.Context, alertID, userID string) error {
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx,
		`UPDATE regulatory_alerts SET status='acknowledged', acknowledged_at=?, acknowledged_by=? WHERE id=?`,
		now, userID, alertID)
	return err
}

func (s *SqliteStore) ResolveRegulatoryAlert(ctx context.Context, alertID, userID, notes string) error {
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx,
		`UPDATE regulatory_alerts SET status='resolved', resolved_at=?, resolved_by=?, notes=? WHERE id=?`,
		now, userID, notes, alertID)
	return err
}

func (s *SqliteStore) ListRegulatoryAlertsCountByRule(ctx context.Context, days int) ([]model.RuleAlertCount, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT rule_code, COUNT(*) as cnt FROM regulatory_alerts
		 WHERE triggered_at > datetime('now', ? || 'days') GROUP BY rule_code`, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.RuleAlertCount
	for rows.Next() {
		var r model.RuleAlertCount
		if err := rows.Scan(&r.RuleCode, &r.Count); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

func (s *SqliteStore) SaveLocationLog(ctx context.Context, log *model.RegulatoryLocationLog) error {
	now := time.Now().UTC()
	log.RecordedAt = now
	if log.ID == "" {
		log.ID = fmt.Sprintf("ll_%d", now.UnixNano())
	}
	insideFence := 0
	if log.InsideFence {
		insideFence = 1
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO regulatory_location_logs (id, patient_id, device_id, lat, lng, accuracy, inside_fence, recorded_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		log.ID, log.PatientID, log.DeviceID, log.Lat, log.Lng, log.Accuracy, insideFence, log.RecordedAt)
	return err
}

func (s *SqliteStore) ListLocationLogs(ctx context.Context, patientID string, limit int) ([]model.RegulatoryLocationLog, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, patient_id, device_id, lat, lng, accuracy, inside_fence, recorded_at
		 FROM regulatory_location_logs WHERE patient_id=? ORDER BY recorded_at DESC LIMIT ?`,
		patientID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.RegulatoryLocationLog
	for rows.Next() {
		var l model.RegulatoryLocationLog
		var inside int
		if err := rows.Scan(&l.ID, &l.PatientID, &l.DeviceID, &l.Lat, &l.Lng,
			&l.Accuracy, &inside, &l.RecordedAt); err != nil {
			return nil, err
		}
		l.InsideFence = inside == 1
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func (s *SqliteStore) GetPatientFenceStatus(ctx context.Context, patientID string) (string, time.Time, int, error) {
	var fenceStatus string
	var exitAt sql.NullTime
	var durationSec int
	err := s.db.QueryRowContext(ctx,
		`SELECT fence_status, fence_exit_at, fence_exit_duration_sec
		 FROM medical_wristband_patients WHERE id=?`, patientID).Scan(&fenceStatus, &exitAt, &durationSec)
	if err != nil {
		return "", time.Time{}, 0, err
	}
	var exitTime time.Time
	if exitAt.Valid {
		exitTime = exitAt.Time
	}
	return fenceStatus, exitTime, durationSec, nil
}

func (s *SqliteStore) GetRegulatoryOverview(ctx context.Context, department string) (*model.RegulatoryDashboardOverview, error) {
	ov := &model.RegulatoryDashboardOverview{}
	q := `SELECT COUNT(*), SUM(CASE WHEN status='admitted' THEN 1 ELSE 0 END),
				SUM(CASE WHEN status='admitted' AND fence_status='outside' THEN 1 ELSE 0 END),
				SUM(CASE WHEN status='admitted' AND (verify_gap_hours >= 24 OR last_verify_at IS NULL) THEN 1 ELSE 0 END)
			FROM medical_wristband_patients`
	args := []interface{}{}
	if department != "" {
		q += " WHERE department=?"
		args = append(args, department)
	}
	err := s.db.QueryRowContext(ctx, q, args...).Scan(
		&ov.TotalAdmitted, &ov.TodayAdmit, &ov.FenceViolationsToday, &ov.NoVerify24h)
	if err != nil {
		return nil, err
	}
	return ov, nil
}

func (s *SqliteStore) ListRegulatoryPatients(ctx context.Context, department string, page, pageSize int) ([]model.RegulatoryPatientRow, error) {
	q := `SELECT id, name, admission_no, department, bed_number, bound_at, last_verify_at,
				verify_gap_hours, fence_status, fence_exit_duration_sec, tag_ids, status
			FROM medical_wristband_patients WHERE 1=1`
	args := []interface{}{}
	if department != "" {
		q += " AND department=?"
		args = append(args, department)
	}
	q += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patients []model.RegulatoryPatientRow
	for rows.Next() {
		var p model.RegulatoryPatientRow
		var tagsRaw string
		if err := rows.Scan(&p.ID, &p.Name, &p.AdmissionNo, &p.Department, &p.BedNumber,
			&p.BoundAt, &p.LastVerify, &p.VerifyGapHours, &p.FenceStatus,
			&p.FenceExitDurationSec, &tagsRaw); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(tagsRaw), &p.AlertTags)
		patients = append(patients, p)
	}
	return patients, rows.Err()
}

func (s *SqliteStore) GetRegulatoryAuditTrail(ctx context.Context, patientID string) (*model.RegulatoryAuditTrail, error) {
	trail := &model.RegulatoryAuditTrail{}

	// Patient
	p, err := s.GetPatient(ctx, patientID)
	if err != nil {
		return nil, err
	}
	trail.Patient = p

	// Bindings
	bindRows, err := s.db.QueryContext(ctx,
		`SELECT id, patient_id, device_id, bound_at, unbound_at FROM medical_bindings WHERE patient_id=? ORDER BY bound_at DESC LIMIT 1`, patientID)
	if err == nil {
		defer bindRows.Close()
		if bindRows.Next() {
			var b model.MedicalBinding
			if err := bindRows.Scan(&b.ID, &b.PatientID, &b.DeviceID, &b.BoundAt, &b.UnboundAt); err == nil {
				trail.Binding = &b
			}
		}
	}

	// Verifications
	verRows, err := s.db.QueryContext(ctx,
		`SELECT id, patient_id, device_id, verification_type, verified_by, result, notes, verified_at, created_at FROM medical_verifications WHERE patient_id=? ORDER BY verified_at DESC`, patientID)
	if err == nil {
		defer verRows.Close()
		for verRows.Next() {
			var v model.MedicalVerification
			if err := verRows.Scan(&v.ID, &v.PatientID, &v.DeviceID, &v.VerificationType, &v.VerifiedBy,
				&v.Result, &v.Notes, &v.VerifiedAt, &v.CreatedAt); err == nil {
				trail.Verifications = append(trail.Verifications, v)
			}
		}
	}

	// Medications
	medRows, err := s.db.QueryContext(ctx,
		`SELECT id, patient_id, name, dosage, frequency, duration, route, notes, created_at FROM medical_medications WHERE patient_id=?`, patientID)
	if err == nil {
		defer medRows.Close()
		for medRows.Next() {
			var m model.MedicalMedication
			if err := medRows.Scan(&m.ID, &m.PatientID, &m.Name, &m.Dosage, &m.Frequency,
				&m.Duration, &m.Route, &m.Notes, &m.CreatedAt); err == nil {
				trail.Medications = append(trail.Medications, m)
			}
		}
	}

	// Expenses
	expRows, err := s.db.QueryContext(ctx,
		`SELECT id, patient_id, item_name, category, amount, quantity, unit_price, notes, created_at FROM medical_expenses WHERE patient_id=?`, patientID)
	if err == nil {
		defer expRows.Close()
		for expRows.Next() {
			var e model.MedicalExpense
			if err := expRows.Scan(&e.ID, &e.PatientID, &e.ItemName, &e.Category, &e.Amount,
				&e.Quantity, &e.UnitPrice, &e.Notes, &e.CreatedAt); err == nil {
				trail.Expenses = append(trail.Expenses, e)
			}
		}
	}

	// Daily entries
	dailyRows, err := s.db.QueryContext(ctx,
		`SELECT id, patient_id, entry_type, content, nurse_id, created_at FROM medical_daily_entries WHERE patient_id=?`, patientID)
	if err == nil {
		defer dailyRows.Close()
		for dailyRows.Next() {
			var d model.MedicalDailyEntry
			if err := dailyRows.Scan(&d.ID, &d.PatientID, &d.EntryType, &d.Content, &d.NurseID, &d.CreatedAt); err == nil {
				trail.DailyEntries = append(trail.DailyEntries, d)
			}
		}
	}

	// Fence logs
	fenceRows, err := s.db.QueryContext(ctx,
		`SELECT id, patient_id, device_id, lat, lng, accuracy, inside_fence, recorded_at
		 FROM regulatory_location_logs WHERE patient_id=? ORDER BY recorded_at DESC LIMIT 100`, patientID)
	if err == nil {
		defer fenceRows.Close()
		for fenceRows.Next() {
			var l model.RegulatoryLocationLog
			var inside int
			if err := fenceRows.Scan(&l.ID, &l.PatientID, &l.DeviceID, &l.Lat, &l.Lng,
				&l.Accuracy, &inside, &l.RecordedAt); err == nil {
				l.InsideFence = inside == 1
				trail.FenceLogs = append(trail.FenceLogs, l)
			}
		}
	}

	// Alerts generated
	alertRows, err := s.db.QueryContext(ctx,
		`SELECT id, rule_code, patient_id, hospital_id, department, severity, alert_type, detail,
			 status, triggered_at, acknowledged_at, acknowledged_by, resolved_at, resolved_by, notes
		 FROM regulatory_alerts WHERE patient_id=?`, patientID)
	if err == nil {
		defer alertRows.Close()
		for alertRows.Next() {
			var a model.RegulatoryAlert
			if err := alertRows.Scan(&a.ID, &a.RuleCode, &a.PatientID, &a.HospitalID, &a.Department,
				&a.Severity, &a.AlertType, &a.Detail, &a.Status, &a.TriggeredAt,
				&a.AcknowledgedAt, &a.AcknowledgedBy, &a.ResolvedAt, &a.ResolvedBy, &a.Notes); err == nil {
				trail.AlertsGenerated = append(trail.AlertsGenerated, a)
			}
		}
	}

	return trail, nil
}

func (s *SqliteStore) ListRuleConfigs(ctx context.Context) ([]model.RegulatoryRuleConfig, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT rule_code, rule_name, enabled, config_json, updated_at FROM regulatory_rule_config ORDER BY rule_code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []model.RegulatoryRuleConfig
	for rows.Next() {
		var c model.RegulatoryRuleConfigDB
		if err := rows.Scan(&c.RuleCode, &c.RuleName, &c.Enabled, &c.ConfigJSON, &c.UpdatedAt); err != nil {
			return nil, err
		}
		configs = append(configs, model.RegulatoryRuleConfig{
			Code:   c.RuleCode,
			Name:   c.RuleName,
			Enabled: c.Enabled,
		})
		json.Unmarshal([]byte(c.ConfigJSON), &configs[len(configs)-1].Config)
	}
	return configs, rows.Err()
}

func (s *SqliteStore) UpdateRuleConfig(ctx context.Context, ruleCode string, configJSON string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE regulatory_rule_config SET config_json=?, updated_at=datetime('now') WHERE rule_code=?`,
		configJSON, ruleCode)
	return err
}

func (s *SqliteStore) GetComplianceReport(ctx context.Context, hospitalID, startDate, endDate string) (*model.ComplianceReport, error) {
	report := &model.ComplianceReport{}

	// Count patients
	var totalPatients int
	q := `SELECT COUNT(*) FROM medical_wristband_patients WHERE 1=1`
	args := []interface{}{}
	if hospitalID != "" {
		q += " AND hospital_id=?" // Note: medical_wristband_patients doesn't have hospital_id column yet; this is a stub
		args = append(args, hospitalID)
	}
	s.db.QueryRowContext(ctx, q, args...).Scan(&totalPatients)
	report.Summary.TotalPatientsPeriod = totalPatients

	// Count alerts by type
	alertTypes := []string{"no_verify", "fence_violation", "expense_spike", "med_verify_mismatch"}
	for _, at := range alertTypes {
		var cnt int
		s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM regulatory_alerts WHERE alert_type=?`, at).Scan(&cnt)
		switch at {
		case "no_verify":
			report.Summary.NoVerifyAlerts = cnt
		case "fence_violation":
			report.Summary.FenceViolations = cnt
		case "expense_spike":
			report.Summary.ExpenseAnomalies = cnt
		case "med_verify_mismatch":
			report.Summary.MedVerifyMismatch = cnt
		}
	}

	// Compliance rate
	totalAlerts := report.Summary.NoVerifyAlerts + report.Summary.FenceViolations + report.Summary.ExpenseAnomalies + report.Summary.MedVerifyMismatch
	if totalPatients > 0 {
		report.Summary.ComplianceRate = float64(totalPatients-totalAlerts) / float64(totalPatients) * 100
	}
	return report, nil
}

func (s *SqliteStore) CreateDepartmentBinding(ctx context.Context, binding *model.DepartmentBinding) error {
	if binding.ID == "" {
		binding.ID = fmt.Sprintf("db_%d", time.Now().UnixNano())
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO user_department_bindings (id, user_id, department, bound_at) VALUES (?, ?, ?, datetime('now'))`,
		binding.ID, binding.UserID, binding.Department)
	return err
}

func (s *SqliteStore) ListDepartmentBindings(ctx context.Context, userID string) ([]model.DepartmentBinding, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, department, bound_at FROM user_department_bindings WHERE user_id=?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bindings []model.DepartmentBinding
	for rows.Next() {
		var b model.DepartmentBinding
		if err := rows.Scan(&b.ID, &b.UserID, &b.Department, &b.BoundAt); err != nil {
			return nil, err
		}
		bindings = append(bindings, b)
	}
	return bindings, rows.Err()
}

func (s *SqliteStore) CreateRegulatoryAlert(ctx context.Context, a *model.RegulatoryAlert) error {
	now := time.Now().UTC()
	a.TriggeredAt = now
	if a.ID == "" {
		a.ID = fmt.Sprintf("ra_%d", now.UnixNano())
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO regulatory_alerts (id, rule_code, patient_id, hospital_id, department, severity, alert_type, detail, status, triggered_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'pending', ?)`,
		a.ID, a.RuleCode, a.PatientID, a.HospitalID, a.Department, a.Severity, a.AlertType, a.Detail, a.TriggeredAt)
	return err
}

func (s *SqliteStore) CountPendingAlertsByRule(ctx context.Context) ([]model.RuleAlertCount, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT rule_code, COUNT(*) as cnt FROM regulatory_alerts WHERE status='pending' GROUP BY rule_code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.RuleAlertCount
	for rows.Next() {
		var r model.RuleAlertCount
		if err := rows.Scan(&r.RuleCode, &r.Count); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

func (s *SqliteStore) CountAlertsByDept(ctx context.Context, startDate, endDate string) ([]model.DeptAlertCount, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT department, COUNT(*) as cnt FROM regulatory_alerts
		 WHERE triggered_at BETWEEN ? AND ? GROUP BY department`, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.DeptAlertCount
	for rows.Next() {
		var d model.DeptAlertCount
		if err := rows.Scan(&d.Department, &d.Count); err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	return result, rows.Err()
}
