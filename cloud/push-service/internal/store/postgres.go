package store

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"eregen.dev/push/internal/model"
)

// Member represents a family account that receives push notifications.
type Member struct {
	UserID      string
	ElderlyID   string
	DeviceToken string // FCM token
	OpenID      string // WeChat open_id
	Phone       string // Mobile number
}

// Postgres provides database access for push-service member lookup.
type Postgres struct {
	db *sql.DB
}

// NewPostgres creates a new Postgres store.
func NewPostgres(db *sql.DB) *Postgres {
	return &Postgres{db: db}
}

// GetFamilyMembersByElderlyID fetches all family accounts linked to an elderly person.
func (p *Postgres) GetFamilyMembers(ctx context.Context, elderlyID string) ([]Member, error) {
	query := `
		SELECT DISTINCT u.id, u.elderly_id, u.device_token, u.open_id, u.phone
		FROM users u
		JOIN user_elderly_links l ON u.id = l.user_id
		WHERE l.elderly_id = $1 AND u.role = 'family'
	`
	rows, err := p.db.QueryContext(ctx, query, elderlyID)
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
func (p *Postgres) GetElderlyByDeviceID(ctx context.Context, deviceID string) (string, error) {
	var elderlyID string
	err := p.db.QueryRowContext(ctx, `
		SELECT ep.id
		FROM devices d
		JOIN elderly_profiles ep ON ep.user_id = d.owner_user_id
		WHERE d.device_id = $1
		LIMIT 1`, deviceID).Scan(&elderlyID)
	if err != nil {
		return "", fmt.Errorf("resolve elderly from device %s: %w", deviceID, err)
	}
	return elderlyID, nil
}

// CreatePushLog inserts a push notification log entry.
func (p *Postgres) CreatePushLog(ctx context.Context, log *model.PushLog) error {
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO push_logs (id, alert_id, elderly_id, channel, status, detail, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		log.ID, log.AlertID, log.ElderlyID, log.Channel, log.Status, log.Detail, log.CreatedAt)
	return err
}

// ListPushLogs returns recent push logs with optional filters.
func (p *Postgres) ListPushLogs(ctx context.Context, alertID, channel string, page, pageSize int) ([]model.PushLog, error) {
	q := `SELECT id, alert_id, elderly_id, channel, status, detail, created_at FROM push_logs WHERE 1=1`
	args := []interface{}{}
	if alertID != "" {
		q += " AND alert_id = $" + strconv.Itoa(len(args)+1)
		args = append(args, alertID)
	}
	if channel != "" {
		q += " AND channel = $" + strconv.Itoa(len(args)+1)
		args = append(args, channel)
	}
	q += " ORDER BY created_at DESC"
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		q += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)
	}
	rows, err := p.db.QueryContext(ctx, q, args...)
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
