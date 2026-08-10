package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"eregen.dev/admin-api/internal/model"
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"
)


// ========== Elderly Profile Management ==========

// ListElderly returns a paginated list of elderly profiles.
func (s *SqliteStore) ListElderly(ctx context.Context, page, pageSize int) ([]model.ElderlyProfile, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, user_id, birth_date, avatar_url, health_tiers, created_at, updated_at
		FROM elderly_profiles ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("list elderly: %w", err)
	}
	defer rows.Close()

	var profiles []model.ElderlyProfile
	for rows.Next() {
		var p model.ElderlyProfile
		var birthRaw, tiersRaw, createdAtStr, updatedAtStr string
		var avatarNull sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &p.UserID, &birthRaw, &avatarNull, &tiersRaw, &createdAtStr, &updatedAtStr); err != nil {
			return nil, fmt.Errorf("scan elderly: %w", err)
		}
		p.CreatedAt = parseTimeOrDefault(createdAtStr, time.Time{})
		p.UpdatedAt = parseTimeOrDefault(updatedAtStr, time.Time{})
		if birthRaw != "" {
			t := parseTimeOrDefault(birthRaw, time.Time{})
			p.BirthDate = &t
		}
		if avatarNull.Valid {
			p.AvatarURL = &avatarNull.String
		}
		if tiersRaw != "" {
			json.Unmarshal([]byte(tiersRaw), &p.HealthTiers)
		}
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}

// GetElderly returns an elderly profile by ID.
func (s *SqliteStore) GetElderly(ctx context.Context, id string) (*model.ElderlyProfile, error) {
	var p model.ElderlyProfile
	var birthRaw, avatarRaw, tiersRaw string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, user_id, birth_date, avatar_url, health_tiers, created_at, updated_at
		FROM elderly_profiles WHERE id = ?`, id).Scan(
		&p.ID, &p.Name, &p.UserID, &birthRaw, &avatarRaw, &tiersRaw, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("elderly not found")
		}
		return nil, fmt.Errorf("get elderly: %w", err)
	}
	if birthRaw != "" {
		if t, err := time.Parse(time.RFC3339, birthRaw); err == nil {
			p.BirthDate = &t
		}
	}
	if avatarRaw != "" {
		p.AvatarURL = &avatarRaw
	}
	if tiersRaw != "" {
		json.Unmarshal([]byte(tiersRaw), &p.HealthTiers)
	}
	return &p, nil
}

// CreateElderly inserts a new elderly profile.
func (s *SqliteStore) CreateElderly(ctx context.Context, name, birthDate, userID string, healthTiers []string, avatarURL string) (*model.ElderlyProfile, error) {
	tiersJSON, _ := json.Marshal(healthTiers)
	id := uuid.New().String()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO elderly_profiles (id, name, user_id, birth_date, avatar_url, health_tiers, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		id, name, userID, birthDate, avatarURL, string(tiersJSON))
	if err != nil {
		return nil, fmt.Errorf("create elderly: %w", err)
	}
	return &model.ElderlyProfile{ID: id, Name: name, UserID: userID, HealthTiers: healthTiers}, nil
}

// UpdateElderly modifies an existing elderly profile.
func (s *SqliteStore) UpdateElderly(ctx context.Context, id, name, birthDate, userID string, healthTiers []string, avatarURL string) (*model.ElderlyProfile, error) {
	tiersJSON, _ := json.Marshal(healthTiers)
	_, err := s.db.ExecContext(ctx,
		`UPDATE elderly_profiles SET name=?, user_id=?, birth_date=?, health_tiers=?, avatar_url=?, updated_at=datetime('now') WHERE id=?`,
		name, userID, birthDate, string(tiersJSON), avatarURL, id)
	if err != nil {
		return nil, fmt.Errorf("update elderly: %w", err)
	}
	return &model.ElderlyProfile{ID: id, Name: name, UserID: userID, HealthTiers: healthTiers}, nil
}

// DeleteElderly removes an elderly profile and its linked devices.
func (s *SqliteStore) DeleteElderly(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM elderly_profiles WHERE id = ?`, id)
	return err
}

// GetElderlyHealthStats returns aggregated health metrics for an elderly person.
func (s *SqliteStore) GetElderlyHealthStats(ctx context.Context, elderlyID string) (*model.HealthStats, error) {
	var stats model.HealthStats
	stats.ElderlyID = elderlyID
	err := s.db.QueryRowContext(ctx, `
		SELECT AVG(hr), MAX(hr), AVG(spo2), SUM(steps), MAX(timestamp)
		FROM health_records WHERE elderly_id = ?`, elderlyID).Scan(
		&stats.AvgHR, &stats.MaxHR, &stats.AvgSpO2, &stats.TotalSteps, &stats.LastSeen)
	if err != nil {
		return nil, fmt.Errorf("get health stats: %w", err)
	}
	return &stats, nil
}

// GetElderlyHealthRecords returns recent health records.
func (s *SqliteStore) GetElderlyHealthRecords(ctx context.Context, elderlyID string, limit int) ([]model.HealthRecordRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, elderly_id, timestamp, hr, spo2, steps, sleep_hours
		FROM health_records WHERE elderly_id = ? ORDER BY timestamp DESC LIMIT ?`, elderlyID, limit)
	if err != nil {
		return nil, fmt.Errorf("list health records: %w", err)
	}
	defer rows.Close()

	var items []model.HealthRecordRow
	for rows.Next() {
		var r model.HealthRecordRow
		if err := rows.Scan(&r.ID, &r.ElderlyID, &r.Timestamp, &r.HR, &r.SpO2, &r.Steps, &r.SleepHours); err != nil {
			return nil, fmt.Errorf("scan health record: %w", err)
		}
		items = append(items, r)
	}
	return items, rows.Err()
}

// GetElderlyMedicationRules returns medication rules for an elderly person.
func (s *SqliteStore) GetElderlyMedicationRules(ctx context.Context, elderlyID string) ([]model.MedicationRuleRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, elderly_id, schedule_time, pill_type, dose_count, days_of_week, active, created_at
		FROM medication_rules WHERE elderly_id = ? ORDER BY schedule_time`, elderlyID)
	if err != nil {
		return nil, fmt.Errorf("list medication rules: %w", err)
	}
	defer rows.Close()

	var items []model.MedicationRuleRow
	for rows.Next() {
		var r model.MedicationRuleRow
		var daysRaw string
		if err := rows.Scan(&r.ID, &r.ElderlyID, &r.ScheduleTime, &r.PillType, &r.DoseCount, &daysRaw, &r.Active, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan medication rule: %w", err)
		}
		json.Unmarshal([]byte(daysRaw), &r.DaysOfWeek)
		items = append(items, r)
	}
	return items, rows.Err()
}

// CreateMedicationRule inserts a new medication rule.
func (s *SqliteStore) CreateMedicationRule(ctx context.Context, elderlyID string, rule *model.MedicationRuleRow) error {
	rule.ID = uuid.New().String()
	daysJSON, _ := json.Marshal(rule.DaysOfWeek)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medication_rules (id, elderly_id, schedule_time, pill_type, dose_count, days_of_week, active)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rule.ID, elderlyID, rule.ScheduleTime, rule.PillType, rule.DoseCount, daysJSON, rule.Active)
	return err
}

// UpdateMedicationRule updates fields of an existing medication rule.
func (s *SqliteStore) UpdateMedicationRule(ctx context.Context, elderlyID, ruleID string, updates map[string]interface{}) error {
	parts := []string{}
	args := []interface{}{}
	for k, v := range updates {
		parts = append(parts, fmt.Sprintf("%s=?", k))
		args = append(args, v)
	}
	args = append(args, elderlyID, ruleID)
	query := fmt.Sprintf("UPDATE medication_rules SET %s WHERE elderly_id=? AND id=?",
		strings.Join(parts, ", "))
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

// DeleteMedicationRule removes a medication rule.
func (s *SqliteStore) DeleteMedicationRule(ctx context.Context, elderlyID, ruleID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM medication_rules WHERE elderly_id=? AND id=?`, elderlyID, ruleID)
	return err
}

// GetElderlyDevices returns devices linked to an elderly person.
func (s *SqliteStore) GetElderlyDevices(ctx context.Context, elderlyID string) ([]model.DeviceSummaryRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT d.id, d.device_id, d.device_type, d.tier, d.status,
		       COALESCE(json_extract(d.settings, '$.fw_version'),'v0.1'),
		       COALESCE(d.last_seen, '0001-01-01')
		FROM devices d JOIN elderly_devices ed ON d.id = ed.device_id
		WHERE ed.elderly_id = ? ORDER BY d.last_seen DESC`, elderlyID)
	if err != nil {
		return nil, fmt.Errorf("list elderly devices: %w", err)
	}
	defer rows.Close()

	var items []model.DeviceSummaryRow
	for rows.Next() {
		var d model.DeviceSummaryRow
		var lastSeenStr string
		if err := rows.Scan(&d.ID, &d.DeviceID, &d.Type, &d.Tier, &d.Status, &d.FirmwareVer, &lastSeenStr); err != nil {
			return nil, fmt.Errorf("scan elderly device: %w", err)
		}
		d.LastSeen = parseTimeOrDefault(lastSeenStr, time.Time{})
		items = append(items, d)
	}
	return items, rows.Err()
}

// GetElderlyLocationHistory returns location history for an elderly person.
func (s *SqliteStore) GetElderlyLocationHistory(ctx context.Context, elderlyID string, limit int) ([]model.LocationPoint, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, elderly_id, lat, lon, accuracy, timestamp
		FROM location_history WHERE elderly_id = ? ORDER BY timestamp DESC LIMIT ?`, elderlyID, limit)
	if err != nil {
		return nil, fmt.Errorf("list location history: %w", err)
	}
	defer rows.Close()

	var items []model.LocationPoint
	for rows.Next() {
		var p model.LocationPoint
		if err := rows.Scan(&p.ID, &p.ElderlyID, &p.Lat, &p.Lon, &p.Accuracy, &p.Timestamp); err != nil {
			return nil, fmt.Errorf("scan location point: %w", err)
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

// GetElderlyAlertHistory returns alert history for an elderly person.
func (s *SqliteStore) GetElderlyAlertHistory(ctx context.Context, elderlyID string, limit int) ([]model.AlertSummaryRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, elderly_id, alert_type, severity, status, created_at
		FROM alerts WHERE elderly_id = ? ORDER BY created_at DESC LIMIT ?`, elderlyID, limit)
	if err != nil {
		return nil, fmt.Errorf("list elderly alerts: %w", err)
	}
	defer rows.Close()

	var items []model.AlertSummaryRow
	for rows.Next() {
		var a model.AlertSummaryRow
		if err := rows.Scan(&a.ID, &a.ElderlyID, &a.AlertType, &a.Severity, &a.Status, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan elderly alert: %w", err)
		}
		items = append(items, a)
	}
	return items, rows.Err()
}

// CreateHealthRecord inserts a new health record.
func (s *SqliteStore) CreateHealthRecord(ctx context.Context, r *model.HealthRecordRow) error {
	r.ID = uuid.New().String()
	if r.Timestamp.IsZero() {
		r.Timestamp = time.Now()
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO health_records (id, elderly_id, timestamp, hr, spo2, steps, sleep_hours)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.ElderlyID, r.Timestamp, r.HR, r.SpO2, r.Steps, r.SleepHours)
	return err
}

// CreateLocation inserts a new location record.
func (s *SqliteStore) CreateLocation(ctx context.Context, loc *model.LocationPoint) error {
	loc.ID = uuid.New().String()
	if loc.Timestamp.IsZero() {
		loc.Timestamp = time.Now()
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO location_history (id, elderly_id, lat, lon, accuracy, timestamp)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		loc.ID, loc.ElderlyID, loc.Lat, loc.Lon, loc.Accuracy, loc.Timestamp)
	return err
}
