package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"eregen.dev/admin-api/internal/model"
	"github.com/google/uuid"
)

// ========== Elderly Profile Management (unified via persons + person_profiles) ==========

func (s *SqliteStore) ListElderly(ctx context.Context, page, pageSize int) ([]model.ElderlyProfile, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.id, p.name, p.phone, p.birth_date, p.avatar_url, p.status, pp.health_risk_level, pp.subscription_tier
		FROM persons p
		JOIN person_profiles pp ON p.id = pp.person_id AND pp.business_chain = 'self'
		ORDER BY p.created_at DESC LIMIT ? OFFSET ?`,
		pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("list elderly: %w", err)
	}
	defer rows.Close()

	var profiles []model.ElderlyProfile
	for rows.Next() {
		var p model.ElderlyProfile
		var birthRaw, statusRaw, riskLevel, subTier string
		var avatarNull sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &p.UserID, &birthRaw, &avatarNull, &statusRaw, &riskLevel, &subTier); err != nil {
			return nil, fmt.Errorf("scan elderly: %w", err)
		}
		p.CreatedAt = time.Now()
		if birthRaw != "" {
			if t, err := time.Parse("2006-01-02", birthRaw); err == nil {
				p.BirthDate = &t
			}
		}
		if avatarNull.Valid {
			p.AvatarURL = &avatarNull.String
		}
		if riskLevel != "" {
			p.HealthTiers = []string{riskLevel}
		}
		if subTier != "" {
			p.HealthTiers = append(p.HealthTiers, subTier)
		}
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}

func (s *SqliteStore) GetElderly(ctx context.Context, id string) (*model.ElderlyProfile, error) {
	var p model.ElderlyProfile
	var birthRaw, avatarRaw, statusRaw, riskLevel, subTier string
	err := s.db.QueryRowContext(ctx, `
		SELECT p.id, p.name, p.phone, p.birth_date, p.avatar_url, p.status, pp.health_risk_level, pp.subscription_tier
		FROM persons p
		JOIN person_profiles pp ON p.id = pp.person_id AND pp.business_chain = 'self'
		WHERE p.id = ?`, id).Scan(
		&p.ID, &p.Name, &p.UserID, &birthRaw, &avatarRaw, &statusRaw, &riskLevel, &subTier)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("elderly not found")
		}
		return nil, fmt.Errorf("get elderly: %w", err)
	}
	p.CreatedAt = time.Now()
	if birthRaw != "" {
		if t, err := time.Parse("2006-01-02", birthRaw); err == nil {
			p.BirthDate = &t
		}
	}
	if avatarRaw != "" {
		p.AvatarURL = &avatarRaw
	}
	if riskLevel != "" {
		p.HealthTiers = []string{riskLevel}
	}
	if subTier != "" {
		p.HealthTiers = append(p.HealthTiers, subTier)
	}
	return &p, nil
}

func (s *SqliteStore) CreateElderly(ctx context.Context, name, birthDate, userID string, healthTiers []string, avatarURL string) (*model.ElderlyProfile, error) {
	id := uuid.New().String()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO persons (id, name, phone, birth_date, avatar_url, status)
		 VALUES (?, ?, ?, ?, ?, 'active')`,
		id, name, userID, birthDate, avatarURL)
	if err != nil {
		return nil, fmt.Errorf("create person: %w", err)
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO person_profiles (person_id, business_chain, subscription_tier, subscription_status, health_risk_level)
		 VALUES (?, 'self', ?, 'active', ?)`,
		id, "starter", "low")
	if err != nil {
		return nil, fmt.Errorf("create profile: %w", err)
	}
	return &model.ElderlyProfile{ID: id, Name: name, UserID: userID, HealthTiers: healthTiers}, nil
}

func (s *SqliteStore) UpdateElderly(ctx context.Context, id, name, birthDate, userID string, healthTiers []string, avatarURL string) (*model.ElderlyProfile, error) {
	_, err := s.db.ExecContext(ctx,
		`UPDATE persons SET name=?, phone=?, birth_date=?, avatar_url=? WHERE id=?`,
		name, userID, birthDate, avatarURL, id)
	if err != nil {
		return nil, fmt.Errorf("update person: %w", err)
	}
	return &model.ElderlyProfile{ID: id, Name: name, UserID: userID, HealthTiers: healthTiers}, nil
}

func (s *SqliteStore) DeleteElderly(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM persons WHERE id = ?`, id)
	return err
}

func (s *SqliteStore) GetElderlyHealthStats(ctx context.Context, elderlyID string) (*model.HealthStats, error) {
	var stats model.HealthStats
	stats.ElderlyID = elderlyID
	err := s.db.QueryRowContext(ctx, `
		SELECT AVG(hr), MAX(hr), AVG(spo2), SUM(steps), MAX(recorded_at)
		FROM health_records WHERE person_id = ?`, elderlyID).Scan(
		&stats.AvgHR, &stats.MaxHR, &stats.AvgSpO2, &stats.TotalSteps, &stats.LastSeen)
	if err != nil {
		return nil, fmt.Errorf("get health stats: %w", err)
	}
	return &stats, nil
}

func (s *SqliteStore) GetElderlyHealthRecords(ctx context.Context, elderlyID string, limit int) ([]model.HealthRecordRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, person_id, timestamp, hr, spo2, steps, sleep_hours
		FROM health_records WHERE person_id = ? ORDER BY timestamp DESC LIMIT ?`, elderlyID, limit)
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

func (s *SqliteStore) GetElderlyMedicationRules(ctx context.Context, elderlyID string) ([]model.MedicationRuleRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, person_id, schedule_time, drug_name, dosage, frequency, days_of_week, active, created_at
		FROM medication_rules WHERE person_id = ? ORDER BY schedule_time`, elderlyID)
	if err != nil {
		return nil, fmt.Errorf("list medication rules: %w", err)
	}
	defer rows.Close()

	var items []model.MedicationRuleRow
	for rows.Next() {
		var r model.MedicationRuleRow
		var daysRaw string
		if err := rows.Scan(&r.ID, &r.ElderlyID, &r.ScheduleTime, &r.DrugName, &r.Dosage, &r.Frequency, &daysRaw, &r.Active, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan medication rule: %w", err)
		}
		json.Unmarshal([]byte(daysRaw), &r.DaysOfWeek)
		items = append(items, r)
	}
	return items, rows.Err()
}

func (s *SqliteStore) CreateMedicationRule(ctx context.Context, elderlyID string, rule *model.MedicationRuleRow) error {
	rule.ID = uuid.New().String()
	daysJSON, _ := json.Marshal(rule.DaysOfWeek)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medication_rules (id, person_id, business_chain, schedule_time, drug_name, dosage, frequency, days_of_week, active)
		 VALUES (?, ?, 'self', ?, ?, ?, ?, ?, ?)`,
		rule.ID, elderlyID, rule.ScheduleTime, rule.DrugName, rule.Dosage, rule.Frequency, daysJSON, rule.Active)
	return err
}

func (s *SqliteStore) UpdateMedicationRule(ctx context.Context, elderlyID, ruleID string, updates map[string]interface{}) error {
	parts := []string{}
	args := []interface{}{}
	for k, v := range updates {
		parts = append(parts, fmt.Sprintf("%s=?", k))
		args = append(args, v)
	}
	args = append(args, elderlyID, ruleID)
	query := fmt.Sprintf("UPDATE medication_rules SET %s WHERE person_id=? AND id=?",
		strings.Join(parts, ", "))
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

func (s *SqliteStore) DeleteMedicationRule(ctx context.Context, elderlyID, ruleID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM medication_rules WHERE person_id=? AND id=?`, elderlyID, ruleID)
	return err
}

func (s *SqliteStore) GetElderlyDevices(ctx context.Context, elderlyID string) ([]model.DeviceSummaryRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT d.id, d.device_id, d.device_type, d.tier, d.status,
		       COALESCE(json_extract(d.settings, '$.fw_version'),'v0.1'),
		       COALESCE(d.last_seen, '0001-01-01')
		FROM devices d JOIN device_bindings db ON d.id = db.device_id
		WHERE db.person_id = ? AND db.business_chain = 'self' ORDER BY d.last_seen DESC`, elderlyID)
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

func (s *SqliteStore) GetElderlyLocationHistory(ctx context.Context, elderlyID string, limit int) ([]model.LocationPoint, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, person_id, lat, lon, accuracy, timestamp
		FROM location_history WHERE person_id = ? ORDER BY timestamp DESC LIMIT ?`, elderlyID, limit)
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

func (s *SqliteStore) GetElderlyAlertHistory(ctx context.Context, elderlyID string, limit int) ([]model.AlertSummaryRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, person_id, alert_type, severity, status, created_at
		FROM alerts WHERE person_id = ? ORDER BY created_at DESC LIMIT ?`, elderlyID, limit)
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

func (s *SqliteStore) CreateHealthRecord(ctx context.Context, r *model.HealthRecordRow) error {
	r.ID = uuid.New().String()
	if r.Timestamp.IsZero() {
		r.Timestamp = time.Now()
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO health_records (id, person_id, business_chain, timestamp, hr, spo2, steps, sleep_hours)
		 VALUES (?, ?, 'self', ?, ?, ?, ?, ?)`,
		r.ID, r.ElderlyID, r.Timestamp, r.HR, r.SpO2, r.Steps, r.SleepHours)
	return err
}

func (s *SqliteStore) CreateLocation(ctx context.Context, loc *model.LocationPoint) error {
	loc.ID = uuid.New().String()
	if loc.Timestamp.IsZero() {
		loc.Timestamp = time.Now()
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO location_history (id, person_id, lat, lon, accuracy, timestamp)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		loc.ID, loc.ElderlyID, loc.Lat, loc.Lon, loc.Accuracy, loc.Timestamp)
	return err
}
