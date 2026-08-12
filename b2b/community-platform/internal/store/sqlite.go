package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"eregen.dev/b2b-community-platform/internal/model"

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

// ---------- Community Event ----------

func (s *SqliteStore) CreateEvent(ctx context.Context, evt *model.CommunityEvent) error {
	evt.ID = generateUUID()
	evt.CreatedAt = time.Now()
	if evt.Status == "" {
		evt.Status = "scheduled"
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO b2b_events (id, name, description, service_type, location,
			start_time, end_time, max_participants, status, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		evt.ID, evt.Name, evt.Description, evt.ServiceType, evt.Location,
		evt.StartTime, evt.EndTime, evt.MaxParticipants, evt.Status, evt.CreatedAt,
	)
	return err
}

func (s *SqliteStore) ListEvents(ctx context.Context, serviceType model.ServiceType, page, pageSize int) ([]model.CommunityEvent, int, error) {
	offset := (page - 1) * pageSize
	var q string
	var args []any
	if serviceType != "" {
		q = `SELECT id, name, description, service_type, location, start_time, end_time,
				max_participants, status, created_at FROM b2b_events
				WHERE service_type = ? ORDER BY start_time DESC LIMIT ? OFFSET ?`
		args = append(args, serviceType, pageSize, offset)
	} else {
		q = `SELECT id, name, description, service_type, location, start_time, end_time,
				max_participants, status, created_at FROM b2b_events
				ORDER BY start_time DESC LIMIT ? OFFSET ?`
		args = append(args, pageSize, offset)
	}

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.CommunityEvent
	for rows.Next() {
		var e model.CommunityEvent
		if err := rows.Scan(&e.ID, &e.Name, &e.Description, &e.ServiceType, &e.Location,
			&e.StartTime, &e.EndTime, &e.MaxParticipants, &e.Status, &e.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, e)
	}

	var total int
	countQ := "SELECT COUNT(*) FROM b2b_events"
	if serviceType != "" {
		countQ += " WHERE service_type = ?"
		s.db.QueryRowContext(ctx, countQ, serviceType).Scan(&total)
	} else {
		s.db.QueryRowContext(ctx, countQ).Scan(&total)
	}
	return list, total, nil
}

// ---------- Event Registration ----------

func (s *SqliteStore) RegisterForEvent(ctx context.Context, reg *model.EventRegistration) error {
	reg.ID = generateUUID()
	reg.RegisteredAt = time.Now()
	if reg.Status == "" {
		reg.Status = "confirmed"
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO b2b_event_registrations (id, event_id, elderly_id, caregiver_id, status, registered_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		reg.ID, reg.EventID, reg.ElderlyID, reg.CaregiverID, reg.Status, reg.RegisteredAt,
	)
	return err
}

func (s *SqliteStore) GetRegistrationsForEvent(ctx context.Context, eventID string) ([]model.EventRegistration, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, event_id, elderly_id, caregiver_id, status, registered_at
		 FROM b2b_event_registrations WHERE event_id = ? ORDER BY registered_at DESC`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var regs []model.EventRegistration
	for rows.Next() {
		var r model.EventRegistration
		if err := rows.Scan(&r.ID, &r.EventID, &r.ElderlyID, &r.CaregiverID, &r.Status, &r.RegisteredAt); err != nil {
			return nil, err
		}
		regs = append(regs, r)
	}
	return regs, rows.Err()
}

// ---------- Health Check Record ----------

func (s *SqliteStore) CreateHealthCheck(ctx context.Context, record *model.HealthCheckRecord) error {
	record.ID = generateUUID()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO b2b_health_checks (id, elderly_id, check_date, bp_systolic, bp_diastolic,
			hr, spo2, weight, height, glucose, notes, checked_by)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.ID, record.ElderlyID, record.CheckDate,
		record.BP_Systolic, record.BP_Diastolic, record.HR, record.SPO2,
		record.Weight, record.Height, record.Glucose, record.Notes, record.CheckedBy,
	)
	return err
}

func (s *SqliteStore) GetHealthChecksForElderly(ctx context.Context, elderlyID string, limit int) ([]model.HealthCheckRecord, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, elderly_id, check_date, bp_systolic, bp_diastolic, hr, spo2,
			weight, height, glucose, notes, checked_by
		 FROM b2b_health_checks WHERE elderly_id = ? ORDER BY check_date DESC LIMIT ?`,
		elderlyID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.HealthCheckRecord
	for rows.Next() {
		var r model.HealthCheckRecord
		if err := rows.Scan(&r.ID, &r.ElderlyID, &r.CheckDate,
			&r.BP_Systolic, &r.BP_Diastolic, &r.HR, &r.SPO2,
			&r.Weight, &r.Height, &r.Glucose, &r.Notes, &r.CheckedBy); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// ---------- Care Plan ----------

func (s *SqliteStore) CreateCarePlan(ctx context.Context, plan *model.CarePlan) error {
	plan.ID = generateUUID()
	plan.CreatedAt = time.Now()
	if plan.Status == "" {
		plan.Status = "active"
	}

	tasksData, _ := json.Marshal(plan.Tasks)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO b2b_care_plans (id, elderly_id, title, description, tasks, assigned_to,
			status, start_date, end_date, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		plan.ID, plan.ElderlyID, plan.Title, plan.Description, tasksData,
		plan.AssignedTo, plan.Status, plan.StartDate, plan.EndDate, plan.CreatedAt,
	)
	return err
}

func (s *SqliteStore) GetCarePlansForElderly(ctx context.Context, elderlyID string) ([]model.CarePlan, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, elderly_id, title, description, tasks, assigned_to, status, start_date, end_date, created_at
		 FROM b2b_care_plans WHERE elderly_id = ? AND status = 'active'`, elderlyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []model.CarePlan
	for rows.Next() {
		var p model.CarePlan
		var data []byte
		if err := rows.Scan(&p.ID, &p.ElderlyID, &p.Title, &p.Description, &data,
			&p.AssignedTo, &p.Status, &p.StartDate, &p.EndDate, &p.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(data, &p.Tasks)
		plans = append(plans, p)
	}
	return plans, rows.Err()
}

// ---------- Event CRUD ----------

func (s *SqliteStore) GetEventByID(ctx context.Context, id string) (*model.CommunityEvent, error) {
	e := &model.CommunityEvent{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, description, service_type, location, start_time, end_time,
			max_participants, status, created_at FROM b2b_events WHERE id = ?`, id).Scan(
		&e.ID, &e.Name, &e.Description, &e.ServiceType, &e.Location,
		&e.StartTime, &e.EndTime, &e.MaxParticipants, &e.Status, &e.CreatedAt,
	)
	return e, err
}

func (s *SqliteStore) DeleteEvent(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM b2b_events WHERE id = ?`, id)
	return err
}

func (s *SqliteStore) CancelEventRegistration(ctx context.Context, eventID, elderlyID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE b2b_event_registrations SET status = 'cancelled' WHERE event_id = ? AND elderly_id = ? AND status = 'confirmed'`,
		eventID, elderlyID,
	)
	return err
}

func (s *SqliteStore) ActiveRegistrationsCount(ctx context.Context, eventID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM b2b_event_registrations WHERE event_id = ? AND status = 'confirmed'`,
		eventID,
	).Scan(&count)
	return count, err
}

// ---------- Care Plan CRUD ----------

func (s *SqliteStore) GetCarePlanByID(ctx context.Context, id string) (*model.CarePlan, error) {
	p := &model.CarePlan{}
	var data []byte
	err := s.db.QueryRowContext(ctx,
		`SELECT id, elderly_id, title, description, tasks, assigned_to, status, start_date, end_date, created_at
		 FROM b2b_care_plans WHERE id = ?`, id).Scan(
		&p.ID, &p.ElderlyID, &p.Title, &p.Description, &data,
		&p.AssignedTo, &p.Status, &p.StartDate, &p.EndDate, &p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(data, &p.Tasks)
	return p, nil
}

func (s *SqliteStore) UpdateCarePlan(ctx context.Context, id string, plan *model.CarePlan) error {
	tasksData, _ := json.Marshal(plan.Tasks)
	_, err := s.db.ExecContext(ctx,
		`UPDATE b2b_care_plans SET title=?, description=?, tasks=?, status=?, start_date=?, end_date=?, created_at=? WHERE id=?`,
		plan.Title, plan.Description, tasksData, plan.Status, plan.StartDate, plan.EndDate, plan.CreatedAt, id,
	)
	return err
}

func (s *SqliteStore) DeleteCarePlan(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM b2b_care_plans WHERE id = ?`, id)
	return err
}

// ---------- Health Check CRUD ----------

func (s *SqliteStore) GetHealthCheckByID(ctx context.Context, id string) (*model.HealthCheckRecord, error) {
	r := &model.HealthCheckRecord{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, elderly_id, check_date, bp_systolic, bp_diastolic, hr, spo2,
			weight, height, glucose, notes, checked_by FROM b2b_health_checks WHERE id = ?`, id).Scan(
		&r.ID, &r.ElderlyID, &r.CheckDate,
		&r.BP_Systolic, &r.BP_Diastolic, &r.HR, &r.SPO2,
		&r.Weight, &r.Height, &r.Glucose, &r.Notes, &r.CheckedBy,
	)
	return r, err
}

func (s *SqliteStore) UpdateHealthCheck(ctx context.Context, id string, record *model.HealthCheckRecord) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE b2b_health_checks SET check_date=?, bp_systolic=?, bp_diastolic=?, hr=?, spo2=?,
			weight=?, height=?, glucose=?, notes=?, checked_by=? WHERE id=?`,
		record.CheckDate, record.BP_Systolic, record.BP_Diastolic, record.HR, record.SPO2,
		record.Weight, record.Height, record.Glucose, record.Notes, record.CheckedBy, id,
	)
	return err
}

func (s *SqliteStore) DeleteHealthCheck(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM b2b_health_checks WHERE id = ?`, id)
	return err
}

// ---------- Migration ----------

func migrate(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS b2b_events (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			service_type TEXT,
			location TEXT,
			start_time TEXT NOT NULL,
			end_time TEXT,
			max_participants INTEGER,
			status TEXT DEFAULT 'scheduled',
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS b2b_event_registrations (
			id TEXT PRIMARY KEY,
			event_id TEXT NOT NULL,
			elderly_id TEXT NOT NULL,
			caregiver_id TEXT,
			status TEXT DEFAULT 'confirmed',
			registered_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS b2b_health_checks (
			id TEXT PRIMARY KEY,
			elderly_id TEXT NOT NULL,
			check_date TEXT NOT NULL,
			bp_systolic REAL,
			bp_diastolic REAL,
			hr REAL,
			spo2 INTEGER,
			weight REAL,
			height REAL,
			glucose REAL,
			notes TEXT,
			checked_by TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS b2b_care_plans (
			id TEXT PRIMARY KEY,
			elderly_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			tasks TEXT DEFAULT '{}',
			assigned_to TEXT,
			status TEXT DEFAULT 'active',
			start_date TEXT,
			end_date TEXT,
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
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uint32(time.Now().UnixNano()),
		uint16(time.Now().Nanosecond()>>8),
		uint16(time.Now().UnixNano()>>16)&0xFFFF,
		uint16(time.Now().UnixNano()>>32)&0xFFFF,
		time.Now().UnixNano())
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
