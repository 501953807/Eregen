package store

import (
	"time"
	"context"
	"database/sql"
	"eregen.dev/admin-api/internal/model"
	"fmt"
)


func (s *PostgresStore) CreateExpense(ctx context.Context, e *model.MedicalExpense) error {

	_, err := s.db.ExecContext(ctx,

		`INSERT INTO medical_expenses (id, patient_id, item_name, category, amount, quantity, unit_price, notes, created_at, updated_at)

		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW(),NOW())`,

		e.ID, e.PatientID, e.ItemName, e.Category, e.Amount, e.Quantity, e.UnitPrice, e.Notes)

	return err

}



func (s *PostgresStore) ListExpenses(ctx context.Context, patientID string, page, pageSize int) ([]model.MedicalExpense, error) {

	rows, err := s.db.QueryContext(ctx, `

		SELECT id, patient_id, item_name, category, amount, quantity, unit_price, notes, created_at, updated_at

		FROM medical_expenses WHERE patient_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,

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



func (s *PostgresStore) CreateMedication(ctx context.Context, m *model.MedicalMedication) error {

	_, err := s.db.ExecContext(ctx,

		`INSERT INTO medical_medications (id, patient_id, name, dosage, frequency, duration, route, notes, created_at, updated_at)

		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW(),NOW())`,

		m.ID, m.PatientID, m.Name, m.Dosage, m.Frequency, m.Duration, m.Route, m.Notes)

	return err

}



func (s *PostgresStore) ListMedications(ctx context.Context, patientID string) ([]model.MedicalMedication, error) {

	rows, err := s.db.QueryContext(ctx, `

		SELECT id, patient_id, name, dosage, frequency, duration, route, notes, created_at, updated_at

		FROM medical_medications WHERE patient_id=$1 ORDER BY created_at DESC`, patientID)

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



func (s *PostgresStore) CreateTestResult(ctx context.Context, r *model.MedicalTestResult) error {

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

		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW(),NOW())`,

		r.ID, r.PatientID, r.TestName, r.Result, r.ReferenceRange, r.Unit, collectedAt, reportedAt, r.Notes)

	return err

}



func (s *PostgresStore) ListTestResults(ctx context.Context, patientID string) ([]model.MedicalTestResult, error) {

	rows, err := s.db.QueryContext(ctx, `

		SELECT id, patient_id, test_name, result, reference_range, unit, collected_at, reported_at, notes, created_at, updated_at

		FROM medical_test_results WHERE patient_id=$1 ORDER BY collected_at DESC`, patientID)

	if err != nil {

		return nil, fmt.Errorf("list test results: %w", err)

	}

	defer rows.Close()



	var items []model.MedicalTestResult

	for rows.Next() {

		var t model.MedicalTestResult

		if err := rows.Scan(&t.ID, &t.PatientID, &t.TestName, &t.Result, &t.ReferenceRange, &t.Unit, &t.CollectedAt, &t.ReportedAt, &t.Notes, &t.CreatedAt, &t.UpdatedAt); err != nil {

			return nil, fmt.Errorf("scan test result: %w", err)

		}

		items = append(items, t)

	}

	return items, rows.Err()

}



func (s *PostgresStore) CreateDailyEntry(ctx context.Context, e *model.MedicalDailyEntry) error {

	_, err := s.db.ExecContext(ctx,

		`INSERT INTO medical_daily_entries (id, patient_id, entry_date, entry_type, content, nurse_id, created_at, updated_at)

		 VALUES ($1,$2,$3,$4,$5,$6,NOW(),NOW())`,

		e.ID, e.PatientID, e.EntryDate, e.EntryType, e.Content, e.NurseID)

	return err

}



func (s *PostgresStore) ListDailyEntries(ctx context.Context, patientID string, date string) ([]model.MedicalDailyEntry, error) {

	var rows *sql.Rows

	var err error

	if date != "" {

		rows, err = s.db.QueryContext(ctx, `

			SELECT id, patient_id, entry_date, entry_type, content, nurse_id, created_at, updated_at

			FROM medical_daily_entries WHERE patient_id=$1 AND entry_date=$2 ORDER BY created_at DESC`, patientID, date)

	} else {

		rows, err = s.db.QueryContext(ctx, `

			SELECT id, patient_id, entry_date, entry_type, content, nurse_id, created_at, updated_at

			FROM medical_daily_entries WHERE patient_id=$1 ORDER BY entry_date DESC, created_at DESC`, patientID)

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



func (s *PostgresStore) CreateVerification(ctx context.Context, v *model.MedicalVerification) error {

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

		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW())`,

		v.ID, v.DeviceID, patientID, v.VerificationType, v.Result, matchedInt, v.VerifiedBy, verifiedAt, v.Notes)

	return err

}



func (s *PostgresStore) ListVerifications(ctx context.Context, page, pageSize int) ([]model.MedicalVerification, error) {

	rows, err := s.db.QueryContext(ctx, `

		SELECT id, device_id, patient_id, verification_type, result, matched, verified_by, verified_at, notes, created_at

		FROM medical_verifications ORDER BY created_at DESC LIMIT $1 OFFSET $2`,

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



func (s *PostgresStore) UpdateVerificationStatus(ctx context.Context, id, status string) error {

	_, err := s.db.ExecContext(ctx, `UPDATE medical_verifications SET status=$1 WHERE id=$2`, status, id)

	return err

}



func (s *PostgresStore) GetTodayVerificationStats(ctx context.Context) (*model.MedicalVerificationStats, error) {

	var stats model.MedicalVerificationStats

	err := s.db.QueryRowContext(ctx, `

		SELECT COUNT(*), SUM(CASE WHEN matched THEN 1 ELSE 0 END), SUM(CASE WHEN NOT matched THEN 1 ELSE 0 END)

		FROM medical_verifications WHERE DATE(verified_at)=CURRENT_DATE`).Scan(

		&stats.Total, &stats.Matched, &stats.Unmatched)

	if err != nil {

		return nil, fmt.Errorf("get verification stats: %w", err)

	}

	return &stats, nil

}



func (s *PostgresStore) GetMedicalStatsOverview(ctx context.Context) (*model.MedicalStatsOverview, error) {

	var overview model.MedicalStatsOverview

	err := s.db.QueryRowContext(ctx, `

		SELECT

			(SELECT COUNT(*) FROM medical_wristband_patients WHERE status='admitted'),

			(SELECT COUNT(*) FROM medical_wristband_patients WHERE DATE(created_at)=CURRENT_DATE),

			(SELECT COUNT(*) FROM medical_wristband_patients WHERE DATE(updated_at)=CURRENT_DATE AND status='discharged'),

			(SELECT COUNT(*) FROM medical_bindings WHERE unbound_at IS NULL),

			(SELECT COUNT(*) FROM medical_wristband_devices)

	`).Scan(

		&overview.ActivePatients, &overview.TodayAdmitted, &overview.TodayDischarged, &overview.BoundDevices, &overview.TotalDevices)

	if err != nil {

		return nil, fmt.Errorf("get medical stats overview: %w", err)

	}

	return &overview, nil

}



func (s *PostgresStore) CreateAlertTagConfig(ctx context.Context, c *model.MedicalAlertTagConfig) error {

	_, err := s.db.ExecContext(ctx,

		`INSERT INTO medical_alert_tag_config (id, tag_name, tag_color, tag_icon, enabled) VALUES ($1,$2,$3,$4,$5)`,

		c.ID, c.TagName, c.TagColor, c.TagIcon, c.Enabled)

	return err

}



func (s *PostgresStore) ListAlertTagConfigs(ctx context.Context) ([]model.MedicalAlertTagConfig, error) {

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

