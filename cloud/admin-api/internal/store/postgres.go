package store

import (
	"context"
	"database/sql"
	"fmt"
	"eregen.dev/admin-api/internal/model"
	"time"

	_ "github.com/lib/pq"
)

// PostgresStore wraps database access for admin operations.
type PostgresStore struct {
	db *sql.DB
}

// NewPostgres opens a connection pool to PostgreSQL.
func NewPostgres(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(fmt.Sprintf("failed to open postgres: %v", err))
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	if err := db.Ping(); err != nil {
		panic(fmt.Sprintf("failed to ping postgres: %v", err))
	}
	return db
}

// NewStore creates a PostgresStore from an existing *sql.DB.
func NewStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// GetDashboardStats returns the top-level metrics for the admin dashboard.
func (s *PostgresStore) GetDashboardStats(ctx context.Context) (*model.DashboardStats, error) {
	var stats model.DashboardStats
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM devices WHERE status='online'`).Scan(&stats.OnlineDevices); err != nil {
		return nil, fmt.Errorf("online devices: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM devices`).Scan(&stats.TotalDevices); err != nil {
		return nil, fmt.Errorf("total devices: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM alerts WHERE status='pending'`).Scan(&stats.ActiveAlerts); err != nil {
		return nil, fmt.Errorf("active alerts: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers); err != nil {
		return nil, fmt.Errorf("total users: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM subscriptions WHERE status='active'`).Scan(&stats.ActiveSubscriptions); err != nil {
		return nil, fmt.Errorf("active subscriptions: %w", err)
	}
	return &stats, nil
}

// ListDevices returns a paginated list of devices with optional filters.
func (s *PostgresStore) ListDevices(ctx context.Context, page, pageSize int, status, devType, tier string) ([]model.DeviceSummary, error) {
	query := `SELECT id, device_id, device_type, tier, status, COALESCE(last_seen, '0001-01-01'),
		(SELECT u.name FROM users u JOIN devices d ON d.owner_user_id = u.id WHERE d.id = devices.id LIMIT 1),
		COALESCE(settings->>'fw_version','v0.1')
		FROM devices WHERE 1=1`
	args := []interface{}{}
	idx := 1
	if status != "" {
		query += fmt.Sprintf(" AND status=$%d", idx)
		args = append(args, status)
		idx++
	}
	if devType != "" {
		query += fmt.Sprintf(" AND device_type=$%d", idx)
		args = append(args, devType)
		idx++
	}
	if tier != "" {
		query += fmt.Sprintf(" AND tier=$%d", idx)
		args = append(args, tier)
		idx++
	}
	query += fmt.Sprintf(" ORDER BY last_seen DESC NULLS LAST LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	defer rows.Close()

	var devices []model.DeviceSummary
	for rows.Next() {
		var d model.DeviceSummary
		if err := rows.Scan(&d.ID, &d.DeviceID, &d.Type, &d.Tier, &d.Status, &d.LastSeen, &d.OwnerName, &d.FirmwareVer); err != nil {
			return nil, fmt.Errorf("scan device: %w", err)
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

// ListUsers returns a paginated list of users with optional role filter.
func (s *PostgresStore) ListUsers(ctx context.Context, page, pageSize int, role string) ([]model.UserSummary, error) {
	query := `SELECT u.id, u.name, u.role, u.created_at,
		(SELECT COUNT(*) FROM devices d WHERE d.owner_user_id = u.id)
		FROM users u WHERE 1=1`
	args := []interface{}{}
	idx := 1
	if role != "" {
		query += fmt.Sprintf(" AND u.role=$%d", idx)
		args = append(args, role)
		idx++
	}
	query += fmt.Sprintf(" ORDER BY u.created_at DESC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []model.UserSummary
	for rows.Next() {
		var u model.UserSummary
		if err := rows.Scan(&u.ID, &u.Name, &u.Role, &u.CreatedAt, &u.Devices); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// ListAlerts returns recent alerts with optional severity and status filters.
func (s *PostgresStore) ListAlerts(ctx context.Context, severity, status string, limit int) ([]model.AlertSummary, error) {
	query := `SELECT a.id, a.elderly_id, a.alert_type, a.severity, a.status, a.created_at,
		COALESCE(d.device_id, '')
		FROM alerts a LEFT JOIN devices d ON a.elderly_id = d.id WHERE 1=1`
	args := []interface{}{}
	idx := 1
	if severity != "" {
		query += fmt.Sprintf(" AND a.severity=$%d", idx)
		args = append(args, severity)
		idx++
	}
	if status != "" {
		query += fmt.Sprintf(" AND a.status=$%d", idx)
		args = append(args, status)
		idx++
	}
	query += fmt.Sprintf(" ORDER BY a.created_at DESC LIMIT $%d", idx)
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}
	defer rows.Close()

	var alerts []model.AlertSummary
	for rows.Next() {
		var a model.AlertSummary
		if err := rows.Scan(&a.ID, &a.ElderlyID, &a.AlertType, &a.Severity, &a.Status, &a.CreatedAt, &a.DeviceID); err != nil {
			return nil, fmt.Errorf("scan alert: %w", err)
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

// SetUserRole updates a user's role.
func (s *PostgresStore) SetUserRole(ctx context.Context, userID, role string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET role = $1 WHERE id = $2`, role, userID)
	return err
}

// UpdateDeviceConfig updates device settings JSONB column.
func (s *PostgresStore) UpdateDeviceConfig(ctx context.Context, deviceID string, config map[string]interface{}) error {
	settings := `{}`
	for k, v := range config {
		if v != nil {
			settings += fmt.Sprintf(`"%s":%v,`, k, v)
		}
	}
	_, err := s.db.ExecContext(ctx, `UPDATE devices SET settings = settings || $1::jsonb WHERE device_id = $2`, settings, deviceID)
	return err
}

// TriggerOTA schedules an OTA update for a device.
func (s *PostgresStore) TriggerOTA(ctx context.Context, deviceID, firmwareURL, sha256Hash string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE devices SET ota_url = $1, ota_hash = $2, ota_status = 'pending' WHERE device_id = $3`,
		firmwareURL, sha256Hash, deviceID)
	return err
}

// ResolveAlert marks an alert as resolved.
func (s *PostgresStore) ResolveAlert(ctx context.Context, alertID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE alerts SET status = 'resolved', resolved_at = NOW() WHERE id = $1`, alertID)
	return err
}

// GetSubscriptionStats returns a per-tier subscription count breakdown.
func (s *PostgresStore) GetSubscriptionStats(ctx context.Context) ([]model.SubscriptionStat, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT plan_tier, COUNT(*)::int,
		       ROUND(COUNT(*)::numeric / NULLIF(SUM(COUNT(*)) OVER (), 0) * 100, 1)
		FROM subscriptions GROUP BY plan_tier ORDER BY COUNT(*) DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("subscription stats: %w", err)
	}
	defer rows.Close()

	var stats []model.SubscriptionStat
	for rows.Next() {
		var st model.SubscriptionStat
		if err := rows.Scan(&st.Tier, &st.Count, &st.Pct); err != nil {
			return nil, fmt.Errorf("scan subscription stat: %w", err)
		}
		stats = append(stats, st)
	}
	return stats, rows.Err()
}

// GetAlertTrend returns alert counts grouped by date and device type.
func (s *PostgresStore) GetAlertTrend(ctx context.Context, days int) ([]model.AlertTrendPoint, error) {
	query := `SELECT DATE(a.created_at) AS alert_date,
		       COUNT(*) FILTER (WHERE d.device_type = 'bracelet')::int AS bracelet_count,
		       COUNT(*) FILTER (WHERE d.device_type = 'pillbox')::int AS pillbox_count
		FROM alerts a LEFT JOIN devices d ON a.elderly_id = d.id
		WHERE a.created_at >= NOW() - (INTERVAL '1 day' * $1)
		GROUP BY DATE(a.created_at) ORDER BY alert_date`
	rows, err := s.db.QueryContext(ctx, query, days)
	if err != nil {
		return nil, fmt.Errorf("alert trend: %w", err)
	}
	defer rows.Close()

	var result []model.AlertTrendPoint
	for rows.Next() {
		var p model.AlertTrendPoint
		if err := rows.Scan(&p.Date, &p.BraceletCount, &p.PillboxCount); err != nil {
			return nil, fmt.Errorf("scan alert trend: %w", err)
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

// GetAlertDistribution returns alert counts by type.
func (s *PostgresStore) GetAlertDistribution(ctx context.Context) ([]model.AlertDistributionItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT alert_type, COUNT(*)::int FROM alerts GROUP BY alert_type ORDER BY COUNT(*) DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("alert distribution: %w", err)
	}
	defer rows.Close()

	colors := map[string]string{
		"sos":        "#ff4d4f",
		"fall":       "#fa541c",
		"med_missed": "#faad14",
		"device_offline": "#1890ff",
		"geofence_breach": "#722ed1",
	}

	var result []model.AlertDistributionItem
	for rows.Next() {
		var item model.AlertDistributionItem
		if err := rows.Scan(&item.Name, &item.Value); err != nil {
			return nil, fmt.Errorf("scan alert dist: %w", err)
		}
		item.Color = colors[item.Name]
		result = append(result, item)
	}
	return result, rows.Err()
}

// GetUserGrowth returns new user counts grouped by month.
func (s *PostgresStore) GetUserGrowth(ctx context.Context, months int) ([]model.UserGrowthPoint, error) {
	query := `SELECT TO_CHAR(DATE_TRUNC('month', created_at), 'YYYY-MM') AS month,
		       COUNT(*)::int AS new_users
		FROM users GROUP BY DATE_TRUNC('month', created_at)
		ORDER BY month DESC LIMIT $1`
	rows, err := s.db.QueryContext(ctx, query, months)
	if err != nil {
		return nil, fmt.Errorf("user growth: %w", err)
	}
	defer rows.Close()

	var result []model.UserGrowthPoint
	for rows.Next() {
		var p model.UserGrowthPoint
		if err := rows.Scan(&p.Month, &p.NewUsers); err != nil {
			return nil, fmt.Errorf("scan user growth: %w", err)
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

// GetDeviceByID returns a single device by its database ID.
func (s *PostgresStore) GetDeviceByID(ctx context.Context, id string) (*model.DeviceDetail, error) {
	var d model.DeviceDetail
	err := s.db.QueryRowContext(ctx, `
		SELECT d.id, d.device_id, d.device_type, d.tier, d.status, COALESCE(d.last_seen, '0001-01-01'),
		       u.name, COALESCE(d.settings->>'fw_version','v0.1'),
		       d.settings,
		       e.name AS elderly_name
		FROM devices d LEFT JOIN users u ON d.owner_user_id = u.id
		LEFT JOIN elderly_profiles e ON d.id = ANY((SELECT ed.device_id FROM elderly_devices ed WHERE ed.elderly_id = e.id LIMIT 1))
		WHERE d.id = $1`, id).Scan(
		&d.ID, &d.DeviceID, &d.Type, &d.Tier, &d.Status, &d.LastSeen,
		&d.OwnerName, &d.FirmwareVer, &d.SettingsJSON, &d.ElderlyName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found")
		}
		return nil, fmt.Errorf("get device: %w", err)
	}
	return &d, nil
}

// UnbindDevice removes a device from its owner and all elderly links.
func (s *PostgresStore) UnbindDevice(ctx context.Context, deviceID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM elderly_devices WHERE device_id = $1;
		 UPDATE devices SET owner_user_id = NULL WHERE device_id = $2`,
		deviceID, deviceID)
	return err
}

// BatchTriggerOTA schedules OTA updates for multiple devices.
func (s *PostgresStore) BatchTriggerOTA(ctx context.Context, deviceIDs, firmwareURL, sha256Hash []string) error {
	for i, id := range deviceIDs {
		url := firmwareURL[i % len(firmwareURL)]
		hash := sha256Hash[i % len(sha256Hash)]
		if _, err := s.db.ExecContext(ctx,
			`UPDATE devices SET ota_url = $1, ota_hash = $2, ota_status = 'pending' WHERE device_id = $3`,
			url, hash, id); err != nil {
			return fmt.Errorf("batch OTA device %s: %w", id, err)
		}
	}
	return nil
}

// CreateFirmwareVersion inserts a new firmware release record.
func (s *PostgresStore) CreateFirmwareVersion(ctx context.Context, v *model.FirmwareVersion) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO firmware_releases (device_type, tier, version, url, sha256_hash, changelog, min_app_version, force_update, active)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,true)`,
		v.DeviceType, v.Tier, v.Version, v.DownloadURL, v.Sha256Hash, v.Changelog, v.MinAppVersion, v.ForceUpdate)
	return err
}

// ListFirmwareVersions returns all firmware versions.
func (s *PostgresStore) ListFirmwareVersions(ctx context.Context) ([]model.FirmwareVersion, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, device_type, tier, version, url, sha256_hash, changelog, min_app_version, force_update, active, created_at
		FROM firmware_releases ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list firmware: %w", err)
	}
	defer rows.Close()

	var result []model.FirmwareVersion
	for rows.Next() {
		var f model.FirmwareVersion
		if err := rows.Scan(&f.ID, &f.DeviceType, &f.Tier, &f.Version, &f.DownloadURL,
			&f.Sha256Hash, &f.Changelog, &f.MinAppVersion, &f.ForceUpdate, &f.IsActive, &f.ReleaseDate); err != nil {
			return nil, fmt.Errorf("scan firmware: %w", err)
		}
		result = append(result, f)
	}
	return result, rows.Err()
}

// DeleteFirmwareVersion soft-deletes a firmware release.
func (s *PostgresStore) DeleteFirmwareVersion(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE firmware_releases SET active = false WHERE id = $1`, id)
	return err
}

// PushOTAJob records an OTA push job.
func (s *PostgresStore) PushOTAJob(ctx context.Context, firmwareID string, deviceIDs []string) error {
	devicesJSON := "[]"
	if len(deviceIDs) > 0 {
		devicesJSON = fmt.Sprintf("%v", deviceIDs) // simplified; use json.Marshal in production
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO ota_jobs (firmware_id, target_devices, progress) VALUES ($1, $2, '{"total":0,"pending":0}')`,
		firmwareID, devicesJSON)
	return err
}

// GetNotificationSettings retrieves system notification config.
func (s *PostgresStore) GetNotificationSettings(ctx context.Context) (map[string]any, error) {
	var jsonb string
	err := s.db.QueryRowContext(ctx, `SELECT COALESCE(setting_value, '{}') FROM system_settings WHERE key = 'notification'`).Scan(&jsonb)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("get notification settings: %w", err)
	}
	return map[string]any{}, nil // default empty
}

// UpdateNotificationSettings persists notification config.
func (s *PostgresStore) UpdateNotificationSettings(ctx context.Context, data map[string]any) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO system_settings (key, setting_value) VALUES ('notification', $1)
		 ON CONFLICT (key) DO UPDATE SET setting_value = $1`,
		`{}`) // placeholder — use json.Marshal in real impl
	return err
}

// ListAPIKeys returns registered B2B API keys.
func (s *PostgresStore) ListAPIKeys(ctx context.Context) ([]model.APIKeySummary, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, key_prefix, expires_at, active, created_at
		FROM b2b_api_keys ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	defer rows.Close()

	var result []model.APIKeySummary
	for rows.Next() {
		var k model.APIKeySummary
		if err := rows.Scan(&k.ID, &k.Name, &k.KeyPrefix, &k.ExpiresAt, &k.Active, &k.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan api key: %w", err)
		}
		result = append(result, k)
	}
	return result, rows.Err()
}

// CreateAPIKey registers a new B2B API key.
func (s *PostgresStore) CreateAPIKey(ctx context.Context, name, keyHash string, expiresAt *time.Time) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO b2b_api_keys (name, key_hash, expires_at, active) VALUES ($1,$2,$3,true) RETURNING id`,
		name, keyHash, expiresAt).Scan(&id)
	return id, err
}

// RevokeAPIKey deactivates a B2B API key.
func (s *PostgresStore) RevokeAPIKey(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE b2b_api_keys SET active = false WHERE id = $1`, id)
	return err
}

// ChangeAdminPassword updates an admin user's password hash.
func (s *PostgresStore) ChangeAdminPassword(ctx context.Context, userID, hash string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET password_hash = $1 WHERE id = $2 AND role = 'admin'`, hash, userID)
	return err
}
