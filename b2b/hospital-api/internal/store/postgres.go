package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"eregen.dev/b2b-hospital-api/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Postgres struct {
	pool *pgxpool.Pool
	log  *zap.Logger
}

func NewPostgres(pool *pgxpool.Pool, log *zap.Logger) *Postgres {
	return &Postgres{pool: pool, log: log}
}

// ---------- Institution ----------

func (s *Postgres) CreateInstitution(ctx context.Context, inst *model.Institution) error {
	inst.ID = uuid.New().String()
	now := time.Now()
	inst.CreatedAt = now
	inst.UpdatedAt = now
	if inst.Status == "" {
		inst.Status = "pending"
	}

	q := `INSERT INTO b2b_institutions (id, name, type, code, contact_name, contact_phone,
		   access_level, status, created_at, updated_at)
		   VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := s.pool.Exec(ctx, q,
		inst.ID, inst.Name, inst.Type, inst.Code,
		inst.ContactName, inst.ContactPhone, inst.AccessLevel, inst.Status,
		now, now,
	)
	return err
}

func (s *Postgres) GetInstitutionByID(ctx context.Context, id string) (*model.Institution, error) {
	inst := &model.Institution{}
	q := `SELECT id, name, type, code, contact_name, contact_phone, access_level, status, created_at, updated_at
		   FROM b2b_institutions WHERE id = $1`
	err := s.pool.QueryRow(ctx, q, id).Scan(
		&inst.ID, &inst.Name, &inst.Type, &inst.Code,
		&inst.ContactName, &inst.ContactPhone, &inst.AccessLevel, &inst.Status,
		&inst.CreatedAt, &inst.UpdatedAt,
	)
	return inst, err
}

func (s *Postgres) GetInstitutionByCode(ctx context.Context, code string) (*model.Institution, error) {
	inst := &model.Institution{}
	q := `SELECT id, name, type, code, contact_name, contact_phone, access_level, status, created_at, updated_at
		   FROM b2b_institutions WHERE code = $1 AND status = 'active'`
	err := s.pool.QueryRow(ctx, q, code).Scan(
		&inst.ID, &inst.Name, &inst.Type, &inst.Code,
		&inst.ContactName, &inst.ContactPhone, &inst.AccessLevel, &inst.Status,
		&inst.CreatedAt, &inst.UpdatedAt,
	)
	return inst, err
}

func (s *Postgres) ListInstitutions(ctx context.Context, page, pageSize int) ([]model.Institution, int, error) {
	offset := (page - 1) * pageSize
	q := fmt.Sprintf(`SELECT id, name, type, code, contact_name, contact_phone, access_level, status, created_at, updated_at
					   FROM b2b_institutions ORDER BY created_at DESC LIMIT $1 OFFSET $2`)
	rows, err := s.pool.Query(ctx, q, pageSize, offset)
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
	s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM b2b_institutions").Scan(&total)
	return list, total, nil
}

// ---------- API Key ----------

func (s *Postgres) CreateAPIKey(ctx context.Context, key *model.InstitutionAPIKey) error {
	key.ID = uuid.New().String()
	key.CreatedAt = time.Now()
	if !key.Active {
		key.Active = true
	}
	q := `INSERT INTO b2b_api_keys (id, institution_id, key_hash, name, expires_at, active, created_at)
		   VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := s.pool.Exec(ctx, q,
		key.ID, key.InstitutionID, key.KeyHash, key.Name, key.ExpiresAt, key.Active, key.CreatedAt,
	)
	return err
}

func (s *Postgres) GetInstitutionByAPIKey(ctx context.Context, keyHash string) (*model.Institution, error) {
	inst := &model.Institution{}
	q := `SELECT i.id, i.name, i.type, i.code, i.contact_name, i.contact_phone,
				i.access_level, i.status, i.created_at, i.updated_at
		   FROM b2b_api_keys k
		   JOIN b2b_institutions i ON i.id = k.institution_id
		   WHERE k.key_hash = $1 AND k.active = true AND k.expires_at > now() AND i.status = 'active'`
	err := s.pool.QueryRow(ctx, q, keyHash).Scan(
		&inst.ID, &inst.Name, &inst.Type, &inst.Code,
		&inst.ContactName, &inst.ContactPhone, &inst.AccessLevel, &inst.Status,
		&inst.CreatedAt, &inst.UpdatedAt,
	)
	return inst, err
}

// ---------- Elderly-Institution Link ----------

func (s *Postgres) LinkElderlyToInstitution(ctx context.Context, link *model.ElderlyInstitutionLink) error {
	link.ID = uuid.New().String()
	link.Active = true
	link.CreatedAt = time.Now()
	link.UpdatedAt = link.CreatedAt

	var data []byte
	if len(link.Notes) > 0 {
		data = link.Notes
	}
	q := `INSERT INTO b2b_elderly_links (id, elderly_id, institution_id, admitted_at, discharged_at,
		   primary_doc, notes, active, created_at, updated_at)
		   VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := s.pool.Exec(ctx, q,
		link.ID, link.ElderlyID, link.InstitutionID,
		link.AdmittedAt, link.DischargedAt, link.PrimaryDoc, data, link.Active,
		link.CreatedAt, link.UpdatedAt,
	)
	return err
}

func (s *Postgres) GetActiveLinksForInstitution(ctx context.Context, instID string) ([]model.ElderlyInstitutionLink, error) {
	q := `SELECT id, elderly_id, institution_id, admitted_at, discharged_at, primary_doc, notes, active, created_at, updated_at
		   FROM b2b_elderly_links WHERE institution_id = $1 AND active = true
		   ORDER BY created_at DESC`
	rows, err := s.pool.Query(ctx, q, instID)
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
	return links, nil
}

func (s *Postgres) GetActiveLinksForElderly(ctx context.Context, elderlyID string) ([]model.ElderlyInstitutionLink, error) {
	q := `SELECT id, elderly_id, institution_id, admitted_at, discharged_at, primary_doc, notes, active, created_at, updated_at
		   FROM b2b_elderly_links WHERE elderly_id = $1 AND active = true
		   ORDER BY created_at DESC`
	rows, err := s.pool.Query(ctx, q, elderlyID)
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
	return links, nil
}

// ---------- Health Data ----------

func (s *Postgres) StoreVitals(ctx context.Context, v *model.VitalSignRecord) error {
	q := `INSERT INTO b2b_vital_signs (id, elderly_id, institution_id, patient_id,
		   heart_rate, spo2, systolic_bp, diastolic_bp, temperature, steps, recorded_at)
		   VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	_, err := s.pool.Exec(ctx, q,
		v.ID, v.ElderlyID, v.InstitutionID, v.PatientID,
		v.HeartRate, v.SPO2, v.SystolicBP, v.DiastolicBP,
		v.Temperature, v.Steps, v.RecordedAt,
	)
	return err
}

func (s *Postgres) BulkStoreVitals(ctx context.Context, vitals []*model.VitalSignRecord) error {
	for _, v := range vitals {
		if err := s.StoreVitals(ctx, v); err != nil {
			s.log.Warn("store vital sign", zap.Error(err))
		}
	}
	return nil
}

func (s *Postgres) GetVitalsForElderly(ctx context.Context, elderlyID string, days int) ([]model.VitalSignRecord, error) {
	q := `SELECT id, elderly_id, institution_id, patient_id,
		   heart_rate, spo2, systolic_bp, diastolic_bp, temperature, steps, recorded_at
		   FROM b2b_vital_signs WHERE elderly_id = $1 AND recorded_at > now() - interval $2 day
		   ORDER BY recorded_at DESC`
	rows, err := s.pool.Query(ctx, q, elderlyID, fmt.Sprintf("%d days", days))
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

func (s *Postgres) LinkElderlyToExternalPatient(ctx context.Context, elderlyID, patientID, eregenID string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO b2b_patient_links (id, external_patient_id, local_elderly_id, created_at)
		 VALUES ($1, $2, $3, now())`,
		uuid.New().String(), patientID, eregenID)
	return err
}

func (s *Postgres) FindElderlyByExternalPatient(ctx context.Context, patientID string) (string, error) {
	var elderlyID string
	err := s.pool.QueryRow(ctx,
		`SELECT local_elderly_id FROM b2b_patient_links WHERE external_patient_id = $1`,
		patientID).Scan(&elderlyID)
	return elderlyID, err
}
