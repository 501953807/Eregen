package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"eregen.dev/api-server/internal/model"

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

// Raw returns a RawDB for dynamic SQL queries.
// SQLite uses ? placeholders directly, no conversion needed.
func (s *SqliteStore) Raw() RawDB {
	return &sqlRaw{db: s.db}
}

// Health checks SQLite connectivity.
func (s *SqliteStore) Health(ctx context.Context) error {
	var val int
	err := s.db.QueryRowContext(ctx, `SELECT 1`).Scan(&val)
	if err != nil {
		return fmt.Errorf("health check: %w", err)
	}
	return nil
}

// ListElderly returns paginated elderly profiles from elderly_profiles table.
func (s *SqliteStore) ListElderly(ctx context.Context, page, pageSize int) ([]model.ElderlyProfile, error) {
	offset := (page - 1) * pageSize
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, user_id, birth_date, avatar_url, health_tiers, created_at, updated_at
		 FROM elderly_profiles ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list elderly: %w", err)
	}
	defer rows.Close()

	var profiles []model.ElderlyProfile
	for rows.Next() {
		var p model.ElderlyProfile
		var birthRaw sql.NullString
		var avatarRaw sql.NullString
		var tiersRaw sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &p.UserID, &birthRaw, &avatarRaw, &tiersRaw, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan elderly: %w", err)
		}
		if birthRaw.Valid {
			if t, err := time.Parse(time.RFC3339, birthRaw.String); err == nil {
				p.BirthDate = &t
			}
		}
		if avatarRaw.Valid {
			p.AvatarURL = &avatarRaw.String
		}
		if tiersRaw.Valid {
			json.Unmarshal([]byte(tiersRaw.String), &p.HealthTiers)
		}
		profiles = append(profiles, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return profiles, nil
}

// ListUsers returns paginated users (legacy interface method).
func (s *SqliteStore) ListUsers(ctx context.Context, page, pageSize int) ([]model.User, int, error) {
	offset := (page - 1) * pageSize
	rows, err := s.db.QueryContext(ctx, `SELECT id, email, phone, password_hash, role, name, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Phone, &u.PasswordHash, &u.Role, &u.Name, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}

	var total int
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
	return users, total, rows.Err()
}

// ListDevicesAdmin returns devices filtered by status (empty means all).
func (s *SqliteStore) ListDevicesAdmin(ctx context.Context, status string) ([]model.DeviceSummary, error) {
	where := "1=1"
	var args []interface{}
	idx := 1

	if status != "" {
		where += " AND status = ?"
		args = append(args, status)
		idx++
	}

	query := fmt.Sprintf(`SELECT id, device_id, device_type, tier, status, last_seen,
	(SELECT u.name FROM users u JOIN devices d ON d.owner_user_id = u.id WHERE d.id = devices.id LIMIT 1) as owner_name,
	COALESCE(json_extract(settings, '$.fw_version'), 'v0.1') as fw_version
	FROM devices WHERE %s ORDER BY last_seen DESC`, where)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	defer rows.Close()

	var devices []model.DeviceSummary
	for rows.Next() {
		var d model.DeviceSummary
		var ownerName, fwVersion string
		if err := rows.Scan(&d.ID, &d.DeviceID, &d.Type, &d.Tier, &d.Status, &d.LastSeen, &ownerName, &fwVersion); err != nil {
			return nil, fmt.Errorf("scan device: %w", err)
		}
		d.OwnerName = ownerName
		d.FirmwareVer = fwVersion
		devices = append(devices, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return devices, nil
}

// GetActiveAlerts returns pending alerts summary.
func (s *SqliteStore) GetActiveAlerts(ctx context.Context) ([]model.AlertSummary, error) {
	query := `SELECT id, elderly_id, alert_type, severity, status, created_at, COALESCE(device_id, '') as device_id
		FROM alerts WHERE status = 'pending' ORDER BY created_at DESC LIMIT 50`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get alerts: %w", err)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return alerts, nil
}

// ValidateToken extracts token from Authorization header or query param.
// For verification mode, any non-empty token passes (placeholder for real auth).
func (s *SqliteStore) ValidateToken(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("empty token")
	}
	// In verification mode, simple token validation — return first user as placeholder
	var userID string
	err := s.db.QueryRowContext(ctx, `SELECT id FROM users WHERE role = 'family' LIMIT 1`).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no family users found")
		}
		return "", fmt.Errorf("validate token: %w", err)
	}
	return userID, nil
}

// ListDailyTasks returns daily tasks for an elderly person on a given date.
func (s *SqliteStore) ListDailyTasks(ctx context.Context, elderlyID string, taskDate string) ([]model.ChronicDailyTask, error) {
	query := `SELECT id, elderly_id, task_type, scheduled_time, completed, completed_at, task_date
		FROM chronic_daily_tasks WHERE elderly_id = ? AND task_date = ?
		ORDER BY scheduled_time ASC`
	rows, err := s.db.QueryContext(ctx, query, elderlyID, taskDate)
	if err != nil {
		return nil, fmt.Errorf("list daily tasks: %w", err)
	}
	defer rows.Close()

	var tasks []model.ChronicDailyTask
	for rows.Next() {
		var t model.ChronicDailyTask
		var completedAtRaw sql.NullString
		if err := rows.Scan(&t.ID, &t.ElderlyID, &t.TaskType, &t.ScheduledTime, &t.Completed, &completedAtRaw, &t.TaskDate); err != nil {
			return nil, fmt.Errorf("scan daily task: %w", err)
		}
		if completedAtRaw.Valid && completedAtRaw.String != "" {
			var parsedAt time.Time
			if parsedAt, err = time.Parse(time.RFC3339, completedAtRaw.String); err == nil {
				t.CompletedAt = &parsedAt
			}
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate daily tasks: %w", err)
	}
	return tasks, nil
}

// UpdateDailyTaskComplete marks a daily task as completed.
func (s *SqliteStore) UpdateDailyTaskComplete(ctx context.Context, taskID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE chronic_daily_tasks SET completed = 1, completed_at = datetime('now') WHERE id = ?`,
		taskID)
	if err != nil {
		return fmt.Errorf("update daily task complete: %w", err)
	}
	return nil
}

// migrate creates tables if they don't exist using schema from admin-api.
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
			phone TEXT,
			open_id TEXT,
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
			metadata TEXT,
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
		`CREATE TABLE IF NOT EXISTS health_records (
			id TEXT PRIMARY KEY,
			elderly_id TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			hr INTEGER,
			spo2 INTEGER,
			steps INTEGER,
			sleep_hours REAL,
			bp_systolic INTEGER,
			bp_diastolic INTEGER
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
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS location_records (
			id TEXT PRIMARY KEY,
			elderly_id TEXT NOT NULL,
			lat REAL NOT NULL,
			lon REAL NOT NULL,
			accuracy REAL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// 血糖检测记录
		`CREATE TABLE IF NOT EXISTS chronic_glucose_records (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			elderly_id TEXT NOT NULL,
			value REAL NOT NULL,
			unit TEXT DEFAULT 'mmol/L',
			test_mode TEXT DEFAULT 'random',
			measurement_time DATETIME NOT NULL,
			detected_at DATETIME DEFAULT (datetime('now')),
			source TEXT DEFAULT 'test_strip',
			quality REAL,
			temperature REAL
		)`,
		// 尿酸检测记录
		`CREATE TABLE IF NOT EXISTS chronic_uric_acid_records (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			elderly_id TEXT NOT NULL,
			value REAL NOT NULL,
			unit TEXT DEFAULT 'μmol/L',
			measurement_time DATETIME NOT NULL,
			detected_at DATETIME DEFAULT (datetime('now')),
			source TEXT DEFAULT 'test_strip'
		)`,
		// 血压记录
		`CREATE TABLE IF NOT EXISTS chronic_bp_records (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			elderly_id TEXT NOT NULL,
			systolic INTEGER NOT NULL,
			diastolic INTEGER NOT NULL,
			pulse INTEGER,
			measurement_time DATETIME NOT NULL,
			detected_at DATETIME DEFAULT (datetime('now'))
		)`,
		// 饮食记录
		`CREATE TABLE IF NOT EXISTS chronic_diet_records (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			elderly_id TEXT NOT NULL,
			meal_type TEXT NOT NULL,
			food_items TEXT NOT NULL,
			total_carbs REAL,
			total_calories REAL,
			recorded_at DATETIME DEFAULT (datetime('now'))
		)`,
		// 运动记录
		`CREATE TABLE IF NOT EXISTS chronic_exercise_records (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			elderly_id TEXT NOT NULL,
			type TEXT NOT NULL,
			duration_min INTEGER,
			calories REAL,
			avg_hr INTEGER,
			max_hr INTEGER,
			recorded_at DATETIME DEFAULT (datetime('now'))
		)`,
		// 每日任务
		`CREATE TABLE IF NOT EXISTS chronic_daily_tasks (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			elderly_id TEXT NOT NULL,
			task_type TEXT NOT NULL,
			scheduled_time TEXT NOT NULL,
			completed INTEGER DEFAULT 0,
			completed_at DATETIME,
			task_date DATE DEFAULT (date('now'))
		)`,
		// 周期报告
		`CREATE TABLE IF NOT EXISTS chronic_health_reports (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			elderly_id TEXT NOT NULL,
			report_type TEXT NOT NULL,
			period_start DATE NOT NULL,
			period_end DATE NOT NULL,
			data_summary TEXT,
			ai_recommendations TEXT,
			generated_at DATETIME DEFAULT (datetime('now'))
		)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w\nSQL: %s", err, migration)
		}
	}
	return nil
}
