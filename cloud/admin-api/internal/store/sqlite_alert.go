package store

import (
	"context"
	"eregen.dev/admin-api/internal/model"
	"fmt"
	"github.com/google/uuid"
	"time"
)


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

// ResolveAlert marks an alert as resolved.
func (s *SqliteStore) ResolveAlert(ctx context.Context, alertID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE alerts SET status = 'resolved', resolved_at = datetime('now') WHERE id = ?`, alertID)
	return err
}

// UpdateAlertStatus updates an alert's status and resolved_at timestamp.
func (s *SqliteStore) UpdateAlertStatus(ctx context.Context, alertID, status string) error {
	if status == "resolved" {
		_, err := s.db.ExecContext(ctx,
			`UPDATE alerts SET status = 'resolved', resolved_at = datetime('now') WHERE id = ?`, alertID)
		return err
	}
	_, err := s.db.ExecContext(ctx,
		`UPDATE alerts SET status = ? WHERE id = ?`, status, alertID)
	return err
}

// CreateAlert inserts a new alert record.
func (s *SqliteStore) CreateAlert(ctx context.Context, a *model.AlertSummary) error {
	a.ID = uuid.New().String()
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now()
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO alerts (id, elderly_id, alert_type, severity, status, message, device_id, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.ElderlyID, a.AlertType, a.Severity, a.Status, "", a.DeviceID, a.CreatedAt)
	return err
}
