package store

import (
	"context"
	"database/sql"
	"eregen.dev/admin-api/internal/model"
	"fmt"
	"time"
)


// CreateExpense inserts a medical expense record.
func (s *SqliteStore) CreateExpense(ctx context.Context, e *model.MedicalExpense) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medical_expenses (id, patient_id, item_name, category, amount, quantity, unit_price, notes, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		e.ID, e.PatientID, e.ItemName, e.Category, e.Amount, e.Quantity, e.UnitPrice, e.Notes)
	return err
}

// ListExpenses returns expenses for a patient.
func (s *SqliteStore) ListExpenses(ctx context.Context, patientID string, page, pageSize int) ([]model.MedicalExpense, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, patient_id, item_name, category, amount, quantity, unit_price, notes, created_at, updated_at
		FROM medical_expenses WHERE patient_id=? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		patientID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("list expenses: %w", err)
	}
	defer rows.Close()

	var items []model.MedicalExpense
	for rows.Next() {
		var e model.MedicalExpense
		if err := rows.Scan(&e.ID, &e.PatientID, &e.ItemName, &e.Category, &e.Amount, &e.Quantity, &e.UnitPrice, &e.Notes, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan expense: %w", err)
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

// CreateMedication inserts a medication record.
func (s *SqliteStore) CreateMedication(ctx context.Context, m *model.MedicalMedication) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medical_medications (id, patient_id, name, dosage, frequency, duration, route, notes, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		m.ID, m.PatientID, m.Name, m.Dosage, m.Frequency, m.Duration, m.Route, m.Notes)
	return err
}

// ListMedications returns medications for a patient.
func (s *SqliteStore) ListMedications(ctx context.Context, patientID string) ([]model.MedicalMedication, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, patient_id, name, dosage, frequency, duration, route, notes, created_at, updated_at
		FROM medical_medications WHERE patient_id=? ORDER BY created_at DESC`, patientID)
	if err != nil {
		return nil, fmt.Errorf("list medications: %w", err)
	}
	defer rows.Close()

	var items []model.MedicalMedication
	for rows.Next() {
		var m model.MedicalMedication
		if err := rows.Scan(&m.ID, &m.PatientID, &m.Name, &m.Dosage, &m.Frequency, &m.Duration, &m.Route, &m.Notes, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan medication: %w", err)
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

// CreateTestResult inserts a test result record.
func (s *SqliteStore) CreateTestResult(ctx context.Context, r *model.MedicalTestResult) error {
	collectedAt := ""
	reportedAt := ""
	if r.CollectedAt != nil {
		collectedAt = r.CollectedAt.Format(time.RFC3339)
	}
	if r.ReportedAt != nil {
		reportedAt = r.ReportedAt.Format(time.RFC3339)
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medical_test_results (id, patient_id, test_name, result, reference_range, unit, collected_at, reported_at, notes, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		r.ID, r.PatientID, r.TestName, r.Result, r.ReferenceRange, r.Unit, collectedAt, reportedAt, r.Notes)
	return err
}

// ListTestResults returns test results for a patient.
func (s *SqliteStore) ListTestResults(ctx context.Context, patientID string) ([]model.MedicalTestResult, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, patient_id, test_name, result, reference_range, unit, collected_at, reported_at, notes, created_at, updated_at
		FROM medical_test_results WHERE patient_id=? ORDER BY collected_at DESC`, patientID)
	if err != nil {
		return nil, fmt.Errorf("list test results: %w", err)
	}
	defer rows.Close()

	var items []model.MedicalTestResult
	for rows.Next() {
		var t model.MedicalTestResult
		var collectedAt, reportedAt string
		if err := rows.Scan(&t.ID, &t.PatientID, &t.TestName, &t.Result, &t.ReferenceRange, &t.Unit, &collectedAt, &reportedAt, &t.Notes, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan test result: %w", err)
		}
		if collectedAt != "" {
			if ct, err := time.Parse(time.RFC3339, collectedAt); err == nil {
				t.CollectedAt = &ct
			}
		}
		if reportedAt != "" {
			if rt, err := time.Parse(time.RFC3339, reportedAt); err == nil {
				t.ReportedAt = &rt
			}
		}
		items = append(items, t)
	}
	return items, rows.Err()
}

// CreateDailyEntry inserts a daily medical entry.
func (s *SqliteStore) CreateDailyEntry(ctx context.Context, e *model.MedicalDailyEntry) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medical_daily_entries (id, patient_id, entry_date, entry_type, content, nurse_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		e.ID, e.PatientID, e.EntryDate, e.EntryType, e.Content, e.NurseID)
	return err
}

// ListDailyEntries returns daily entries for a patient.
func (s *SqliteStore) ListDailyEntries(ctx context.Context, patientID string, date string) ([]model.MedicalDailyEntry, error) {
	query := `SELECT id, patient_id, entry_date, entry_type, content, nurse_id, created_at, updated_at
		FROM medical_daily_entries WHERE patient_id=?`
	var rows *sql.Rows
	var err error
	if date != "" {
		rows, err = s.db.QueryContext(ctx, query+` AND entry_date=? ORDER BY created_at DESC`, patientID, date)
	} else {
		rows, err = s.db.QueryContext(ctx, query+` ORDER BY entry_date DESC, created_at DESC`, patientID)
	}
	if err != nil {
		return nil, fmt.Errorf("list daily entries: %w", err)
	}
	defer rows.Close()

	var items []model.MedicalDailyEntry
	for rows.Next() {
		var e model.MedicalDailyEntry
		if err := rows.Scan(&e.ID, &e.PatientID, &e.EntryDate, &e.EntryType, &e.Content, &e.NurseID, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan daily entry: %w", err)
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

// CreateVerification inserts a verification record.
func (s *SqliteStore) CreateVerification(ctx context.Context, v *model.MedicalVerification) error {
	matchedInt := 0
	if v.Matched {
		matchedInt = 1
	}
	verifiedAt := ""
	if v.VerifiedAt != nil {
		verifiedAt = v.VerifiedAt.Format(time.RFC3339)
	}
	patientID := ""
	if v.PatientID != nil {
		patientID = *v.PatientID
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medical_verifications (id, device_id, patient_id, verification_type, result, matched, verified_by, verified_at, notes, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))`,
		v.ID, v.DeviceID, patientID, v.VerificationType, v.Result, matchedInt, v.VerifiedBy, verifiedAt, v.Notes)
	return err
}

// ListVerifications returns verification records with pagination.
func (s *SqliteStore) ListVerifications(ctx context.Context, page, pageSize int) ([]model.MedicalVerification, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, device_id, patient_id, verification_type, result, matched, verified_by, verified_at, notes, created_at
		FROM medical_verifications ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("list verifications: %w", err)
	}
	defer rows.Close()

	var items []model.MedicalVerification
	for rows.Next() {
		var v model.MedicalVerification
		var matchedInt int
		var patientID string
		if err := rows.Scan(&v.ID, &v.DeviceID, &patientID, &v.VerificationType, &v.Result, &matchedInt, &v.VerifiedBy, &v.VerifiedAt, &v.Notes, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan verification: %w", err)
		}
		v.Matched = matchedInt != 0
		if patientID != "" {
			v.PatientID = &patientID
		}
		items = append(items, v)
	}
	return items, rows.Err()
}

// UpdateVerificationStatus updates verification status.
func (s *SqliteStore) UpdateVerificationStatus(ctx context.Context, id, status string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE medical_verifications SET status=? WHERE id=?`, status, id)
	return err
}

// GetTodayVerificationStats returns today's verification statistics.
func (s *SqliteStore) GetTodayVerificationStats(ctx context.Context) (*model.MedicalVerificationStats, error) {
	var stats model.MedicalVerificationStats
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*), SUM(CASE WHEN matched=1 THEN 1 ELSE 0 END), SUM(CASE WHEN matched=0 THEN 1 ELSE 0 END)
		FROM medical_verifications WHERE DATE(verified_at)=DATE('now')`).Scan(
		&stats.Total, &stats.Matched, &stats.Unmatched)
	if err != nil {
		return nil, fmt.Errorf("get verification stats: %w", err)
	}
	return &stats, nil
}

// GetMedicalStatsOverview returns overall medical statistics.
func (s *SqliteStore) GetMedicalStatsOverview(ctx context.Context) (*model.MedicalStatsOverview, error) {
	var overview model.MedicalStatsOverview
	err := s.db.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM medical_wristband_patients WHERE status='admitted'),
			(SELECT COUNT(*) FROM medical_wristband_patients WHERE DATE(created_at)=DATE('now')),
			(SELECT COUNT(*) FROM medical_wristband_patients WHERE DATE(updated_at)=DATE('now') AND status='discharged'),
			(SELECT COUNT(*) FROM medical_bindings WHERE unbound_at IS NULL),
			(SELECT COUNT(*) FROM medical_wristband_devices)
	`).Scan(
		&overview.ActivePatients, &overview.TodayAdmitted, &overview.TodayDischarged, &overview.BoundDevices, &overview.TotalDevices)
	if err != nil {
		return nil, fmt.Errorf("get medical stats overview: %w", err)
	}
	return &overview, nil
}

// CreateAlertTagConfig creates an alert tag configuration.
func (s *SqliteStore) CreateAlertTagConfig(ctx context.Context, c *model.MedicalAlertTagConfig) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medical_alert_tag_config (id, tag_name, tag_color, tag_icon, enabled) VALUES (?, ?, ?, ?, ?)`,
		c.ID, c.TagName, c.TagColor, c.TagIcon, c.Enabled)
	return err
}

// ListAlertTagConfigs returns all alert tag configurations.
func (s *SqliteStore) ListAlertTagConfigs(ctx context.Context) ([]model.MedicalAlertTagConfig, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tag_name, tag_color, tag_icon, enabled, created_at, updated_at
		FROM medical_alert_tag_config ORDER BY tag_name`)
	if err != nil {
		return nil, fmt.Errorf("list alert tag configs: %w", err)
	}
	defer rows.Close()

	var items []model.MedicalAlertTagConfig
	for rows.Next() {
		var c model.MedicalAlertTagConfig
		if err := rows.Scan(&c.ID, &c.TagName, &c.TagColor, &c.TagIcon, &c.Enabled, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan alert tag config: %w", err)
		}
		items = append(items, c)
	}
	return items, rows.Err()
}
