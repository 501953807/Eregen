package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// ElderlySummary is a lightweight row for the elderly list view.
type ElderlySummary struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	UserID      string     `json:"user_id"`
	AvatarURL   string     `json:"avatar_url,omitempty"`
	BirthDate   *time.Time `json:"birth_date,omitempty"`
	HealthTiers []string   `json:"health_tiers"`
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   string     `json:"updated_at"`
}

// HealthStats aggregates recent health metrics for an elderly person.
type HealthStats struct {
	ElderlyID string    `json:"elderly_id"`
	AvgHR     *float64  `json:"avg_hr,omitempty"`
	MaxHR     *int      `json:"max_hr,omitempty"`
	AvgSpO2   *float64  `json:"avg_spo2,omitempty"`
	TotalSteps *int64   `json:"total_steps,omitempty"`
	LastSeen  time.Time `json:"last_seen"`
}

// HealthRecordRow represents a single health record from the database.
type HealthRecordRow struct {
	ID        string    `json:"id"`
	ElderlyID string    `json:"elderly_id"`
	Timestamp time.Time `json:"timestamp"`
	HR        *int      `json:"hr,omitempty"`
	SpO2      *int      `json:"spo2,omitempty"`
	Steps     *int64    `json:"steps,omitempty"`
	SleepHours *float64  `json:"sleep_hours,omitempty"`
}

// MedicationRuleRow represents a medication rule.
type MedicationRuleRow struct {
	ID         string   `json:"id"`
	ElderlyID  string   `json:"elderly_id"`
	ScheduleTime string `json:"schedule_time"`
	DoseCount  int      `json:"dose_count"`
	PillType   string   `json:"pill_type"`
	DaysOfWeek []int    `json:"days_of_week"`
	Active     bool     `json:"active"`
	CreatedAt  string   `json:"created_at"`
}

// DeviceSummaryRow is a device linked to an elderly person.
type DeviceSummaryRow struct {
	ID          string    `json:"id"`
	DeviceID    string    `json:"device_id"`
	Type        string    `json:"type"`
	Tier        string    `json:"tier"`
	Status      string    `json:"status"`
	FirmwareVer string    `json:"firmware_version"`
	LastSeen    time.Time `json:"last_seen"`
}

// LocationPoint represents a location record.
type LocationPoint struct {
	ID        string    `json:"id"`
	ElderlyID string    `json:"elderly_id"`
	Lat       float64   `json:"lat"`
	Lon       float64   `json:"lon"`
	Accuracy  *float64  `json:"accuracy,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// AlertSummaryRow represents an alert.
type AlertSummaryRow struct {
	ID        string    `json:"id"`
	ElderlyID string    `json:"elderly_id"`
	AlertType string    `json:"alert_type"`
	Severity  string    `json:"severity"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// ListElderly returns paginated elderly profiles.
func (s *PostgresStore) ListElderly(ctx context.Context, page, pageSize int) ([]ElderlySummary, error) {
	query := `SELECT id, name, COALESCE(user_id, ''), COALESCE(avatar_url, ''),
		COALESCE(birth_date, '0001-01-01'), health_tiers, created_at, updated_at
		FROM elderly_profiles ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := s.db.QueryContext(ctx, query, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("list elderly: %w", err)
	}
	defer rows.Close()

	var profiles []ElderlySummary
	for rows.Next() {
		var p ElderlySummary
		var zeroTime time.Time
		var tiersRaw interface{}
		err := rows.Scan(&p.ID, &p.Name, &p.UserID, &p.AvatarURL, &p.BirthDate, &tiersRaw, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan elderly: %w", err)
		}
		if p.BirthDate != nil && *p.BirthDate == zeroTime {
			p.BirthDate = nil
		}
		// Parse health_tiers from JSON array
		switch v := tiersRaw.(type) {
		case []byte:
			json.Unmarshal(v, &p.HealthTiers)
		case string:
			json.Unmarshal([]byte(v), &p.HealthTiers)
		}
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}

// GetElderly returns a single elderly profile by ID.
func (s *PostgresStore) GetElderly(ctx context.Context, id string) (*ElderlySummary, error) {
	var p ElderlySummary
	var tiersRaw interface{}
	var zeroTime time.Time

	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, COALESCE(user_id, ''), COALESCE(avatar_url, ''),
			COALESCE(birth_date, '0001-01-01'), health_tiers, created_at, updated_at
		FROM elderly_profiles WHERE id = $1`, id).Scan(
		&p.ID, &p.Name, &p.UserID, &p.AvatarURL, &p.BirthDate, &tiersRaw, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("get elderly: %w", err)
	}
	if p.BirthDate != nil && *p.BirthDate == zeroTime {
		p.BirthDate = nil
	}
	switch v := tiersRaw.(type) {
	case []byte:
		json.Unmarshal(v, &p.HealthTiers)
	case string:
		json.Unmarshal([]byte(v), &p.HealthTiers)
	}
	return &p, nil
}

// CreateElderly inserts a new elderly profile.
func (s *PostgresStore) CreateElderly(ctx context.Context, name, birthDate, userID string, healthTiers []string, avatarURL string) (*ElderlySummary, error) {
	tiersJSON, _ := json.Marshal(healthTiers)
	var bd *time.Time
	if birthDate != "" {
		t, err := time.Parse("2006-01-02", birthDate)
		if err == nil {
			bd = &t
		}
	}

	now := time.Now().Format(time.RFC3339)
	var p ElderlySummary
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO elderly_profiles (name, birth_date, user_id, health_tiers, avatar_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		name, bd, userID, tiersJSON, avatarURL, now, now).Scan(&p.ID)
	if err != nil {
		return nil, fmt.Errorf("create elderly: %w", err)
	}
	p.Name = name
	p.UserID = userID
	p.AvatarURL = avatarURL
	p.HealthTiers = healthTiers
	p.CreatedAt = now
	p.UpdatedAt = now
	if bd != nil {
		p.BirthDate = bd
	}
	return &p, nil
}

// UpdateElderly updates an existing elderly profile.
func (s *PostgresStore) UpdateElderly(ctx context.Context, id, name, birthDate, userID string, healthTiers []string, avatarURL string) (*ElderlySummary, error) {
	tiersJSON, _ := json.Marshal(healthTiers)
	var bd *time.Time
	if birthDate != "" {
		t, err := time.Parse("2006-01-02", birthDate)
		if err == nil {
			bd = &t
		}
	}

	now := time.Now().Format(time.RFC3339)
	var p ElderlySummary
	err := s.db.QueryRowContext(ctx, `
		UPDATE elderly_profiles SET name=$1, birth_date=$2, user_id=$3, health_tiers=$4,
			avatar_url=$5, updated_at=$6 WHERE id=$7 RETURNING id`,
		name, bd, userID, tiersJSON, avatarURL, now, id).Scan(&p.ID)
	if err != nil {
		return nil, fmt.Errorf("update elderly: %w", err)
	}
	p.Name = name
	p.UserID = userID
	p.AvatarURL = avatarURL
	p.HealthTiers = healthTiers
	p.CreatedAt = now
	p.UpdatedAt = now
	if bd != nil {
		p.BirthDate = bd
	}
	return &p, nil
}

// DeleteElderly removes an elderly profile.
func (s *PostgresStore) DeleteElderly(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM elderly_profiles WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete elderly: %w", err)
	}
	return nil
}

// GetElderlyHealthStats returns aggregated health statistics.
func (s *PostgresStore) GetElderlyHealthStats(ctx context.Context, elderlyID string) (*HealthStats, error) {
	var stats HealthStats
	stats.ElderlyID = elderlyID

	err := s.db.QueryRowContext(ctx, `
		SELECT AVG(hr)::float8, MAX(hr), AVG(spo2)::float8, SUM(steps)::bigint, MAX(timestamp)
		FROM health_records WHERE elderly_id = $1`, elderlyID).Scan(
		&stats.AvgHR, &stats.MaxHR, &stats.AvgSpO2, &stats.TotalSteps, &stats.LastSeen)
	if err != nil {
		return nil, fmt.Errorf("health stats: %w", err)
	}
	return &stats, nil
}

// GetElderlyHealthRecords returns recent health records.
func (s *PostgresStore) GetElderlyHealthRecords(ctx context.Context, elderlyID string, limit int) ([]HealthRecordRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, elderly_id, timestamp, hr, spo2, steps, sleep_hours
		FROM health_records WHERE elderly_id = $1 ORDER BY timestamp DESC LIMIT $2`,
		elderlyID, limit)
	if err != nil {
		return nil, fmt.Errorf("health records: %w", err)
	}
	defer rows.Close()

	var records []HealthRecordRow
	for rows.Next() {
		var r HealthRecordRow
		if err := rows.Scan(&r.ID, &r.ElderlyID, &r.Timestamp, &r.HR, &r.SpO2, &r.Steps, &r.SleepHours); err != nil {
			return nil, fmt.Errorf("scan health record: %w", err)
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// GetElderlyMedicationRules returns medication rules.
func (s *PostgresStore) GetElderlyMedicationRules(ctx context.Context, elderlyID string) ([]MedicationRuleRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, elderly_id, schedule_time, dose_count, pill_type, days_of_week, active, created_at
		FROM medication_rules WHERE elderly_id = $1 ORDER BY schedule_time`, elderlyID)
	if err != nil {
		return nil, fmt.Errorf("med rules: %w", err)
	}
	defer rows.Close()

	var rules []MedicationRuleRow
	for rows.Next() {
		var r MedicationRuleRow
		var daysRaw interface{}
		if err := rows.Scan(&r.ID, &r.ElderlyID, &r.ScheduleTime, &r.DoseCount, &r.PillType, &daysRaw, &r.Active, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan med rule: %w", err)
		}
		switch v := daysRaw.(type) {
		case []byte:
			json.Unmarshal(v, &r.DaysOfWeek)
		case string:
			json.Unmarshal([]byte(v), &r.DaysOfWeek)
		}
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

// GetElderlyDevices returns devices linked to an elderly person.
func (s *PostgresStore) GetElderlyDevices(ctx context.Context, elderlyID string) ([]DeviceSummaryRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT d.id, d.device_id, d.device_type, d.tier, d.status,
			COALESCE(d.settings->>'fw_version', 'v0.1'),
			COALESCE(d.last_seen, '0001-01-01')
		FROM devices d JOIN elderly_devices ed ON d.id = ed.device_id
		WHERE ed.elderly_id = $1 ORDER BY d.last_seen DESC NULLS LAST`, elderlyID)
	if err != nil {
		return nil, fmt.Errorf("elderly devices: %w", err)
	}
	defer rows.Close()

	var devices []DeviceSummaryRow
	for rows.Next() {
		var d DeviceSummaryRow
		var zeroTime time.Time
		if err := rows.Scan(&d.ID, &d.DeviceID, &d.Type, &d.Tier, &d.Status, &d.FirmwareVer, &d.LastSeen); err != nil {
			return nil, fmt.Errorf("scan device: %w", err)
		}
		if d.LastSeen == zeroTime {
			d.LastSeen = time.Time{}
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

// GetElderlyLocationHistory returns location history.
func (s *PostgresStore) GetElderlyLocationHistory(ctx context.Context, elderlyID string, limit int) ([]LocationPoint, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, elderly_id, lat, lon, accuracy, timestamp
		FROM location_records WHERE elderly_id = $1 ORDER BY timestamp DESC LIMIT $2`,
		elderlyID, limit)
	if err != nil {
		return nil, fmt.Errorf("location history: %w", err)
	}
	defer rows.Close()

	var locations []LocationPoint
	for rows.Next() {
		var l LocationPoint
		if err := rows.Scan(&l.ID, &l.ElderlyID, &l.Lat, &l.Lon, &l.Accuracy, &l.Timestamp); err != nil {
			return nil, fmt.Errorf("scan location: %w", err)
		}
		locations = append(locations, l)
	}
	return locations, rows.Err()
}

// GetElderlyAlertHistory returns alert history.
func (s *PostgresStore) GetElderlyAlertHistory(ctx context.Context, elderlyID string, limit int) ([]AlertSummaryRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, elderly_id, alert_type, severity, status, created_at
		FROM alerts WHERE elderly_id = $1 ORDER BY created_at DESC LIMIT $2`,
		elderlyID, limit)
	if err != nil {
		return nil, fmt.Errorf("alert history: %w", err)
	}
	defer rows.Close()

	var alerts []AlertSummaryRow
	for rows.Next() {
		var a AlertSummaryRow
		if err := rows.Scan(&a.ID, &a.ElderlyID, &a.AlertType, &a.Severity, &a.Status, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan alert: %w", err)
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}
