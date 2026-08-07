package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"eregen.dev/admin-api/internal/model"
	"fmt"
)


// ========== Medical Wristband Methods ==========

// CreatePatient inserts a new patient record.
func (s *SqliteStore) CreatePatient(ctx context.Context, p *model.MedicalPatient) error {
	tagsJSON, _ := json.Marshal(p.TagIDs)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medical_wristband_patients (id, admission_no, name, gender, age, department, bed_number, blood_type, allergies, special_conditions, tag_ids, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		p.ID, p.AdmissionNo, p.Name, p.Gender, p.Age, p.Department, p.BedNumber, p.BloodType, p.Allergies, p.SpecialConditions, string(tagsJSON), p.Status)
	return err
}

// GetPatient returns a patient by ID.
func (s *SqliteStore) GetPatient(ctx context.Context, id string) (*model.MedicalPatient, error) {
	var p model.MedicalPatient
	var tagsRaw string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, admission_no, name, gender, age, department, bed_number, blood_type, allergies, special_conditions, tag_ids, status, created_at, updated_at
		FROM medical_wristband_patients WHERE id = ?`, id).Scan(
		&p.ID, &p.AdmissionNo, &p.Name, &p.Gender, &p.Age, &p.Department, &p.BedNumber,
		&p.BloodType, &p.Allergies, &p.SpecialConditions, &tagsRaw, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("patient not found")
		}
		return nil, fmt.Errorf("get patient: %w", err)
	}
	json.Unmarshal([]byte(tagsRaw), &p.TagIDs)
	return &p, nil
}

// ListPatients returns paginated patients.
func (s *SqliteStore) ListPatients(ctx context.Context, page, pageSize int, status string) ([]model.MedicalPatient, error) {
	query := `SELECT id, admission_no, name, gender, age, department, bed_number, status, created_at, updated_at
		FROM medical_wristband_patients WHERE 1=1`
	args := []interface{}{}
	idx := 1
	if status != "" {
		query += fmt.Sprintf(" AND status=?")
		args = append(args, status)
		idx++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT ? OFFSET ?")
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list patients: %w", err)
	}
	defer rows.Close()

	var patients []model.MedicalPatient
	for rows.Next() {
		var p model.MedicalPatient
		if err := rows.Scan(&p.ID, &p.AdmissionNo, &p.Name, &p.Gender, &p.Age, &p.Department, &p.BedNumber, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan patient: %w", err)
		}
		patients = append(patients, p)
	}
	return patients, rows.Err()
}

// UpdatePatient modifies a patient.
func (s *SqliteStore) UpdatePatient(ctx context.Context, p *model.MedicalPatient) error {
	tagsJSON, _ := json.Marshal(p.TagIDs)
	_, err := s.db.ExecContext(ctx,
		`UPDATE medical_wristband_patients SET admission_no=?, name=?, gender=?, age=?, department=?, bed_number=?, blood_type=?, allergies=?, special_conditions=?, tag_ids=?, status=?, updated_at=datetime('now') WHERE id=?`,
		p.AdmissionNo, p.Name, p.Gender, p.Age, p.Department, p.BedNumber, p.BloodType, p.Allergies, p.SpecialConditions, string(tagsJSON), p.Status, p.ID)
	return err
}

// DeletePatient soft-deletes a patient (marks as discharged).
func (s *SqliteStore) DeletePatient(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE medical_wristband_patients SET status='discharged', updated_at=datetime('now') WHERE id=?`, id)
	return err
}

// GetPatientByAdmissionNo returns a patient by admission number.
func (s *SqliteStore) GetPatientByAdmissionNo(ctx context.Context, admissionNo string) (*model.MedicalPatient, error) {
	var p model.MedicalPatient
	var tagsRaw string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, admission_no, name, gender, age, department, bed_number, blood_type, allergies, special_conditions, tag_ids, status, created_at, updated_at
		FROM medical_wristband_patients WHERE admission_no=?`, admissionNo).Scan(
		&p.ID, &p.AdmissionNo, &p.Name, &p.Gender, &p.Age, &p.Department, &p.BedNumber,
		&p.BloodType, &p.Allergies, &p.SpecialConditions, &tagsRaw, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("patient not found")
		}
		return nil, err
	}
	json.Unmarshal([]byte(tagsRaw), &p.TagIDs)
	return &p, nil
}

// BatchImportPatients imports multiple patients.
func (s *SqliteStore) BatchImportPatients(ctx context.Context, patients []model.MedicalPatient) error {
	for _, p := range patients {
		if err := s.CreatePatient(ctx, &p); err != nil {
			return fmt.Errorf("import patient %s: %w", p.Name, err)
		}
	}
	return nil
}

// GetPatientHistory returns treatment history for a patient.
func (s *SqliteStore) GetPatientHistory(ctx context.Context, patientID string) (*model.MedicalPatientHistory, error) {
	entries, err := s.ListDailyEntries(ctx, patientID, "")
	if err != nil {
		return nil, err
	}
	return &model.MedicalPatientHistory{DailyEntries: entries}, nil
}
