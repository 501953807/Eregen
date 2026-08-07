// Package store provides SQLite implementation of the Store interface.
package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// parseTimeOrDefault parses a time string in various formats, returning a default value on error.
func parseTimeOrDefault(s string, defaultVal time.Time) time.Time {
	if s == "" {
		return defaultVal
	}
	// Try RFC3339 first
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	// Try common datetime formats
	for _, layout := range []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05Z",
		"2006-01-02T15:04:05Z07:00",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return defaultVal
}

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
			phone TEXT UNIQUE,
			role TEXT DEFAULT 'family' CHECK (role IN ('family', 'elderly', 'institution', 'admin')),
			password_hash TEXT NOT NULL,
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
			last_verify_at DATETIME,
			verify_gap_hours INTEGER DEFAULT 0,
			fence_status TEXT DEFAULT 'inside',
			fence_exit_at DATETIME,
			fence_exit_duration_sec INTEGER DEFAULT 0,
			allergies TEXT,
			special_conditions TEXT,
			tag_ids TEXT DEFAULT '[]',
			status TEXT DEFAULT 'admitted',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
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
		// Clinical workflow tables
		`CREATE TABLE IF NOT EXISTS hospital_admissions (
			id TEXT PRIMARY KEY,
			patient_id TEXT NOT NULL,
			admission_no TEXT NOT NULL,
			bed_no TEXT NOT NULL,
			department TEXT NOT NULL,
			diagnosis TEXT,
			emergency_contact TEXT,
			allergies TEXT,
			admitted_at TEXT NOT NULL DEFAULT (datetime('now')),
			expected_discharge_at TEXT,
			discharged_at TEXT,
			discharge_type TEXT,
			transferred_to TEXT,
			notes TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS ward_rounds (
			id TEXT PRIMARY KEY,
			patient_id TEXT NOT NULL,
			nurse_id TEXT NOT NULL,
			blood_pressure TEXT,
			heart_rate INTEGER,
			spo2 INTEGER,
			temperature REAL,
			weight REAL,
			notes TEXT,
			observations TEXT,
			completed_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
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

// Ensure SqliteStore implements Store interface
var _ Store = (*SqliteStore)(nil)
