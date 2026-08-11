package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"eregen.dev/api-server/internal/model"

	"github.com/google/uuid"
)

// ========== User methods ==========

func (s *SqliteStore) CreateUser(ctx context.Context, u *model.User) error {
	u.ID = uuid.New().String()
	u.CreatedAt = time.Now()
	u.UpdatedAt = u.CreatedAt
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO users (id, email, phone, open_id, password_hash, role, name, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.Email, u.Phone, u.OpenID, u.PasswordHash, u.Role, u.Name, u.CreatedAt, u.UpdatedAt)
	return err
}

func (s *SqliteStore) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	u := &model.User{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, phone, open_id, password_hash, role, name, created_at, updated_at FROM users WHERE id = ?`, id).Scan(
		&u.ID, &u.Email, &u.Phone, &u.OpenID, &u.PasswordHash, &u.Role, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *SqliteStore) GetUserByPhone(ctx context.Context, phone string) (*model.User, error) {
	u := &model.User{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, phone, password_hash, role, name, created_at, updated_at FROM users WHERE phone = ?`, phone).Scan(
		&u.ID, &u.Email, &u.Phone, &u.PasswordHash, &u.Role, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (s *SqliteStore) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	u := &model.User{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, phone, open_id, password_hash, role, name, created_at, updated_at FROM users WHERE email = ?`, email).Scan(
		&u.ID, &u.Email, &u.Phone, &u.OpenID, &u.PasswordHash, &u.Role, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (s *SqliteStore) GetUserByOpenID(ctx context.Context, openID string) (*model.User, error) {
	u := &model.User{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, phone, open_id, password_hash, role, name, created_at, updated_at FROM users WHERE open_id = ?`, openID).Scan(
		&u.ID, &u.Email, &u.Phone, &u.OpenID, &u.PasswordHash, &u.Role, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (s *SqliteStore) UpdateUser(ctx context.Context, id string, name, phone, email *string) error {
	set := []string{"updated_at = datetime('now')"}
	args := []any{}
	if name != nil {
		set = append(set, "name = ?")
		args = append(args, *name)
	}
	if phone != nil {
		set = append(set, "phone = ?")
		args = append(args, *phone)
	}
	if email != nil {
		set = append(set, "email = ?")
		args = append(args, *email)
	}
	args = append(args, id)
	_, err := s.db.ExecContext(ctx,
		"UPDATE users SET "+strings.Join(set, ", ")+" WHERE id = ?", args...)
	return err
}

func (s *SqliteStore) DeleteUser(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, userID)
	return err
}

func (s *SqliteStore) ListUsers(ctx context.Context, page, pageSize int) ([]model.User, int, error) {
	offset := (page - 1) * pageSize
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, email, phone, open_id, password_hash, role, name, created_at, updated_at FROM users LIMIT ? OFFSET ?`,
		pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Phone, &u.OpenID, &u.PasswordHash, &u.Role, &u.Name, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	var total int
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
	return users, total, rows.Err()
}

func (s *SqliteStore) UpdateUserRole(ctx context.Context, userID, role string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET role = ?, updated_at = datetime('now') WHERE id = ?`, role, userID)
	return err
}

// ========== ElderlyProfile methods ==========

func (s *SqliteStore) CreateElderlyProfile(ctx context.Context, ep *model.ElderlyProfile) error {
	ep.ID = uuid.New().String()
	ep.CreatedAt = time.Now()
	ep.UpdatedAt = ep.CreatedAt
	data, _ := json.Marshal(ep.HealthTiers)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO elderly_profiles (id, user_id, name, birth_date, avatar_url, health_tiers, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		ep.ID, ep.UserID, ep.Name, ep.BirthDate, ep.AvatarURL, data, ep.CreatedAt, ep.UpdatedAt)
	return err
}

func (s *SqliteStore) GetElderlyProfile(ctx context.Context, elderlyID string) (*model.ElderlyProfile, error) {
	ep := &model.ElderlyProfile{}
	var data []byte
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, birth_date, avatar_url, health_tiers, created_at, updated_at
		 FROM elderly_profiles WHERE id = ?`, elderlyID).Scan(
		&ep.ID, &ep.UserID, &ep.Name, &ep.BirthDate, &ep.AvatarURL, &data, &ep.CreatedAt, &ep.UpdatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(data, &ep.HealthTiers)
	return ep, nil
}

func (s *SqliteStore) UpdateElderlyProfile(ctx context.Context, elderlyID string, req *model.UpdateElderlyRequest) error {
	set := []string{"updated_at = datetime('now')"}
	args := []any{}
	if req.Name != "" {
		set = append(set, "name = ?")
		args = append(args, req.Name)
	}
	if req.BirthDate != nil {
		set = append(set, "birth_date = ?")
		args = append(args, *req.BirthDate)
	}
	if req.AvatarURL != nil {
		set = append(set, "avatar_url = ?")
		args = append(args, *req.AvatarURL)
	}
	if len(req.HealthTiers) > 0 {
		data, _ := json.Marshal(req.HealthTiers)
		set = append(set, "health_tiers = ?")
		args = append(args, data)
	}
	args = append(args, elderlyID)
	_, err := s.db.ExecContext(ctx,
		"UPDATE elderly_profiles SET "+strings.Join(set, ", ")+" WHERE id = ?", args...)
	return err
}

func (s *SqliteStore) ListElderlyProfiles(ctx context.Context, userID string, page, pageSize int) ([]model.ElderlyProfile, int, error) {
	offset := (page - 1) * pageSize
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, name, birth_date, avatar_url, health_tiers, created_at, updated_at
		 FROM elderly_profiles WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var profiles []model.ElderlyProfile
	for rows.Next() {
		var ep model.ElderlyProfile
		var data []byte
		if err := rows.Scan(&ep.ID, &ep.UserID, &ep.Name, &ep.BirthDate, &ep.AvatarURL, &data, &ep.CreatedAt, &ep.UpdatedAt); err != nil {
			return nil, 0, err
		}
		json.Unmarshal(data, &ep.HealthTiers)
		profiles = append(profiles, ep)
	}
	var total int
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM elderly_profiles WHERE user_id = ?`, userID).Scan(&total)
	return profiles, total, rows.Err()
}

func (s *SqliteStore) GetElderlyProfilesByUserID(ctx context.Context, userID string) ([]model.ElderlyProfile, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, name, birth_date, avatar_url, health_tiers, created_at, updated_at
		 FROM elderly_profiles WHERE user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var profiles []model.ElderlyProfile
	for rows.Next() {
		var ep model.ElderlyProfile
		var data []byte
		if err := rows.Scan(&ep.ID, &ep.UserID, &ep.Name, &ep.BirthDate, &ep.AvatarURL, &data, &ep.CreatedAt, &ep.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(data, &ep.HealthTiers)
		profiles = append(profiles, ep)
	}
	return profiles, rows.Err()
}

func (s *SqliteStore) GetElderlyIDsByUserID(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM elderly_profiles WHERE user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *SqliteStore) GetDeviceByElderlyID(ctx context.Context, elderlyID string) (string, error) {
	var userID string
	err := s.db.QueryRowContext(ctx, `SELECT user_id FROM elderly_profiles WHERE id = ?`, elderlyID).Scan(&userID)
	if err != nil {
		return "", err
	}
	var deviceID string
	err = s.db.QueryRowContext(ctx,
		`SELECT device_id FROM devices WHERE owner_user_id = ? AND device_type = 'pillbox' ORDER BY created_at DESC LIMIT 1`,
		userID).Scan(&deviceID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return deviceID, err
}

func (s *SqliteStore) CheckElderlyAccess(ctx context.Context, elderlyID, userID string) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM elderly_profiles WHERE id = ? AND user_id = ?`, elderlyID, userID).Scan(&count)
	return count > 0, err
}

func (s *SqliteStore) MedRuleElderlyID(ctx context.Context, ruleID string) (string, error) {
	var elderlyID string
	err := s.db.QueryRowContext(ctx, `SELECT elderly_id FROM medication_rules WHERE id = ?`, ruleID).Scan(&elderlyID)
	return elderlyID, err
}

// ========== Device methods ==========

func (s *SqliteStore) CreateDevice(ctx context.Context, d *model.Device) error {
	d.ID = uuid.New().String()
	d.CreatedAt = time.Now()
	d.UpdatedAt = d.CreatedAt
	if d.Status == "" {
		d.Status = model.DeviceOffline
	}
	settingsJSON, _ := json.Marshal(d.Settings)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO devices (id, device_id, device_type, tier, owner_user_id, status, last_seen, created_at, updated_at, settings)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		d.ID, d.DeviceID, d.DeviceType, d.Tier, d.OwnerUserID, d.Status, d.LastSeen, d.CreatedAt, d.UpdatedAt, settingsJSON)
	return err
}

func (s *SqliteStore) ListDevices(ctx context.Context, ownerID string, deviceType *string, page, pageSize int) ([]model.Device, int, error) {
	where := "owner_user_id = ?"
	args := []any{ownerID}
	idx := 2
	if deviceType != nil && *deviceType != "" {
		where += fmt.Sprintf(" AND device_type = ?")
		args = append(args, *deviceType)
		idx++
	}
	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`SELECT id, device_id, device_type, tier, owner_user_id, status, last_seen, created_at, updated_at, settings
		FROM devices WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?`, where)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var devices []model.Device
	for rows.Next() {
		d, err := scanDeviceSQLite(rows)
		if err != nil {
			return nil, 0, err
		}
		devices = append(devices, *d)
	}
	countArgs := make([]any, len(args)-2)
	copy(countArgs, args)
	var count int
	s.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM devices WHERE %s", where), countArgs...).Scan(&count)
	return devices, count, rows.Err()
}

func (s *SqliteStore) GetDevice(ctx context.Context, deviceID string) (*model.Device, error) {
	var d model.Device
	var settingsJSON []byte
	err := s.db.QueryRowContext(ctx,
		`SELECT id, device_id, device_type, tier, owner_user_id, status, last_seen, created_at, updated_at, settings
		 FROM devices WHERE device_id = ?`, deviceID).Scan(
		&d.ID, &d.DeviceID, &d.DeviceType, &d.Tier, &d.OwnerUserID,
		&d.Status, &d.LastSeen, &d.CreatedAt, &d.UpdatedAt, &settingsJSON)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(settingsJSON, &d.Settings)
	return &d, nil
}

func (s *SqliteStore) GetDeviceByDeviceID(ctx context.Context, deviceID string) (*model.Device, error) {
	return s.GetDevice(ctx, deviceID)
}

func (s *SqliteStore) UpdateDeviceSettings(ctx context.Context, deviceID string, settings map[string]any) error {
	data, _ := json.Marshal(settings)
	_, err := s.db.ExecContext(ctx,
		`UPDATE devices SET settings = ?, updated_at = datetime('now'), status = 'online', last_seen = datetime('now') WHERE device_id = ?`,
		data, deviceID)
	return err
}

func (s *SqliteStore) DeleteDevice(ctx context.Context, deviceID, ownerID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM devices WHERE device_id = ? AND owner_user_id = ?`, deviceID, ownerID)
	return err
}

func (s *SqliteStore) BindDevice(ctx context.Context, deviceID, ownerUserID, deviceType, tier string) (*model.Device, error) {
	d := &model.Device{
		DeviceID: deviceID, DeviceType: deviceType, Tier: tier,
		OwnerUserID: ownerUserID, Status: model.DeviceOffline, Settings: map[string]any{},
	}
	settingsJSON, _ := json.Marshal(d.Settings)
	_, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO devices (id, device_id, device_type, tier, owner_user_id, status, last_seen, created_at, updated_at, settings)
		 VALUES (?, ?, ?, ?, ?, ?, NULL, datetime('now'), datetime('now'), ?)`,
		uuid.New().String(), deviceID, deviceType, tier, ownerUserID, settingsJSON)
	if err != nil {
		return nil, err
	}
	return s.GetDevice(ctx, deviceID)
}

func (s *SqliteStore) AdminDeleteDevice(ctx context.Context, deviceID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM devices WHERE device_id = ?`, deviceID)
	return err
}

func (s *SqliteStore) AdminDeviceList(ctx context.Context, deviceType, tier, status string, page, pageSize int) ([]model.Device, int, error) {
	where := "1=1"
	args := []any{}
	idx := 1
	if deviceType != "" {
		where += fmt.Sprintf(" AND device_type = ?")
		args = append(args, deviceType)
		idx++
	}
	if tier != "" {
		where += fmt.Sprintf(" AND tier = ?")
		args = append(args, tier)
		idx++
	}
	if status != "" {
		where += fmt.Sprintf(" AND status = ?")
		args = append(args, status)
		idx++
	}
	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`SELECT id, device_id, device_type, tier, owner_user_id, status, last_seen, created_at, updated_at, settings
		FROM devices WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?`, where)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var devices []model.Device
	for rows.Next() {
		d, err := scanDeviceSQLite(rows)
		if err != nil {
			return nil, 0, err
		}
		devices = append(devices, *d)
	}
	countArgs := make([]any, len(args)-2)
	copy(countArgs, args)
	var count int
	s.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM devices WHERE %s", where), countArgs...).Scan(&count)
	return devices, count, rows.Err()
}

// ========== HealthRecord methods ==========

func (s *SqliteStore) CreateHealthRecord(ctx context.Context, r *model.HealthRecord) error {
	r.ID = uuid.New().String()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO health_records (id, elderly_id, timestamp, hr, spo2, steps, sleep_hours, bp_systolic, bp_diastolic)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.ElderlyID, r.Timestamp, r.HR, r.SPO2, r.Steps, r.SleepHours, r.BPSystolic, r.BPDiastolic)
	return err
}

func (s *SqliteStore) GetHealthSummary(ctx context.Context, elderlyID string, day time.Time) (*model.HealthRecord, error) {
	r := &model.HealthRecord{}
	start := day.Format("2006-01-02")
	end := day.AddDate(0, 0, 1).Format("2006-01-02")
	err := s.db.QueryRowContext(ctx,
		`SELECT id, elderly_id, MAX(timestamp), AVG(hr), MIN(spo2), COALESCE(SUM(steps),0),
				MAX(sleep_hours), MAX(bp_systolic), MAX(bp_diastolic)
		 FROM health_records WHERE elderly_id = ? AND timestamp >= ? AND timestamp < ?
		 GROUP BY elderly_id`,
		elderlyID, start, end).Scan(
		&r.ID, &r.ElderlyID, &r.Timestamp, &r.HR, &r.SPO2, &r.Steps, &r.SleepHours, &r.BPSystolic, &r.BPDiastolic)
	return r, err
}

func (s *SqliteStore) GetHealthHistory(ctx context.Context, elderlyID string, days int) ([]model.HealthRecord, error) {
	until := time.Now()
	from := until.Add(-time.Duration(days) * 24 * time.Hour)
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, elderly_id, date(timestamp), AVG(hr), MIN(spo2), COALESCE(SUM(steps),0),
				MAX(sleep_hours), MAX(bp_systolic), MAX(bp_diastolic)
		 FROM health_records WHERE elderly_id = ? AND timestamp >= ? AND timestamp <= ?
		 GROUP BY elderly_id, date(timestamp) ORDER BY timestamp DESC`,
		elderlyID, from, until)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []model.HealthRecord
	for rows.Next() {
		var r model.HealthRecord
		if err := rows.Scan(&r.ID, &r.ElderlyID, &r.Timestamp, &r.HR, &r.SPO2, &r.Steps, &r.SleepHours, &r.BPSystolic, &r.BPDiastolic); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

func (s *SqliteStore) GetHealthTrend(ctx context.Context, elderlyID, metric string, days int) ([]model.HealthRecord, error) {
	_ = metric
	until := time.Now()
	from := until.Add(-time.Duration(days) * 24 * time.Hour)
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, elderly_id, timestamp, hr, spo2, steps, sleep_hours, bp_systolic, bp_diastolic
		 FROM health_records WHERE elderly_id = ? AND timestamp >= ? AND timestamp <= ?
		 ORDER BY timestamp ASC`, elderlyID, from, until)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []model.HealthRecord
	for rows.Next() {
		var r model.HealthRecord
		if err := rows.Scan(&r.ID, &r.ElderlyID, &r.Timestamp, &r.HR, &r.SPO2, &r.Steps, &r.SleepHours, &r.BPSystolic, &r.BPDiastolic); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

func (s *SqliteStore) GetHealthRecordsByElderlyID(ctx context.Context, elderlyID string, from, until time.Time) ([]model.HealthRecord, error) {
	return s.GetHealthHistory(ctx, elderlyID, int(until.Sub(from).Hours()/24))
}

func (s *SqliteStore) LatestHealthByElderlyID(ctx context.Context, elderlyID string, since time.Time) (*model.HealthRecord, error) {
	r := &model.HealthRecord{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, elderly_id, timestamp, hr, spo2, steps, sleep_hours, bp_systolic, bp_diastolic
		 FROM health_records WHERE elderly_id = ? AND timestamp >= ?
		 ORDER BY timestamp DESC LIMIT 1`, elderlyID, since).Scan(
		&r.ID, &r.ElderlyID, &r.Timestamp, &r.HR, &r.SPO2, &r.Steps, &r.SleepHours, &r.BPSystolic, &r.BPDiastolic)
	return r, err
}

func (s *SqliteStore) HealthRecordsByElderlyIDs(ctx context.Context, elderIDs []string, days int) ([]model.HealthRecord, error) {
	if len(elderIDs) == 0 {
		return nil, nil
	}
	until := time.Now()
	from := until.Add(-time.Duration(days) * 24 * time.Hour)
	placeholders := make([]string, len(elderIDs))
	args := make([]any, len(elderIDs)+2)
	for i, id := range elderIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	args[len(elderIDs)] = from
	args[len(elderIDs)+1] = until
	query := fmt.Sprintf(`SELECT id, elderly_id, timestamp, hr, spo2, steps, sleep_hours, bp_systolic, bp_diastolic
		FROM health_records WHERE elderly_id IN (%s) AND timestamp >= ? AND timestamp <= ?
		ORDER BY timestamp DESC`, strings.Join(placeholders, ","))
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []model.HealthRecord
	for rows.Next() {
		var r model.HealthRecord
		if err := rows.Scan(&r.ID, &r.ElderlyID, &r.Timestamp, &r.HR, &r.SPO2, &r.Steps, &r.SleepHours, &r.BPSystolic, &r.BPDiastolic); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

func (s *SqliteStore) HealthTrendByElderlyID(ctx context.Context, elderlyID string, days int) (avgHR, avgSpO2, totalSteps int64, lastHR, lastSpO2 *int, err error) {
	until := time.Now()
	from := until.Add(-time.Duration(days) * 24 * time.Hour)
	err = s.db.QueryRowContext(ctx,
		`SELECT COALESCE(AVG(hr),0), COALESCE(AVG(spo2),0), COALESCE(SUM(steps),0), MAX(hr), MAX(spo2)
		 FROM health_records WHERE elderly_id = ? AND timestamp >= ? AND timestamp <= ?`,
		elderlyID, from, until).Scan(&avgHR, &avgSpO2, &totalSteps, &lastHR, &lastSpO2)
	return
}

func (s *SqliteStore) GetElderlyName(ctx context.Context, elderlyID string) (string, error) {
	var name string
	err := s.db.QueryRowContext(ctx, `SELECT name FROM elderly_profiles WHERE id = ?`, elderlyID).Scan(&name)
	return name, err
}

// ========== LocationRecord methods ==========

func (s *SqliteStore) CreateLocationRecord(ctx context.Context, r *model.LocationRecord) error {
	r.ID = uuid.New().String()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO location_records (id, elderly_id, timestamp, lat, lon, accuracy)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		r.ID, r.ElderlyID, r.Timestamp, r.Lat, r.Lon, r.Accuracy)
	return err
}

func (s *SqliteStore) GetLatestLocation(ctx context.Context, elderlyID string) (*model.LocationRecord, error) {
	r := &model.LocationRecord{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, elderly_id, timestamp, lat, lon, accuracy FROM location_records
		 WHERE elderly_id = ? ORDER BY timestamp DESC LIMIT 1`, elderlyID).Scan(
		&r.ID, &r.ElderlyID, &r.Timestamp, &r.Lat, &r.Lon, &r.Accuracy)
	return r, err
}

func (s *SqliteStore) GetLocationHistory(ctx context.Context, elderlyID string, from, until time.Time) ([]model.LocationRecord, error) {
	return s.GetLocationHistoryByElderlyID(ctx, elderlyID, from, until)
}

func (s *SqliteStore) GetLocationHistoryByElderlyID(ctx context.Context, elderlyID string, from, until time.Time) ([]model.LocationRecord, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, elderly_id, timestamp, lat, lon, accuracy FROM location_records
		 WHERE elderly_id = ? AND timestamp >= ? AND timestamp <= ? ORDER BY timestamp DESC`,
		elderlyID, from, until)
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

// ========== MedicationRule methods ==========

func (s *SqliteStore) CreateMedicationRule(ctx context.Context, mr *model.MedicationRule) error {
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
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medication_rules (id, elderly_id, schedule_time, dose_count, pill_type, days_of_week, active, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		mr.ID, mr.ElderlyID, mr.ScheduleTime, mr.DoseCount, mr.PillType, data, mr.Active, mr.CreatedAt, mr.UpdatedAt)
	return err
}

func (s *SqliteStore) ListMedicationRules(ctx context.Context, elderlyID string) ([]model.MedicationRule, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, elderly_id, schedule_time, dose_count, pill_type, days_of_week, active, created_at, updated_at
		 FROM medication_rules WHERE elderly_id = ? AND active = 1 ORDER BY schedule_time ASC`, elderlyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRulesSQLite(rows)
}

func (s *SqliteStore) GetMedicationRule(ctx context.Context, ruleID string) (*model.MedicationRule, error) {
	mr := &model.MedicationRule{}
	var data []byte
	err := s.db.QueryRowContext(ctx,
		`SELECT id, elderly_id, schedule_time, dose_count, pill_type, days_of_week, active, created_at, updated_at
		 FROM medication_rules WHERE id = ?`, ruleID).Scan(
		&mr.ID, &mr.ElderlyID, &mr.ScheduleTime, &mr.DoseCount, &mr.PillType, &data, &mr.Active, &mr.CreatedAt, &mr.UpdatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(data, &mr.DaysOfWeek)
	return mr, nil
}

func (s *SqliteStore) UpdateMedicationRule(ctx context.Context, ruleID string, req *model.CreateMedicationRuleRequest) error {
	data, _ := json.Marshal(req.DaysOfWeek)
	_, err := s.db.ExecContext(ctx,
		`UPDATE medication_rules SET schedule_time = ?, dose_count = ?, pill_type = ?, days_of_week = ?, active = ?, updated_at = datetime('now') WHERE id = ?`,
		req.ScheduleTime, req.DoseCount, req.PillType, data, req.Active, ruleID)
	return err
}

func (s *SqliteStore) DeleteMedicationRule(ctx context.Context, ruleID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM medication_rules WHERE id = ?`, ruleID)
	return err
}

func (s *SqliteStore) GetMedicationRulesByElderlyID(ctx context.Context, elderlyID string) ([]model.MedicationRule, error) {
	return s.ListMedicationRules(ctx, elderlyID)
}

// ========== MedStatusRecord methods ==========

func (s *SqliteStore) CreateMedStatusRecord(ctx context.Context, r *model.MedStatusRecord) error {
	r.ID = uuid.New().String()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO med_status_records (id, rule_id, elderly_id, taken_at, taken, missed_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		r.ID, r.RuleID, r.ElderlyID, r.TakenAt, r.Taken, r.MissedAt)
	return err
}

func (s *SqliteStore) GetTodayMedStatus(ctx context.Context, elderlyID string) ([]model.MedStatusRecord, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.Add(24 * time.Hour)
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, rule_id, elderly_id, taken_at, taken, missed_at FROM med_status_records
		 WHERE elderly_id = ? AND taken_at >= ? AND taken_at < ? ORDER BY taken_at ASC`,
		elderlyID, start, end)
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

func (s *SqliteStore) GetMedicationHistory(ctx context.Context, elderlyID string, days int) ([]model.MedStatusRecord, error) {
	until := time.Now()
	from := until.Add(-time.Duration(days) * 24 * time.Hour)
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, rule_id, elderly_id, taken_at, taken, missed_at FROM med_status_records
		 WHERE elderly_id = ? AND taken_at >= ? AND taken_at <= ? ORDER BY taken_at DESC`,
		elderlyID, from, until)
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

func (s *SqliteStore) CreateMedTakeRecord(ctx context.Context, ruleID, elderlyID string) error {
	return s.CreateMedStatusRecord(ctx, &model.MedStatusRecord{RuleID: ruleID, ElderlyID: elderlyID, Taken: true, TakenAt: time.Now()})
}

// ========== Alert methods ==========

func (s *SqliteStore) CreateAlert(ctx context.Context, a *model.Alert) error {
	a.ID = uuid.New().String()
	a.CreatedAt = time.Now()
	if a.Status == "" {
		a.Status = model.AlertPending
	}
	metaJSON, _ := json.Marshal(a.Metadata)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO alerts (id, elderly_id, alert_type, severity, status, metadata, created_at, resolved_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.ElderlyID, a.AlertType, a.Severity, a.Status, metaJSON, a.CreatedAt, a.ResolvedAt)
	return err
}

func (s *SqliteStore) ListAlerts(ctx context.Context, elderIDs []string, filter *model.AlertFilter, page, pageSize int) ([]model.Alert, int, error) {
	if len(elderIDs) == 0 {
		return nil, 0, nil
	}
	placeholders := make([]string, len(elderIDs))
	args := make([]any, len(elderIDs))
	for i, id := range elderIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	where := fmt.Sprintf("elderly_id IN (%s)", strings.Join(placeholders, ","))
	idx := len(elderIDs) + 1
	if filter != nil {
		if filter.Severity != nil {
			where += fmt.Sprintf(" AND severity = ?")
			args = append(args, *filter.Severity)
			idx++
		}
		if filter.Status != nil {
			where += fmt.Sprintf(" AND status = ?")
			args = append(args, *filter.Status)
			idx++
		}
	}
	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)
	query := fmt.Sprintf(`SELECT id, elderly_id, alert_type, severity, status, metadata, created_at, resolved_at
		FROM alerts WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?`, where)
	rows, err := s.db.QueryContext(ctx, query, args...)
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
		json.Unmarshal(metaJSON, &a.Metadata)
		alerts = append(alerts, a)
	}
	countArgs := make([]any, len(args)-2)
	copy(countArgs, args)
	var count int
	s.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM alerts WHERE %s", where), countArgs...).Scan(&count)
	return alerts, count, rows.Err()
}

func (s *SqliteStore) GetAlert(ctx context.Context, alertID string) (*model.Alert, error) {
	a := &model.Alert{}
	var metaJSON []byte
	err := s.db.QueryRowContext(ctx,
		`SELECT id, elderly_id, alert_type, severity, status, metadata, created_at, resolved_at FROM alerts WHERE id = ?`,
		alertID).Scan(&a.ID, &a.ElderlyID, &a.AlertType, &a.Severity, &a.Status, &metaJSON, &a.CreatedAt, &a.ResolvedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(metaJSON, &a.Metadata)
	return a, nil
}

func (s *SqliteStore) UpdateAlert(ctx context.Context, alertID string, status model.AlertStatus) error {
	_, err := s.db.ExecContext(ctx, `UPDATE alerts SET status = ?, resolved_at = datetime('now') WHERE id = ?`, status, alertID)
	return err
}

func (s *SqliteStore) ResolveAlertByID(ctx context.Context, alertID string) error {
	return s.UpdateAlert(ctx, alertID, model.AlertResolved)
}

func (s *SqliteStore) GetAlertsByElderlyID(ctx context.Context, elderlyID string, from, until time.Time) ([]model.Alert, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, elderly_id, alert_type, severity, status, metadata, created_at, resolved_at
		 FROM alerts WHERE elderly_id = ? AND created_at >= ? AND created_at <= ? ORDER BY created_at DESC`,
		elderlyID, from, until)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var alerts []model.Alert
	for rows.Next() {
		var a model.Alert
		var metaJSON []byte
		if err := rows.Scan(&a.ID, &a.ElderlyID, &a.AlertType, &a.Severity, &a.Status, &metaJSON, &a.CreatedAt, &a.ResolvedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(metaJSON, &a.Metadata)
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

func (s *SqliteStore) GetAlertElderlyID(ctx context.Context, alertID string) (string, error) {
	var elderlyID string
	err := s.db.QueryRowContext(ctx, `SELECT elderly_id FROM alerts WHERE id = ?`, alertID).Scan(&elderlyID)
	return elderlyID, err
}

// ========== Geofence methods ==========

func (s *SqliteStore) CreateGeofence(ctx context.Context, gf *model.Geofence) error {
	gf.ID = uuid.New().String()
	gf.CreatedAt = time.Now()
	gf.UpdatedAt = gf.CreatedAt
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO geofences (id, elderly_id, name, latitude, longitude, radius_meters, active, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		gf.ID, gf.ElderlyID, gf.Name, gf.Latitude, gf.Longitude, gf.RadiusMeters, gf.Active, gf.CreatedAt, gf.UpdatedAt)
	return err
}

func (s *SqliteStore) ListGeofences(ctx context.Context, elderlyID string) ([]model.Geofence, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, elderly_id, name, latitude, longitude, radius_meters, active, created_at, updated_at
		 FROM geofences WHERE elderly_id = ? ORDER BY name ASC`, elderlyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var fences []model.Geofence
	for rows.Next() {
		var g model.Geofence
		if err := rows.Scan(&g.ID, &g.ElderlyID, &g.Name, &g.Latitude, &g.Longitude, &g.RadiusMeters, &g.Active, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		fences = append(fences, g)
	}
	return fences, rows.Err()
}

func (s *SqliteStore) UpdateGeofence(ctx context.Context, id string, req *model.UpdateGeofenceRequest) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE geofences SET name=?, latitude=?, longitude=?, radius_meters=?, active=?, updated_at=datetime('now') WHERE id=?`,
		req.Name, req.Lat, req.Lon, req.RadiusMeters, req.Active, id)
	return err
}

func (s *SqliteStore) DeleteGeofence(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM geofences WHERE id = ?`, id)
	return err
}

// ========== Subscription methods ==========

func (s *SqliteStore) CreateSubscription(ctx context.Context, sub *model.Subscription) error {
	sub.ID = uuid.New().String()
	sub.StartDate = time.Now()
	sub.EndDate = sub.StartDate.AddDate(0, 1, 0)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO subscriptions (id, user_id, plan_tier, status, start_date, end_date, auto_renew, payment_method, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		sub.ID, sub.UserID, sub.PlanTier, sub.Status, sub.StartDate, sub.EndDate, sub.AutoRenew, sub.PaymentMethod)
	return err
}

func (s *SqliteStore) GetSubscription(ctx context.Context, userID string) (*model.Subscription, error) {
	sub := &model.Subscription{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, plan_tier, status, start_date, end_date, auto_renew, payment_method, created_at, updated_at
		 FROM subscriptions WHERE user_id = ? ORDER BY created_at DESC LIMIT 1`, userID).Scan(
		&sub.ID, &sub.UserID, &sub.PlanTier, &sub.Status, &sub.StartDate, &sub.EndDate,
		&sub.AutoRenew, &sub.PaymentMethod, &sub.CreatedAt, &sub.UpdatedAt)
	return sub, err
}

func (s *SqliteStore) ListSubscriptions(ctx context.Context, userID string, page, pageSize int) ([]model.Subscription, int, error) {
	offset := (page - 1) * pageSize
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, plan_tier, status, start_date, end_date, auto_renew, payment_method, created_at, updated_at
		 FROM subscriptions WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var subs []model.Subscription
	for rows.Next() {
		var sub model.Subscription
		if err := rows.Scan(&sub.ID, &sub.UserID, &sub.PlanTier, &sub.Status, &sub.StartDate, &sub.EndDate,
			&sub.AutoRenew, &sub.PaymentMethod, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
			return nil, 0, err
		}
		subs = append(subs, sub)
	}
	var total int
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM subscriptions WHERE user_id = ?`, userID).Scan(&total)
	return subs, total, rows.Err()
}

// ========== FirmwareRelease methods ==========

func (s *SqliteStore) CreateFirmwareRelease(ctx context.Context, r *model.FirmwareRelease) error {
	r.ID = uuid.New().String()
	r.CreatedAt = time.Now()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO firmware_releases (id, device_type, tier, version, url, sha256_hash, signature, changelog, min_app_version, force_update, active, settings, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, datetime('now'))`,
		r.ID, r.DeviceType, r.Tier, r.Version, r.URL, r.Sha256Hash, r.Signature, r.Changelog,
		r.MinAppVersion, r.ForceUpdate, r.Active)
	return err
}

func (s *SqliteStore) ListFirmwareReleases(ctx context.Context, deviceType, tier string) ([]model.FirmwareRelease, error) {
	where := "1=1"
	args := []any{}
	idx := 1
	if deviceType != "" {
		where += fmt.Sprintf(" AND device_type = ?")
		args = append(args, deviceType)
		idx++
	}
	if tier != "" {
		where += fmt.Sprintf(" AND tier = ?")
		args = append(args, tier)
		idx++
	}
	rows, err := s.db.QueryContext(ctx,
		fmt.Sprintf(`SELECT id, device_type, tier, version, url, sha256_hash, signature, changelog, min_app_version, force_update, active, created_at
			FROM firmware_releases WHERE %s ORDER BY created_at DESC`, where), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var releases []model.FirmwareRelease
	for rows.Next() {
		var r model.FirmwareRelease
		if err := rows.Scan(&r.ID, &r.DeviceType, &r.Tier, &r.Version, &r.URL, &r.Sha256Hash,
			&r.Signature, &r.Changelog, &r.MinAppVersion, &r.ForceUpdate, &r.Active, &r.CreatedAt); err != nil {
			return nil, err
		}
		releases = append(releases, r)
	}
	return releases, rows.Err()
}

func (s *SqliteStore) GetFirmwareRelease(ctx context.Context, id string) (*model.FirmwareRelease, error) {
	r := &model.FirmwareRelease{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, device_type, tier, version, url, sha256_hash, signature, changelog, min_app_version, force_update, active, created_at
		 FROM firmware_releases WHERE id = ?`, id).Scan(
		&r.ID, &r.DeviceType, &r.Tier, &r.Version, &r.URL, &r.Sha256Hash,
		&r.Signature, &r.Changelog, &r.MinAppVersion, &r.ForceUpdate, &r.Active, &r.CreatedAt)
	return r, err
}

// ========== OTAJob methods ==========

func (s *SqliteStore) CreateOTAJob(ctx context.Context, j *model.OTAJob) error {
	j.ID = uuid.New().String()
	j.CreatedAt = time.Now()
	targetJSON, _ := json.Marshal(j.TargetDevices)
	progressJSON, _ := json.Marshal(j.Progress)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO ota_jobs (id, firmware_id, target_devices, progress, created_at)
		 VALUES (?, ?, ?, ?, datetime('now'))`,
		j.ID, j.FirmwareID, targetJSON, progressJSON)
	return err
}

func (s *SqliteStore) GetOTAJob(ctx context.Context, id string) (*model.OTAJob, error) {
	j := &model.OTAJob{}
	var targetJSON, progressJSON []byte
	err := s.db.QueryRowContext(ctx,
		`SELECT id, firmware_id, target_devices, progress, created_at FROM ota_jobs WHERE id = ?`, id).Scan(
		&j.ID, &j.FirmwareID, &targetJSON, &progressJSON, &j.CreatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(targetJSON, &j.TargetDevices)
	json.Unmarshal(progressJSON, &j.Progress)
	return j, nil
}

func (s *SqliteStore) UpdateOTAJobProgress(ctx context.Context, jobID string, fn UpdateOTAJobProgressFn) error {
	j, err := s.GetOTAJob(ctx, jobID)
	if err != nil {
		return err
	}
	fn(j.Progress)
	progressJSON, _ := json.Marshal(j.Progress)
	_, err = s.db.ExecContext(ctx, `UPDATE ota_jobs SET progress = ? WHERE id = ?`, progressJSON, jobID)
	return err
}

// ========== Admin stats ==========

func (s *SqliteStore) AdminStatsOverview(ctx context.Context) (*StatsOverview, error) {
	svo := &StatsOverview{}
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM devices WHERE status = 'online' AND last_seen > datetime('now', '-5 minutes')`).Scan(&svo.OnlineDevices)
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE role IN ('family', 'elderly', 'institution')`).Scan(&svo.TotalUsers)
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM alerts WHERE status = 'pending'`).Scan(&svo.AlertCount)
	var subRate float64
	s.db.QueryRowContext(ctx,
		`SELECT CASE WHEN COUNT(*) = 0 THEN 0.0 ELSE ROUND(CAST(SUM(CASE WHEN plan_tier IN ('premium','enterprise') THEN 1 ELSE 0 END) AS REAL) / COUNT(*), 4) END
		 FROM subscriptions WHERE status = 'active'`).Scan(&subRate)
	svo.SubscriptionRate = subRate * 100
	return svo, nil
}

func (s *SqliteStore) AdminStatsAlertTrend(ctx context.Context, days int) ([]AlertTrendPoint, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT date(created_at) as date, COUNT(*) FROM alerts
		 WHERE created_at >= datetime('now', ?)
		 GROUP BY date(created_at) ORDER BY date ASC`, fmt.Sprintf("-%d days", days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var points []AlertTrendPoint
	for rows.Next() {
		var p AlertTrendPoint
		if err := rows.Scan(&p.Date, &p.BraceletCount); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, rows.Err()
}

func (s *SqliteStore) AdminStatsAlertDistribution(ctx context.Context) ([]AlertDistributionItem, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT alert_type, COUNT(*) as cnt FROM alerts GROUP BY alert_type ORDER BY cnt DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AlertDistributionItem
	for rows.Next() {
		var it AlertDistributionItem
		if err := rows.Scan(&it.Name, &it.Value); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (s *SqliteStore) AdminStatsUserGrowth(ctx context.Context, months int) ([]UserGrowthPoint, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT strftime('%Y-%m', created_at) as month, COUNT(*) FROM users
		 WHERE created_at >= datetime('now', ?)
		 GROUP BY month ORDER BY month ASC`, fmt.Sprintf("-%d months", months))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var points []UserGrowthPoint
	for rows.Next() {
		var p UserGrowthPoint
		if err := rows.Scan(&p.Month, &p.NewUsers); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, rows.Err()
}

func (s *SqliteStore) SubscriptionTierStats(ctx context.Context) ([]struct{ Tier string; Count int }, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT plan_tier, COUNT(*) as cnt FROM subscriptions WHERE status = 'active' GROUP BY plan_tier`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type stat struct{ Tier string; Count int }
	var results []stat
	for rows.Next() {
		var s stat
		if err := rows.Scan(&s.Tier, &s.Count); err != nil {
			return nil, err
		}
		results = append(results, s)
	}
	return results, rows.Err()
}

// ========== Helper functions for SQLite ==========

func scanDeviceSQLite(rows interface {
	Scan(dest ...any) error
}) (*model.Device, error) {
	var d model.Device
	var settingsJSON []byte
	err := rows.Scan(&d.ID, &d.DeviceID, &d.DeviceType, &d.Tier, &d.OwnerUserID,
		&d.Status, &d.LastSeen, &d.CreatedAt, &d.UpdatedAt, &settingsJSON)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(settingsJSON, &d.Settings)
	return &d, nil
}

func scanRulesSQLite(rows interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
	Err() error
}) ([]model.MedicationRule, error) {
	var rules []model.MedicationRule
	for rows.Next() {
		var mr model.MedicationRule
		var data []byte
		if err := rows.Scan(&mr.ID, &mr.ElderlyID, &mr.ScheduleTime, &mr.DoseCount,
			&mr.PillType, &data, &mr.Active, &mr.CreatedAt, &mr.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(data, &mr.DaysOfWeek)
		rules = append(rules, mr)
	}
	return rules, rows.Err()
}
