package store

import (
	"context"
	"database/sql"
	"fmt"

	"eregen.dev/push/internal/model"
)

// Store provides database access for push-service member lookup.
// Supports both PostgreSQL and SQLite.
type Store struct {
	db *sql.DB
}

// NewStore opens a database connection based on storage type.
func NewStore(dbType, dsn, sqlitePath string) (*Store, error) {
	var db *sql.DB
	var err error
	switch dbType {
	case "postgres":
		db, err = sql.Open("postgres", dsn)
	default:
		db, err = sql.Open("sqlite", sqlitePath)
	}
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	return &Store{db: db}, nil
}

// Close shuts down the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// Member represents a family account that receives push notifications.
type Member struct {
	UserID      string
	ElderlyID   string
	DeviceToken string // FCM token
	OpenID      string // WeChat open_id
	Phone       string // Mobile number
}

// GetFamilyMembersByElderlyID fetches all family accounts linked to an elderly person.
func (s *Store) GetFamilyMembers(ctx context.Context, elderlyID string) ([]Member, error) {
	query := `
		SELECT DISTINCT u.id, u.elderly_id, u.device_token, u.open_id, u.phone
		FROM users u
		JOIN user_elderly_links l ON u.id = l.user_id
		WHERE l.elderly_id = ? AND u.role = 'family'
	`
	rows, err := s.db.QueryContext(ctx, query, elderlyID)
	if err != nil {
		return nil, fmt.Errorf("query family members: %w", err)
	}
	defer rows.Close()

	var members []Member
	for rows.Next() {
		var m Member
		err := rows.Scan(&m.UserID, &m.ElderlyID, &m.DeviceToken, &m.OpenID, &m.Phone)
		if err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		members = append(members, m)
	}
	return members, nil
}

// GetElderlyByDeviceID resolves an elderly profile ID from a device serial (BR-XXXX / PX-XXXX).
// Joins: devices → elderly_profiles via owner_user_id.
func (s *Store) GetElderlyByDeviceID(ctx context.Context, deviceID string) (string, error) {
	var elderlyID string
	err := s.db.QueryRowContext(ctx, `
		SELECT ep.id
		FROM devices d
		JOIN elderly_profiles ep ON ep.user_id = d.owner_user_id
		WHERE d.device_id = ?
		LIMIT 1`, deviceID).Scan(&elderlyID)
	if err != nil {
		return "", fmt.Errorf("resolve elderly from device %s: %w", deviceID, err)
	}
	return elderlyID, nil
}

// CreatePushLog inserts a push notification log entry.
func (s *Store) CreatePushLog(ctx context.Context, log *model.PushLog) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO push_logs (id, alert_id, elderly_id, channel, status, detail, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		log.ID, log.AlertID, log.ElderlyID, log.Channel, log.Status, log.Detail, log.CreatedAt)
	return err
}

// ListPushLogs returns recent push logs with optional filters.
func (s *Store) ListPushLogs(ctx context.Context, alertID, channel string, page, pageSize int) ([]model.PushLog, error) {
	q := `SELECT id, alert_id, elderly_id, channel, status, detail, created_at FROM push_logs WHERE 1=1`
	args := []interface{}{}
	if alertID != "" {
		q += " AND alert_id = ?"
		args = append(args, alertID)
	}
	if channel != "" {
		q += " AND channel = ?"
		args = append(args, channel)
	}
	q += " ORDER BY created_at DESC"
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		q += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)
	}
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list push logs: %w", err)
	}
	defer rows.Close()
	var logs []model.PushLog
	for rows.Next() {
		var l model.PushLog
		if err := rows.Scan(&l.ID, &l.AlertID, &l.ElderlyID, &l.Channel, &l.Status, &l.Detail, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan push log: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
