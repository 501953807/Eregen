package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"eregen.dev/b2b-hospital-api/internal/model"

	_ "modernc.org/sqlite"
)

// SqliteStore implements Database interface using SQLite.
type SqliteStore struct {
	db *sql.DB
}

// NewSqlite creates a new SQLite store.
func NewSqlite(path string) (*SqliteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite: %w", err)
	}
	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate sqlite: %w", err)
	}
	return &SqliteStore{db: db}, nil
}

// Close closes the database connection.
func (s *SqliteStore) Close() error {
	return s.db.Close()
}

// ---------- Institution ----------

func (s *SqliteStore) CreateInstitution(ctx context.Context, inst *model.Institution) error {
	inst.ID = generateUUID()
	now := time.Now()
	inst.CreatedAt = now
	inst.UpdatedAt = now
	if inst.Status == "" {
		inst.Status = "pending"
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO b2b_institutions (id, name, type, code, contact_name, contact_phone,
			access_level, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		inst.ID, inst.Name, inst.Type, inst.Code,
		inst.ContactName, inst.ContactPhone, inst.AccessLevel, inst.Status,
		inst.CreatedAt, inst.UpdatedAt,
	)
	return err
}

func (s *SqliteStore) GetInstitutionByID(ctx context.Context, id string) (*model.Institution, error) {
	inst := &model.Institution{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, type, code, contact_name, contact_phone, access_level, status, created_at, updated_at
		 FROM b2b_institutions WHERE id = ?`, id).Scan(
		&inst.ID, &inst.Name, &inst.Type, &inst.Code,
		&inst.ContactName, &inst.ContactPhone, &inst.AccessLevel, &inst.Status,
		&inst.CreatedAt, &inst.UpdatedAt,
	)
	return inst, err
}

func (s *SqliteStore) GetInstitutionByCode(ctx context.Context, code string) (*model.Institution, error) {
	inst := &model.Institution{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, type, code, contact_name, contact_phone, access_level, status, created_at, updated_at
		 FROM b2b_institutions WHERE code = ? AND status = 'active'`, code).Scan(
		&inst.ID, &inst.Name, &inst.Type, &inst.Code,
		&inst.ContactName, &inst.ContactPhone, &inst.AccessLevel, &inst.Status,
		&inst.CreatedAt, &inst.UpdatedAt,
	)
	return inst, err
}

func (s *SqliteStore) ListInstitutions(ctx context.Context, page, pageSize int) ([]model.Institution, int, error) {
	offset := (page - 1) * pageSize
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, type, code, contact_name, contact_phone, access_level, status, created_at, updated_at
		 FROM b2b_institutions ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.Institution
	for rows.Next() {
		var i model.Institution
		if err := rows.Scan(&i.ID, &i.Name, &i.Type, &i.Code, &i.ContactName, &i.ContactPhone,
			&i.AccessLevel, &i.Status, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, i)
	}

	var total int
	s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM b2b_institutions").Scan(&total)
	return list, total, rows.Err()
}

func (s *SqliteStore) UpdateInstitution(ctx context.Context, id string, inst *model.Institution) error {
	inst.UpdatedAt = time.Now()
	_, err := s.db.ExecContext(ctx,
		`UPDATE b2b_institutions SET name=?, type=?, code=?, contact_name=?, contact_phone=?,
			access_level=?, status=?, updated_at=? WHERE id=?`,
		inst.Name, inst.Type, inst.Code,
		inst.ContactName, inst.ContactPhone, inst.AccessLevel, inst.Status,
		inst.UpdatedAt, id,
	)
	return err
}

// ---------- API Key ----------

func (s *SqliteStore) CreateAPIKey(ctx context.Context, key *model.InstitutionAPIKey) error {
	key.ID = generateUUID()
	key.CreatedAt = time.Now()
	if !key.Active {
		key.Active = true
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO b2b_api_keys (id, institution_id, key_hash, name, expires_at, active, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		key.ID, key.InstitutionID, key.KeyHash, key.Name, key.ExpiresAt, key.Active, key.CreatedAt,
	)
	return err
}

func (s *SqliteStore) GetInstitutionByAPIKey(ctx context.Context, keyHash string) (*model.Institution, error) {
	inst := &model.Institution{}
	err := s.db.QueryRowContext(ctx,
		`SELECT i.id, i.name, i.type, i.code, i.contact_name, i.contact_phone,
				i.access_level, i.status, i.created_at, i.updated_at
		 FROM b2b_api_keys k
		 JOIN b2b_institutions i ON i.id = k.institution_id
		 WHERE k.key_hash = ? AND k.active = 1 AND k.expires_at > datetime('now') AND i.status = 'active'`,
		keyHash).Scan(
		&inst.ID, &inst.Name, &inst.Type, &inst.Code,
		&inst.ContactName, &inst.ContactPhone, &inst.AccessLevel, &inst.Status,
		&inst.CreatedAt, &inst.UpdatedAt,
	)
	return inst, err
}

// ---------- Elderly-Institution Link ----------

func (s *SqliteStore) LinkElderlyToInstitution(ctx context.Context, link *model.ElderlyInstitutionLink) error {
	link.ID = generateUUID()
	link.Active = true
	link.CreatedAt = time.Now()
	link.UpdatedAt = link.CreatedAt

	var data []byte
	if len(link.Notes) > 0 {
		data = link.Notes
	} else {
		data = []byte("{}")
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO b2b_elderly_links (id, elderly_id, institution_id, admitted_at, discharged_at,
			primary_doc, notes, active, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		link.ID, link.ElderlyID, link.InstitutionID,
		link.AdmittedAt, link.DischargedAt, link.PrimaryDoc, data, link.Active,
		link.CreatedAt, link.UpdatedAt,
	)
	return err
}

func (s *SqliteStore) GetActiveLinksForInstitution(ctx context.Context, instID string) ([]model.ElderlyInstitutionLink, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, elderly_id, institution_id, admitted_at, discharged_at, primary_doc, notes, active, created_at, updated_at
		 FROM b2b_elderly_links WHERE institution_id = ? AND active = 1
		 ORDER BY created_at DESC`, instID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []model.ElderlyInstitutionLink
	for rows.Next() {
		var l model.ElderlyInstitutionLink
		var data []byte
		if err := rows.Scan(&l.ID, &l.ElderlyID, &l.InstitutionID, &l.AdmittedAt,
			&l.DischargedAt, &l.PrimaryDoc, &data, &l.Active, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		if len(data) > 0 {
			json.Unmarshal(data, &l.Notes)
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

func (s *SqliteStore) GetActiveLinksForElderly(ctx context.Context, elderlyID string) ([]model.ElderlyInstitutionLink, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, elderly_id, institution_id, admitted_at, discharged_at, primary_doc, notes, active, created_at, updated_at
		 FROM b2b_elderly_links WHERE elderly_id = ? AND active = 1
		 ORDER BY created_at DESC`, elderlyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []model.ElderlyInstitutionLink
	for rows.Next() {
		var l model.ElderlyInstitutionLink
		var data []byte
		if err := rows.Scan(&l.ID, &l.ElderlyID, &l.InstitutionID, &l.AdmittedAt,
			&l.DischargedAt, &l.PrimaryDoc, &data, &l.Active, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		if len(data) > 0 {
			json.Unmarshal(data, &l.Notes)
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

// ---------- Health Data ----------

func (s *SqliteStore) StoreVitals(ctx context.Context, v *model.VitalSignRecord) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO b2b_vital_signs (id, elderly_id, institution_id, patient_id,
			heart_rate, spo2, systolic_bp, diastolic_bp, temperature, steps, recorded_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		v.ID, v.ElderlyID, v.InstitutionID, v.PatientID,
		v.HeartRate, v.SPO2, v.SystolicBP, v.DiastolicBP,
		v.Temperature, v.Steps, v.RecordedAt,
	)
	return err
}

func (s *SqliteStore) BulkStoreVitals(ctx context.Context, vitals []*model.VitalSignRecord) error {
	for _, v := range vitals {
		if err := s.StoreVitals(ctx, v); err != nil {
			fmt.Printf("store vital sign error: %v\n", err)
		}
	}
	return nil
}

func (s *SqliteStore) GetVitalsForElderly(ctx context.Context, elderlyID string, days int) ([]model.VitalSignRecord, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, elderly_id, institution_id, patient_id,
			heart_rate, spo2, systolic_bp, diastolic_bp, temperature, steps, recorded_at
		 FROM b2b_vital_signs WHERE elderly_id = ? AND recorded_at > datetime('now', ?)
		 ORDER BY recorded_at DESC`,
		elderlyID, fmt.Sprintf("-%d days", days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vitals []model.VitalSignRecord
	for rows.Next() {
		var v model.VitalSignRecord
		if err := rows.Scan(&v.ID, &v.ElderlyID, &v.InstitutionID, &v.PatientID,
			&v.HeartRate, &v.SPO2, &v.SystolicBP, &v.DiastolicBP,
			&v.Temperature, &v.Steps, &v.RecordedAt); err != nil {
			return nil, err
		}
		vitals = append(vitals, v)
	}
	return vitals, rows.Err()
}

func (s *SqliteStore) LinkElderlyToExternalPatient(ctx context.Context, elderlyID, patientID, eregenID string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO b2b_patient_links (id, external_patient_id, local_elderly_id, created_at)
		 VALUES (?, ?, ?, datetime('now'))`,
		generateUUID(), patientID, eregenID)
	return err
}

func (s *SqliteStore) FindElderlyByExternalPatient(ctx context.Context, patientID string) (string, error) {
	var elderlyID string
	err := s.db.QueryRowContext(ctx,
		`SELECT local_elderly_id FROM b2b_patient_links WHERE external_patient_id = ?`,
		patientID).Scan(&elderlyID)
	return elderlyID, err
}

// ---------- Diagnoses ----------

func (s *SqliteStore) StoreDiagnoses(ctx context.Context, records []*model.DiagnosisRecord) error {
	for _, r := range records {
		if err := s.storeSingleDiagnosis(ctx, r); err != nil {
			fmt.Printf("store diagnosis error: %v\n", err)
		}
	}
	return nil
}

func (s *SqliteStore) storeSingleDiagnosis(ctx context.Context, r *model.DiagnosisRecord) error {
	r.ID = generateUUID()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO b2b_diagnoses (id, elderly_id, institution_id, patient_id,
			diagnosis_code, diagnosis_name, severity, diagnosed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.ElderlyID, r.InstitutionID, r.PatientID,
		r.DiagnosisCode, r.DiagnosisName, r.Severity, r.DiagnosedAt,
	)
	return err
}

func (s *SqliteStore) GetDiagnosesForElderly(ctx context.Context, elderlyID string, days int) ([]model.DiagnosisRecord, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, elderly_id, institution_id, patient_id, diagnosis_code, diagnosis_name, severity, diagnosed_at
		 FROM b2b_diagnoses WHERE elderly_id = ? AND diagnosed_at > datetime('now', ?)
		 ORDER BY diagnosed_at DESC`,
		elderlyID, fmt.Sprintf("-%d days", days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.DiagnosisRecord
	for rows.Next() {
		var r model.DiagnosisRecord
		if err := rows.Scan(&r.ID, &r.ElderlyID, &r.InstitutionID, &r.PatientID,
			&r.DiagnosisCode, &r.DiagnosisName, &r.Severity, &r.DiagnosedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// ---------- Medications ----------

func (s *SqliteStore) StoreMedications(ctx context.Context, records []*model.MedicationRecord) error {
	for _, r := range records {
		if err := s.storeSingleMedication(ctx, r); err != nil {
			fmt.Printf("store medication error: %v\n", err)
		}
	}
	return nil
}

func (s *SqliteStore) storeSingleMedication(ctx context.Context, r *model.MedicationRecord) error {
	r.ID = generateUUID()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO b2b_medications (id, elderly_id, institution_id, patient_id,
			medication_name, dose, frequency, route, duration, prescribed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.ElderlyID, r.InstitutionID, r.PatientID,
		r.MedicationName, r.Dose, r.Frequency, r.Route, r.Duration, r.PrescribedAt,
	)
	return err
}

func (s *SqliteStore) GetMedicationsForElderly(ctx context.Context, elderlyID string) ([]model.MedicationRecord, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, elderly_id, institution_id, patient_id, medication_name, dose, frequency, route, duration, prescribed_at
		 FROM b2b_medications WHERE elderly_id = ? ORDER BY prescribed_at DESC`,
		elderlyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.MedicationRecord
	for rows.Next() {
		var r model.MedicationRecord
		if err := rows.Scan(&r.ID, &r.ElderlyID, &r.InstitutionID, &r.PatientID,
			&r.MedicationName, &r.Dose, &r.Frequency, &r.Route, &r.Duration, &r.PrescribedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// ---------- Migration ----------

func migrate(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS b2b_institutions (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			code TEXT UNIQUE,
			contact_name TEXT,
			contact_phone TEXT,
			access_level TEXT DEFAULT 'read',
			status TEXT DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS b2b_api_keys (
			id TEXT PRIMARY KEY,
			institution_id TEXT NOT NULL REFERENCES b2b_institutions(id),
			key_hash TEXT UNIQUE NOT NULL,
			name TEXT,
			expires_at DATETIME,
			active INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS b2b_elderly_links (
			id TEXT PRIMARY KEY,
			elderly_id TEXT NOT NULL,
			institution_id TEXT NOT NULL REFERENCES b2b_institutions(id),
			admitted_at DATETIME,
			discharged_at DATETIME,
			primary_doc TEXT,
			notes TEXT DEFAULT '{}',
			active INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS b2b_vital_signs (
			id TEXT PRIMARY KEY,
			elderly_id TEXT NOT NULL,
			institution_id TEXT NOT NULL,
			patient_id TEXT,
			heart_rate INTEGER,
			spo2 INTEGER,
			systolic_bp INTEGER,
			diastolic_bp INTEGER,
			temperature REAL,
			steps INTEGER,
			recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS b2b_patient_links (
			id TEXT PRIMARY KEY,
			external_patient_id TEXT NOT NULL,
			local_elderly_id TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS b2b_diagnoses (
			id TEXT PRIMARY KEY,
			elderly_id TEXT NOT NULL,
			institution_id TEXT NOT NULL,
			patient_id TEXT,
			diagnosis_code TEXT,
			diagnosis_name TEXT,
			severity TEXT,
			diagnosed_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS b2b_medications (
			id TEXT PRIMARY KEY,
			elderly_id TEXT NOT NULL,
			institution_id TEXT NOT NULL,
			patient_id TEXT,
			medication_name TEXT,
			dose TEXT,
			frequency TEXT,
			route TEXT,
			duration TEXT,
			prescribed_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			if !contains(err.Error(), "duplicate column name") &&
				!contains(err.Error(), "already exists") {
				return fmt.Errorf("migration failed: %w\nSQL: %s", err, migration)
			}
		}
	}
	return nil
}

func generateUUID() string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uint32(time.Now().UnixNano()),
		uint16(time.Now().Nanosecond()>>8),
		uint16(time.Now().UnixNano()>>16)&0xFFFF,
		uint16(time.Now().UnixNano()>>32)&0xFFFF,
		time.Now().UnixNano())
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Compile-time assertions
var _ Store = (*PostgresStore)(nil)
var _ Store = (*SqliteStore)(nil)

