// © 2026 Eregen (颐贞). All rights reserved.

// Package store provides PostgreSQL persistence for device data.
package store

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store wraps pgx pool access to the PostgreSQL database.
type Store struct {
	pool *pgxpool.Pool
}

// New creates a new Store connected to the given DSN.
func New(ctx context.Context, dsn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	s := &Store{pool: pool}
	if err := s.ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres ping: %w", err)
	}
	return s, nil
}

func (s *Store) ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// Close releases the database connection pool.
func (s *Store) Close() { s.pool.Close() }

// DeviceExists checks whether a device ID is registered in the system.
func (s *Store) DeviceExists(ctx context.Context, deviceID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM devices WHERE device_id = $1)", deviceID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check device %s: %w", deviceID, err)
	}
	return exists, nil
}

// InsertHealthRecord stores a health reading into the health_data table.
func (s *Store) InsertHealthRecord(ctx context.Context, deviceID string, hr, spo2, steps, sleep int, ts int64) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO health_data (dev_id, heart_rate, spo2, steps, sleep_minutes, recorded_at)
		 VALUES ($1, $2, $3, $4, $5, to_timestamp($6))`,
		deviceID, hr, spo2, steps, sleep, ts,
	)
	if err != nil {
		return fmt.Errorf("insert health: %w", err)
	}
	return nil
}

// InsertMedStatusRecord stores a pillbox medication status event.
func (s *Store) InsertMedStatusRecord(ctx context.Context, deviceID string, compartment int, taken bool, ts int64) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO med_status (dev_id, compartment, taken, recorded_at)
		 VALUES ($1, $2, $3, to_timestamp($4))`,
		deviceID, compartment, taken, ts,
	)
	if err != nil {
		return fmt.Errorf("insert med_status: %w", err)
	}
	return nil
}

// InsertLocationRecord stores a GPS location update.
func (s *Store) InsertLocationRecord(ctx context.Context, deviceID string, lat, lon float64, accuracy int, ts int64) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO location_data (dev_id, latitude, longitude, accuracy, recorded_at)
		 VALUES ($1, $2, $3, $4, to_timestamp($5))`,
		deviceID, lat, lon, accuracy, ts,
	)
	if err != nil {
		return fmt.Errorf("insert location: %w", err)
	}
	return nil
}

// RecentHeartbeat returns the last heartbeat timestamp for a device.
func (s *Store) RecentHeartbeat(ctx context.Context, deviceID string) (time.Time, error) {
	var t time.Time
	err := s.pool.QueryRow(ctx,
		"SELECT MAX(recorded_at) FROM heartbeat_data WHERE dev_id = $1", deviceID,
	).Scan(&t)
	if err != nil {
		return time.Time{}, fmt.Errorf("query heartbeat: %w", err)
	}
	return t, nil
}

// RegisterDeviceAuto creates a pending device record if it does not exist yet.
// Returns true+nil when a new record was created, false+nil when it already existed.
func (s *Store) RegisterDeviceAuto(ctx context.Context, deviceID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM devices WHERE device_id = $1)", deviceID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check device %s: %w", deviceID, err)
	}
	if exists {
		return false, nil
	}

	q := `INSERT INTO devices (device_id, device_type, tier, status, owner_user_id, settings, created_at, updated_at)
		  VALUES ($1, $2, $3, 'pending', NULL, '{}', now(), now())`
	deviceType := "bracelet"
	tier := "starter"
	if len(deviceID) >= 3 && deviceID[:2] == "PX" {
		deviceType = "pillbox"
	}
	_, err = s.pool.Exec(ctx, q, deviceID, deviceType, tier)
	if err != nil {
		return false, fmt.Errorf("register device %s: %w", deviceID, err)
	}
	log.Printf("AUTO-REGISTERED device %s (type=%s, tier=%s)", deviceID, deviceType, tier)
	return true, nil
}
