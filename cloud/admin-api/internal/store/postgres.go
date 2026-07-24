package store

import (
	"context"
	"database/sql"
	"encoding/json"
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

// ========== Medical Wristband Methods ==========

func (s *PostgresStore) CreatePatient(ctx context.Context, p *model.MedicalPatient) error {
	tagsJSON, _ := json.Marshal(p.TagIDs)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medical_wristband_patients (id, admission_no, name, gender, age, department, bed_number, blood_type, allergies, special_conditions, tag_ids, status, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,NOW(),NOW())`,
		p.ID, p.AdmissionNo, p.Name, p.Gender, p.Age, p.Department, p.BedNumber, p.BloodType, p.Allergies, p.SpecialConditions, string(tagsJSON), p.Status)
	return err
}

func (s *PostgresStore) GetPatient(ctx context.Context, id string) (*model.MedicalPatient, error) {
	var p model.MedicalPatient
	var tagsRaw string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, admission_no, name, gender, age, department, bed_number, blood_type, allergies, special_conditions, tag_ids, status, created_at, updated_at
		FROM medical_wristband_patients WHERE id = $1`, id).Scan(
		&p.ID, &p.AdmissionNo, &p.Name, &p.Gender, &p.Age, &p.Department, &p.BedNumber,
		&p.BloodType, &p.Allergies, &p.SpecialConditions, &tagsRaw, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("patient not found")
		}
		return nil, fmt.Errorf("get patient: %w", err)
	}
	json.Unmarshal([]byte(tagsRaw), &p.TagIDs)
	return &p, nil
}

func (s *PostgresStore) ListPatients(ctx context.Context, page, pageSize int, status string) ([]model.MedicalPatient, error) {
	query := `SELECT id, admission_no, name, gender, age, department, bed_number, status, created_at, updated_at
		FROM medical_wristband_patients WHERE 1=1`
	var args []interface{}
	idx := 1
	if status != "" {
		query += fmt.Sprintf(" AND status=$%d", idx)
		args = append(args, status)
		idx++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list patients: %w", err)
	}
	defer rows.Close()

	var patients []model.MedicalPatient
	for rows.Next() {
		var p model.MedicalPatient
		if err := rows.Scan(&p.ID, &p.AdmissionNo, &p.Name, &p.Gender, &p.Age, &p.Department, &p.BedNumber, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan patient: %w", err)
		}
		patients = append(patients, p)
	}
	return patients, rows.Err()
}

func (s *PostgresStore) UpdatePatient(ctx context.Context, p *model.MedicalPatient) error {
	tagsJSON, _ := json.Marshal(p.TagIDs)
	_, err := s.db.ExecContext(ctx,
		`UPDATE medical_wristband_patients SET admission_no=$1, name=$2, gender=$3, age=$4, department=$5, bed_number=$6, blood_type=$7, allergies=$8, special_conditions=$9, tag_ids=$10, status=$11, updated_at=NOW() WHERE id=$12`,
		p.AdmissionNo, p.Name, p.Gender, p.Age, p.Department, p.BedNumber, p.BloodType, p.Allergies, p.SpecialConditions, string(tagsJSON), p.Status, p.ID)
	return err
}

func (s *PostgresStore) DeletePatient(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE medical_wristband_patients SET status='discharged', updated_at=NOW() WHERE id=$1`, id)
	return err
}

func (s *PostgresStore) BindWristband(ctx context.Context, patientID, deviceID string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medical_bindings (patient_id, device_id, bound_at) VALUES ($1,$2,NOW())
		 ON CONFLICT DO NOTHING`, patientID, deviceID)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`UPDATE medical_wristband_devices SET bound_patient_id=$1, status='bound', updated_at=NOW() WHERE id=$2`, patientID, deviceID)
	return err
}

func (s *PostgresStore) UnbindWristband(ctx context.Context, bindingID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE medical_bindings SET unbound_at=NOW() WHERE id=$1 AND unbound_at IS NULL`, bindingID)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`UPDATE medical_wristband_devices SET bound_patient_id=NULL, status='idle', updated_at=NOW() WHERE id IN (SELECT device_id FROM medical_bindings WHERE id=$1)`, bindingID)
	return err
}

func (s *PostgresStore) ClearWristband(ctx context.Context, deviceID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE medical_wristband_devices SET bound_patient_id=NULL, status='cleared', updated_at=NOW() WHERE id=$1`, deviceID)
	return err
}

func (s *PostgresStore) ListWristbands(ctx context.Context, page, pageSize int, status string) ([]model.MedicalWristbandDevice, error) {
	query := `SELECT id, device_id, firmware_version, status, bound_patient_id, created_at, updated_at
		FROM medical_wristband_devices WHERE 1=1`
	var args []interface{}
	idx := 1
	if status != "" {
		query += fmt.Sprintf(" AND status=$%d", idx)
		args = append(args, status)
		idx++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list wristbands: %w", err)
	}
	defer rows.Close()

	var devices []model.MedicalWristbandDevice
	for rows.Next() {
		var d model.MedicalWristbandDevice
		if err := rows.Scan(&d.ID, &d.DeviceID, &d.FirmwareVersion, &d.Status, &d.BoundPatientID, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan wristband: %w", err)
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

func (s *PostgresStore) GetWristbandFirmware(ctx context.Context, deviceID string) (string, error) {
	var fw string
	err := s.db.QueryRowContext(ctx, `SELECT firmware_version FROM medical_wristband_devices WHERE device_id=$1`, deviceID).Scan(&fw)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("wristband not found")
		}
		return "", err
	}
	return fw, nil
}

func (s *PostgresStore) WriteToWristband(ctx context.Context, deviceID, data string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE medical_wristband_devices SET firmware_version=$1, updated_at=NOW() WHERE device_id=$2`,
		data, deviceID)
	return err
}

func (s *PostgresStore) CreateExpense(ctx context.Context, e *model.MedicalExpense) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medical_expenses (id, patient_id, item_name, category, amount, quantity, unit_price, notes, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW(),NOW())`,
		e.ID, e.PatientID, e.ItemName, e.Category, e.Amount, e.Quantity, e.UnitPrice, e.Notes)
	return err
}

func (s *PostgresStore) ListExpenses(ctx context.Context, patientID string, page, pageSize int) ([]model.MedicalExpense, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, patient_id, item_name, category, amount, quantity, unit_price, notes, created_at, updated_at
		FROM medical_expenses WHERE patient_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		patientID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("list expenses: %w", err)
	}
	defer rows.Close()

	var items []model.MedicalExpense
	for rows.Next() {
		var e model.MedicalExpense
		if err := rows.Scan(&e.ID, &e.PatientID, &e.ItemName, &e.Category, &e.Amount, &e.Quantity, &e.UnitPrice, &e.Notes, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan expense: %w", err)
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

func (s *PostgresStore) CreateMedication(ctx context.Context, m *model.MedicalMedication) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medical_medications (id, patient_id, name, dosage, frequency, duration, route, notes, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW(),NOW())`,
		m.ID, m.PatientID, m.Name, m.Dosage, m.Frequency, m.Duration, m.Route, m.Notes)
	return err
}

func (s *PostgresStore) ListMedications(ctx context.Context, patientID string) ([]model.MedicalMedication, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, patient_id, name, dosage, frequency, duration, route, notes, created_at, updated_at
		FROM medical_medications WHERE patient_id=$1 ORDER BY created_at DESC`, patientID)
	if err != nil {
		return nil, fmt.Errorf("list medications: %w", err)
	}
	defer rows.Close()

	var items []model.MedicalMedication
	for rows.Next() {
		var m model.MedicalMedication
		if err := rows.Scan(&m.ID, &m.PatientID, &m.Name, &m.Dosage, &m.Frequency, &m.Duration, &m.Route, &m.Notes, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan medication: %w", err)
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

func (s *PostgresStore) CreateTestResult(ctx context.Context, r *model.MedicalTestResult) error {
	collectedAt := ""
	reportedAt := ""
	if r.CollectedAt != nil {
		collectedAt = r.CollectedAt.Format(time.RFC3339)
	}
	if r.ReportedAt != nil {
		reportedAt = r.ReportedAt.Format(time.RFC3339)
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medical_test_results (id, patient_id, test_name, result, reference_range, unit, collected_at, reported_at, notes, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW(),NOW())`,
		r.ID, r.PatientID, r.TestName, r.Result, r.ReferenceRange, r.Unit, collectedAt, reportedAt, r.Notes)
	return err
}

func (s *PostgresStore) ListTestResults(ctx context.Context, patientID string) ([]model.MedicalTestResult, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, patient_id, test_name, result, reference_range, unit, collected_at, reported_at, notes, created_at, updated_at
		FROM medical_test_results WHERE patient_id=$1 ORDER BY collected_at DESC`, patientID)
	if err != nil {
		return nil, fmt.Errorf("list test results: %w", err)
	}
	defer rows.Close()

	var items []model.MedicalTestResult
	for rows.Next() {
		var t model.MedicalTestResult
		if err := rows.Scan(&t.ID, &t.PatientID, &t.TestName, &t.Result, &t.ReferenceRange, &t.Unit, &t.CollectedAt, &t.ReportedAt, &t.Notes, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan test result: %w", err)
		}
		items = append(items, t)
	}
	return items, rows.Err()
}

func (s *PostgresStore) CreateDailyEntry(ctx context.Context, e *model.MedicalDailyEntry) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medical_daily_entries (id, patient_id, entry_date, entry_type, content, nurse_id, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,NOW(),NOW())`,
		e.ID, e.PatientID, e.EntryDate, e.EntryType, e.Content, e.NurseID)
	return err
}

func (s *PostgresStore) ListDailyEntries(ctx context.Context, patientID string, date string) ([]model.MedicalDailyEntry, error) {
	var rows *sql.Rows
	var err error
	if date != "" {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, patient_id, entry_date, entry_type, content, nurse_id, created_at, updated_at
			FROM medical_daily_entries WHERE patient_id=$1 AND entry_date=$2 ORDER BY created_at DESC`, patientID, date)
	} else {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, patient_id, entry_date, entry_type, content, nurse_id, created_at, updated_at
			FROM medical_daily_entries WHERE patient_id=$1 ORDER BY entry_date DESC, created_at DESC`, patientID)
	}
	if err != nil {
		return nil, fmt.Errorf("list daily entries: %w", err)
	}
	defer rows.Close()

	var items []model.MedicalDailyEntry
	for rows.Next() {
		var e model.MedicalDailyEntry
		if err := rows.Scan(&e.ID, &e.PatientID, &e.EntryDate, &e.EntryType, &e.Content, &e.NurseID, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan daily entry: %w", err)
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

func (s *PostgresStore) CreateVerification(ctx context.Context, v *model.MedicalVerification) error {
	matchedInt := 0
	if v.Matched {
		matchedInt = 1
	}
	verifiedAt := ""
	if v.VerifiedAt != nil {
		verifiedAt = v.VerifiedAt.Format(time.RFC3339)
	}
	patientID := ""
	if v.PatientID != nil {
		patientID = *v.PatientID
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medical_verifications (id, device_id, patient_id, verification_type, result, matched, verified_by, verified_at, notes, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW())`,
		v.ID, v.DeviceID, patientID, v.VerificationType, v.Result, matchedInt, v.VerifiedBy, verifiedAt, v.Notes)
	return err
}

func (s *PostgresStore) ListVerifications(ctx context.Context, page, pageSize int) ([]model.MedicalVerification, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, device_id, patient_id, verification_type, result, matched, verified_by, verified_at, notes, created_at
		FROM medical_verifications ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("list verifications: %w", err)
	}
	defer rows.Close()

	var items []model.MedicalVerification
	for rows.Next() {
		var v model.MedicalVerification
		var matchedInt int
		var patientID string
		if err := rows.Scan(&v.ID, &v.DeviceID, &patientID, &v.VerificationType, &v.Result, &matchedInt, &v.VerifiedBy, &v.VerifiedAt, &v.Notes, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan verification: %w", err)
		}
		v.Matched = matchedInt != 0
		if patientID != "" {
			v.PatientID = &patientID
		}
		items = append(items, v)
	}
	return items, rows.Err()
}

func (s *PostgresStore) UpdateVerificationStatus(ctx context.Context, id, status string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE medical_verifications SET status=$1 WHERE id=$2`, status, id)
	return err
}

func (s *PostgresStore) GetTodayVerificationStats(ctx context.Context) (*model.MedicalVerificationStats, error) {
	var stats model.MedicalVerificationStats
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*), SUM(CASE WHEN matched THEN 1 ELSE 0 END), SUM(CASE WHEN NOT matched THEN 1 ELSE 0 END)
		FROM medical_verifications WHERE DATE(verified_at)=CURRENT_DATE`).Scan(
		&stats.Total, &stats.Matched, &stats.Unmatched)
	if err != nil {
		return nil, fmt.Errorf("get verification stats: %w", err)
	}
	return &stats, nil
}

func (s *PostgresStore) GetMedicalStatsOverview(ctx context.Context) (*model.MedicalStatsOverview, error) {
	var overview model.MedicalStatsOverview
	err := s.db.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM medical_wristband_patients WHERE status='admitted'),
			(SELECT COUNT(*) FROM medical_wristband_patients WHERE DATE(created_at)=CURRENT_DATE),
			(SELECT COUNT(*) FROM medical_wristband_patients WHERE DATE(updated_at)=CURRENT_DATE AND status='discharged'),
			(SELECT COUNT(*) FROM medical_bindings WHERE unbound_at IS NULL),
			(SELECT COUNT(*) FROM medical_wristband_devices)
	`).Scan(
		&overview.ActivePatients, &overview.TodayAdmitted, &overview.TodayDischarged, &overview.BoundDevices, &overview.TotalDevices)
	if err != nil {
		return nil, fmt.Errorf("get medical stats overview: %w", err)
	}
	return &overview, nil
}

func (s *PostgresStore) GetPatientByAdmissionNo(ctx context.Context, admissionNo string) (*model.MedicalPatient, error) {
	var p model.MedicalPatient
	var tagsRaw string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, admission_no, name, gender, age, department, bed_number, blood_type, allergies, special_conditions, tag_ids, status, created_at, updated_at
		FROM medical_wristband_patients WHERE admission_no=$1`, admissionNo).Scan(
		&p.ID, &p.AdmissionNo, &p.Name, &p.Gender, &p.Age, &p.Department, &p.BedNumber,
		&p.BloodType, &p.Allergies, &p.SpecialConditions, &tagsRaw, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("patient not found")
		}
		return nil, err
	}
	json.Unmarshal([]byte(tagsRaw), &p.TagIDs)
	return &p, nil
}

func (s *PostgresStore) BatchImportPatients(ctx context.Context, patients []model.MedicalPatient) error {
	for _, p := range patients {
		if err := s.CreatePatient(ctx, &p); err != nil {
			return fmt.Errorf("import patient %s: %w", p.Name, err)
		}
	}
	return nil
}

func (s *PostgresStore) GetPatientHistory(ctx context.Context, patientID string) (*model.MedicalPatientHistory, error) {
	entries, err := s.ListDailyEntries(ctx, patientID, "")
	if err != nil {
		return nil, err
	}
	return &model.MedicalPatientHistory{DailyEntries: entries}, nil
}

func (s *PostgresStore) CreateAlertTagConfig(ctx context.Context, c *model.MedicalAlertTagConfig) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medical_alert_tag_config (id, tag_name, tag_color, tag_icon, enabled) VALUES ($1,$2,$3,$4,$5)`,
		c.ID, c.TagName, c.TagColor, c.TagIcon, c.Enabled)
	return err
}

func (s *PostgresStore) ListAlertTagConfigs(ctx context.Context) ([]model.MedicalAlertTagConfig, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, tag_name, tag_color, tag_icon, enabled, created_at, updated_at
		FROM medical_alert_tag_config ORDER BY tag_name`)
	if err != nil {
		return nil, fmt.Errorf("list alert tag configs: %w", err)
	}
	defer rows.Close()

	var items []model.MedicalAlertTagConfig
	for rows.Next() {
		var c model.MedicalAlertTagConfig
		if err := rows.Scan(&c.ID, &c.TagName, &c.TagColor, &c.TagIcon, &c.Enabled, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan alert tag config: %w", err)
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

// ========== Regulatory stub implementations (PostgresStore) ==========

func (p *PostgresStore) CreateFenceConfig(ctx context.Context, fc *model.RegulatoryFenceConfig) error {
	now := time.Now().UTC(); fc.CreatedAt = now; fc.UpdatedAt = now
	if fc.ID == "" { fc.ID = fmt.Sprintf("fc_%d", now.UnixNano()) }
	_, err := p.db.ExecContext(ctx, `INSERT INTO regulatory_fence_config (id,hospital_id,hospital_name,center_lat,center_lng,radius_meters,enabled,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, fc.ID, fc.HospitalID, fc.HospitalName, fc.CenterLat, fc.CenterLng, fc.RadiusMeters, fc.Enabled, fc.CreatedAt, fc.UpdatedAt)
	return err
}
func (p *PostgresStore) GetFenceConfig(ctx context.Context, hospitalID string) (*model.RegulatoryFenceConfig, error) {
	var fc model.RegulatoryFenceConfig; var enabled int
	err := p.db.QueryRowContext(ctx, `SELECT id,hospital_id,hospital_name,center_lat,center_lng,radius_meters,enabled,created_at,updated_at FROM regulatory_fence_config WHERE hospital_id=$1`, hospitalID).Scan(&fc.ID, &fc.HospitalID, &fc.HospitalName, &fc.CenterLat, &fc.CenterLng, &fc.RadiusMeters, &enabled, &fc.CreatedAt, &fc.UpdatedAt)
	fc.Enabled = enabled == 1; return &fc, err
}
func (p *PostgresStore) UpdateFenceConfig(ctx context.Context, fc *model.RegulatoryFenceConfig) error {
	fc.UpdatedAt = time.Now().UTC()
	_, err := p.db.ExecContext(ctx, `UPDATE regulatory_fence_config SET hospital_name=$1,center_lat=$2,center_lng=$3,radius_meters=$4,enabled=$5,updated_at=$6 WHERE hospital_id=$7`, fc.HospitalName, fc.CenterLat, fc.CenterLng, fc.RadiusMeters, fc.Enabled, fc.UpdatedAt, fc.HospitalID)
	return err
}
func (p *PostgresStore) ListRegulatoryAlerts(ctx context.Context, ruleCode, level, status, department string, page, pageSize int) ([]model.RegulatoryAlert, error) { return nil, nil }
func (p *PostgresStore) GetRegulatoryAlert(ctx context.Context, alertID string) (*model.RegulatoryAlert, error) { return nil, nil }
func (p *PostgresStore) AcknowledgeAlert(ctx context.Context, alertID, userID string) error {
	_, err := p.db.ExecContext(ctx, `UPDATE regulatory_alerts SET status='acknowledged',acknowledged_at=NOW(),acknowledged_by=$1 WHERE id=$2`, userID, alertID)
	return err
}
func (p *PostgresStore) ResolveRegulatoryAlert(ctx context.Context, alertID, userID, notes string) error {
	_, err := p.db.ExecContext(ctx, `UPDATE regulatory_alerts SET status='resolved',resolved_at=NOW(),resolved_by=$1,notes=$2 WHERE id=$3`, userID, notes, alertID)
	return err
}
func (p *PostgresStore) ListRegulatoryAlertsCountByRule(ctx context.Context, days int) ([]model.RuleAlertCount, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT rule_code,COUNT(*) FROM regulatory_alerts WHERE triggered_at > NOW()-($1||' days') GROUP BY rule_code`, days)
	if err != nil { return nil, err }
	defer rows.Close()
	var result []model.RuleAlertCount
	for rows.Next() { var r model.RuleAlertCount; rows.Scan(&r.RuleCode, &r.Count); result = append(result, r) }
	return result, rows.Err()
}
func (p *PostgresStore) SaveLocationLog(ctx context.Context, log *model.RegulatoryLocationLog) error {
	log.RecordedAt = time.Now().UTC()
	if log.ID == "" { log.ID = fmt.Sprintf("ll_%d", log.RecordedAt.UnixNano()) }
	insideFence := 0; if log.InsideFence { insideFence = 1 }
	_, err := p.db.ExecContext(ctx, `INSERT INTO regulatory_location_logs (id,patient_id,device_id,lat,lng,accuracy,inside_fence,recorded_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, log.ID, log.PatientID, log.DeviceID, log.Lat, log.Lng, log.Accuracy, insideFence, log.RecordedAt)
	return err
}
func (p *PostgresStore) ListLocationLogs(ctx context.Context, patientID string, limit int) ([]model.RegulatoryLocationLog, error) { return nil, nil }
func (p *PostgresStore) GetPatientFenceStatus(ctx context.Context, patientID string) (string, time.Time, int, error) { return "inside", time.Time{}, 0, nil }
func (p *PostgresStore) GetRegulatoryOverview(ctx context.Context, department string) (*model.RegulatoryDashboardOverview, error) { return &model.RegulatoryDashboardOverview{}, nil }
func (p *PostgresStore) ListRegulatoryPatients(ctx context.Context, department string, page, pageSize int) ([]model.RegulatoryPatientRow, error) { return nil, nil }
func (p *PostgresStore) GetRegulatoryAuditTrail(ctx context.Context, patientID string) (*model.RegulatoryAuditTrail, error) { return nil, nil }
func (p *PostgresStore) ListRuleConfigs(ctx context.Context) ([]model.RegulatoryRuleConfig, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT rule_code,rule_name,enabled,config_json,updated_at FROM regulatory_rule_config ORDER BY rule_code`)
	if err != nil { return nil, err }
	defer rows.Close()
	var result []model.RegulatoryRuleConfig
	for rows.Next() {
		var cfg model.RegulatoryRuleConfigDB
		rows.Scan(&cfg.RuleCode, &cfg.RuleName, &cfg.Enabled, &cfg.ConfigJSON, &cfg.UpdatedAt)
		result = append(result, model.RegulatoryRuleConfig{Code: cfg.RuleCode, Name: cfg.RuleName, Enabled: cfg.Enabled})
	}
	return result, rows.Err()
}
func (p *PostgresStore) UpdateRuleConfig(ctx context.Context, ruleCode string, configJSON string) error {
	_, err := p.db.ExecContext(ctx, `UPDATE regulatory_rule_config SET config_json=$1,updated_at=NOW() WHERE rule_code=$2`, configJSON, ruleCode)
	return err
}
func (p *PostgresStore) GetComplianceReport(ctx context.Context, hospitalID, startDate, endDate string) (*model.ComplianceReport, error) { return &model.ComplianceReport{}, nil }
func (p *PostgresStore) CreateDepartmentBinding(ctx context.Context, binding *model.DepartmentBinding) error {
	if binding.ID == "" { binding.ID = fmt.Sprintf("db_%d", time.Now().UnixNano()) }
	_, err := p.db.ExecContext(ctx, `INSERT INTO user_department_bindings (id,user_id,department,bound_at) VALUES ($1,$2,$3,NOW())`, binding.ID, binding.UserID, binding.Department)
	return err
}
func (p *PostgresStore) ListDepartmentBindings(ctx context.Context, userID string) ([]model.DepartmentBinding, error) { return nil, nil }
func (p *PostgresStore) CreateRegulatoryAlert(ctx context.Context, alert *model.RegulatoryAlert) error {
	alert.TriggeredAt = time.Now().UTC()
	if alert.ID == "" { alert.ID = fmt.Sprintf("ra_%d", alert.TriggeredAt.UnixNano()) }
	_, err := p.db.ExecContext(ctx, `INSERT INTO regulatory_alerts (id,rule_code,patient_id,hospital_id,department,severity,alert_type,detail,status,triggered_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'pending',$9)`, alert.ID, alert.RuleCode, alert.PatientID, alert.HospitalID, alert.Department, alert.Severity, alert.AlertType, alert.Detail, alert.TriggeredAt)
	return err
}
func (p *PostgresStore) CountPendingAlertsByRule(ctx context.Context) ([]model.RuleAlertCount, error) { return nil, nil }
func (p *PostgresStore) CountAlertsByDept(ctx context.Context, startDate, endDate string) ([]model.DeptAlertCount, error) { return nil, nil }

// ========== Community stub implementations (PostgresStore) ==========

func (p *PostgresStore) CreateCommunityElder(ctx context.Context, e *model.CommunityElder) error {
	now := time.Now().UTC(); e.CreatedAt = now; e.UpdatedAt = now
	if e.ID == "" { e.ID = fmt.Sprintf("ce_%d", now.UnixNano()) }
	_, err := p.db.ExecContext(ctx, `INSERT INTO community_elders (id,name,id_card,gender,age,address,emergency_contact,bank_account,hospital_id,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, e.ID, e.Name, e.IDCard, e.Gender, e.Age, e.Address, e.EmergencyContact, e.BankAccount, e.HospitalID, e.Status, e.CreatedAt, e.UpdatedAt)
	return err
}
func (p *PostgresStore) GetCommunityElder(ctx context.Context, id string) (*model.CommunityElder, error) {
	var e model.CommunityElder
	err := p.db.QueryRowContext(ctx, `SELECT id,name,id_card,gender,age,address,emergency_contact,bank_account,hospital_id,status,created_at,updated_at,deactivated_at,deactivated_reason FROM community_elders WHERE id=$1`, id).Scan(&e.ID, &e.Name, &e.IDCard, &e.Gender, &e.Age, &e.Address, &e.EmergencyContact, &e.BankAccount, &e.HospitalID, &e.Status, &e.CreatedAt, &e.UpdatedAt, &e.DeactivatedAt, &e.DeactivatedReason)
	return &e, err
}
func (p *PostgresStore) ListCommunityElders(ctx context.Context, page, pageSize int, status string) ([]model.CommunityElder, error) { return nil, nil }
func (p *PostgresStore) UpdateCommunityElder(ctx context.Context, e *model.CommunityElder) error {
	e.UpdatedAt = time.Now().UTC()
	_, err := p.db.ExecContext(ctx, `UPDATE community_elders SET name=$1,id_card=$2,gender=$3,age=$4,address=$5,emergency_contact=$6,bank_account=$7,hospital_id=$8,status=$9,updated_at=$10,deactivated_at=$11,deactivated_reason=$12 WHERE id=$13`, e.Name, e.IDCard, e.Gender, e.Age, e.Address, e.EmergencyContact, e.BankAccount, e.HospitalID, e.Status, e.UpdatedAt, e.DeactivatedAt, e.DeactivatedReason, e.ID)
	return err
}
func (p *PostgresStore) DeleteCommunityElder(ctx context.Context, id string) error {
	_, err := p.db.ExecContext(ctx, `UPDATE community_elders SET status='deactivated',deactivated_at=NOW(),deactivated_reason='deleted' WHERE id=$1`, id)
	return err
}
func (p *PostgresStore) BulkUpsertCommunityElders(ctx context.Context, elders []model.CommunityElder) error { return nil }
func (p *PostgresStore) GetCommunityElderStats(ctx context.Context) (*model.CommunityElderStats, error) { return &model.CommunityElderStats{}, nil }
func (p *PostgresStore) CreateCommunityDevice(ctx context.Context, d *model.CommunityWristbandDevice) error {
	now := time.Now().UTC(); d.CreatedAt = now; d.UpdatedAt = now
	if d.ID == "" { d.ID = fmt.Sprintf("cd_%d", now.UnixNano()) }
	_, err := p.db.ExecContext(ctx, `INSERT INTO community_wristband_devices (id,device_id,firmware_version,mode,status,last_seen,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, d.ID, d.DeviceID, d.FirmwareVersion, d.Mode, d.Status, d.LastSeen, d.CreatedAt, d.UpdatedAt)
	return err
}
func (p *PostgresStore) GetCommunityDevice(ctx context.Context, deviceID string) (*model.CommunityWristbandDevice, error) {
	var d model.CommunityWristbandDevice
	err := p.db.QueryRowContext(ctx, `SELECT id,device_id,firmware_version,mode,status,last_seen,created_at,updated_at FROM community_wristband_devices WHERE device_id=$1`, deviceID).Scan(&d.ID, &d.DeviceID, &d.FirmwareVersion, &d.Mode, &d.Status, &d.LastSeen, &d.CreatedAt, &d.UpdatedAt)
	return &d, err
}
func (p *PostgresStore) ListCommunityDevices(ctx context.Context, page, pageSize int, status string) ([]model.CommunityWristbandDevice, error) { return nil, nil }
func (p *PostgresStore) UpdateCommunityDevice(ctx context.Context, d *model.CommunityWristbandDevice) error {
	d.UpdatedAt = time.Now().UTC()
	_, err := p.db.ExecContext(ctx, `UPDATE community_wristband_devices SET firmware_version=$1,status=$2,last_seen=$3,updated_at=$4 WHERE id=$5`, d.FirmwareVersion, d.Status, d.LastSeen, d.UpdatedAt, d.ID)
	return err
}
func (p *PostgresStore) BindCommunityElderDevice(ctx context.Context, elderID, deviceID string) error {
	id := fmt.Sprintf("cb_%d", time.Now().UnixNano())
	_, err := p.db.ExecContext(ctx, `INSERT INTO community_elder_bindings (id,elder_id,device_id,bound_at) VALUES ($1,$2,$3,NOW())`, id, elderID, deviceID)
	return err
}
func (p *PostgresStore) UnbindCommunityElderDevice(ctx context.Context, bindingID string) error {
	_, err := p.db.ExecContext(ctx, `UPDATE community_elder_bindings SET unbound_at=NOW() WHERE id=$1`, bindingID)
	return err
}
func (p *PostgresStore) CreateWelfareTagConfig(ctx context.Context, c *model.CommunityWelfareTagConfig) error {
	now := time.Now().UTC(); c.CreatedAt = now; c.UpdatedAt = now
	if c.ID == "" { c.ID = fmt.Sprintf("wtc_%d", now.UnixNano()) }
	_, err := p.db.ExecContext(ctx, `INSERT INTO community_welfare_tag_config (id,tag_code,tag_name,issuer,renewal_period_days,benefit_amount,enabled) VALUES ($1,$2,$3,$4,$5,$6,$7)`, c.ID, c.TagCode, c.TagName, c.Issuer, c.RenewalPeriodDays, c.BenefitAmount, c.Enabled)
	return err
}
func (p *PostgresStore) UpdateWelfareTagConfig(ctx context.Context, c *model.CommunityWelfareTagConfig) error {
	c.UpdatedAt = time.Now().UTC()
	_, err := p.db.ExecContext(ctx, `UPDATE community_welfare_tag_config SET tag_name=$1,issuer=$2,renewal_period_days=$3,benefit_amount=$4,enabled=$5,updated_at=$6 WHERE tag_code=$7`, c.TagName, c.Issuer, c.RenewalPeriodDays, c.BenefitAmount, c.Enabled, c.UpdatedAt, c.TagCode)
	return err
}
func (p *PostgresStore) ListWelfareTagConfigs(ctx context.Context) ([]model.CommunityWelfareTagConfig, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT id,tag_code,tag_name,issuer,renewal_period_days,benefit_amount,enabled,created_at,updated_at FROM community_welfare_tag_config ORDER BY tag_code`)
	if err != nil { return nil, err }
	defer rows.Close()
	var items []model.CommunityWelfareTagConfig
	for rows.Next() {
		var c model.CommunityWelfareTagConfig
		rows.Scan(&c.ID, &c.TagCode, &c.TagName, &c.Issuer, &c.RenewalPeriodDays, &c.BenefitAmount, &c.Enabled, &c.CreatedAt, &c.UpdatedAt)
		items = append(items, c)
	}
	return items, rows.Err()
}
func (p *PostgresStore) GetWelfareTagConfig(ctx context.Context, tagCode string) (*model.CommunityWelfareTagConfig, error) {
	var c model.CommunityWelfareTagConfig
	err := p.db.QueryRowContext(ctx, `SELECT id,tag_code,tag_name,issuer,renewal_period_days,benefit_amount,enabled,created_at,updated_at FROM community_welfare_tag_config WHERE tag_code=$1`, tagCode).Scan(&c.ID, &c.TagCode, &c.TagName, &c.Issuer, &c.RenewalPeriodDays, &c.BenefitAmount, &c.Enabled, &c.CreatedAt, &c.UpdatedAt)
	return &c, err
}
func (p *PostgresStore) AssignWelfareTag(ctx context.Context, welfare *model.CommunityElderWelfare) error {
	now := time.Now().UTC(); welfare.EffectiveAt = now
	if welfare.ID == "" { welfare.ID = fmt.Sprintf("ewf_%d", now.UnixNano()) }
	_, err := p.db.ExecContext(ctx, `INSERT INTO community_elder_welfare (id,elder_id,tag_code,valid_from,valid_to,certified_by,certification_doc,effective_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, welfare.ID, welfare.ElderID, welfare.TagCode, welfare.ValidFrom, welfare.ValidTo, welfare.CertifiedBy, welfare.CertificationDoc, welfare.EffectiveAt)
	return err
}
func (p *PostgresStore) RevokeWelfareTag(ctx context.Context, elderID, tagCode string) error {
	_, err := p.db.ExecContext(ctx, `UPDATE community_elder_welfare SET revoked_at=NOW() WHERE elder_id=$1 AND tag_code=$2 AND revoked_at IS NULL`, elderID, tagCode)
	return err
}
func (p *PostgresStore) ListElderWelfareTags(ctx context.Context, elderID string) ([]model.CommunityElderWelfare, error) { return nil, nil }
func (p *PostgresStore) CreateSigninRecord(ctx context.Context, sRec *model.CommunitySigninRecord) error {
	sRec.SigninTime = time.Now().UTC()
	if sRec.ID == "" { sRec.ID = fmt.Sprintf("sr_%d", sRec.SigninTime.UnixNano()) }
	_, err := p.db.ExecContext(ctx, `INSERT INTO community_signin_records (id,elder_id,device_id,hospital_id,pharmacist_id,signin_time,period,activated_tags,is_medical_signin,is_welfare_signin,notes) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, sRec.ID, sRec.ElderID, sRec.DeviceID, sRec.HospitalID, sRec.PharmacistID, sRec.SigninTime, sRec.Period, sRec.ActivatedTags, sRec.IsMedicalSignin, sRec.IsWelfareSignin, sRec.Notes)
	return err
}
func (p *PostgresStore) ListSigninRecords(ctx context.Context, elderID, period, hospitalID string, page, pageSize int) ([]model.CommunitySigninRecord, error) { return nil, nil }
func (p *PostgresStore) GetSigninSummary(ctx context.Context, elderID, period string) (*model.CommunitySigninRecord, error) { return nil, nil }
func (p *PostgresStore) CreatePharmacyLog(ctx context.Context, pLog *model.CommunityPharmacyLog) error {
	pLog.DispenseTime = time.Now().UTC()
	if pLog.ID == "" { pLog.ID = fmt.Sprintf("pl_%d", pLog.DispenseTime.UnixNano()) }
	_, err := p.db.ExecContext(ctx, `INSERT INTO community_pharmacy_logs (id,elder_id,device_id,hospital_id,pharmacist_id,dispense_time,period,items,total_cost,insurance_covered,self_pay,notes) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, pLog.ID, pLog.ElderID, pLog.DeviceID, pLog.HospitalID, pLog.PharmacistID, pLog.DispenseTime, pLog.Period, pLog.Items, pLog.TotalCost, pLog.InsuranceCovered, pLog.SelfPay, pLog.Notes)
	return err
}
func (p *PostgresStore) ListPharmacyLogs(ctx context.Context, elderID, period string, page, pageSize int) ([]model.CommunityPharmacyLog, error) { return nil, nil }
func (p *PostgresStore) CreateMinzhengSync(ctx context.Context, m *model.CommunityMinzhengSync) error {
	now := time.Now().UTC(); m.CreatedAt = now
	if m.ID == "" { m.ID = fmt.Sprintf("ms_%d", now.UnixNano()) }
	_, err := p.db.ExecContext(ctx, `INSERT INTO community_minzheng_sync (id,source,filename,imported_count,matched_count,pending_review_count,error_count,status,created_at,completed_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, m.ID, m.Source, m.Filename, m.ImportedCount, m.MatchedCount, m.PendingReviewCount, m.ErrorCount, m.Status, m.CreatedAt, m.CompletedAt)
	return err
}
func (p *PostgresStore) ListMinzhengSync(ctx context.Context, page, pageSize int) ([]model.CommunityMinzhengSync, error) { return nil, nil }
func (p *PostgresStore) GetLatestMinzhengSync(ctx context.Context) (*model.CommunityMinzhengSync, error) { return nil, nil }
func (p *PostgresStore) CreateBatchPayment(ctx context.Context, pmt *model.CommunityBatchPayment) error {
	now := time.Now().UTC(); pmt.CreatedAt = now
	if pmt.ID == "" { pmt.ID = fmt.Sprintf("bp_%d", now.UnixNano()) }
	_, err := p.db.ExecContext(ctx, `INSERT INTO community_batch_payments (id,batch_id,period,pay_type,elder_id,amount,bank_account,status,failure_reason,executed_at,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, pmt.ID, pmt.BatchID, pmt.Period, pmt.PayType, pmt.ElderID, pmt.Amount, pmt.BankAccount, pmt.Status, pmt.FailureReason, pmt.ExecutedAt, pmt.CreatedAt)
	return err
}
func (p *PostgresStore) BulkCreateBatchPayments(ctx context.Context, payments []model.CommunityBatchPayment) error { return nil }
func (p *PostgresStore) ListBatchPayments(ctx context.Context, batchID string, page, pageSize int) ([]model.CommunityBatchPayment, error) { return nil, nil }
func (p *PostgresStore) UpdateBatchPaymentStatus(ctx context.Context, id, status string, failureReason string) error {
	_, err := p.db.ExecContext(ctx, `UPDATE community_batch_payments SET status=$1,failure_reason=$2,executed_at=NOW() WHERE id=$3`, status, failureReason, id)
	return err
}
func (p *PostgresStore) CountPendingPayments(ctx context.Context) (int64, error) {
	var count int64
	err := p.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM community_batch_payments WHERE status='pending'`).Scan(&count)
	return count, err
}
