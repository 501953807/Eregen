package store

import (
	"context"
	"database/sql"
	"fmt"
	"eregen.dev/admin-api/internal/model"

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
		var s model.SubscriptionStat
		if err := rows.Scan(&s.Tier, &s.Count, &s.Pct); err != nil {
			return nil, fmt.Errorf("scan subscription stat: %w", err)
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}
