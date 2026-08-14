// © 2026 Eregen (颐贞). All rights reserved.

// Package store provides database persistence for device data.
// Supports both PostgreSQL and SQLite backends.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

// Store wraps database access (PostgreSQL or SQLite).
type Store struct {
	db     *sql.DB
	isPostgres bool
}

// NewPostgres creates a Store connected to PostgreSQL.
func NewPostgres(ctx context.Context, dsn string) (*Store, error) {
	pool, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	if err := pool.PingContext(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres ping: %w", err)
	}
	return &Store{db: pool, isPostgres: true}, nil
}

// NewSQLite opens a SQLite database and runs migrations.
func NewSQLite(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("sqlite ping: %w", err)
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite migrate: %w", err)
	}
	return &Store{db: db, isPostgres: false}, nil
}

// Close releases the database connection.
func (s *Store) Close() { s.db.Close() }

// IsPostgres returns true if using PostgreSQL backend.
func (s *Store) IsPostgres() bool {
	return s.isPostgres
}

// DeviceExists checks whether a device ID is registered in the system.
func (s *Store) DeviceExists(ctx context.Context, deviceID string) (bool, error) {
	var exists bool
	if s.isPostgres {
		err := s.db.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM devices WHERE device_id = $1)", deviceID,
		).Scan(&exists)
		if err != nil {
			return false, fmt.Errorf("check device %s: %w", deviceID, err)
		}
	} else {
		var existsInt int
		err := s.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM devices WHERE device_id = ?", deviceID,
		).Scan(&existsInt)
		if err != nil {
			return false, fmt.Errorf("check device %s: %w", deviceID, err)
		}
		exists = existsInt > 0
	}
	return exists, nil
}

// InsertHealthRecord stores a health reading into the health_data table.
func (s *Store) InsertHealthRecord(ctx context.Context, deviceID string, hr, spo2, steps, sleep int, ts int64) error {
	if s.isPostgres {
		_, err := s.db.ExecContext(ctx,
			`INSERT INTO health_data (dev_id, heart_rate, spo2, steps, sleep_minutes, recorded_at)
			 VALUES ($1, $2, $3, $4, $5, to_timestamp($6))`,
			deviceID, hr, spo2, steps, sleep, ts,
		)
		if err != nil {
			return fmt.Errorf("insert health: %w", err)
		}
	} else {
		_, err := s.db.ExecContext(ctx,
			`INSERT INTO health_data (dev_id, heart_rate, spo2, steps, sleep_minutes, recorded_at)
			 VALUES (?, ?, ?, ?, ?, datetime(?))`,
			deviceID, hr, spo2, steps, sleep, time.Unix(ts, 0).Format("2006-01-02 15:04:05"),
		)
		if err != nil {
			return fmt.Errorf("insert health: %w", err)
		}
	}
	return nil
}

// InsertMedStatusRecord stores a pillbox medication status event.
func (s *Store) InsertMedStatusRecord(ctx context.Context, deviceID string, compartment int, taken bool, ts int64) error {
	if s.isPostgres {
		_, err := s.db.ExecContext(ctx,
			`INSERT INTO med_status (dev_id, compartment, taken, recorded_at)
			 VALUES ($1, $2, $3, to_timestamp($4))`,
			deviceID, compartment, taken, ts,
		)
		if err != nil {
			return fmt.Errorf("insert med_status: %w", err)
		}
	} else {
		takenStr := "false"
		if taken {
			takenStr = "true"
		}
		_, err := s.db.ExecContext(ctx,
			`INSERT INTO med_status (dev_id, compartment, taken, recorded_at)
			 VALUES (?, ?, ?, datetime(?))`,
			deviceID, compartment, takenStr, time.Unix(ts, 0).Format("2006-01-02 15:04:05"),
		)
		if err != nil {
			return fmt.Errorf("insert med_status: %w", err)
		}
	}
	return nil
}

// InsertLocationRecord stores a GPS location update.
func (s *Store) InsertLocationRecord(ctx context.Context, deviceID string, lat, lon float64, accuracy int, ts int64) error {
	if s.isPostgres {
		_, err := s.db.ExecContext(ctx,
			`INSERT INTO location_data (dev_id, latitude, longitude, accuracy, recorded_at)
			 VALUES ($1, $2, $3, $4, to_timestamp($5))`,
			deviceID, lat, lon, accuracy, ts,
		)
		if err != nil {
			return fmt.Errorf("insert location: %w", err)
		}
	} else {
		_, err := s.db.ExecContext(ctx,
			`INSERT INTO location_data (dev_id, latitude, longitude, accuracy, recorded_at)
			 VALUES (?, ?, ?, ?, datetime(?))`,
			deviceID, lat, lon, accuracy, time.Unix(ts, 0).Format("2006-01-02 15:04:05"),
		)
		if err != nil {
			return fmt.Errorf("insert location: %w", err)
		}
	}
	return nil
}

// RecentHeartbeat returns the last heartbeat timestamp for a device.
func (s *Store) RecentHeartbeat(ctx context.Context, deviceID string) (time.Time, error) {
	var ts int64
	var err error
	if s.isPostgres {
		err = s.db.QueryRowContext(ctx,
			"SELECT MAX(EXTRACT(EPOCH FROM recorded_at)) FROM heartbeat_data WHERE dev_id = $1", deviceID,
		).Scan(&ts)
	} else {
		var tsStr string
		err = s.db.QueryRowContext(ctx,
			"SELECT MAX(recorded_at) FROM heartbeat_data WHERE dev_id = ?", deviceID,
		).Scan(&tsStr)
		if err == nil && tsStr != "" {
			t, parseErr := time.Parse("2006-01-02 15:04:05", tsStr)
			if parseErr == nil {
				return t, nil
			}
		}
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("query heartbeat: %w", err)
	}
	return time.Unix(ts, 0), nil
}

// RegisterDeviceAuto creates a pending device record if it does not exist yet.
// Returns true+nil when a new record was created, false+nil when it already existed.
func (s *Store) RegisterDeviceAuto(ctx context.Context, deviceID string) (bool, error) {
	var exists bool
	if s.isPostgres {
		err := s.db.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM devices WHERE device_id = $1)", deviceID,
		).Scan(&exists)
		if err != nil {
			return false, fmt.Errorf("check device %s: %w", deviceID, err)
		}
	} else {
		var existsInt int
		err := s.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM devices WHERE device_id = ?", deviceID,
		).Scan(&existsInt)
		if err != nil {
			return false, fmt.Errorf("check device %s: %w", deviceID, err)
		}
		exists = existsInt > 0
	}
	if exists {
		return false, nil
	}

	deviceType := "bracelet"
	tier := "starter"
	if len(deviceID) >= 3 && deviceID[:2] == "PX" {
		deviceType = "pillbox"
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	if s.isPostgres {
		_, err := s.db.ExecContext(ctx,
			`INSERT INTO devices (device_id, device_type, tier, status, owner_user_id, settings, created_at, updated_at)
			  VALUES ($1, $2, $3, 'pending', NULL, '{}', now(), now())`,
			deviceID, deviceType, tier,
		)
		if err != nil {
			return false, fmt.Errorf("register device %s: %w", deviceID, err)
		}
	} else {
		_, err := s.db.ExecContext(ctx,
			`INSERT INTO devices (device_id, device_type, tier, status, owner_user_id, settings, created_at, updated_at)
			  VALUES (?, ?, ?, 'pending', NULL, '{}', ?, ?)`,
			deviceID, deviceType, tier, now, now,
		)
		if err != nil {
			return false, fmt.Errorf("register device %s: %w", deviceID, err)
		}
	}
	fmt.Fprintf(os.Stderr, "AUTO-REGISTERED device %s (type=%s, tier=%s)\n", deviceID, deviceType, tier)
	return true, nil
}

// migrate creates tables if they don't exist using SQLite-compatible SQL.
func migrate(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS devices (
			device_id TEXT PRIMARY KEY,
			device_type TEXT NOT NULL,
			tier TEXT NOT NULL,
			status TEXT DEFAULT 'offline',
			last_seen TEXT,
			owner_user_id TEXT,
			settings TEXT DEFAULT '{}',
			ota_url TEXT,
			ota_hash TEXT,
			ota_status TEXT DEFAULT '',
			created_at TEXT DEFAULT (datetime('now')),
			updated_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS heartbeat_data (
			dev_id TEXT NOT NULL,
			recorded_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS health_data (
			dev_id TEXT NOT NULL,
			heart_rate INTEGER,
			spo2 INTEGER,
			steps INTEGER,
			sleep_minutes INTEGER,
			recorded_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS location_data (
			dev_id TEXT NOT NULL,
			latitude REAL,
			longitude REAL,
			accuracy INTEGER,
			recorded_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS med_status (
			dev_id TEXT NOT NULL,
			compartment INTEGER,
			taken BOOLEAN,
			recorded_at TEXT NOT NULL
		)`,
	}
	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w\nSQL: %s", err, migration)
		}
	}
	return nil
}
