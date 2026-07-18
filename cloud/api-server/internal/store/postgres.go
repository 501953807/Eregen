package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"eregen.dev/api-server/internal/model"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Postgres wraps all database operations using pgxpool.
type Postgres struct {
	pool *pgxpool.Pool
	log  *zap.Logger
}

// NewPostgres creates a new repository backed by pgxpool.
func NewPostgres(pool *pgxpool.Pool, log *zap.Logger) *Postgres {
	return &Postgres{pool: pool, log: log}
}

// Pool returns the underlying pgxpool for raw queries.
func (p *Postgres) Pool() *pgxpool.Pool {
	return p.pool
}

// ---------- User ----------

func (p *Postgres) CreateUser(ctx context.Context, u *model.User) error {
	u.ID = uuid.New().String()
	u.CreatedAt = time.Now()
	u.UpdatedAt = u.CreatedAt

	q := `INSERT INTO users (id, email, phone, password_hash, role, name, created_at, updated_at)
		  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := p.pool.Exec(ctx, q,
		u.ID, u.Email, u.Phone, u.PasswordHash, u.Role, u.Name, u.CreatedAt, u.UpdatedAt,
	)
	return err
}

func (p *Postgres) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	u := &model.User{}
	q := `SELECT id, email, phone, password_hash, role, name, created_at, updated_at FROM users WHERE id = $1`
	err := p.pool.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.Email, &u.Phone, &u.PasswordHash, &u.Role, &u.Name, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (p *Postgres) GetUserByPhone(ctx context.Context, phone string) (*model.User, error) {
	u := &model.User{}
	q := `SELECT id, email, phone, password_hash, role, name, created_at, updated_at FROM users WHERE phone = $1`
	err := p.pool.QueryRow(ctx, q, phone).Scan(
		&u.ID, &u.Email, &u.Phone, &u.PasswordHash, &u.Role, &u.Name, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (p *Postgres) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	u := &model.User{}
	q := `SELECT id, email, phone, password_hash, role, name, created_at, updated_at FROM users WHERE email = $1`
	err := p.pool.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.Email, &u.Phone, &u.PasswordHash, &u.Role, &u.Name, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (p *Postgres) UpdateUser(ctx context.Context, id string, name, phone, email *string) error {
	parts := []string{"updated_at = now()"}
	args := []any{id}
	idx := 2

	if name != nil {
		parts = append(parts, fmt.Sprintf("name = $%d", idx))
		args = append(args, *name)
		idx++
	}
	if phone != nil {
		parts = append(parts, fmt.Sprintf("phone = $%d", idx))
		args = append(args, *phone)
		idx++
	}
	if email != nil {
		parts = append(parts, fmt.Sprintf("email = $%d", idx))
		args = append(args, *email)
		idx++
	}

	q := "UPDATE users SET " + parts[0]
	for i := 1; i < len(parts); i++ {
		q += ", " + parts[i]
	}
	q += " WHERE id = $1"
	_, err := p.pool.Exec(ctx, q, args...)
	return err
}

// ---------- ElderlyProfile ----------

func (p *Postgres) CreateElderlyProfile(ctx context.Context, ep *model.ElderlyProfile) error {
	ep.ID = uuid.New().String()
	ep.CreatedAt = time.Now()
	ep.UpdatedAt = ep.CreatedAt

	data, _ := json.Marshal(ep.HealthTiers)
	q := `INSERT INTO elderly_profiles (id, user_id, name, birth_date, avatar_url, health_tiers, created_at, updated_at)
		  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := p.pool.Exec(ctx, q,
		ep.ID, ep.UserID, ep.Name, ep.BirthDate, ep.AvatarURL, pq.Array(data),
		ep.CreatedAt, ep.UpdatedAt,
	)
	return err
}

func (p *Postgres) GetElderlyProfile(ctx context.Context, elderlyID string) (*model.ElderlyProfile, error) {
	ep := &model.ElderlyProfile{}
	q := `SELECT id, user_id, name, birth_date, avatar_url, health_tiers, created_at, updated_at
		  FROM elderly_profiles WHERE id = $1`
	var data pq.ByteaArray
	err := p.pool.QueryRow(ctx, q, elderlyID).Scan(
		&ep.ID, &ep.UserID, &ep.Name, &ep.BirthDate, &ep.AvatarURL, &data,
		&ep.CreatedAt, &ep.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		json.Unmarshal(data[0], &ep.HealthTiers)
	}
	return ep, nil
}

func (p *Postgres) UpdateElderlyProfile(ctx context.Context, elderlyID string, req *model.UpdateElderlyRequest) error {
	parts := []string{"updated_at = now()"}
	var args []any
	idx := 1

	if req.Name != "" {
		parts = append(parts, fmt.Sprintf("name = $%d", idx))
		args = append(args, req.Name)
		idx++
	}
	if req.BirthDate != nil {
		t, _ := time.Parse("2006-01-02", *req.BirthDate)
		parts = append(parts, fmt.Sprintf("birth_date = $%d", idx))
		args = append(args, t)
		idx++
	}
	if req.AvatarURL != nil {
		parts = append(parts, fmt.Sprintf("avatar_url = $%d", idx))
		args = append(args, *req.AvatarURL)
		idx++
	}
	if len(req.HealthTiers) > 0 {
		data, _ := json.Marshal(req.HealthTiers)
		parts = append(parts, fmt.Sprintf("health_tiers = $%d", idx))
		args = append(args, pq.Array(data))
		idx++
	}
	args = append(args, elderlyID)
	parts = append(parts, "id = $"+fmt.Sprintf("%d", idx))

	q := "UPDATE elderly_profiles SET " + parts[0]
	for i := 1; i < len(parts); i++ {
		q += ", " + parts[i]
	}
	_, err := p.pool.Exec(ctx, q, args...)
	return err
}

// ListElderlyProfiles returns a paginated list of elderly profiles for a given user.
func (p *Postgres) ListElderlyProfiles(ctx context.Context, userID string, page, pageSize int) ([]model.ElderlyProfile, int, error) {
	offset := (page - 1) * pageSize

	countQ := "SELECT COUNT(*) FROM elderly_profiles WHERE user_id = $1"
	var total int
	if err := p.pool.QueryRow(ctx, countQ, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := `SELECT id, user_id, name, birth_date, avatar_url, health_tiers, created_at, updated_at
		  FROM elderly_profiles WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := p.pool.Query(ctx, q, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var profiles []model.ElderlyProfile
	for rows.Next() {
		var ep model.ElderlyProfile
		var data pq.ByteaArray
		if err := rows.Scan(&ep.ID, &ep.UserID, &ep.Name, &ep.BirthDate, &ep.AvatarURL, &data, &ep.CreatedAt, &ep.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if len(data) > 0 {
			json.Unmarshal(data[0], &ep.HealthTiers)
		}
		profiles = append(profiles, ep)
	}
	return profiles, total, rows.Err()
}

// ---------- Device ----------

func (p *Postgres) CreateDevice(ctx context.Context, d *model.Device) error {
	d.ID = uuid.New().String()
	d.CreatedAt = time.Now()
	d.UpdatedAt = d.CreatedAt
	if d.Status == "" {
		d.Status = model.DeviceOffline
	}

	settingsJSON, _ := json.Marshal(d.Settings)
	q := `INSERT INTO devices (id, device_id, device_type, tier, owner_user_id, status, last_seen, created_at, updated_at, settings)
		  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := p.pool.Exec(ctx, q,
		d.ID, d.DeviceID, d.DeviceType, d.Tier, d.OwnerUserID,
		d.Status, d.LastSeen, d.CreatedAt, d.UpdatedAt, settingsJSON,
	)
	return err
}

func (p *Postgres) ListDevices(ctx context.Context, ownerID string, deviceType *string, page, pageSize int) ([]model.Device, int, error) {
	where := "owner_user_id = $1"
	var args []any
	args = append(args, ownerID)
	idx := 2

	if deviceType != nil && *deviceType != "" {
		where += fmt.Sprintf(" AND device_type = $%d", idx)
		args = append(args, *deviceType)
		idx++
	}

	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)
	q := fmt.Sprintf("SELECT id, device_id, device_type, tier, owner_user_id, status, last_seen, created_at, updated_at, settings FROM devices WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d", where, idx, idx+1)

	rows, err := p.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var devices []model.Device
	for rows.Next() {
		d, scanErr := scanDevice(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		devices = append(devices, *d)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	countArgs := make([]any, len(args)-2)
	copy(countArgs, args)
	countQ := fmt.Sprintf("SELECT COUNT(*) FROM devices WHERE %s", where)
	var count int
	err = p.pool.QueryRow(ctx, countQ, countArgs...).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	return devices, count, nil
}

func (p *Postgres) GetDevice(ctx context.Context, deviceID string) (*model.Device, error) {
	q := `SELECT id, device_id, device_type, tier, owner_user_id, status, last_seen, created_at, updated_at, settings
		  FROM devices WHERE device_id = $1`
	row := p.pool.QueryRow(ctx, q, deviceID)
	return scanDevice(row)
}

func (p *Postgres) UpdateDeviceSettings(ctx context.Context, deviceID string, settings map[string]any) error {
	data, _ := json.Marshal(settings)
	q := `UPDATE devices SET settings = $1, updated_at = now(), status = 'online', last_seen = now()
		  WHERE device_id = $2`
	_, err := p.pool.Exec(ctx, q, data, deviceID)
	return err
}

func (p *Postgres) DeleteDevice(ctx context.Context, deviceID, ownerID string) error {
	q := `DELETE FROM devices WHERE device_id = $1 AND owner_user_id = $2`
	_, err := p.pool.Exec(ctx, q, deviceID, ownerID)
	return err
}

// BindDevice registers a new device under the given user and returns it.
func (p *Postgres) BindDevice(ctx context.Context, deviceID, ownerUserID, deviceType, tier string) (*model.Device, error) {
	d := &model.Device{
		DeviceID:    deviceID,
		DeviceType:  deviceType,
		Tier:        tier,
		OwnerUserID: ownerUserID,
		Status:      model.DeviceOffline,
		Settings:    map[string]any{},
	}

	settingsJSON, _ := json.Marshal(d.Settings)
	q := `INSERT INTO devices (id, device_id, device_type, tier, owner_user_id, status, last_seen, created_at, updated_at, settings)
		  VALUES ($1, $2, $3, $4, $5, $6, NULL, $7, $8, $9)
		  ON CONFLICT (device_id) DO NOTHING`
	_, err := p.pool.Exec(ctx, q,
		uuid.New().String(), d.DeviceID, d.DeviceType, d.Tier, d.OwnerUserID,
		d.Status, d.CreatedAt, d.UpdatedAt, settingsJSON,
	)
	if err != nil {
		return nil, err
	}

	// Return the bound device (may already exist from a prior bind)
	existing, getErr := p.GetDevice(ctx, deviceID)
	if getErr != nil {
		return d, nil
	}
	return existing, nil
}

// ---------- HealthRecord ----------

func (p *Postgres) CreateHealthRecord(ctx context.Context, r *model.HealthRecord) error {
	r.ID = uuid.New().String()
	q := `INSERT INTO health_records (id, elderly_id, timestamp, hr, spo2, steps, sleep_hours, bp_systolic, bp_diastolic)
		  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := p.pool.Exec(ctx, q,
		r.ID, r.ElderlyID, r.Timestamp, r.HR, r.SPO2, r.Steps,
		r.SleepHours, r.BPSystolic, r.BPDiastolic,
	)
	return err
}

func (p *Postgres) GetHealthSummary(ctx context.Context, elderlyID string, day time.Time) (*model.HealthRecord, error) {
	r := &model.HealthRecord{}
	start := day.Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour)
	q := `SELECT id, elderly_id, MAX(timestamp) as timestamp,
				AVG(hr)::int as hr, MIN(spo2)::int as spo2, COALESCE(SUM(steps),0)::bigint as steps,
				MAX(sleep_hours) as sleep_hours, MAX(bp_systolic) as bp_systolic, MAX(bp_diastolic) as bp_diastolic
			FROM health_records WHERE elderly_id = $1 AND timestamp >= $2 AND timestamp < $3
			GROUP BY elderly_id`
	err := p.pool.QueryRow(ctx, q, elderlyID, start, end).Scan(
		&r.ID, &r.ElderlyID, &r.Timestamp, &r.HR, &r.SPO2, &r.Steps,
		&r.SleepHours, &r.BPSystolic, &r.BPDiastolic,
	)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (p *Postgres) GetHealthHistory(ctx context.Context, elderlyID string, days int) ([]model.HealthRecord, error) {
	until := time.Now()
	from := until.Add(-time.Duration(days) * 24 * time.Hour)

	q := `SELECT id, elderly_id, date_trunc('day', timestamp) as timestamp,
				AVG(hr)::int as hr, MIN(spo2)::int as spo2, COALESCE(SUM(steps),0)::bigint as steps,
				MAX(sleep_hours) as sleep_hours, MAX(bp_systolic) as bp_systolic, MAX(bp_diastolic) as bp_diastolic
			FROM health_records WHERE elderly_id = $1 AND timestamp >= $2 AND timestamp <= $3
			GROUP BY elderly_id, date_trunc('day', timestamp)
			ORDER BY timestamp DESC`

	rows, err := p.pool.Query(ctx, q, elderlyID, from, until)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.HealthRecord
	for rows.Next() {
		var r model.HealthRecord
		if err := rows.Scan(&r.ID, &r.ElderlyID, &r.Timestamp, &r.HR, &r.SPO2, &r.Steps,
			&r.SleepHours, &r.BPSystolic, &r.BPDiastolic); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

func (p *Postgres) GetHealthTrend(ctx context.Context, elderlyID, metric string, days int) ([]model.HealthRecord, error) {
	_ = metric
	until := time.Now()
	from := until.Add(-time.Duration(days) * 24 * time.Hour)

	q := `SELECT id, elderly_id, timestamp, hr, spo2, steps, sleep_hours, bp_systolic, bp_diastolic
			FROM health_records WHERE elderly_id = $1 AND timestamp >= $2 AND timestamp <= $3
			ORDER BY timestamp ASC`

	rows, err := p.pool.Query(ctx, q, elderlyID, from, until)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.HealthRecord
	for rows.Next() {
		var r model.HealthRecord
		if err := rows.Scan(&r.ID, &r.ElderlyID, &r.Timestamp, &r.HR, &r.SPO2, &r.Steps,
			&r.SleepHours, &r.BPSystolic, &r.BPDiastolic); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// ---------- LocationRecord ----------

func (p *Postgres) CreateLocationRecord(ctx context.Context, r *model.LocationRecord) error {
	r.ID = uuid.New().String()
	q := `INSERT INTO location_records (id, elderly_id, timestamp, lat, lon, accuracy)
		  VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := p.pool.Exec(ctx, q,
		r.ID, r.ElderlyID, r.Timestamp, r.Lat, r.Lon, r.Accuracy,
	)
	return err
}

func (p *Postgres) GetLatestLocation(ctx context.Context, elderlyID string) (*model.LocationRecord, error) {
	r := &model.LocationRecord{}
	q := `SELECT id, elderly_id, timestamp, lat, lon, accuracy
		  FROM location_records WHERE elderly_id = $1
		  ORDER BY timestamp DESC LIMIT 1`
	err := p.pool.QueryRow(ctx, q, elderlyID).Scan(
		&r.ID, &r.ElderlyID, &r.Timestamp, &r.Lat, &r.Lon, &r.Accuracy,
	)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (p *Postgres) GetLocationHistory(ctx context.Context, elderlyID string, from, until time.Time) ([]model.LocationRecord, error) {
	q := `SELECT id, elderly_id, timestamp, lat, lon, accuracy
		  FROM location_records WHERE elderly_id = $1 AND timestamp >= $2 AND timestamp <= $3
		  ORDER BY timestamp DESC`

	rows, err := p.pool.Query(ctx, q, elderlyID, from, until)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.LocationRecord
	for rows.Next() {
		var r model.LocationRecord
		if err := rows.Scan(&r.ID, &r.ElderlyID, &r.Timestamp, &r.Lat, &r.Lon, &r.Accuracy); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// ---------- MedicationRule ----------

func (p *Postgres) CreateMedicationRule(ctx context.Context, mr *model.MedicationRule) error {
	mr.ID = uuid.New().String()
	mr.CreatedAt = time.Now()
	mr.UpdatedAt = mr.CreatedAt
	if !mr.Active {
		mr.Active = true
	}
	if len(mr.DaysOfWeek) == 0 {
		mr.DaysOfWeek = []int{1, 2, 3, 4, 5, 6, 7}
	}

	data, _ := json.Marshal(mr.DaysOfWeek)
	q := `INSERT INTO medication_rules (id, elderly_id, schedule_time, dose_count, pill_type, days_of_week, active, created_at, updated_at)
		  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := p.pool.Exec(ctx, q,
		mr.ID, mr.ElderlyID, mr.ScheduleTime, mr.DoseCount, mr.PillType,
		pq.Array(data), mr.Active, mr.CreatedAt, mr.UpdatedAt,
	)
	return err
}

func (p *Postgres) ListMedicationRules(ctx context.Context, elderlyID string) ([]model.MedicationRule, error) {
	q := `SELECT id, elderly_id, schedule_time, dose_count, pill_type, days_of_week, active, created_at, updated_at
		  FROM medication_rules WHERE elderly_id = $1 AND active = true
		  ORDER BY schedule_time ASC`

	rows, err := p.pool.Query(ctx, q, elderlyID)
	if err != nil {
		return nil, err
	}
	return scanRules(rows)
}

func (p *Postgres) GetMedicationRule(ctx context.Context, ruleID string) (*model.MedicationRule, error) {
	mr := &model.MedicationRule{}
	q := `SELECT id, elderly_id, schedule_time, dose_count, pill_type, days_of_week, active, created_at, updated_at
		  FROM medication_rules WHERE id = $1`
	var data pq.ByteaArray
	err := p.pool.QueryRow(ctx, q, ruleID).Scan(
		&mr.ID, &mr.ElderlyID, &mr.ScheduleTime, &mr.DoseCount, &mr.PillType,
		&data, &mr.Active, &mr.CreatedAt, &mr.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		json.Unmarshal(data[0], &mr.DaysOfWeek)
	}
	return mr, nil
}

func (p *Postgres) UpdateMedicationRule(ctx context.Context, ruleID string, req *model.CreateMedicationRuleRequest) error {
	data, _ := json.Marshal(req.DaysOfWeek)
	q := `UPDATE medication_rules SET schedule_time = $1, dose_count = $2, pill_type = $3,
		  days_of_week = $4, active = $5, updated_at = now() WHERE id = $6`
	_, err := p.pool.Exec(ctx, q,
		req.ScheduleTime, req.DoseCount, req.PillType, pq.Array(data), req.Active, ruleID,
	)
	return err
}

func (p *Postgres) DeleteMedicationRule(ctx context.Context, ruleID string) error {
	q := `DELETE FROM medication_rules WHERE id = $1`
	_, err := p.pool.Exec(ctx, q, ruleID)
	return err
}

// ---------- MedStatusRecord ----------

func (p *Postgres) CreateMedStatusRecord(ctx context.Context, r *model.MedStatusRecord) error {
	r.ID = uuid.New().String()
	q := `INSERT INTO med_status_records (id, rule_id, elderly_id, taken_at, taken, missed_at)
		  VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := p.pool.Exec(ctx, q,
		r.ID, r.RuleID, r.ElderlyID, r.TakenAt, r.Taken, r.MissedAt,
	)
	return err
}

func (p *Postgres) GetTodayMedStatus(ctx context.Context, elderlyID string) ([]model.MedStatusRecord, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.Add(24 * time.Hour)

	q := `SELECT id, rule_id, elderly_id, taken_at, taken, missed_at
			FROM med_status_records WHERE elderly_id = $1 AND taken_at >= $2 AND taken_at < $3
			ORDER BY taken_at ASC`

	rows, err := p.pool.Query(ctx, q, elderlyID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.MedStatusRecord
	for rows.Next() {
		var r model.MedStatusRecord
		if err := rows.Scan(&r.ID, &r.RuleID, &r.ElderlyID, &r.TakenAt, &r.Taken, &r.MissedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

func (p *Postgres) GetMedicationHistory(ctx context.Context, elderlyID string, days int) ([]model.MedStatusRecord, error) {
	until := time.Now()
	from := until.Add(-time.Duration(days) * 24 * time.Hour)

	q := `SELECT id, rule_id, elderly_id, taken_at, taken, missed_at
		  FROM med_status_records WHERE elderly_id = $1 AND taken_at >= $2 AND taken_at <= $3
		  ORDER BY taken_at DESC`

	rows, err := p.pool.Query(ctx, q, elderlyID, from, until)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.MedStatusRecord
	for rows.Next() {
		var r model.MedStatusRecord
		if err := rows.Scan(&r.ID, &r.RuleID, &r.ElderlyID, &r.TakenAt, &r.Taken, &r.MissedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// ---------- Alert ----------

func (p *Postgres) CreateAlert(ctx context.Context, a *model.Alert) error {
	a.ID = uuid.New().String()
	a.CreatedAt = time.Now()
	if a.Status == "" {
		a.Status = model.AlertPending
	}

	metaJSON, _ := json.Marshal(a.Metadata)
	q := `INSERT INTO alerts (id, elderly_id, alert_type, severity, status, metadata, created_at, resolved_at)
		  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := p.pool.Exec(ctx, q,
		a.ID, a.ElderlyID, a.AlertType, a.Severity, a.Status, metaJSON,
		a.CreatedAt, a.ResolvedAt,
	)
	return err
}

func (p *Postgres) ListAlerts(ctx context.Context, elderIDs []string, filter *model.AlertFilter, page, pageSize int) ([]model.Alert, int, error) {
	where := "elderly_id = ANY($1)"
	var args []any
	args = append(args, pq.Array(elderIDs))
	idx := 2

	if filter != nil {
		if filter.Severity != nil {
			where += fmt.Sprintf(" AND severity = $%d", idx)
			args = append(args, *filter.Severity)
			idx++
		}
		if filter.Status != nil {
			where += fmt.Sprintf(" AND status = $%d", idx)
			args = append(args, *filter.Status)
			idx++
		}
	}

	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)
	q := fmt.Sprintf("SELECT id, elderly_id, alert_type, severity, status, metadata, created_at, resolved_at FROM alerts WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d", where, idx, idx+1)

	rows, err := p.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var alerts []model.Alert
	for rows.Next() {
		var a model.Alert
		var metaJSON []byte
		if err := rows.Scan(&a.ID, &a.ElderlyID, &a.AlertType, &a.Severity, &a.Status, &metaJSON, &a.CreatedAt, &a.ResolvedAt); err != nil {
			return nil, 0, err
		}
		if len(metaJSON) > 0 {
			json.Unmarshal(metaJSON, &a.Metadata)
		}
		alerts = append(alerts, a)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	countArgs := make([]any, len(args)-2)
	copy(countArgs, args)
	countQ := fmt.Sprintf("SELECT COUNT(*) FROM alerts WHERE %s", where)
	var count int
	err = p.pool.QueryRow(ctx, countQ, countArgs...).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	return alerts, count, nil
}

func (p *Postgres) GetAlert(ctx context.Context, alertID string) (*model.Alert, error) {
	a := &model.Alert{}
	q := `SELECT id, elderly_id, alert_type, severity, status, metadata, created_at, resolved_at
		  FROM alerts WHERE id = $1`
	var metaJSON []byte
	err := p.pool.QueryRow(ctx, q, alertID).Scan(
		&a.ID, &a.ElderlyID, &a.AlertType, &a.Severity, &a.Status, &metaJSON,
		&a.CreatedAt, &a.ResolvedAt,
	)
	if err != nil {
		return nil, err
	}
	if len(metaJSON) > 0 {
		json.Unmarshal(metaJSON, &a.Metadata)
	}
	return a, nil
}

func (p *Postgres) UpdateAlert(ctx context.Context, alertID string, status model.AlertStatus) error {
	q := `UPDATE alerts SET status = $1, resolved_at = now() WHERE id = $2`
	_, err := p.pool.Exec(ctx, q, status, alertID)
	return err
}

// ---------- Geofence ----------

func (p *Postgres) CreateGeofence(ctx context.Context, gf *model.Geofence) error {
	gf.ID = uuid.New().String()
	gf.CreatedAt = time.Now()
	gf.UpdatedAt = gf.CreatedAt
	q := `INSERT INTO geofences (id, elderly_id, name, latitude, longitude, radius_meters, active, created_at, updated_at)
		  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := p.pool.Exec(ctx, q,
		gf.ID, gf.ElderlyID, gf.Name, gf.Latitude, gf.Longitude,
		gf.RadiusMeters, gf.Active, gf.CreatedAt, gf.UpdatedAt,
	)
	return err
}

func (p *Postgres) ListGeofences(ctx context.Context, elderlyID string) ([]model.Geofence, error) {
	q := `SELECT id, elderly_id, name, latitude, longitude, radius_meters, active, created_at, updated_at
		  FROM geofences WHERE elderly_id = $1 ORDER BY name ASC`
	rows, err := p.pool.Query(ctx, q, elderlyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fences []model.Geofence
	for rows.Next() {
		var g model.Geofence
		if err := rows.Scan(&g.ID, &g.ElderlyID, &g.Name, &g.Latitude, &g.Longitude,
			&g.RadiusMeters, &g.Active, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		fences = append(fences, g)
	}
	return fences, rows.Err()
}

func (p *Postgres) UpdateGeofence(ctx context.Context, id string, req *model.UpdateGeofenceRequest) error {
	q := `UPDATE geofences SET name=$1, latitude=$2, longitude=$3, radius_meters=$4, active=$5, updated_at=now()
		  WHERE id=$6`
	_, err := p.pool.Exec(ctx, q, req.Name, req.Lat, req.Lon, req.RadiusMeters, req.Active, id)
	return err
}

func (p *Postgres) DeleteGeofence(ctx context.Context, id string) error {
	_, err := p.pool.Exec(ctx, `DELETE FROM geofences WHERE id = $1`, id)
	return err
}

// GetDeviceByElderlyID finds the device linked to an elderly profile via owner user.
func (p *Postgres) GetDeviceByElderlyID(ctx context.Context, elderlyID string) (string, error) {
	var userID string
	err := p.pool.QueryRow(ctx, `SELECT user_id FROM elderly_profiles WHERE id = $1`, elderlyID).Scan(&userID)
	if err != nil {
		return "", err
	}
	var deviceID string
	err = p.pool.QueryRow(ctx, `SELECT device_id FROM devices WHERE owner_user_id = $1 AND device_type = 'pillbox' ORDER BY created_at DESC LIMIT 1`, userID).Scan(&deviceID)
	if err == sql.ErrNoRows {
		return "", nil // no pillbox linked
	}
	return deviceID, err
}

// ---------- Subscription ----------

func (p *Postgres) CreateSubscription(ctx context.Context, s *model.Subscription) error {
	s.ID = uuid.New().String()
	s.StartDate = time.Now()
	s.EndDate = s.StartDate.AddDate(0, 1, 0)

	q := `INSERT INTO subscriptions (id, user_id, plan_tier, status, start_date, end_date)
		  VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := p.pool.Exec(ctx, q,
		s.ID, s.UserID, s.PlanTier, s.Status, s.StartDate, s.EndDate,
	)
	return err
}

func (p *Postgres) GetSubscription(ctx context.Context, userID string) (*model.Subscription, error) {
	s := &model.Subscription{}
	q := `SELECT id, user_id, plan_tier, status, start_date, end_date
		  FROM subscriptions WHERE user_id = $1 AND status = 'active'
		  ORDER BY end_date DESC LIMIT 1`
	err := p.pool.QueryRow(ctx, q, userID).Scan(
		&s.ID, &s.UserID, &s.PlanTier, &s.Status, &s.StartDate, &s.EndDate,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// ---------- helpers ----------

func scanDevice(row any) (*model.Device, error) {
	type scanner interface {
		Scan(dest ...any) error
	}
	s, ok := row.(scanner)
	if !ok {
		return nil, fmt.Errorf("row does not implement Scan")
	}
	d := &model.Device{}
	var settingsJSON []byte
	err := s.Scan(&d.ID, &d.DeviceID, &d.DeviceType, &d.Tier, &d.OwnerUserID,
		&d.Status, &d.LastSeen, &d.CreatedAt, &d.UpdatedAt, &settingsJSON)
	if err != nil {
		return nil, err
	}
	if len(settingsJSON) > 0 {
		json.Unmarshal(settingsJSON, &d.Settings)
	}
	return d, nil
}

func scanRules(rows any) ([]model.MedicationRule, error) {
	type closer interface {
		Close() error
	}
	type iterator interface {
		Next() bool
		Scan(dest ...any) error
		Err() error
	}
	c, _ := rows.(closer)
	if c != nil {
		defer c.Close()
	}
	it, ok := rows.(iterator)
	if !ok {
		return nil, fmt.Errorf("rows does not implement iterator")
	}
	var rules []model.MedicationRule
	for it.Next() {
		var mr model.MedicationRule
		var data pq.ByteaArray
		if err := it.Scan(&mr.ID, &mr.ElderlyID, &mr.ScheduleTime, &mr.DoseCount,
			&mr.PillType, &data, &mr.Active, &mr.CreatedAt, &mr.UpdatedAt); err != nil {
			return nil, err
		}
		if len(data) > 0 {
			json.Unmarshal(data[0], &mr.DaysOfWeek)
		}
		rules = append(rules, mr)
	}
	return rules, it.Err()
}
