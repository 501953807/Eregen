package store

import (
	"context"
	"eregen.dev/admin-api/internal/model"
	"fmt"
	"time"
)


func (p *PostgresStore) CreateFenceConfig(ctx context.Context, fc *model.RegulatoryFenceConfig) error {

	now := time.Now().UTC(); fc.CreatedAt = now; fc.UpdatedAt = now

	if fc.ID == "" { fc.ID = fmt.Sprintf("fc_%d", now.UnixNano()) }

	_, err := p.db.ExecContext(ctx, `INSERT INTO regulatory_fence_config (id,hospital_id,hospital_name,center_lat,center_lng,radius_meters,enabled,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, fc.ID, fc.HospitalID, fc.HospitalName, fc.CenterLat, fc.CenterLng, fc.RadiusMeters, fc.Enabled, fc.CreatedAt, fc.UpdatedAt)

	return err

}

func (p *PostgresStore) GetFenceConfig(ctx context.Context, hospitalID string) (*model.RegulatoryFenceConfig, error) {

	var fc model.RegulatoryFenceConfig; var enabled int

	err := p.db.QueryRowContext(ctx, `SELECT id,hospital_id,hospital_name,center_lat,center_lng,radius_meters,enabled,created_at,updated_at FROM regulatory_fence_config WHERE hospital_id=$1`, hospitalID).Scan(&fc.ID, &fc.HospitalID, &fc.HospitalName, &fc.CenterLat, &fc.CenterLng, &fc.RadiusMeters, &enabled, &fc.CreatedAt, &fc.UpdatedAt)

	fc.Enabled = enabled == 1; return &fc, err

}

func (p *PostgresStore) UpdateFenceConfig(ctx context.Context, fc *model.RegulatoryFenceConfig) error {

	fc.UpdatedAt = time.Now().UTC()

	_, err := p.db.ExecContext(ctx, `UPDATE regulatory_fence_config SET hospital_name=$1,center_lat=$2,center_lng=$3,radius_meters=$4,enabled=$5,updated_at=$6 WHERE hospital_id=$7`, fc.HospitalName, fc.CenterLat, fc.CenterLng, fc.RadiusMeters, fc.Enabled, fc.UpdatedAt, fc.HospitalID)

	return err

}

func (p *PostgresStore) ListRegulatoryAlerts(ctx context.Context, ruleCode, level, status, department string, page, pageSize int) ([]model.RegulatoryAlert, error) {

	q := `SELECT id, rule_code, COALESCE(patient_id::text,''), hospital_id, department, severity, alert_type, detail, status, triggered_at, acknowledged_at, acknowledged_by, resolved_at, resolved_by, notes

		 FROM regulatory_alerts WHERE 1=1`

	args := []interface{}{}

	if ruleCode != "" { q += " AND rule_code=$" + fmt.Sprintf("%d", len(args)+1); args = append(args, ruleCode) }

	if level != "" { q += " AND severity=$" + fmt.Sprintf("%d", len(args)+1); args = append(args, level) }

	if status != "" { q += " AND status=$" + fmt.Sprintf("%d", len(args)+1); args = append(args, status) }

	if department != "" { q += " AND department=$" + fmt.Sprintf("%d", len(args)+1); args = append(args, department) }

	q += " ORDER BY triggered_at DESC LIMIT $" + fmt.Sprintf("%d", len(args)+1) + " OFFSET $" + fmt.Sprintf("%d", len(args)+2)

	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := p.db.QueryContext(ctx, q, args...)

	if err != nil { return nil, err }

	defer rows.Close()

	var items []model.RegulatoryAlert

	for rows.Next() {

		var a model.RegulatoryAlert

		if err := rows.Scan(&a.ID, &a.RuleCode, &a.PatientID, &a.HospitalID, &a.Department, &a.Severity, &a.AlertType, &a.Detail, &a.Status, &a.TriggeredAt, &a.AcknowledgedAt, &a.AcknowledgedBy, &a.ResolvedAt, &a.ResolvedBy, &a.Notes); err != nil { return nil, err }

		items = append(items, a)

	}

	return items, rows.Err()

}

func (p *PostgresStore) GetRegulatoryAlert(ctx context.Context, alertID string) (*model.RegulatoryAlert, error) {

	var a model.RegulatoryAlert

	err := p.db.QueryRowContext(ctx, `SELECT id, rule_code, COALESCE(patient_id::text,''), hospital_id, department, severity, alert_type, detail, status, triggered_at, acknowledged_at, acknowledged_by, resolved_at, resolved_by, notes FROM regulatory_alerts WHERE id=$1`, alertID).Scan(&a.ID, &a.RuleCode, &a.PatientID, &a.HospitalID, &a.Department, &a.Severity, &a.AlertType, &a.Detail, &a.Status, &a.TriggeredAt, &a.AcknowledgedAt, &a.AcknowledgedBy, &a.ResolvedAt, &a.ResolvedBy, &a.Notes)

	return &a, err

}

func (p *PostgresStore) AcknowledgeAlert(ctx context.Context, alertID, userID string) error {

	_, err := p.db.ExecContext(ctx, `UPDATE regulatory_alerts SET status='acknowledged',acknowledged_at=NOW(),acknowledged_by=$1 WHERE id=$2`, userID, alertID)

	return err

}

func (p *PostgresStore) ResolveRegulatoryAlert(ctx context.Context, alertID, userID, notes string) error {

	_, err := p.db.ExecContext(ctx, `UPDATE regulatory_alerts SET status='resolved',resolved_at=NOW(),resolved_by=$1,notes=$2 WHERE id=$3`, userID, notes, alertID)

	return err

}

func (p *PostgresStore) ListRegulatoryAlertsCountByRule(ctx context.Context, days int) ([]model.RuleAlertCount, error) {

	rows, err := p.db.QueryContext(ctx, `SELECT rule_code,COUNT(*) FROM regulatory_alerts WHERE triggered_at > NOW()-($1||' days') GROUP BY rule_code`, days)

	if err != nil { return nil, err }

	defer rows.Close()

	var result []model.RuleAlertCount

	for rows.Next() { var r model.RuleAlertCount; rows.Scan(&r.RuleCode, &r.Count); result = append(result, r) }

	return result, rows.Err()

}

func (p *PostgresStore) SaveLocationLog(ctx context.Context, log *model.RegulatoryLocationLog) error {

	log.RecordedAt = time.Now().UTC()

	if log.ID == "" { log.ID = fmt.Sprintf("ll_%d", log.RecordedAt.UnixNano()) }

	insideFence := 0; if log.InsideFence { insideFence = 1 }

	_, err := p.db.ExecContext(ctx, `INSERT INTO regulatory_location_logs (id,patient_id,device_id,lat,lng,accuracy,inside_fence,recorded_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, log.ID, log.PatientID, log.DeviceID, log.Lat, log.Lng, log.Accuracy, insideFence, log.RecordedAt)

	return err

}

func (p *PostgresStore) ListLocationLogs(ctx context.Context, patientID string, limit int) ([]model.RegulatoryLocationLog, error) {

	rows, err := p.db.QueryContext(ctx, `SELECT id, patient_id, device_id, lat, lng, accuracy, inside_fence, recorded_at FROM regulatory_location_logs WHERE patient_id=$1 ORDER BY recorded_at DESC LIMIT $2`, patientID, limit)

	if err != nil { return nil, err }

	defer rows.Close()

	var items []model.RegulatoryLocationLog

	for rows.Next() {

		var l model.RegulatoryLocationLog

		var acc float64

		var insideFence int

		if err := rows.Scan(&l.ID, &l.PatientID, &l.DeviceID, &l.Lat, &l.Lng, &acc, &insideFence, &l.RecordedAt); err != nil { return nil, err }

		l.Accuracy = &acc

		l.InsideFence = insideFence == 1

		items = append(items, l)

	}

	return items, rows.Err()

}

func (p *PostgresStore) GetPatientFenceStatus(ctx context.Context, patientID string) (string, time.Time, int, error) {

	var fenceStatus string

	var latestAt time.Time

	var exitSec int

	err := p.db.QueryRowContext(ctx, `

		SELECT COALESCE((SELECT status FROM regulatory_fence_config WHERE hospital_id IN (SELECT hospital_id FROM medical_bindings mb JOIN medical_wristband_patients p ON p.id=mb.patient_id WHERE p.id=$1) LIMIT 1),'unknown'),

		       COALESCE((SELECT MAX(recorded_at) FROM regulatory_location_logs WHERE patient_id=$1), NOW()),

		       COALESCE((SELECT EXTRACT(EPOCH FROM (NOW() - MAX(recorded_at)))::int FROM regulatory_location_logs WHERE patient_id=$1), 0)

	`, patientID).Scan(&fenceStatus, &latestAt, &exitSec)

	return fenceStatus, latestAt, exitSec, err

}

func (p *PostgresStore) GetRegulatoryOverview(ctx context.Context, department string) (*model.RegulatoryDashboardOverview, error) {

	overview := &model.RegulatoryDashboardOverview{}

	if department != "" {

		p.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM medical_wristband_patients WHERE department=$1 AND status='admitted'`, department).Scan(&overview.TotalAdmitted)

		p.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM hospital_admissions WHERE department=$1 AND admitted_at::date=$2`, department, time.Now().Format("2006-01-02")).Scan(&overview.TodayAdmit)

		p.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM hospital_admissions WHERE department=$1 AND discharged_at IS NOT NULL AND discharged_at::date=$2`, department, time.Now().Format("2006-01-02")).Scan(&overview.TodayDischarge)

	} else {

		p.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM medical_wristband_patients WHERE status='admitted'`).Scan(&overview.TotalAdmitted)

	}

	rows, _ := p.db.QueryContext(ctx, `

		SELECT p.department, COUNT(DISTINCT p.id), COUNT(DISTINCT a.id)

		FROM medical_wristband_patients p LEFT JOIN regulatory_alerts a ON a.patient_id=p.id AND a.status='pending'

		WHERE p.status='admitted' GROUP BY p.department`)

	defer rows.Close()

	for rows.Next() {

		var s model.RegulatoryDeptStat

		if err := rows.Scan(&s.Name, &s.Count, &s.AlertCount); err == nil { overview.ByDepartment = append(overview.ByDepartment, s) }

	}

	p.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM regulatory_location_logs WHERE recorded_at > NOW()-'24 hours' AND inside_fence=false`).Scan(&overview.FenceViolationsToday)

	p.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM medical_verifications v JOIN medical_wristband_patients p ON p.id=v.patient_id WHERE p.status='admitted' AND v.status='pending' AND v.created_at > NOW()-'24 hours'`).Scan(&overview.NoVerify24h)

	return overview, nil

}

func (p *PostgresStore) ListRegulatoryPatients(ctx context.Context, department string, page, pageSize int) ([]model.RegulatoryPatientRow, error) {

	q := `SELECT p.id, p.name, p.admission_no, p.department, p.bed_number, p.admitted_at

			FROM medical_wristband_patients p WHERE p.status='admitted'`

	args := []interface{}{}

	if department != "" { q += " AND p.department=$" + fmt.Sprintf("%d", len(args)+1); args = append(args, department) }

	q += " ORDER BY p.admitted_at DESC LIMIT $" + fmt.Sprintf("%d", len(args)+1) + " OFFSET $" + fmt.Sprintf("%d", len(args)+2)

	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := p.db.QueryContext(ctx, q, args...)

	if err != nil { return nil, err }

	defer rows.Close()

	var items []model.RegulatoryPatientRow

	for rows.Next() {

		var r model.RegulatoryPatientRow

		if err := rows.Scan(&r.ID, &r.Name, &r.AdmissionNo, &r.Department, &r.BedNumber, &r.BoundAt); err != nil { return nil, err }

		items = append(items, r)

	}

	return items, rows.Err()

}

func (p *PostgresStore) GetRegulatoryAuditTrail(ctx context.Context, patientID string) (*model.RegulatoryAuditTrail, error) {

	trail := &model.RegulatoryAuditTrail{}

	var admission model.HospitalAdmission

	err := p.db.QueryRowContext(ctx, `SELECT id, patient_id, admission_no, bed_no, department, diagnosis, emergency_contact, allergies, admitted_at, discharged_at FROM hospital_admissions WHERE patient_id=$1 ORDER BY admitted_at DESC LIMIT 1`, patientID).Scan(

		&admission.ID, &admission.PatientID, &admission.AdmissionNo, &admission.BedNo, &admission.Department, &admission.Diagnosis, &admission.EmergencyContact, &admission.Allergies, &admission.AdmittedAt, &admission.DischargedAt)

	if err != nil { return nil, err }

	var patient model.MedicalPatient

	p.db.QueryRowContext(ctx, `SELECT id, admission_no, name, gender, age, department, bed_number, blood_type, allergies, special_conditions, status, created_at, updated_at FROM medical_wristband_patients WHERE id=$1`, patientID).Scan(

		&patient.ID, &patient.AdmissionNo, &patient.Name, &patient.Gender, &patient.Age, &patient.Department, &patient.BedNumber, &patient.BloodType, &patient.Allergies, &patient.SpecialConditions, &patient.Status, &patient.CreatedAt, &patient.UpdatedAt)

	trail.Patient = &patient

	rows, _ := p.db.QueryContext(ctx, `SELECT id, patient_id, device_id, verification_type, result, matched, verified_by, verified_at, notes, created_at FROM medical_verifications WHERE patient_id=$1 ORDER BY created_at DESC`, patientID)

	defer rows.Close()

	for rows.Next() {

		var v model.MedicalVerification

		if err := rows.Scan(&v.ID, &v.PatientID, &v.DeviceID, &v.VerificationType, &v.Result, &v.Matched, &v.VerifiedBy, &v.VerifiedAt, &v.Notes, &v.CreatedAt); err == nil { trail.Verifications = append(trail.Verifications, v) }

	}

	rows2, _ := p.db.QueryContext(ctx, `SELECT id, patient_id, name, dosage, frequency, duration, notes, created_at, updated_at FROM medical_medications WHERE patient_id=$1`, patientID)

	defer rows2.Close()

	for rows2.Next() {

		var m model.MedicalMedication

		if err := rows2.Scan(&m.ID, &m.PatientID, &m.Name, &m.Dosage, &m.Frequency, &m.Duration, &m.Notes, &m.CreatedAt, &m.UpdatedAt); err == nil { trail.Medications = append(trail.Medications, m) }

	}

	rows3, _ := p.db.QueryContext(ctx, `SELECT id, patient_id, item_name, category, amount, quantity, unit_price, notes, created_at, updated_at FROM medical_expenses WHERE patient_id=$1 ORDER BY created_at DESC`, patientID)

	defer rows3.Close()

	for rows3.Next() {

		var e model.MedicalExpense

		if err := rows3.Scan(&e.ID, &e.PatientID, &e.ItemName, &e.Category, &e.Amount, &e.Quantity, &e.UnitPrice, &e.Notes, &e.CreatedAt, &e.UpdatedAt); err == nil { trail.Expenses = append(trail.Expenses, e) }

	}

	rows4, _ := p.db.QueryContext(ctx, `SELECT id, patient_id, entry_date, content, nurse_id, created_at, updated_at FROM medical_daily_entries WHERE patient_id=$1 ORDER BY entry_date DESC`, patientID)

	defer rows4.Close()

	for rows4.Next() {

		var d model.MedicalDailyEntry

		if err := rows4.Scan(&d.ID, &d.PatientID, &d.EntryDate, &d.Content, &d.NurseID, &d.CreatedAt, &d.UpdatedAt); err == nil { trail.DailyEntries = append(trail.DailyEntries, d) }

	}

	rows5, _ := p.db.QueryContext(ctx, `SELECT id, patient_id, device_id, lat, lng, accuracy, inside_fence, recorded_at FROM regulatory_location_logs WHERE patient_id=$1 ORDER BY recorded_at DESC LIMIT 50`, patientID)

	defer rows5.Close()

	for rows5.Next() {

		var l model.RegulatoryLocationLog

		var acc float64

		var insideFence int

		if err := rows5.Scan(&l.ID, &l.PatientID, &l.DeviceID, &l.Lat, &l.Lng, &acc, &insideFence, &l.RecordedAt); err == nil { l.Accuracy = &acc; l.InsideFence = insideFence == 1; trail.FenceLogs = append(trail.FenceLogs, l) }

	}

	rows6, _ := p.db.QueryContext(ctx, `SELECT id, rule_code, COALESCE(patient_id::text,''), hospital_id, department, severity, alert_type, detail, status, triggered_at FROM regulatory_alerts WHERE patient_id=$1 ORDER BY triggered_at DESC`, patientID)

	defer rows6.Close()

	for rows6.Next() {

		var a model.RegulatoryAlert

		if err := rows6.Scan(&a.ID, &a.RuleCode, &a.PatientID, &a.HospitalID, &a.Department, &a.Severity, &a.AlertType, &a.Detail, &a.Status, &a.TriggeredAt); err == nil { trail.AlertsGenerated = append(trail.AlertsGenerated, a) }

	}

	return trail, nil

}

func (p *PostgresStore) ListRuleConfigs(ctx context.Context) ([]model.RegulatoryRuleConfig, error) {

	rows, err := p.db.QueryContext(ctx, `SELECT rule_code,rule_name,enabled,config_json,updated_at FROM regulatory_rule_config ORDER BY rule_code`)

	if err != nil { return nil, err }

	defer rows.Close()

	var result []model.RegulatoryRuleConfig

	for rows.Next() {

		var cfg model.RegulatoryRuleConfigDB

		rows.Scan(&cfg.RuleCode, &cfg.RuleName, &cfg.Enabled, &cfg.ConfigJSON, &cfg.UpdatedAt)

		result = append(result, model.RegulatoryRuleConfig{Code: cfg.RuleCode, Name: cfg.RuleName, Enabled: cfg.Enabled})

	}

	return result, rows.Err()

}

func (p *PostgresStore) UpdateRuleConfig(ctx context.Context, ruleCode string, configJSON string) error {

	_, err := p.db.ExecContext(ctx, `UPDATE regulatory_rule_config SET config_json=$1,updated_at=NOW() WHERE rule_code=$2`, configJSON, ruleCode)

	return err

}

func (p *PostgresStore) GetComplianceReport(ctx context.Context, hospitalID, startDate, endDate string) (*model.ComplianceReport, error) {

	report := &model.ComplianceReport{}

	today := time.Now().UTC()

	qStart, qEnd := "2020-01-01", today.Format("2006-01-02")

	if startDate != "" { qStart = startDate }

	if endDate != "" { qEnd = endDate }

	p.db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT patient_id) FROM medical_wristband_patients WHERE admitted_at BETWEEN $1 AND $2 AND discharged_at IS NOT NULL`, qStart, qEnd).Scan(&report.Summary.TotalPatientsPeriod)

	p.db.QueryRowContext(ctx, `SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (discharged_at - admitted_at))/86400), 0)::float8 FROM hospital_admissions WHERE admitted_at BETWEEN $1 AND $2 AND discharged_at IS NOT NULL`, qStart, qEnd).Scan(&report.Summary.AvgStayDays)

	p.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM regulatory_alerts WHERE triggered_at BETWEEN $1 AND $2 AND alert_type='fence_violation'`, qStart, qEnd).Scan(&report.Summary.FenceViolations)

	p.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM regulatory_alerts WHERE triggered_at BETWEEN $1 AND $2 AND alert_type='no_verify'`, qStart, qEnd).Scan(&report.Summary.NoVerifyAlerts)

	if report.Summary.TotalPatientsPeriod > 0 {

		report.Summary.ComplianceRate = float64(report.Summary.TotalPatientsPeriod-report.Summary.FenceViolations-report.Summary.NoVerifyAlerts) / float64(report.Summary.TotalPatientsPeriod) * 100

	}

	rows, _ := p.db.QueryContext(ctx, `

		SELECT p.department, COUNT(DISTINCT p.id), COUNT(DISTINCT a.id),

		       ROUND(CAST((COUNT(DISTINCT p.id) - COUNT(DISTINCT a.id)) AS float) / NULLIF(COUNT(DISTINCT p.id), 0) * 100, 2)

		FROM medical_wristband_patients p LEFT JOIN regulatory_alerts a ON a.patient_id=p.id

		WHERE p.admitted_at BETWEEN $1 AND $2 GROUP BY p.department`, qStart, qEnd)

	defer rows.Close()

	for rows.Next() {

		var b model.ComplianceDeptBreakdown

		if err := rows.Scan(&b.Name, &b.TotalPatients, &b.Alerts, &b.ComplianceRate); err == nil { report.DepartmentBreakdown = append(report.DepartmentBreakdown, b) }

	}

	return report, nil

}

func (p *PostgresStore) ListDepartmentBindings(ctx context.Context, userID string) ([]model.DepartmentBinding, error) {

	rows, err := p.db.QueryContext(ctx, `SELECT id, user_id, department, bound_at FROM user_department_bindings WHERE user_id=$1 ORDER BY bound_at DESC`, userID)

	if err != nil { return nil, err }

	defer rows.Close()

	var items []model.DepartmentBinding

	for rows.Next() {

		var b model.DepartmentBinding

		if err := rows.Scan(&b.ID, &b.UserID, &b.Department, &b.BoundAt); err != nil { return nil, err }

		items = append(items, b)

	}

	return items, rows.Err()

}

func (p *PostgresStore) CountPendingAlertsByRule(ctx context.Context) ([]model.RuleAlertCount, error) {

	rows, err := p.db.QueryContext(ctx, `SELECT rule_code, COUNT(*) FROM regulatory_alerts WHERE status='pending' GROUP BY rule_code`)

	if err != nil { return nil, err }

	defer rows.Close()

	var items []model.RuleAlertCount

	for rows.Next() {

		var r model.RuleAlertCount

		if err := rows.Scan(&r.RuleCode, &r.Count); err != nil { return nil, err }

		items = append(items, r)

	}

	return items, rows.Err()

}

func (p *PostgresStore) CountAlertsByDept(ctx context.Context, startDate, endDate string) ([]model.DeptAlertCount, error) {

	qStart, qEnd := "2020-01-01", time.Now().Format("2006-01-02")

	if startDate != "" { qStart = startDate }

	if endDate != "" { qEnd = endDate }

	rows, err := p.db.QueryContext(ctx, `SELECT department, COUNT(*) FROM regulatory_alerts WHERE triggered_at BETWEEN $1 AND $2 GROUP BY department`, qStart, qEnd)

	if err != nil { return nil, err }

	defer rows.Close()

	var items []model.DeptAlertCount

	for rows.Next() {

		var d model.DeptAlertCount

		if err := rows.Scan(&d.Department, &d.Count); err != nil { return nil, err }

		items = append(items, d)

	}

	return items, rows.Err()

}

func (p *PostgresStore) CreateDepartmentBinding(ctx context.Context, binding *model.DepartmentBinding) error {

	if binding.ID == "" { binding.ID = fmt.Sprintf("db_%d", time.Now().UnixNano()) }

	_, err := p.db.ExecContext(ctx, `INSERT INTO user_department_bindings (id,user_id,department,bound_at) VALUES ($1,$2,$3,NOW())`, binding.ID, binding.UserID, binding.Department)

	return err

}

func (p *PostgresStore) CreateRegulatoryAlert(ctx context.Context, alert *model.RegulatoryAlert) error {

	alert.TriggeredAt = time.Now().UTC()

	if alert.ID == "" { alert.ID = fmt.Sprintf("ra_%d", alert.TriggeredAt.UnixNano()) }

	var patientID interface{} = nil

	if alert.PatientID != nil { patientID = *alert.PatientID }

	_, err := p.db.ExecContext(ctx, `INSERT INTO regulatory_alerts (id,rule_code,patient_id,hospital_id,department,severity,alert_type,detail,status,triggered_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'pending',$9)`, alert.ID, alert.RuleCode, patientID, alert.HospitalID, alert.Department, alert.Severity, alert.AlertType, alert.Detail, alert.TriggeredAt)

	return err

}



// ========== Community stub implementations (PostgresStore) ==========

