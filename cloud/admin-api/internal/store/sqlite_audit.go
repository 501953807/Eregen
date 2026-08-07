package store

import (
	"context"
	"encoding/json"
	"eregen.dev/admin-api/internal/model"
	"github.com/google/uuid"
	"time"
)


// ========== Audit Log Methods ==========

func (s *SqliteStore) CreateAuditLog(ctx context.Context, log *model.AuditLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	detailsJSON, _ := json.Marshal(log.Details)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO audit_log (id, actor_id, action, resource_type, resource_id, details, ip_address, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		log.ID, log.UserID, log.Action, log.Resource, log.ResourceID,
		string(detailsJSON), log.IP, log.CreatedAt)
	return err
}

func (s *SqliteStore) ListAuditLogs(ctx context.Context, limit int) ([]model.AuditLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, actor_id, action, resource_type, resource_id, details, ip_address, created_at
		FROM audit_log ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []model.AuditLog
	for rows.Next() {
		var l model.AuditLog
		var detailsJSON string
		if err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.Resource, &l.ResourceID, &detailsJSON, &l.IP, &l.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(detailsJSON), &l.Details)
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func (s *SqliteStore) ListAuditLogsByUser(ctx context.Context, userID string, limit int) ([]model.AuditLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, actor_id, action, resource_type, resource_id, details, ip_address, created_at
		FROM audit_log WHERE actor_id=? ORDER BY created_at DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []model.AuditLog
	for rows.Next() {
		var l model.AuditLog
		var detailsJSON string
		if err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.Resource, &l.ResourceID, &detailsJSON, &l.IP, &l.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(detailsJSON), &l.Details)
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func (s *SqliteStore) ListAuditLogsByAction(ctx context.Context, action string, limit int) ([]model.AuditLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, actor_id, action, resource_type, resource_id, details, ip_address, created_at
		FROM audit_log WHERE action=? ORDER BY created_at DESC LIMIT ?`, action, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []model.AuditLog
	for rows.Next() {
		var l model.AuditLog
		var detailsJSON string
		if err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.Resource, &l.ResourceID, &detailsJSON, &l.IP, &l.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(detailsJSON), &l.Details)
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
