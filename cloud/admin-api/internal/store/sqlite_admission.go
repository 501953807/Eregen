package store

import (
	"context"
	"database/sql"
	"eregen.dev/admin-api/internal/model"
	"fmt"
	"time"
)


// ========== Clinical Workflow Methods ==========

func (s *SqliteStore) CreateAdmission(ctx context.Context, a *model.HospitalAdmission) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO hospital_admissions (id, patient_id, admission_no, bed_no, department, diagnosis,
		 emergency_contact, allergies, admitted_at, expected_discharge_at, discharge_type, transferred_to, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), ?, ?, ?, ?)`,
		a.ID, a.PatientID, a.AdmissionNo, a.BedNo, a.Department, a.Diagnosis,
		a.EmergencyContact, a.Allergies, a.ExpectedDischargeAt, a.DischargeType, a.TransferredTo, a.Notes)
	return err
}

func (s *SqliteStore) GetAdmission(ctx context.Context, id string) (*model.HospitalAdmission, error) {
	var a model.HospitalAdmission
	var admittedAtStr, expectedDischarge, dischargedAt, transferTo string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, patient_id, admission_no, bed_no, department, diagnosis, emergency_contact,
		 allergies, admitted_at, expected_discharge_at, discharged_at, discharge_type, transferred_to, notes
		 FROM hospital_admissions WHERE id = ?`, id).Scan(
		&a.ID, &a.PatientID, &a.AdmissionNo, &a.BedNo, &a.Department, &a.Diagnosis,
		&a.EmergencyContact, &a.Allergies, &admittedAtStr, &expectedDischarge, &dischargedAt,
		&a.DischargeType, &transferTo, &a.Notes)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("admission not found")
		}
		return nil, fmt.Errorf("get admission: %w", err)
	}
	if t, err := time.Parse("2006-01-02 15:04:05", admittedAtStr); err == nil {
		a.AdmittedAt = t
	}
	if expectedDischarge != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", expectedDischarge); err == nil {
			a.ExpectedDischargeAt = &t
		}
	}
	if dischargedAt != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", dischargedAt); err == nil {
			a.DischargedAt = &t
		}
	}
	a.TransferredTo = transferTo
	return &a, nil
}

func (s *SqliteStore) ListAdmissions(ctx context.Context, page, pageSize int, department, status string) ([]model.HospitalAdmission, error) {
	query := `SELECT id, patient_id, admission_no, bed_no, department, diagnosis, emergency_contact,
		  allergies, admitted_at, expected_discharge_at, discharged_at, discharge_type, transferred_to, notes
		  FROM hospital_admissions WHERE 1=1`
	var args []interface{}
	if department != "" {
		query += fmt.Sprintf(" AND department=?")
		args = append(args, department)
	}
	if status != "" {
		query += fmt.Sprintf(" AND (discharged_at IS NULL OR discharge_type != ?)")
		args = append(args, status)
	}
	query += fmt.Sprintf(" ORDER BY admitted_at DESC LIMIT ? OFFSET ?")
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list admissions: %w", err)
	}
	defer rows.Close()

	var items []model.HospitalAdmission
	for rows.Next() {
		var a model.HospitalAdmission
		var admittedAtStr, expectedDischargeStr, dischargedAtStr, transferTo string
		if err := rows.Scan(&a.ID, &a.PatientID, &a.AdmissionNo, &a.BedNo, &a.Department,
			&a.Diagnosis, &a.EmergencyContact, &a.Allergies, &admittedAtStr,
			&expectedDischargeStr, &dischargedAtStr, &a.DischargeType, &transferTo, &a.Notes); err != nil {
			return nil, fmt.Errorf("scan admission: %w", err)
		}
		if t, err := time.Parse("2006-01-02 15:04:05", admittedAtStr); err == nil {
			a.AdmittedAt = t
		}
		if expectedDischargeStr != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", expectedDischargeStr); err == nil {
				a.ExpectedDischargeAt = &t
			}
		}
		if dischargedAtStr != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", dischargedAtStr); err == nil {
				a.DischargedAt = &t
			}
		}
		a.TransferredTo = transferTo
		items = append(items, a)
	}
	return items, rows.Err()
}

func (s *SqliteStore) CompleteAdmission(ctx context.Context, id, dischargeType, notes, transferredTo string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE hospital_admissions SET discharged_at=datetime('now'), discharge_type=?, notes=?, transferred_to=? WHERE id=?`,
		dischargeType, notes, transferredTo, id)
	if err != nil {
		return err
	}
	// Also get patient_id to update patient status
	var patientID string
	err = s.db.QueryRowContext(ctx, `SELECT patient_id FROM hospital_admissions WHERE id=?`, id).Scan(&patientID)
	if err == nil {
		s.db.ExecContext(ctx, `UPDATE medical_wristband_patients SET status='discharged', updated_at=datetime('now') WHERE id=?`, patientID)
		// Unbind any wristband bindings for this patient
		s.db.ExecContext(ctx, `UPDATE medical_bindings SET unbound_at=datetime('now') WHERE patient_id=? AND unbound_at IS NULL`, patientID)
		s.db.ExecContext(ctx, `UPDATE medical_wristband_devices SET bound_patient_id=NULL, status='idle' WHERE id IN (SELECT device_id FROM medical_bindings WHERE patient_id=? AND unbound_at IS NULL)`, patientID)
	}
	return nil
}

func (s *SqliteStore) CreateWardRound(ctx context.Context, w *model.WardRoundEntry) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO ward_rounds (id, patient_id, nurse_id, blood_pressure, heart_rate, spo2,
		 temperature, weight, notes, observations, completed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))`,
		w.ID, w.PatientID, w.NurseID, w.BloodPressure, w.HeartRate, w.SpO2,
		w.Temperature, w.Weight, w.Notes, w.Observations)
	return err
}

func (s *SqliteStore) ListWardRounds(ctx context.Context, patientID string) ([]model.WardRoundEntry, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, patient_id, nurse_id, blood_pressure, heart_rate, spo2, temperature, weight, notes, observations, completed_at
		 FROM ward_rounds WHERE patient_id=? ORDER BY completed_at DESC`, patientID)
	if err != nil {
		return nil, fmt.Errorf("list ward rounds: %w", err)
	}
	defer rows.Close()

	var items []model.WardRoundEntry
	for rows.Next() {
		var w model.WardRoundEntry
		if err := rows.Scan(&w.ID, &w.PatientID, &w.NurseID, &w.BloodPressure, &w.HeartRate,
			&w.SpO2, &w.Temperature, &w.Weight, &w.Notes, &w.Observations, &w.CompletedAt); err != nil {
			return nil, fmt.Errorf("scan ward round: %w", err)
		}
		items = append(items, w)
	}
	return items, rows.Err()
}

func (s *SqliteStore) EvaluateRegulatoryRules(ctx context.Context, event string, data map[string]string) ([]*model.RegulatoryRuleResult, error) {
	var results []*model.RegulatoryRuleResult
	now := time.Now().UTC()

	switch event {
	case "patient_admitted":
		// R01: Bed-fraud detection -- check if patient has active wristband binding
		patientID := data["patient_id"]
		var bindingCount int
		s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM medical_bindings WHERE patient_id=? AND unbound_at IS NULL`, patientID).Scan(&bindingCount)
		if bindingCount == 0 {
			results = append(results, &model.RegulatoryRuleResult{
				RuleCode: "R01", Severity: "P1", PatientID: patientID,
				Message: "Patient admitted without active wristband binding", TriggeredAt: now,
			})
		}
	case "patient_discharged":
		// R08: Post-discharge data retention -- handled by marking admissions
	case "ward_round_completed":
		// No specific rules triggered
	case "verification_scan":
		// R05: Medication-verification mismatch
		scanType := data["scan_type"]
		if scanType == "medication" {
			patientID := data["patient_id"]
			var bindingCount int
			s.db.QueryRowContext(ctx,
				`SELECT COUNT(*) FROM medical_bindings mb JOIN medical_wristband_patients p ON p.id=mb.patient_id
				 WHERE p.id=? AND p.status='admitted' AND mb.unbound_at IS NULL`, patientID).Scan(&bindingCount)
			if bindingCount == 0 {
				results = append(results, &model.RegulatoryRuleResult{
					RuleCode: "R05", Severity: "P2", PatientID: patientID,
					Message: "Medication verification without active wristband binding", TriggeredAt: now,
				})
			}
		}
	}
	return results, nil
}
