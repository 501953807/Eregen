package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"eregen.dev/b2b-insurance-integration/internal/model"

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

// ---------- Insurance Provider ----------

func (s *Postgres) UpdateProvider(ctx context.Context, p *model.InsuranceProvider) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE b2b_insurance_providers SET name=$1, code=$2, api_endpoint=$3, active=$4, updated_at=now() WHERE id=$5`,
		p.Name, p.Code, p.APIEndpoint, p.Active, p.ID,
	)
	return err
}

func (s *Postgres) CreateProvider(ctx context.Context, p *model.InsuranceProvider) error {
	p.ID = uuid.New().String()
	p.CreatedAt = time.Now()
	p.Active = true
	q := `INSERT INTO b2b_insurance_providers (id, name, code, api_endpoint, api_key, secret, active, created_at)
		   VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := s.pool.Exec(ctx, q,
		p.ID, p.Name, p.Code, p.APIEndpoint, p.APIKey, p.Secret, p.Active, p.CreatedAt,
	)
	return err
}

func (s *Postgres) GetProviderByID(ctx context.Context, id string) (*model.InsuranceProvider, error) {
	p := &model.InsuranceProvider{}
	q := `SELECT id, name, code, api_endpoint, active, created_at FROM b2b_insurance_providers WHERE id = $1`
	err := s.pool.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.Name, &p.Code, &p.APIEndpoint, &p.Active, &p.CreatedAt,
	)
	return p, err
}

func (s *Postgres) ListProviders(ctx context.Context, page, pageSize int) ([]model.InsuranceProvider, int, error) {
	offset := (page - 1) * pageSize
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, code, api_endpoint, active, created_at FROM b2b_insurance_providers ORDER BY name LIMIT $1 OFFSET $2`,
		pageSize, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.InsuranceProvider
	for rows.Next() {
		var p model.InsuranceProvider
		if err := rows.Scan(&p.ID, &p.Name, &p.Code, &p.APIEndpoint, &p.Active, &p.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, p)
	}

	var total int
	s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM b2b_insurance_providers").Scan(&total)
	return list, total, nil
}

// ---------- Policy ----------

func (s *Postgres) CreatePolicy(ctx context.Context, policy *model.Policy) error {
	policy.ID = uuid.New().String()
	policy.CreatedAt = time.Now()
	if policy.Status == "" {
		policy.Status = "active"
	}
	q := `INSERT INTO b2b_policies (id, elderly_id, provider_id, plan_name, plan_code, policy_number,
		   start_date, end_date, coverage_limit, premium, status, created_at)
		   VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`
	_, err := s.pool.Exec(ctx, q,
		policy.ID, policy.ElderlyID, policy.ProviderID, policy.PlanName, policy.PlanCode,
		policy.PolicyNumber, policy.StartDate, policy.EndDate, policy.CoverageLimit,
		policy.Premium, policy.Status, policy.CreatedAt,
	)
	return err
}

func (s *Postgres) GetPoliciesForElderly(ctx context.Context, elderlyID string) ([]model.Policy, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, elderly_id, provider_id, plan_name, plan_code, policy_number,
				start_date, end_date, coverage_limit, premium, status, created_at
			 FROM b2b_policies WHERE elderly_id = $1 ORDER BY end_date DESC`,
		elderlyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []model.Policy
	for rows.Next() {
		var p model.Policy
		if err := rows.Scan(&p.ID, &p.ElderlyID, &p.ProviderID, &p.PlanName, &p.PlanCode,
			&p.PolicyNumber, &p.StartDate, &p.EndDate, &p.CoverageLimit,
			&p.Premium, &p.Status, &p.CreatedAt); err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, nil
}

func (s *Postgres) GetPolicyByID(ctx context.Context, id string) (*model.Policy, error) {
	p := &model.Policy{}
	q := `SELECT id, elderly_id, provider_id, plan_name, plan_code, policy_number,
		   start_date, end_date, coverage_limit, premium, status, created_at
		   FROM b2b_policies WHERE id = $1`
	err := s.pool.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.ElderlyID, &p.ProviderID, &p.PlanName, &p.PlanCode,
		&p.PolicyNumber, &p.StartDate, &p.EndDate, &p.CoverageLimit,
		&p.Premium, &p.Status, &p.CreatedAt,
	)
	return p, err
}

// ---------- Claim ----------

func (s *Postgres) CreateClaim(ctx context.Context, claim *model.InsuranceClaim) error {
	claim.ID = uuid.New().String()
	claim.CreatedAt = time.Now()
	claim.UpdatedAt = claim.CreatedAt
	if claim.Status == "" {
		claim.Status = model.ClaimPending
	}

	data, err := json.Marshal(claim.EvidenceFiles)
	if err != nil {
		data = []byte("[]")
	}
	q := `INSERT INTO b2b_claims (id, elderly_id, family_member_id, provider_id, claim_type, status,
		   incident_date, claim_amount, coverage_limit, description, evidence_files, created_at, updated_at)
		   VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`
	_, err = s.pool.Exec(ctx, q,
		claim.ID, claim.ElderlyID, claim.FamilyMemberID, claim.ProviderID,
		claim.ClaimType, claim.Status, claim.IncidentDate, claim.ClaimAmount,
		claim.CoverageLimit, claim.Description, data, claim.CreatedAt, claim.UpdatedAt,
	)
	return err
}

func (s *Postgres) UpdateClaimStatus(ctx context.Context, claimID string, status model.ClaimStatus, notes string) error {
	now := time.Now()
	q := `UPDATE b2b_claims SET status = $1, reviewed_at = $2, reviewer_notes = $3, updated_at = $4 WHERE id = $5`
	_, err := s.pool.Exec(ctx, q, status, &now, notes, now, claimID)
	return err
}

func (s *Postgres) GetClaimsForElderly(ctx context.Context, elderlyID string) ([]model.InsuranceClaim, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, elderly_id, family_member_id, provider_id, claim_type, status, incident_date,
				claim_amount, coverage_limit, description, evidence_files, submitted_at, reviewed_at,
				reviewer_notes, created_at, updated_at
			 FROM b2b_claims WHERE elderly_id = $1 ORDER BY created_at DESC`,
		elderlyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var claims []model.InsuranceClaim
	for rows.Next() {
		var c model.InsuranceClaim
		var data []byte
		if err := rows.Scan(&c.ID, &c.ElderlyID, &c.FamilyMemberID, &c.ProviderID,
			&c.ClaimType, &c.Status, &c.IncidentDate, &c.ClaimAmount, &c.CoverageLimit,
			&c.Description, &data, &c.SubmittedAt, &c.ReviewedAt, &c.ReviewerNotes,
			&c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(data, &c.EvidenceFiles)
		claims = append(claims, c)
	}
	return claims, nil
}

func (s *Postgres) GetClaimByID(ctx context.Context, claimID string) (*model.InsuranceClaim, error) {
	c := &model.InsuranceClaim{}
	q := `SELECT id, elderly_id, family_member_id, provider_id, claim_type, status, incident_date,
		   claim_amount, coverage_limit, description, evidence_files, submitted_at, reviewed_at,
		   reviewer_notes, created_at, updated_at FROM b2b_claims WHERE id = $1`
	var data []byte
	err := s.pool.QueryRow(ctx, q, claimID).Scan(
		&c.ID, &c.ElderlyID, &c.FamilyMemberID, &c.ProviderID, &c.ClaimType, &c.Status,
		&c.IncidentDate, &c.ClaimAmount, &c.CoverageLimit, &c.Description, &data,
		&c.SubmittedAt, &c.ReviewedAt, &c.ReviewerNotes, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(data, &c.EvidenceFiles)
	return c, nil
}

func (s *Postgres) ListClaims(ctx context.Context, status model.ClaimStatus, page, pageSize int) ([]model.InsuranceClaim, int, error) {
	offset := (page - 1) * pageSize
	var q string
	var args []any
	if status != "" {
		q = fmt.Sprintf(`SELECT id, elderly_id, family_member_id, provider_id, claim_type, status, incident_date,
			   claim_amount, coverage_limit, description, evidence_files, submitted_at, reviewed_at,
			   reviewer_notes, created_at, updated_at FROM b2b_claims
			   WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`)
		args = append(args, status, pageSize, offset)
	} else {
		q = fmt.Sprintf(`SELECT id, elderly_id, family_member_id, provider_id, claim_type, status, incident_date,
			   claim_amount, coverage_limit, description, evidence_files, submitted_at, reviewed_at,
			   reviewer_notes, created_at, updated_at FROM b2b_claims
			   ORDER BY created_at DESC LIMIT $1 OFFSET $2`)
		args = append(args, pageSize, offset)
	}

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var claims []model.InsuranceClaim
	for rows.Next() {
		var c model.InsuranceClaim
		var data []byte
		if err := rows.Scan(&c.ID, &c.ElderlyID, &c.FamilyMemberID, &c.ProviderID,
			&c.ClaimType, &c.Status, &c.IncidentDate, &c.ClaimAmount, &c.CoverageLimit,
			&c.Description, &data, &c.SubmittedAt, &c.ReviewedAt, &c.ReviewerNotes,
			&c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		json.Unmarshal(data, &c.EvidenceFiles)
		claims = append(claims, c)
	}

	var total int
	countQ := "SELECT COUNT(*) FROM b2b_claims"
	if status != "" {
		countQ += " WHERE status = $1"
		s.pool.QueryRow(ctx, countQ, status).Scan(&total)
	} else {
		s.pool.QueryRow(ctx, countQ).Scan(&total)
	}
	return claims, total, nil
}

// ---------- Evidence File ----------

func (s *Postgres) AddEvidenceFile(ctx context.Context, file *model.EvidenceFile) error {
	file.ID = uuid.New().String()
	file.UploadedAt = time.Now()
	q := `INSERT INTO b2b_evidence_files (id, claim_id, file_type, file_name, file_url, uploaded_at)
		   VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := s.pool.Exec(ctx, q,
		file.ID, file.ClaimID, file.FileType, file.FileName, file.FileURL, file.UploadedAt,
	)
	return err
}

func (s *Postgres) GetEvidenceForClaim(ctx context.Context, claimID string) ([]model.EvidenceFile, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, claim_id, file_type, file_name, file_url, uploaded_at FROM b2b_evidence_files WHERE claim_id = $1`,
		claimID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []model.EvidenceFile
	for rows.Next() {
		var f model.EvidenceFile
		if err := rows.Scan(&f.ID, &f.ClaimID, &f.FileType, &f.FileName, &f.FileURL, &f.UploadedAt); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}

// ---------- Health Data Export ----------

func (s *Postgres) CreateExport(ctx context.Context, export *model.HealthDataExport) error {
	export.ID = uuid.New().String()
	export.GeneratedAt = time.Now()
	export.Status = "generating"
	q := `INSERT INTO b2b_health_exports (id, elderly_id, claim_id, export_type, period_start, period_end,
		   file_url, generated_at, status)
		   VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := s.pool.Exec(ctx, q,
		export.ID, export.ElderlyID, export.ClaimID, export.ExportType,
		export.PeriodStart, export.PeriodEnd, export.FileURL, export.GeneratedAt, export.Status,
	)
	return err
}

func (s *Postgres) MarkExportReady(ctx context.Context, exportID string, fileURL string) error {
	q := `UPDATE b2b_health_exports SET status = 'ready', file_url = $1, updated_at = now() WHERE id = $2`
	_, err := s.pool.Exec(ctx, q, fileURL, exportID)
	return err
}

func (s *Postgres) GetExportByID(ctx context.Context, id string) (*model.HealthDataExport, error) {
	e := &model.HealthDataExport{}
	q := `SELECT id, elderly_id, claim_id, export_type, period_start, period_end,
		   file_url, generated_at, status FROM b2b_health_exports WHERE id = $1`
	err := s.pool.QueryRow(ctx, q, id).Scan(
		&e.ID, &e.ElderlyID, &e.ClaimID, &e.ExportType,
		&e.PeriodStart, &e.PeriodEnd, &e.FileURL, &e.GeneratedAt, &e.Status,
	)
	return e, err
}

func (s *Postgres) UpdatePolicy(ctx context.Context, policy *model.Policy) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE b2b_policies SET plan_name=$1, plan_code=$2, policy_number=$3,
			 start_date=$4, end_date=$5, coverage_limit=$6, premium=$7, status=$8, updated_at=now()
		 WHERE id=$9`,
		policy.PlanName, policy.PlanCode, policy.PolicyNumber,
		policy.StartDate, policy.EndDate, policy.CoverageLimit,
		policy.Premium, policy.Status, policy.ID,
	)
	return err
}

// ---------- Premium Reminder ----------

func (s *Postgres) CreateReminder(ctx context.Context, reminder *model.PremiumReminder) error {
	reminder.ID = uuid.New().String()
	reminder.CreatedAt = time.Now()
	q := `INSERT INTO b2b_premium_reminders (id, policy_id, elderly_id, family_id, remind_date, amount, sent, created_at)
		   VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := s.pool.Exec(ctx, q,
		reminder.ID, reminder.PolicyID, reminder.ElderlyID, reminder.FamilyID,
		reminder.RemindDate, reminder.Amount, reminder.Sent, reminder.CreatedAt,
	)
	return err
}

func (s *Postgres) GetUpcomingReminders(ctx context.Context, daysAhead int) ([]model.PremiumReminder, error) {
	threshold := time.Now().AddDate(0, 0, daysAhead)
	rows, err := s.pool.Query(ctx,
		`SELECT id, policy_id, elderly_id, family_id, remind_date, amount, sent, created_at
		   FROM b2b_premium_reminders WHERE remind_date <= $1 AND sent = false ORDER BY remind_date ASC`,
		threshold,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []model.PremiumReminder
	for rows.Next() {
		var r model.PremiumReminder
		if err := rows.Scan(&r.ID, &r.PolicyID, &r.ElderlyID, &r.FamilyID,
			&r.RemindDate, &r.Amount, &r.Sent, &r.CreatedAt); err != nil {
			return nil, err
		}
		reminders = append(reminders, r)
	}
	return reminders, nil
}

// ---------- Helper ----------
