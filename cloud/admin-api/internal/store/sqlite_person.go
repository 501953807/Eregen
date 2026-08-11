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

// Helper functions for date handling
func formatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

func parseDateOrNil(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}

func nullableString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// PersonStore implementation

func (s *SqliteStore) CreatePerson(ctx context.Context, p *model.Person) error {
	p.ID = uuid.New().String()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO persons (id, id_card, name, gender, birth_date, phone, emergency_contact, address, avatar_url, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.IDCard, p.Name, p.Gender,
		formatDate(p.BirthDate), p.Phone, p.EmergencyContact, p.Address,
		nullableString(p.AvatarURL), p.Status)
	return err
}

func (s *SqliteStore) GetPerson(ctx context.Context, id string) (*model.Person, error) {
	var p model.Person
	var birthRaw, avatarRaw sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT id, id_card, name, gender, birth_date, phone, emergency_contact, address, avatar_url, status, created_at, updated_at
		 FROM persons WHERE id = ?`, id).Scan(
		&p.ID, &p.IDCard, &p.Name, &p.Gender, &birthRaw, &p.Phone, &p.EmergencyContact,
		&p.Address, &avatarRaw, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("person not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get person: %w", err)
	}
	p.BirthDate = parseDateOrNil(birthRaw.String)
	if avatarRaw.Valid && avatarRaw.String != "" {
		p.AvatarURL = &avatarRaw.String
	}
	return &p, nil
}

func (s *SqliteStore) GetPersonByIDCard(ctx context.Context, idCard string) (*model.Person, error) {
	var p model.Person
	var birthRaw, avatarRaw sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT id, id_card, name, gender, birth_date, phone, emergency_contact, address, avatar_url, status, created_at, updated_at
		 FROM persons WHERE id_card = ?`, idCard).Scan(
		&p.ID, &p.IDCard, &p.Name, &p.Gender, &birthRaw, &p.Phone, &p.EmergencyContact,
		&p.Address, &avatarRaw, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get person by id_card: %w", err)
	}
	p.BirthDate = parseDateOrNil(birthRaw.String)
	if avatarRaw.Valid && avatarRaw.String != "" {
		p.AvatarURL = &avatarRaw.String
	}
	return &p, nil
}

func (s *SqliteStore) ListPersons(ctx context.Context, page, pageSize int, businessChain, status string) ([]model.Person, error) {
	query := `SELECT id, id_card, name, gender, birth_date, phone, emergency_contact, address, avatar_url, status, created_at, updated_at
			  FROM persons WHERE 1=1`
	args := []any{}
	if businessChain != "" {
		query += ` AND id IN (SELECT person_id FROM person_profiles WHERE business_chain = ?)`
		args = append(args, businessChain)
	}
	if status != "" {
		query += ` AND status = ?`
		args = append(args, status)
	}
	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list persons: %w", err)
	}
	defer rows.Close()

	var persons []model.Person
	for rows.Next() {
		var p model.Person
		var birthRaw, avatarRaw sql.NullString
		if err := rows.Scan(&p.ID, &p.IDCard, &p.Name, &p.Gender, &birthRaw, &p.Phone, &p.EmergencyContact,
			&p.Address, &avatarRaw, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan person: %w", err)
		}
		p.BirthDate = parseDateOrNil(birthRaw.String)
		if avatarRaw.Valid && avatarRaw.String != "" {
			p.AvatarURL = &avatarRaw.String
		}
		persons = append(persons, p)
	}
	return persons, rows.Err()
}

func (s *SqliteStore) UpdatePerson(ctx context.Context, id string, updates map[string]any) error {
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
		fmt.Sprintf("UPDATE persons SET %s, updated_at = datetime('now') WHERE id = ?", strings.Join(setClauses, ", ")), args...)
	return err
}

func (s *SqliteStore) DeletePerson(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM persons WHERE id = ?`, id)
	return err
}

// Profile methods

func (s *SqliteStore) CreateProfile(ctx context.Context, pp *model.PersonProfile) error {
	pp.CreatedAt = time.Now()
	pp.UpdatedAt = time.Now()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO person_profiles (person_id, business_chain, subscription_tier, subscription_status,
		 subscription_start, subscription_end, health_risk_level, admission_no, department, bed_number,
		 blood_type, attending_doctor, diagnosis, admission_date, expected_discharge_date, discharge_date,
		 discharge_type, hospital_id, hospital_id_community, minzheng_certified, subsidy_type,
		 certification_date, certification_doc, next_review_date, linked_person_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		pp.PersonID, pp.BusinessChain, pp.SubscriptionTier, pp.SubscriptionStatus,
		formatDate(pp.SubscriptionStart), formatDate(pp.SubscriptionEnd), pp.HealthRiskLevel,
		pp.AdmissionNo, pp.Department, pp.BedNumber, pp.BloodType, pp.AttendingDoctor, pp.Diagnosis,
		formatDate(pp.AdmissionDate), formatDate(pp.ExpectedDischarge), formatDate(pp.DischargeDate),
		pp.DischargeType, pp.HospitalID, pp.HospitalIDCommunity, pp.MinzhengCertified, pp.SubsidyType,
		formatDate(pp.CertificationDate), pp.CertificationDoc, formatDate(pp.NextReviewDate), pp.LinkedPersonID)
	return err
}

func (s *SqliteStore) GetProfile(ctx context.Context, personID string, chain model.BusinessChain) (*model.PersonProfile, error) {
	var pp model.PersonProfile
	var subStartRaw, subEndRaw, admDateRaw, expDischRaw, dischDateRaw, certDateRaw, nextRevRaw string
	err := s.db.QueryRowContext(ctx,
		`SELECT person_id, business_chain, subscription_tier, subscription_status, subscription_start, subscription_end,
		 health_risk_level, admission_no, department, bed_number, blood_type, attending_doctor, diagnosis,
		 admission_date, expected_discharge_date, discharge_date, discharge_type, hospital_id, hospital_id_community,
		 minzheng_certified, subsidy_type, certification_date, certification_doc, next_review_date, linked_person_id,
		 created_at, updated_at
		 FROM person_profiles WHERE person_id = ? AND business_chain = ?`, personID, chain).Scan(
		&pp.PersonID, &pp.BusinessChain, &pp.SubscriptionTier, &pp.SubscriptionStatus, &subStartRaw, &subEndRaw,
		&pp.HealthRiskLevel, &pp.AdmissionNo, &pp.Department, &pp.BedNumber, &pp.BloodType, &pp.AttendingDoctor,
		&pp.Diagnosis, &admDateRaw, &expDischRaw, &dischDateRaw, &pp.DischargeType, &pp.HospitalID,
		&pp.HospitalIDCommunity, &pp.MinzhengCertified, &pp.SubsidyType, &certDateRaw, &pp.CertificationDoc,
		&nextRevRaw, &pp.CreatedAt, &pp.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	pp.SubscriptionStart = parseDateOrNil(subStartRaw)
	pp.SubscriptionEnd = parseDateOrNil(subEndRaw)
	pp.AdmissionDate = parseDateOrNil(admDateRaw)
	pp.ExpectedDischarge = parseDateOrNil(expDischRaw)
	pp.DischargeDate = parseDateOrNil(dischDateRaw)
	pp.CertificationDate = parseDateOrNil(certDateRaw)
	pp.NextReviewDate = parseDateOrNil(nextRevRaw)
	return &pp, nil
}

func (s *SqliteStore) ListProfiles(ctx context.Context, chain model.BusinessChain) ([]model.PersonProfile, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT person_id, business_chain, subscription_tier, subscription_status, subscription_start, subscription_end,
		 health_risk_level, admission_no, department, bed_number, blood_type, attending_doctor, diagnosis,
		 admission_date, expected_discharge_date, discharge_date, discharge_type, hospital_id, hospital_id_community,
		 minzheng_certified, subsidy_type, certification_date, certification_doc, next_review_date, linked_person_id,
		 created_at, updated_at
		 FROM person_profiles WHERE business_chain = ? ORDER BY created_at DESC`, chain)
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}
	defer rows.Close()

	var profiles []model.PersonProfile
	for rows.Next() {
		var pp model.PersonProfile
		var subStartRaw, subEndRaw, admDateRaw, expDischRaw, dischDateRaw, certDateRaw, nextRevRaw string
		if err := rows.Scan(&pp.PersonID, &pp.BusinessChain, &pp.SubscriptionTier, &pp.SubscriptionStatus,
			&subStartRaw, &subEndRaw, &pp.HealthRiskLevel, &pp.AdmissionNo, &pp.Department, &pp.BedNumber,
			&pp.BloodType, &pp.AttendingDoctor, &pp.Diagnosis, &admDateRaw, &expDischRaw, &dischDateRaw,
			&pp.DischargeType, &pp.HospitalID, &pp.HospitalIDCommunity, &pp.MinzhengCertified, &pp.SubsidyType,
			&certDateRaw, &pp.CertificationDoc, &nextRevRaw, &pp.CreatedAt, &pp.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan profile: %w", err)
		}
		pp.SubscriptionStart = parseDateOrNil(subStartRaw)
		pp.SubscriptionEnd = parseDateOrNil(subEndRaw)
		pp.AdmissionDate = parseDateOrNil(admDateRaw)
		pp.ExpectedDischarge = parseDateOrNil(expDischRaw)
		pp.DischargeDate = parseDateOrNil(dischDateRaw)
		pp.CertificationDate = parseDateOrNil(certDateRaw)
		pp.NextReviewDate = parseDateOrNil(nextRevRaw)
		profiles = append(profiles, pp)
	}
	return profiles, rows.Err()
}

func (s *SqliteStore) UpdateProfile(ctx context.Context, pp *model.PersonProfile) error {
	pp.UpdatedAt = time.Now()
	_, err := s.db.ExecContext(ctx,
		`UPDATE person_profiles SET subscription_tier=?, subscription_status=?, subscription_start=?, subscription_end=?,
		 health_risk_level=?, admission_no=?, department=?, bed_number=?, blood_type=?, attending_doctor=?,
		 diagnosis=?, admission_date=?, expected_discharge_date=?, discharge_date=?, discharge_type=?,
		 hospital_id=?, hospital_id_community=?, minzheng_certified=?, subsidy_type=?, certification_date=?,
		 certification_doc=?, next_review_date=?, linked_person_id=?, updated_at=?
		 WHERE person_id=? AND business_chain=?`,
		pp.SubscriptionTier, pp.SubscriptionStatus, formatDate(pp.SubscriptionStart), formatDate(pp.SubscriptionEnd),
		pp.HealthRiskLevel, pp.AdmissionNo, pp.Department, pp.BedNumber, pp.BloodType, pp.AttendingDoctor,
		pp.Diagnosis, formatDate(pp.AdmissionDate), formatDate(pp.ExpectedDischarge), formatDate(pp.DischargeDate),
		pp.DischargeType, pp.HospitalID, pp.HospitalIDCommunity, pp.MinzhengCertified, pp.SubsidyType,
		formatDate(pp.CertificationDate), pp.CertificationDoc, formatDate(pp.NextReviewDate), pp.LinkedPersonID,
		pp.UpdatedAt, pp.PersonID, pp.BusinessChain)
	return err
}

// Welfare tag methods

func (s *SqliteStore) AssignPersonWelfareTag(ctx context.Context, wt *model.PersonWelfareTag) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO person_welfare_tags (person_id, tag_code, valid_from, valid_to)
		 VALUES (?, ?, ?, ?)`, wt.PersonID, wt.TagCode, wt.ValidFrom, wt.ValidTo)
	return err
}

func (s *SqliteStore) RevokePersonWelfareTag(ctx context.Context, personID, tagCode string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM person_welfare_tags WHERE person_id = ? AND tag_code = ?`, personID, tagCode)
	return err
}

func (s *SqliteStore) ListPersonWelfareTags(ctx context.Context, personID string) ([]model.PersonWelfareTag, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT person_id, tag_code, valid_from, valid_to FROM person_welfare_tags WHERE person_id = ?`, personID)
	if err != nil {
		return nil, fmt.Errorf("list welfare tags: %w", err)
	}
	defer rows.Close()

	var tags []model.PersonWelfareTag
	for rows.Next() {
		var wt model.PersonWelfareTag
		if err := rows.Scan(&wt.PersonID, &wt.TagCode, &wt.ValidFrom, &wt.ValidTo); err != nil {
			return nil, fmt.Errorf("scan welfare tag: %w", err)
		}
		tags = append(tags, wt)
	}
	return tags, rows.Err()
}
