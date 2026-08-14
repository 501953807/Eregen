package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"eregen.dev/b2b-insurance-integration/internal/model"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// SqliteStore implements Store interface using SQLite.
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

// ---------- Insurance Provider ----------

func (s *SqliteStore) UpdateProvider(ctx context.Context, p *model.InsuranceProvider) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE b2b_insurance_providers SET name=?, code=?, api_endpoint=?, active=?, updated_at=datetime('now') WHERE id=?`,
		p.Name, p.Code, p.APIEndpoint, p.Active, p.ID,
	)
	return err
}

func (s *SqliteStore) CreateProvider(ctx context.Context, p *model.InsuranceProvider) error {
	p.ID = generateUUID()
	p.CreatedAt = time.Now()
	p.Active = true
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO b2b_insurance_providers (id, name, code, api_endpoint, api_key, secret, active, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.Code, p.APIEndpoint, p.APIKey, p.Secret, p.Active, p.CreatedAt,
	)
	return err
}

func (s *SqliteStore) GetProviderByID(ctx context.Context, id string) (*model.InsuranceProvider, error) {
	p := &model.InsuranceProvider{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, code, api_endpoint, active, created_at FROM b2b_insurance_providers WHERE id = ?`, id).Scan(
		&p.ID, &p.Name, &p.Code, &p.APIEndpoint, &p.Active, &p.CreatedAt,
	)
	return p, err
}

func (s *SqliteStore) ListProviders(ctx context.Context, page, pageSize int) ([]model.InsuranceProvider, int, error) {
	offset := (page - 1) * pageSize
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, code, api_endpoint, active, created_at FROM b2b_insurance_providers ORDER BY name LIMIT ? OFFSET ?`,
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
	s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM b2b_insurance_providers").Scan(&total)
	return list, total, nil
}

// ---------- Policy ----------

func (s *SqliteStore) CreatePolicy(ctx context.Context, policy *model.Policy) error {
	policy.ID = generateUUID()
	policy.CreatedAt = time.Now()
	if policy.Status == "" {
		policy.Status = "active"
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO b2b_policies (id, elderly_id, provider_id, plan_name, plan_code, policy_number,
			start_date, end_date, coverage_limit, premium, status, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		policy.ID, policy.ElderlyID, policy.ProviderID, policy.PlanName, policy.PlanCode,
		policy.PolicyNumber, policy.StartDate, policy.EndDate, policy.CoverageLimit,
		policy.Premium, policy.Status, policy.CreatedAt,
	)
	return err
}

func (s *SqliteStore) GetPoliciesForElderly(ctx context.Context, elderlyID string) ([]model.Policy, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, elderly_id, provider_id, plan_name, plan_code, policy_number,
				start_date, end_date, coverage_limit, premium, status, created_at
			 FROM b2b_policies WHERE elderly_id = ? ORDER BY end_date DESC`,
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
	return policies, rows.Err()
}

func (s *SqliteStore) GetPolicyByID(ctx context.Context, id string) (*model.Policy, error) {
	p := &model.Policy{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, elderly_id, provider_id, plan_name, plan_code, policy_number,
				start_date, end_date, coverage_limit, premium, status, created_at
			 FROM b2b_policies WHERE id = ?`, id).Scan(
		&p.ID, &p.ElderlyID, &p.ProviderID, &p.PlanName, &p.PlanCode,
		&p.PolicyNumber, &p.StartDate, &p.EndDate, &p.CoverageLimit,
		&p.Premium, &p.Status, &p.CreatedAt,
	)
	return p, err
}

func (s *SqliteStore) UpdatePolicy(ctx context.Context, policy *model.Policy) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE b2b_policies SET plan_name=?, plan_code=?, policy_number=?,
			 start_date=?, end_date=?, coverage_limit=?, premium=?, status=?, updated_at=datetime('now')
		 WHERE id=?`,
		policy.PlanName, policy.PlanCode, policy.PolicyNumber,
		policy.StartDate, policy.EndDate, policy.CoverageLimit,
		policy.Premium, policy.Status, policy.ID,
	)
	return err
}

// ---------- Claim ----------

func (s *SqliteStore) CreateClaim(ctx context.Context, claim *model.InsuranceClaim) error {
	claim.ID = generateUUID()
	claim.CreatedAt = time.Now()
	claim.UpdatedAt = claim.CreatedAt
	if claim.Status == "" {
		claim.Status = model.ClaimPending
	}

	data, _ := json.Marshal(claim.EvidenceFiles)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO b2b_claims (id, elderly_id, family_member_id, provider_id, claim_type, status,
			 incident_date, claim_amount, coverage_limit, description, evidence_files, submitted_at, reviewed_at,
			 reviewer_notes, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		claim.ID, claim.ElderlyID, claim.FamilyMemberID, claim.ProviderID,
		claim.ClaimType, claim.Status, claim.IncidentDate, claim.ClaimAmount,
		claim.CoverageLimit, claim.Description, data, claim.SubmittedAt, claim.ReviewedAt,
		claim.ReviewerNotes, claim.CreatedAt, claim.UpdatedAt,
	)
	return err
}

func (s *SqliteStore) UpdateClaimStatus(ctx context.Context, claimID string, status model.ClaimStatus, notes string) error {
	now := time.Now()
	_, err := s.db.ExecContext(ctx,
		`UPDATE b2b_claims SET status = ?, reviewed_at = ?, reviewer_notes = ?, updated_at = datetime('now') WHERE id = ?`,
		status, &now, notes, claimID,
	)
	return err
}

func (s *SqliteStore) GetClaimsForElderly(ctx context.Context, elderlyID string) ([]model.InsuranceClaim, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, elderly_id, family_member_id, provider_id, claim_type, status, incident_date,
				claim_amount, coverage_limit, description, evidence_files, submitted_at, reviewed_at,
				reviewer_notes, created_at, updated_at
			 FROM b2b_claims WHERE elderly_id = ? ORDER BY created_at DESC`,
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
	return claims, rows.Err()
}

func (s *SqliteStore) GetClaimByID(ctx context.Context, claimID string) (*model.InsuranceClaim, error) {
	c := &model.InsuranceClaim{}
	var data []byte
	err := s.db.QueryRowContext(ctx,
		`SELECT id, elderly_id, family_member_id, provider_id, claim_type, status, incident_date,
			 claim_amount, coverage_limit, description, evidence_files, submitted_at, reviewed_at,
			 reviewer_notes, created_at, updated_at FROM b2b_claims WHERE id = ?`, claimID).Scan(
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

func (s *SqliteStore) ListClaims(ctx context.Context, status model.ClaimStatus, page, pageSize int) ([]model.InsuranceClaim, int, error) {
	offset := (page - 1) * pageSize
	var q string
	var args []any
	if status != "" {
		q = `SELECT id, elderly_id, family_member_id, provider_id, claim_type, status, incident_date,
			   claim_amount, coverage_limit, description, evidence_files, submitted_at, reviewed_at,
			   reviewer_notes, created_at, updated_at FROM b2b_claims
			   WHERE status = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
		args = append(args, status, pageSize, offset)
	} else {
		q = `SELECT id, elderly_id, family_member_id, provider_id, claim_type, status, incident_date,
			   claim_amount, coverage_limit, description, evidence_files, submitted_at, reviewed_at,
			   reviewer_notes, created_at, updated_at FROM b2b_claims
			   ORDER BY created_at DESC LIMIT ? OFFSET ?`
		args = append(args, pageSize, offset)
	}

	rows, err := s.db.QueryContext(ctx, q, args...)
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
		countQ += " WHERE status = ?"
		s.db.QueryRowContext(ctx, countQ, status).Scan(&total)
	} else {
		s.db.QueryRowContext(ctx, countQ).Scan(&total)
	}
	return claims, total, nil
}

// ---------- Evidence File ----------

func (s *SqliteStore) AddEvidenceFile(ctx context.Context, file *model.EvidenceFile) error {
	file.ID = generateUUID()
	file.UploadedAt = time.Now()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO b2b_evidence_files (id, claim_id, file_type, file_name, file_url, uploaded_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		file.ID, file.ClaimID, file.FileType, file.FileName, file.FileURL, file.UploadedAt,
	)
	return err
}

func (s *SqliteStore) GetEvidenceForClaim(ctx context.Context, claimID string) ([]model.EvidenceFile, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, claim_id, file_type, file_name, file_url, uploaded_at FROM b2b_evidence_files WHERE claim_id = ?`,
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
	return files, rows.Err()
}

// ---------- Health Data Export ----------

func (s *SqliteStore) CreateExport(ctx context.Context, export *model.HealthDataExport) error {
	export.ID = generateUUID()
	export.GeneratedAt = time.Now()
	export.Status = "generating"
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO b2b_health_exports (id, elderly_id, claim_id, export_type, period_start, period_end,
			file_url, generated_at, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		export.ID, export.ElderlyID, export.ClaimID, export.ExportType,
		export.PeriodStart, export.PeriodEnd, export.FileURL, export.GeneratedAt, export.Status,
	)
	return err
}

func (s *SqliteStore) MarkExportReady(ctx context.Context, exportID string, fileURL string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE b2b_health_exports SET status = 'ready', file_url = ?, updated_at = datetime('now') WHERE id = ?`,
		fileURL, exportID,
	)
	return err
}

func (s *SqliteStore) GetExportByID(ctx context.Context, id string) (*model.HealthDataExport, error) {
	e := &model.HealthDataExport{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, elderly_id, claim_id, export_type, period_start, period_end,
			 file_url, generated_at, status FROM b2b_health_exports WHERE id = ?`, id).Scan(
		&e.ID, &e.ElderlyID, &e.ClaimID, &e.ExportType,
		&e.PeriodStart, &e.PeriodEnd, &e.FileURL, &e.GeneratedAt, &e.Status,
	)
	return e, err
}

// ---------- Premium Reminder ----------

func (s *SqliteStore) CreateReminder(ctx context.Context, reminder *model.PremiumReminder) error {
	reminder.ID = generateUUID()
	reminder.CreatedAt = time.Now()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO b2b_premium_reminders (id, policy_id, elderly_id, family_id, remind_date, amount, sent, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		reminder.ID, reminder.PolicyID, reminder.ElderlyID, reminder.FamilyID,
		reminder.RemindDate, reminder.Amount, reminder.Sent, reminder.CreatedAt,
	)
	return err
}

func (s *SqliteStore) GetUpcomingReminders(ctx context.Context, daysAhead int) ([]model.PremiumReminder, error) {
	threshold := time.Now().AddDate(0, 0, daysAhead)
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, policy_id, elderly_id, family_id, remind_date, amount, sent, created_at
		 FROM b2b_premium_reminders WHERE remind_date <= ? AND sent = 0 ORDER BY remind_date ASC`,
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
	return reminders, rows.Err()
}

// ---------- Migration ----------

func migrate(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS b2b_insurance_providers (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			code TEXT UNIQUE,
			api_endpoint TEXT,
			api_key TEXT,
			secret TEXT,
			active INTEGER DEFAULT 1,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS b2b_policies (
			id TEXT PRIMARY KEY,
			elderly_id TEXT NOT NULL,
			provider_id TEXT NOT NULL,
			plan_name TEXT,
			plan_code TEXT,
			policy_number TEXT,
			start_date TEXT NOT NULL,
			end_date TEXT NOT NULL,
			coverage_limit REAL,
			premium REAL,
			status TEXT DEFAULT 'active',
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS b2b_claims (
			id TEXT PRIMARY KEY,
			elderly_id TEXT NOT NULL,
			family_member_id TEXT,
			provider_id TEXT NOT NULL,
			claim_type TEXT,
			status TEXT DEFAULT 'pending',
			incident_date TEXT,
			claim_amount REAL,
			coverage_limit REAL,
			description TEXT,
			evidence_files TEXT DEFAULT '[]',
			submitted_at TEXT,
			reviewed_at TEXT,
			reviewer_notes TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS b2b_evidence_files (
			id TEXT PRIMARY KEY,
			claim_id TEXT NOT NULL,
			file_type TEXT,
			file_name TEXT,
			file_url TEXT,
			uploaded_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS b2b_health_exports (
			id TEXT PRIMARY KEY,
			elderly_id TEXT NOT NULL,
			claim_id TEXT,
			export_type TEXT,
			period_start TEXT NOT NULL,
			period_end TEXT NOT NULL,
			file_url TEXT,
			generated_at TEXT DEFAULT CURRENT_TIMESTAMP,
			status TEXT DEFAULT 'generating'
		)`,
		`CREATE TABLE IF NOT EXISTS b2b_premium_reminders (
			id TEXT PRIMARY KEY,
			policy_id TEXT NOT NULL,
			elderly_id TEXT NOT NULL,
			family_id TEXT,
			remind_date TEXT NOT NULL,
			amount REAL,
			sent INTEGER DEFAULT 0,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
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
	return uuid.New().String()
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Compile-time assertions
var _ Store = (*SqliteStore)(nil)
