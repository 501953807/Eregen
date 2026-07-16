package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"eregen.dev/b2b-community-platform/internal/model"

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

// ---------- Community Event ----------

func (s *Postgres) CreateEvent(ctx context.Context, evt *model.CommunityEvent) error {
	evt.ID = uuid.New().String()
	evt.CreatedAt = time.Now()
	if evt.Status == "" {
		evt.Status = "scheduled"
	}
	q := `INSERT INTO b2b_events (id, name, description, service_type, location,
		   start_time, end_time, max_participants, status, created_at)
		   VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := s.pool.Exec(ctx, q,
		evt.ID, evt.Name, evt.Description, evt.ServiceType, evt.Location,
		evt.StartTime, evt.EndTime, evt.MaxParticipants, evt.Status, evt.CreatedAt,
	)
	return err
}

func (s *Postgres) ListEvents(ctx context.Context, serviceType model.ServiceType, page, pageSize int) ([]model.CommunityEvent, int, error) {
	offset := (page - 1) * pageSize
	var q string
	var args []any
	if serviceType != "" {
		q = fmt.Sprintf(`SELECT id, name, description, service_type, location, start_time, end_time,
						   max_participants, status, created_at FROM b2b_events
						   WHERE service_type = $1 ORDER BY start_time DESC LIMIT $2 OFFSET $3`)
		args = append(args, serviceType, pageSize, offset)
	} else {
		q = fmt.Sprintf(`SELECT id, name, description, service_type, location, start_time, end_time,
						   max_participants, status, created_at FROM b2b_events
						   ORDER BY start_time DESC LIMIT $1 OFFSET $2`)
		args = append(args, pageSize, offset)
	}

	rows, err := s.pool.Query(ctx, q, args...)
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
		countQ += " WHERE service_type = $1"
		s.pool.QueryRow(ctx, countQ, serviceType).Scan(&total)
	} else {
		s.pool.QueryRow(ctx, countQ).Scan(&total)
	}
	return list, total, nil
}

// ---------- Event Registration ----------

func (s *Postgres) RegisterForEvent(ctx context.Context, reg *model.EventRegistration) error {
	reg.ID = uuid.New().String()
	reg.RegisteredAt = time.Now()
	if reg.Status == "" {
		reg.Status = "confirmed"
	}
	q := `INSERT INTO b2b_event_registrations (id, event_id, elderly_id, caregiver_id, status, registered_at)
		   VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := s.pool.Exec(ctx, q,
		reg.ID, reg.EventID, reg.ElderlyID, reg.CaregiverID, reg.Status, reg.RegisteredAt,
	)
	return err
}

func (s *Postgres) GetRegistrationsForEvent(ctx context.Context, eventID string) ([]model.EventRegistration, error) {
	q := `SELECT id, event_id, elderly_id, caregiver_id, status, registered_at
		   FROM b2b_event_registrations WHERE event_id = $1 ORDER BY registered_at DESC`
	rows, err := s.pool.Query(ctx, q, eventID)
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
	return regs, nil
}

// ---------- Health Check Record ----------

func (s *Postgres) CreateHealthCheck(ctx context.Context, record *model.HealthCheckRecord) error {
	record.ID = uuid.New().String()
	q := `INSERT INTO b2b_health_checks (id, elderly_id, check_date, bp_systolic, bp_diastolic,
		   hr, spo2, weight, height, glucose, notes, checked_by)
		   VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`
	_, err := s.pool.Exec(ctx, q,
		record.ID, record.ElderlyID, record.CheckDate,
		record.BP_Systolic, record.BP_Diastolic, record.HR, record.SPO2,
		record.Weight, record.Height, record.Glucose, record.Notes, record.CheckedBy,
	)
	return err
}

func (s *Postgres) GetHealthChecksForElderly(ctx context.Context, elderlyID string, limit int) ([]model.HealthCheckRecord, error) {
	q := `SELECT id, elderly_id, check_date, bp_systolic, bp_diastolic, hr, spo2,
		   weight, height, glucose, notes, checked_by
		   FROM b2b_health_checks WHERE elderly_id = $1 ORDER BY check_date DESC LIMIT $2`
	rows, err := s.pool.Query(ctx, q, elderlyID, limit)
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
	return records, nil
}

// ---------- Care Plan ----------

func (s *Postgres) CreateCarePlan(ctx context.Context, plan *model.CarePlan) error {
	plan.ID = uuid.New().String()
	plan.CreatedAt = time.Now()
	if plan.Status == "" {
		plan.Status = "active"
	}

	tasksData, _ := json.Marshal(plan.Tasks)
	q := `INSERT INTO b2b_care_plans (id, elderly_id, title, description, tasks, assigned_to,
		   status, start_date, end_date, created_at)
		   VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := s.pool.Exec(ctx, q,
		plan.ID, plan.ElderlyID, plan.Title, plan.Description, tasksData,
		plan.AssignedTo, plan.Status, plan.StartDate, plan.EndDate, plan.CreatedAt,
	)
	return err
}

func (s *Postgres) GetCarePlansForElderly(ctx context.Context, elderlyID string) ([]model.CarePlan, error) {
	q := `SELECT id, elderly_id, title, description, tasks, assigned_to, status, start_date, end_date, created_at
		   FROM b2b_care_plans WHERE elderly_id = $1 AND status = 'active'`
	rows, err := s.pool.Query(ctx, q, elderlyID)
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
	return plans, nil
}
