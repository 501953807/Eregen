// Package store provides SQLite implementation of the Store interface.
package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"eregen.dev/admin-api/internal/model"
	"time"

	_ "modernc.org/sqlite"
)

// SqliteStore wraps database access for admin operations using SQLite.
type SqliteStore struct {
	db *sql.DB
}

// NewSqlite opens a connection to a SQLite database and runs migrations.
func NewSqlite(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite: %w", err)
	}
	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate sqlite: %w", err)
	}
	return db, nil
}

// NewSqliteStore creates a SqliteStore from an existing *sql.DB.
func NewSqliteStore(db *sql.DB) *SqliteStore {
	return &SqliteStore{db: db}
}

func migrate(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS devices (
			id TEXT PRIMARY KEY,
			device_id TEXT UNIQUE NOT NULL,
			device_type TEXT NOT NULL,
			tier TEXT NOT NULL,
			status TEXT DEFAULT 'offline',
			last_seen DATETIME,
			owner_user_id TEXT,
			settings TEXT DEFAULT '{}',
			ota_url TEXT,
			ota_hash TEXT,
			ota_status TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE,
			role TEXT DEFAULT 'user',
			password_hash TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS alerts (
			id TEXT PRIMARY KEY,
			elderly_id TEXT,
			alert_type TEXT NOT NULL,
			severity TEXT NOT NULL,
			status TEXT DEFAULT 'pending',
			message TEXT,
			device_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			resolved_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS elderly_profiles (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			user_id TEXT,
			birth_date DATE,
			health_tiers TEXT DEFAULT '[]',
			avatar_url TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS elderly_devices (
			id TEXT PRIMARY KEY,
			elderly_id TEXT NOT NULL,
			device_id TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (elderly_id) REFERENCES elderly_profiles(id),
			FOREIGN KEY (device_id) REFERENCES devices(id)
		)`,
		`CREATE TABLE IF NOT EXISTS health_records (
			id TEXT PRIMARY KEY,
			elderly_id TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			hr INTEGER,
			spo2 INTEGER,
			steps INTEGER,
			sleep_hours REAL,
			FOREIGN KEY (elderly_id) REFERENCES elderly_profiles(id)
		)`,
		`CREATE TABLE IF NOT EXISTS medication_rules (
			id TEXT PRIMARY KEY,
			elderly_id TEXT NOT NULL,
			schedule_time TEXT NOT NULL,
			pill_type TEXT DEFAULT 'capsule',
			dose_count INTEGER DEFAULT 1,
			days_of_week TEXT DEFAULT '[]',
			active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (elderly_id) REFERENCES elderly_profiles(id)
		)`,
		`CREATE TABLE IF NOT EXISTS location_history (
			id TEXT PRIMARY KEY,
			elderly_id TEXT NOT NULL,
			lat REAL NOT NULL,
			lon REAL NOT NULL,
			accuracy REAL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (elderly_id) REFERENCES elderly_profiles(id)
		)`,
		`CREATE TABLE IF NOT EXISTS subscriptions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			plan_tier TEXT NOT NULL,
			status TEXT DEFAULT 'active',
			starts_at DATETIME,
			expires_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS firmware_releases (
			id TEXT PRIMARY KEY,
			device_type TEXT NOT NULL,
			tier TEXT NOT NULL,
			version TEXT NOT NULL,
			url TEXT NOT NULL,
			sha256_hash TEXT NOT NULL,
			changelog TEXT DEFAULT '',
			min_app_version TEXT DEFAULT '',
			force_update BOOLEAN DEFAULT 0,
			active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS ota_jobs (
			id TEXT PRIMARY KEY,
			firmware_id TEXT NOT NULL,
			target_devices TEXT DEFAULT '[]',
			progress TEXT DEFAULT '{}',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS system_settings (
			key TEXT PRIMARY KEY,
			setting_value TEXT DEFAULT '{}'
		)`,
		`CREATE TABLE IF NOT EXISTS b2b_api_keys (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			key_hash TEXT NOT NULL,
			key_prefix TEXT,
			expires_at DATETIME,
			active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Medical wristband tables
		`CREATE TABLE IF NOT EXISTS medical_wristband_patients (
			id TEXT PRIMARY KEY,
			admission_no TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			gender TEXT,
			age INTEGER,
			department TEXT,
			bed_number TEXT,
			blood_type TEXT,
			allergies TEXT,
			special_conditions TEXT,
			tag_ids TEXT DEFAULT '[]',
			status TEXT DEFAULT 'admitted',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Regulatory: extend medical_wristband_patients with fence/verify fields
		`ALTER TABLE medical_wristband_patients ADD COLUMN last_verify_at DATETIME`,
		`ALTER TABLE medical_wristband_patients ADD COLUMN verify_gap_hours INTEGER DEFAULT 0`,
		`ALTER TABLE medical_wristband_patients ADD COLUMN fence_status TEXT DEFAULT 'inside'`,
		`ALTER TABLE medical_wristband_patients ADD COLUMN fence_exit_at DATETIME`,
		`ALTER TABLE medical_wristband_patients ADD COLUMN fence_exit_duration_sec INTEGER DEFAULT 0`,
		`CREATE TABLE IF NOT EXISTS medical_wristband_devices (
			id TEXT PRIMARY KEY,
			device_id TEXT UNIQUE NOT NULL,
			firmware_version TEXT DEFAULT '',
			status TEXT DEFAULT 'idle',
			bound_patient_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS medical_bindings (
			id TEXT PRIMARY KEY,
			patient_id TEXT NOT NULL,
			device_id TEXT NOT NULL,
			bound_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			unbound_at DATETIME,
			FOREIGN KEY (patient_id) REFERENCES medical_wristband_patients(id),
			FOREIGN KEY (device_id) REFERENCES medical_wristband_devices(id)
		)`,
		`CREATE TABLE IF NOT EXISTS medical_expenses (
			id TEXT PRIMARY KEY,
			patient_id TEXT NOT NULL,
			item_name TEXT NOT NULL,
			category TEXT,
			amount REAL,
			quantity INTEGER DEFAULT 1,
			unit_price REAL,
			notes TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (patient_id) REFERENCES medical_wristband_patients(id)
		)`,
		`CREATE TABLE IF NOT EXISTS medical_medications (
			id TEXT PRIMARY KEY,
			patient_id TEXT NOT NULL,
			name TEXT NOT NULL,
			dosage TEXT,
			frequency TEXT,
			duration TEXT,
			route TEXT,
			notes TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (patient_id) REFERENCES medical_wristband_patients(id)
		)`,
		`CREATE TABLE IF NOT EXISTS medical_test_results (
			id TEXT PRIMARY KEY,
			patient_id TEXT NOT NULL,
			test_name TEXT NOT NULL,
			result TEXT,
			reference_range TEXT,
			unit TEXT,
			collected_at DATETIME,
			reported_at DATETIME,
			notes TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (patient_id) REFERENCES medical_wristband_patients(id)
		)`,
		`CREATE TABLE IF NOT EXISTS medical_daily_entries (
			id TEXT PRIMARY KEY,
			patient_id TEXT NOT NULL,
			entry_date DATE NOT NULL,
			entry_type TEXT NOT NULL,
			content TEXT NOT NULL,
			nurse_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (patient_id) REFERENCES medical_wristband_patients(id)
		)`,
		`CREATE TABLE IF NOT EXISTS medical_verifications (
			id TEXT PRIMARY KEY,
			device_id TEXT NOT NULL,
			patient_id TEXT,
			verification_type TEXT NOT NULL,
			result TEXT,
			matched BOOLEAN DEFAULT 0,
			verified_by TEXT,
			verified_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			notes TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (device_id) REFERENCES medical_wristband_devices(id),
			FOREIGN KEY (patient_id) REFERENCES medical_wristband_patients(id)
		)`,
		`CREATE TABLE IF NOT EXISTS medical_alert_tag_config (
			id TEXT PRIMARY KEY,
			tag_name TEXT UNIQUE NOT NULL,
			tag_color TEXT DEFAULT '#ff4d4f',
			tag_icon TEXT DEFAULT 'alert',
			enabled BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Regulatory closure tables
		`CREATE TABLE IF NOT EXISTS regulatory_fence_config (
			id TEXT PRIMARY KEY, hospital_id TEXT NOT NULL, hospital_name TEXT NOT NULL,
			center_lat REAL NOT NULL, center_lng REAL NOT NULL, radius_meters INTEGER DEFAULT 200,
			enabled INTEGER DEFAULT 1, created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, UNIQUE(hospital_id)
		)`,
		`CREATE TABLE IF NOT EXISTS regulatory_location_logs (
			id TEXT PRIMARY KEY, patient_id TEXT NOT NULL REFERENCES medical_wristband_patients(id),
			device_id TEXT NOT NULL, lat REAL NOT NULL, lng REAL NOT NULL, accuracy REAL,
			location_source TEXT DEFAULT 'gps' CHECK (location_source IN ('gps','base_station')),
			inside_fence INTEGER DEFAULT 1, recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_rll_source ON regulatory_location_logs(location_source, recorded_at)`,
		`CREATE INDEX IF NOT EXISTS idx_rll_patient ON regulatory_location_logs(patient_id)`,
		`CREATE INDEX IF NOT EXISTS idx_rll_time ON regulatory_location_logs(recorded_at)`,
		`CREATE INDEX IF NOT EXISTS idx_rll_fence ON regulatory_location_logs(inside_fence, recorded_at)`,
		`CREATE TABLE IF NOT EXISTS regulatory_alerts (
			id TEXT PRIMARY KEY, rule_code TEXT NOT NULL, patient_id TEXT REFERENCES medical_wristband_patients(id),
			hospital_id TEXT, department TEXT, severity TEXT CHECK (severity IN ('low','medium','high')),
			alert_type TEXT NOT NULL CHECK (alert_type IN ('no_verify','fence_violation','fake_admission',
				'expense_spike','med_verify_mismatch','frequent_transfer','device_disconnect','post_discharge')),
			detail TEXT NOT NULL, status TEXT DEFAULT 'pending' CHECK (status IN ('pending','acknowledged','resolved','false_positive')),
			triggered_at DATETIME DEFAULT CURRENT_TIMESTAMP, acknowledged_at DATETIME, acknowledged_by TEXT,
			resolved_at DATETIME, resolved_by TEXT, notes TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_ra_rule ON regulatory_alerts(rule_code)`,
		`CREATE INDEX IF NOT EXISTS idx_ra_status ON regulatory_alerts(status)`,
		`CREATE INDEX IF NOT EXISTS idx_ra_patient ON regulatory_alerts(patient_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ra_triggered ON regulatory_alerts(triggered_at)`,
		`CREATE INDEX IF NOT EXISTS idx_ra_dept ON regulatory_alerts(department, status)`,
		`CREATE TABLE IF NOT EXISTS user_department_bindings (
			id TEXT PRIMARY KEY, user_id TEXT NOT NULL REFERENCES users(id),
			department TEXT NOT NULL, bound_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, department)
		)`,
		`CREATE TABLE IF NOT EXISTS regulatory_rule_config (
			rule_code TEXT PRIMARY KEY, rule_name TEXT NOT NULL,
			enabled INTEGER DEFAULT 1, config_json TEXT DEFAULT '{}',
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Community elderly tables
		`CREATE TABLE IF NOT EXISTS community_elders (
			id TEXT PRIMARY KEY, name TEXT NOT NULL, id_card TEXT UNIQUE NOT NULL,
			gender INTEGER CHECK (gender IN (0,1,2)), age INTEGER, address TEXT,
			emergency_contact TEXT, bank_account TEXT, hospital_id TEXT,
			status TEXT DEFAULT 'active' CHECK (status IN ('active','deactivated','deceased')),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deactivated_at DATETIME, deactivated_reason TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_ce_id_card ON community_elders(id_card)`,
		`CREATE INDEX IF NOT EXISTS idx_ce_status ON community_elders(status)`,
		`CREATE INDEX IF NOT EXISTS idx_ce_hospital ON community_elders(hospital_id)`,
		`CREATE TABLE IF NOT EXISTS community_wristband_devices (
			id TEXT PRIMARY KEY, device_id TEXT UNIQUE NOT NULL, firmware_version TEXT,
			mode TEXT DEFAULT 'community' CHECK (mode IN ('hospital','community')),
			status TEXT DEFAULT 'active' CHECK (status IN ('active','inactive','retired')),
			last_seen DATETIME, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS community_elder_bindings (
			id TEXT PRIMARY KEY, elder_id TEXT NOT NULL REFERENCES community_elders(id),
			device_id TEXT NOT NULL REFERENCES community_wristband_devices(id),
			bound_at DATETIME DEFAULT CURRENT_TIMESTAMP, unbound_at DATETIME,
			UNIQUE(elder_id, device_id)
		)`,
		`CREATE TABLE IF NOT EXISTS community_welfare_tag_config (
			id TEXT PRIMARY KEY, tag_code TEXT UNIQUE NOT NULL, tag_name TEXT NOT NULL,
			issuer TEXT NOT NULL, renewal_period_days INTEGER, benefit_amount REAL,
			enabled INTEGER DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS community_elder_welfare (
			id TEXT PRIMARY KEY, elder_id TEXT NOT NULL REFERENCES community_elders(id),
			tag_code TEXT NOT NULL REFERENCES community_welfare_tag_config(tag_code),
			valid_from DATE NOT NULL, valid_to DATE NOT NULL, certified_by TEXT,
			certification_doc TEXT, effective_at DATETIME DEFAULT CURRENT_TIMESTAMP, revoked_at DATETIME,
			UNIQUE(elder_id, tag_code, valid_from, valid_to)
		)`,
		`CREATE TABLE IF NOT EXISTS community_signin_records (
			id TEXT PRIMARY KEY, elder_id TEXT NOT NULL REFERENCES community_elders(id),
			device_id TEXT NOT NULL, hospital_id TEXT NOT NULL, pharmacist_id TEXT,
			signin_time DATETIME DEFAULT CURRENT_TIMESTAMP, period TEXT NOT NULL,
			activated_tags TEXT DEFAULT '[]', is_medical_signin INTEGER DEFAULT 1,
			is_welfare_signin INTEGER DEFAULT 1, notes TEXT,
			UNIQUE(elder_id, device_id, period)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_csr_elder ON community_signin_records(elder_id)`,
		`CREATE INDEX IF NOT EXISTS idx_csr_period ON community_signin_records(period)`,
		`CREATE TABLE IF NOT EXISTS community_pharmacy_logs (
			id TEXT PRIMARY KEY, elder_id TEXT NOT NULL REFERENCES community_elders(id),
			device_id TEXT, hospital_id TEXT NOT NULL, pharmacist_id TEXT,
			dispense_time DATETIME DEFAULT CURRENT_TIMESTAMP, period TEXT NOT NULL,
			items TEXT NOT NULL, total_cost REAL DEFAULT 0, insurance_covered REAL DEFAULT 0,
			self_pay REAL DEFAULT 0, notes TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_cpl_elder ON community_pharmacy_logs(elder_id)`,
		`CREATE INDEX IF NOT EXISTS idx_cpl_period ON community_pharmacy_logs(period)`,
		`CREATE TABLE IF NOT EXISTS community_minzheng_sync (
			id TEXT PRIMARY KEY, source TEXT NOT NULL, filename TEXT,
			imported_count INTEGER DEFAULT 0, matched_count INTEGER DEFAULT 0,
			pending_review_count INTEGER DEFAULT 0, error_count INTEGER DEFAULT 0,
			status TEXT DEFAULT 'processing' CHECK (status IN ('processing','completed','failed')),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP, completed_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS community_batch_payments (
			id TEXT PRIMARY KEY, batch_id TEXT NOT NULL, period TEXT NOT NULL,
			pay_type TEXT NOT NULL, elder_id TEXT NOT NULL REFERENCES community_elders(id),
			amount REAL NOT NULL, bank_account TEXT,
			status TEXT DEFAULT 'pending' CHECK (status IN ('pending','success','failed','retrying')),
			failure_reason TEXT, executed_at DATETIME, created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_cbp_batch ON community_batch_payments(batch_id)`,
		`CREATE INDEX IF NOT EXISTS idx_cbp_status ON community_batch_payments(status)`,
		// Pre-seed: community welfare tag config
		`INSERT OR IGNORE INTO community_welfare_tag_config (tag_code, tag_name, issuer, renewal_period_days, benefit_amount) VALUES
			('orphan', '孤寡老人', '民政局', 365, 0),
			('poverty_level_1', '特困一级', '民政局', 365, 800),
			('poverty_level_2', '特困二级', '民政局', 365, 500),
			('disability_level_1', '残疾一级', '残联', 365, 1200),
			('disability_level_2', '残疾二级', '残联', 365, 800),
			('special_disease', '特病补助', '医保局', 180, 2000),
			('bus_discount', '乘车补贴', '民政局', 365, 360),
			('elder_care_subsidy', '高龄津贴', '民政局', 365, 200),
			('nursing_subsidy', '护理补贴', '民政局', 365, 600)`,
		// Pre-seed: default regulatory rule configs
		`INSERT OR IGNORE INTO regulatory_rule_config (rule_code, rule_name, enabled, config_json) VALUES
			('R01', '挂床住院', 1, '{"max_verify_gap_hours":24,"severity":"high"}'),
			('R02', '电子围栏越界', 1, '{"max_fence_exit_minutes":30,"severity":"high"}'),
			('R03', '虚假入院', 1, '{"bind_duration_hours":48,"severity":"medium"}'),
			('R04', '费用突增', 1, '{"expense_multiplier":3,"severity":"medium"}'),
			('R05', '用药与核验不匹配', 1, '{"severity":"medium"}'),
			('R06', '频繁转科', 1, '{"transfers_per_week":3,"severity":"low"}'),
			('R07', '腕带异常断开', 1, '{"disconnect_hours":2,"severity":"high"}'),
			('R08', '长期不在院', 1, '{"severity":"low"}'),
			('R_C01', '重复领取福利', 1, '{"overlap_days":30,"severity":"high"}'),
			('R_C02', '跨社区医院互认', 1, '{"enabled":1,"severity":"low"}'),
			('R_C03', '冒领嫌疑', 1, '{"id_card_mismatch":1,"severity":"high"}'),
			('R_C04', '福利标签超期未续', 1, '{"grace_days":7,"severity":"medium"}'),
			('R_C05', '签到-发药时间差异常', 1, '{"max_gap_hours":24,"severity":"medium"}'),
			('R_C06', '批量发放失败重试超限', 1, '{"max_retries":3,"severity":"high"}'),
			('R_C07', '僵尸账户', 1, '{"inactive_days":180,"severity":"low"}'),
			('R_C08', '死亡后仍激活', 1, '{"severity":"high"}')`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w\nSQL: %s", err, migration)
		}
	}
	return nil
}

// GetDashboardStats returns the top-level metrics for the admin dashboard.
func (s *SqliteStore) GetDashboardStats(ctx context.Context) (*model.DashboardStats, error) {
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
func (s *SqliteStore) ListDevices(ctx context.Context, page, pageSize int, status, devType, tier string) ([]model.DeviceSummary, error) {
	query := `SELECT id, device_id, device_type, tier, status, COALESCE(last_seen, '0001-01-01'),
		(SELECT u.name FROM users u JOIN devices d ON d.owner_user_id = u.id WHERE d.id = devices.id LIMIT 1),
		COALESCE(json_extract(settings, '$.fw_version'),'v0.1')
		FROM devices WHERE 1=1`
	args := []interface{}{}
	idx := 1
	if status != "" {
		query += " AND status=?"
		args = append(args, status)
		idx++
	}
	if devType != "" {
		query += " AND device_type=?"
		args = append(args, devType)
		idx++
	}
	if tier != "" {
		query += " AND tier=?"
		args = append(args, tier)
		idx++
	}
	query += fmt.Sprintf(" ORDER BY last_seen DESC LIMIT ? OFFSET ?")
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
func (s *SqliteStore) ListUsers(ctx context.Context, page, pageSize int, role string) ([]model.UserSummary, error) {
	query := `SELECT u.id, u.name, u.role, u.created_at,
		(SELECT COUNT(*) FROM devices d WHERE d.owner_user_id = u.id)
		FROM users u WHERE 1=1`
	args := []interface{}{}
	idx := 1
	if role != "" {
		query += fmt.Sprintf(" AND u.role=?")
		args = append(args, role)
		idx++
	}
	query += fmt.Sprintf(" ORDER BY u.created_at DESC LIMIT ? OFFSET ?")
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
func (s *SqliteStore) ListAlerts(ctx context.Context, severity, status string, limit int) ([]model.AlertSummary, error) {
	query := `SELECT a.id, a.elderly_id, a.alert_type, a.severity, a.status, a.created_at,
		COALESCE(d.device_id, '')
		FROM alerts a LEFT JOIN devices d ON a.elderly_id = d.id WHERE 1=1`
	args := []interface{}{}
	idx := 1
	if severity != "" {
		query += fmt.Sprintf(" AND a.severity=?")
		args = append(args, severity)
		idx++
	}
	if status != "" {
		query += fmt.Sprintf(" AND a.status=?")
		args = append(args, status)
		idx++
	}
	query += fmt.Sprintf(" ORDER BY a.created_at DESC LIMIT ?")
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
func (s *SqliteStore) SetUserRole(ctx context.Context, userID, role string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET role = ? WHERE id = ?`, role, userID)
	return err
}

// UpdateDeviceConfig updates device settings JSON column.
func (s *SqliteStore) UpdateDeviceConfig(ctx context.Context, deviceID string, config map[string]interface{}) error {
	settingsJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `UPDATE devices SET settings = json_patch(COALESCE(settings, '{}'), ?) WHERE device_id = ?`, string(settingsJSON), deviceID)
	return err
}

// TriggerOTA schedules an OTA update for a device.
func (s *SqliteStore) TriggerOTA(ctx context.Context, deviceID, firmwareURL, sha256Hash string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE devices SET ota_url = ?, ota_hash = ?, ota_status = 'pending' WHERE device_id = ?`,
		firmwareURL, sha256Hash, deviceID)
	return err
}

// ResolveAlert marks an alert as resolved.
func (s *SqliteStore) ResolveAlert(ctx context.Context, alertID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE alerts SET status = 'resolved', resolved_at = datetime('now') WHERE id = ?`, alertID)
	return err
}

// GetSubscriptionStats returns a per-tier subscription count breakdown.
func (s *SqliteStore) GetSubscriptionStats(ctx context.Context) ([]model.SubscriptionStat, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT plan_tier, COUNT(*),
		       ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM subscriptions), 1)
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
func (s *SqliteStore) GetAlertTrend(ctx context.Context, days int) ([]model.AlertTrendPoint, error) {
	query := `SELECT DATE(a.created_at) AS alert_date,
		   SUM(CASE WHEN d.device_type = 'bracelet' THEN 1 ELSE 0 END) AS bracelet_count,
		   SUM(CASE WHEN d.device_type = 'pillbox' THEN 1 ELSE 0 END) AS pillbox_count
		FROM alerts a LEFT JOIN devices d ON a.elderly_id = d.id
		WHERE a.created_at >= datetime('now', '-' || ? || ' days')
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
func (s *SqliteStore) GetAlertDistribution(ctx context.Context) ([]model.AlertDistributionItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT alert_type, COUNT(*) FROM alerts GROUP BY alert_type ORDER BY COUNT(*) DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("alert distribution: %w", err)
	}
	defer rows.Close()

	colors := map[string]string{
		"sos":              "#ff4d4f",
		"fall":             "#fa541c",
		"med_missed":       "#faad14",
		"device_offline":   "#1890ff",
		"geofence_breach":  "#722ed1",
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
func (s *SqliteStore) GetUserGrowth(ctx context.Context, months int) ([]model.UserGrowthPoint, error) {
	query := `SELECT strftime('%Y-%m', created_at) AS month,
	       COUNT(*) AS new_users
		FROM users GROUP BY strftime('%Y-%m', created_at)
		ORDER BY month DESC LIMIT ?`
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
func (s *SqliteStore) GetDeviceByID(ctx context.Context, id string) (*model.DeviceDetail, error) {
	var d model.DeviceDetail
	err := s.db.QueryRowContext(ctx, `
		SELECT d.id, d.device_id, d.device_type, d.tier, d.status, COALESCE(d.last_seen, '0001-01-01'),
		       u.name, COALESCE(json_extract(d.settings, '$.fw_version'),'v0.1'),
		       d.settings,
		       e.name AS elderly_name
			FROM devices d LEFT JOIN users u ON d.owner_user_id = u.id
			LEFT JOIN elderly_devices ed ON d.id = ed.device_id
			LEFT JOIN elderly_profiles e ON ed.elderly_id = e.id
			WHERE d.id = ?`, id).Scan(
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
func (s *SqliteStore) UnbindDevice(ctx context.Context, deviceID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM elderly_devices WHERE device_id = ?;
		 UPDATE devices SET owner_user_id = NULL WHERE device_id = ?`,
		deviceID, deviceID)
	return err
}

// BatchTriggerOTA schedules OTA updates for multiple devices.
func (s *SqliteStore) BatchTriggerOTA(ctx context.Context, deviceIDs, firmwareURL, sha256Hash []string) error {
	for i, id := range deviceIDs {
		url := firmwareURL[i%len(firmwareURL)]
		hash := sha256Hash[i%len(sha256Hash)]
		if _, err := s.db.ExecContext(ctx,
			`UPDATE devices SET ota_url = ?, ota_hash = ?, ota_status = 'pending' WHERE device_id = ?`,
			url, hash, id); err != nil {
			return fmt.Errorf("batch OTA device %s: %w", id, err)
		}
	}
	return nil
}

// CreateFirmwareVersion inserts a new firmware release record.
func (s *SqliteStore) CreateFirmwareVersion(ctx context.Context, v *model.FirmwareVersion) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO firmware_releases (device_type, tier, version, url, sha256_hash, changelog, min_app_version, force_update, active)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1)`,
		v.DeviceType, v.Tier, v.Version, v.DownloadURL, v.Sha256Hash, v.Changelog, v.MinAppVersion, v.ForceUpdate)
	return err
}

// ListFirmwareVersions returns all firmware versions.
func (s *SqliteStore) ListFirmwareVersions(ctx context.Context) ([]model.FirmwareVersion, error) {
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
func (s *SqliteStore) DeleteFirmwareVersion(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE firmware_releases SET active = 0 WHERE id = ?`, id)
	return err
}

// PushOTAJob records an OTA push job.
func (s *SqliteStore) PushOTAJob(ctx context.Context, firmwareID string, deviceIDs []string) error {
	devicesJSON := "[]"
	if len(deviceIDs) > 0 {
		data, _ := json.Marshal(deviceIDs)
		devicesJSON = string(data)
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO ota_jobs (firmware_id, target_devices, progress) VALUES (?, ?, '{"total":0,"pending":0}')`,
		firmwareID, devicesJSON)
	return err
}

// GetNotificationSettings retrieves system notification config.
func (s *SqliteStore) GetNotificationSettings(ctx context.Context) (map[string]any, error) {
	var value string
	err := s.db.QueryRowContext(ctx, `SELECT COALESCE(setting_value, '{}') FROM system_settings WHERE key = 'notification'`).Scan(&value)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("get notification settings: %w", err)
	}
	var result map[string]any
	if value != "" {
		json.Unmarshal([]byte(value), &result)
	}
	if result == nil {
		result = map[string]any{}
	}
	return result, nil
}

// UpdateNotificationSettings persists notification config.
func (s *SqliteStore) UpdateNotificationSettings(ctx context.Context, data map[string]any) error {
	value, _ := json.Marshal(data)
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO system_settings (key, setting_value) VALUES ('notification', ?)`,
		string(value))
	return err
}

// ListAPIKeys returns registered B2B API keys.
func (s *SqliteStore) ListAPIKeys(ctx context.Context) ([]model.APIKeySummary, error) {
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
func (s *SqliteStore) CreateAPIKey(ctx context.Context, name, keyHash string, expiresAt *time.Time) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO b2b_api_keys (name, key_hash, expires_at, active) VALUES (?, ?, ?, 1) RETURNING id`,
		name, keyHash, expiresAt).Scan(&id)
	return id, err
}

// RevokeAPIKey deactivates a B2B API key.
func (s *SqliteStore) RevokeAPIKey(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE b2b_api_keys SET active = 0 WHERE id = ?`, id)
	return err
}

// ChangeAdminPassword updates an admin user's password hash.
func (s *SqliteStore) ChangeAdminPassword(ctx context.Context, userID, hash string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET password_hash = ? WHERE id = ? AND role = 'admin'`, hash, userID)
	return err
}

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
		var birthRaw, avatarRaw, tiersRaw string
		if err := rows.Scan(&p.ID, &p.Name, &p.UserID, &birthRaw, &avatarRaw, &tiersRaw, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan elderly: %w", err)
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
	var id string
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO elderly_profiles (name, user_id, birth_date, health_tiers, avatar_url, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now')) RETURNING id`,
		name, userID, birthDate, string(tiersJSON), avatarURL).Scan(&id)
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
		if err := rows.Scan(&d.ID, &d.DeviceID, &d.Type, &d.Tier, &d.Status, &d.FirmwareVer, &d.LastSeen); err != nil {
			return nil, fmt.Errorf("scan elderly device: %w", err)
		}
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

// ========== Medical Wristband Methods ==========

// CreatePatient inserts a new patient record.
func (s *SqliteStore) CreatePatient(ctx context.Context, p *model.MedicalPatient) error {
	tagsJSON, _ := json.Marshal(p.TagIDs)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medical_wristband_patients (id, admission_no, name, gender, age, department, bed_number, blood_type, allergies, special_conditions, tag_ids, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		p.ID, p.AdmissionNo, p.Name, p.Gender, p.Age, p.Department, p.BedNumber, p.BloodType, p.Allergies, p.SpecialConditions, string(tagsJSON), p.Status)
	return err
}

// GetPatient returns a patient by ID.
func (s *SqliteStore) GetPatient(ctx context.Context, id string) (*model.MedicalPatient, error) {
	var p model.MedicalPatient
	var tagsRaw string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, admission_no, name, gender, age, department, bed_number, blood_type, allergies, special_conditions, tag_ids, status, created_at, updated_at
		FROM medical_wristband_patients WHERE id = ?`, id).Scan(
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

// ListPatients returns paginated patients.
func (s *SqliteStore) ListPatients(ctx context.Context, page, pageSize int, status string) ([]model.MedicalPatient, error) {
	query := `SELECT id, admission_no, name, gender, age, department, bed_number, status, created_at, updated_at
		FROM medical_wristband_patients WHERE 1=1`
	args := []interface{}{}
	idx := 1
	if status != "" {
		query += fmt.Sprintf(" AND status=?")
		args = append(args, status)
		idx++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT ? OFFSET ?")
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

// UpdatePatient modifies a patient.
func (s *SqliteStore) UpdatePatient(ctx context.Context, p *model.MedicalPatient) error {
	tagsJSON, _ := json.Marshal(p.TagIDs)
	_, err := s.db.ExecContext(ctx,
		`UPDATE medical_wristband_patients SET admission_no=?, name=?, gender=?, age=?, department=?, bed_number=?, blood_type=?, allergies=?, special_conditions=?, tag_ids=?, status=?, updated_at=datetime('now') WHERE id=?`,
		p.AdmissionNo, p.Name, p.Gender, p.Age, p.Department, p.BedNumber, p.BloodType, p.Allergies, p.SpecialConditions, string(tagsJSON), p.Status, p.ID)
	return err
}

// DeletePatient soft-deletes a patient (marks as discharged).
func (s *SqliteStore) DeletePatient(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE medical_wristband_patients SET status='discharged', updated_at=datetime('now') WHERE id=?`, id)
	return err
}

// BindWristband binds a device to a patient.
func (s *SqliteStore) BindWristband(ctx context.Context, patientID, deviceID string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medical_bindings (patient_id, device_id, bound_at) VALUES (?, ?, datetime('now'))
		 ON CONFLICT DO NOTHING`, patientID, deviceID)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`UPDATE medical_wristband_devices SET bound_patient_id=?, status='bound', updated_at=datetime('now') WHERE id=?`, patientID, deviceID)
	return err
}

// UnbindWristband unbinds a device from a patient.
func (s *SqliteStore) UnbindWristband(ctx context.Context, bindingID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE medical_bindings SET unbound_at=datetime('now') WHERE id=? AND unbound_at IS NULL`, bindingID)
	if err != nil {
		return err
	}
	// Also clear device binding
	_, err = s.db.ExecContext(ctx,
		`UPDATE medical_wristband_devices SET bound_patient_id=NULL, status='idle', updated_at=datetime('now') WHERE id IN (SELECT device_id FROM medical_bindings WHERE id=?)`, bindingID)
	return err
}

// ClearWristband clears all data from a wristband device.
func (s *SqliteStore) ClearWristband(ctx context.Context, deviceID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE medical_wristband_devices SET bound_patient_id=NULL, status='cleared', updated_at=datetime('now') WHERE id=?`, deviceID)
	return err
}

// ListWristbands returns all wristband devices with optional pagination/status.
func (s *SqliteStore) ListWristbands(ctx context.Context, page, pageSize int, status string) ([]model.MedicalWristbandDevice, error) {
	query := `SELECT id, device_id, firmware_version, status, bound_patient_id, created_at, updated_at
		FROM medical_wristband_devices WHERE 1=1`
	args := []interface{}{}
	idx := 1
	if status != "" {
		query += fmt.Sprintf(" AND status=?")
		args = append(args, status)
		idx++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT ? OFFSET ?")
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

// GetWristbandFirmware returns firmware version for a device.
func (s *SqliteStore) GetWristbandFirmware(ctx context.Context, deviceID string) (string, error) {
	var fw string
	err := s.db.QueryRowContext(ctx, `SELECT firmware_version FROM medical_wristband_devices WHERE device_id=?`, deviceID).Scan(&fw)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("wristband not found")
		}
		return "", err
	}
	return fw, nil
}

// WriteToWristband pushes data to a wristband device (stub).
func (s *SqliteStore) WriteToWristband(ctx context.Context, deviceID, data string) error {
	// In production, this would trigger MQTT message to device
	_, err := s.db.ExecContext(ctx,
		`UPDATE medical_wristband_devices SET firmware_version=?, updated_at=datetime('now') WHERE device_id=?`,
		data, deviceID)
	return err
}

// CreateExpense inserts a medical expense record.
func (s *SqliteStore) CreateExpense(ctx context.Context, e *model.MedicalExpense) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medical_expenses (id, patient_id, item_name, category, amount, quantity, unit_price, notes, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		e.ID, e.PatientID, e.ItemName, e.Category, e.Amount, e.Quantity, e.UnitPrice, e.Notes)
	return err
}

// ListExpenses returns expenses for a patient.
func (s *SqliteStore) ListExpenses(ctx context.Context, patientID string, page, pageSize int) ([]model.MedicalExpense, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, patient_id, item_name, category, amount, quantity, unit_price, notes, created_at, updated_at
		FROM medical_expenses WHERE patient_id=? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
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

// CreateMedication inserts a medication record.
func (s *SqliteStore) CreateMedication(ctx context.Context, m *model.MedicalMedication) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medical_medications (id, patient_id, name, dosage, frequency, duration, route, notes, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		m.ID, m.PatientID, m.Name, m.Dosage, m.Frequency, m.Duration, m.Route, m.Notes)
	return err
}

// ListMedications returns medications for a patient.
func (s *SqliteStore) ListMedications(ctx context.Context, patientID string) ([]model.MedicalMedication, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, patient_id, name, dosage, frequency, duration, route, notes, created_at, updated_at
		FROM medical_medications WHERE patient_id=? ORDER BY created_at DESC`, patientID)
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

// CreateTestResult inserts a test result record.
func (s *SqliteStore) CreateTestResult(ctx context.Context, r *model.MedicalTestResult) error {
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
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		r.ID, r.PatientID, r.TestName, r.Result, r.ReferenceRange, r.Unit, collectedAt, reportedAt, r.Notes)
	return err
}

// ListTestResults returns test results for a patient.
func (s *SqliteStore) ListTestResults(ctx context.Context, patientID string) ([]model.MedicalTestResult, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, patient_id, test_name, result, reference_range, unit, collected_at, reported_at, notes, created_at, updated_at
		FROM medical_test_results WHERE patient_id=? ORDER BY collected_at DESC`, patientID)
	if err != nil {
		return nil, fmt.Errorf("list test results: %w", err)
	}
	defer rows.Close()

	var items []model.MedicalTestResult
	for rows.Next() {
		var t model.MedicalTestResult
		var collectedAt, reportedAt string
		if err := rows.Scan(&t.ID, &t.PatientID, &t.TestName, &t.Result, &t.ReferenceRange, &t.Unit, &collectedAt, &reportedAt, &t.Notes, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan test result: %w", err)
		}
		if collectedAt != "" {
			if ct, err := time.Parse(time.RFC3339, collectedAt); err == nil {
				t.CollectedAt = &ct
			}
		}
		if reportedAt != "" {
			if rt, err := time.Parse(time.RFC3339, reportedAt); err == nil {
				t.ReportedAt = &rt
			}
		}
		items = append(items, t)
	}
	return items, rows.Err()
}

// CreateDailyEntry inserts a daily medical entry.
func (s *SqliteStore) CreateDailyEntry(ctx context.Context, e *model.MedicalDailyEntry) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medical_daily_entries (id, patient_id, entry_date, entry_type, content, nurse_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		e.ID, e.PatientID, e.EntryDate, e.EntryType, e.Content, e.NurseID)
	return err
}

// ListDailyEntries returns daily entries for a patient.
func (s *SqliteStore) ListDailyEntries(ctx context.Context, patientID string, date string) ([]model.MedicalDailyEntry, error) {
	query := `SELECT id, patient_id, entry_date, entry_type, content, nurse_id, created_at, updated_at
		FROM medical_daily_entries WHERE patient_id=?`
	var rows *sql.Rows
	var err error
	if date != "" {
		rows, err = s.db.QueryContext(ctx, query+` AND entry_date=? ORDER BY created_at DESC`, patientID, date)
	} else {
		rows, err = s.db.QueryContext(ctx, query+` ORDER BY entry_date DESC, created_at DESC`, patientID)
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

// CreateVerification inserts a verification record.
func (s *SqliteStore) CreateVerification(ctx context.Context, v *model.MedicalVerification) error {
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
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))`,
		v.ID, v.DeviceID, patientID, v.VerificationType, v.Result, matchedInt, v.VerifiedBy, verifiedAt, v.Notes)
	return err
}

// ListVerifications returns verification records with pagination.
func (s *SqliteStore) ListVerifications(ctx context.Context, page, pageSize int) ([]model.MedicalVerification, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, device_id, patient_id, verification_type, result, matched, verified_by, verified_at, notes, created_at
		FROM medical_verifications ORDER BY created_at DESC LIMIT ? OFFSET ?`,
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

// UpdateVerificationStatus updates verification status.
func (s *SqliteStore) UpdateVerificationStatus(ctx context.Context, id, status string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE medical_verifications SET status=? WHERE id=?`, status, id)
	return err
}

// GetTodayVerificationStats returns today's verification statistics.
func (s *SqliteStore) GetTodayVerificationStats(ctx context.Context) (*model.MedicalVerificationStats, error) {
	var stats model.MedicalVerificationStats
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*), SUM(CASE WHEN matched=1 THEN 1 ELSE 0 END), SUM(CASE WHEN matched=0 THEN 1 ELSE 0 END)
		FROM medical_verifications WHERE DATE(verified_at)=DATE('now')`).Scan(
		&stats.Total, &stats.Matched, &stats.Unmatched)
	if err != nil {
		return nil, fmt.Errorf("get verification stats: %w", err)
	}
	return &stats, nil
}

// GetMedicalStatsOverview returns overall medical statistics.
func (s *SqliteStore) GetMedicalStatsOverview(ctx context.Context) (*model.MedicalStatsOverview, error) {
	var overview model.MedicalStatsOverview
	err := s.db.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM medical_wristband_patients WHERE status='admitted'),
			(SELECT COUNT(*) FROM medical_wristband_patients WHERE DATE(created_at)=DATE('now')),
			(SELECT COUNT(*) FROM medical_wristband_patients WHERE DATE(updated_at)=DATE('now') AND status='discharged'),
			(SELECT COUNT(*) FROM medical_bindings WHERE unbound_at IS NULL),
			(SELECT COUNT(*) FROM medical_wristband_devices)
	`).Scan(
		&overview.ActivePatients, &overview.TodayAdmitted, &overview.TodayDischarged, &overview.BoundDevices, &overview.TotalDevices)
	if err != nil {
		return nil, fmt.Errorf("get medical stats overview: %w", err)
	}
	return &overview, nil
}

// GetPatientByAdmissionNo returns a patient by admission number.
func (s *SqliteStore) GetPatientByAdmissionNo(ctx context.Context, admissionNo string) (*model.MedicalPatient, error) {
	var p model.MedicalPatient
	var tagsRaw string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, admission_no, name, gender, age, department, bed_number, blood_type, allergies, special_conditions, tag_ids, status, created_at, updated_at
		FROM medical_wristband_patients WHERE admission_no=?`, admissionNo).Scan(
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

// BatchImportPatients imports multiple patients.
func (s *SqliteStore) BatchImportPatients(ctx context.Context, patients []model.MedicalPatient) error {
	for _, p := range patients {
		if err := s.CreatePatient(ctx, &p); err != nil {
			return fmt.Errorf("import patient %s: %w", p.Name, err)
		}
	}
	return nil
}

// GetPatientHistory returns treatment history for a patient.
func (s *SqliteStore) GetPatientHistory(ctx context.Context, patientID string) (*model.MedicalPatientHistory, error) {
	entries, err := s.ListDailyEntries(ctx, patientID, "")
	if err != nil {
		return nil, err
	}
	return &model.MedicalPatientHistory{DailyEntries: entries}, nil
}

// CreateAlertTagConfig creates an alert tag configuration.
func (s *SqliteStore) CreateAlertTagConfig(ctx context.Context, c *model.MedicalAlertTagConfig) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medical_alert_tag_config (id, tag_name, tag_color, tag_icon, enabled) VALUES (?, ?, ?, ?, ?)`,
		c.ID, c.TagName, c.TagColor, c.TagIcon, c.Enabled)
	return err
}

// ListAlertTagConfigs returns all alert tag configurations.
func (s *SqliteStore) ListAlertTagConfigs(ctx context.Context) ([]model.MedicalAlertTagConfig, error) {
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

// Ensure SqliteStore implements Store interface
var _ Store = (*SqliteStore)(nil)
