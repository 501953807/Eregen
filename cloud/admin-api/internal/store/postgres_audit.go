package store

import (
	"context"
	"encoding/json"
	"eregen.dev/admin-api/internal/model"
	"github.com/google/uuid"
	"time"
)


func (p *PostgresStore) CreateAuditLog(ctx context.Context, log *model.AuditLog) error {

	if log.ID == "" {

		log.ID = uuid.New().String()

	}

	if log.CreatedAt.IsZero() {

		log.CreatedAt = time.Now()

	}

	detailsJSON, _ := json.Marshal(log.Details)

	_, err := p.db.ExecContext(ctx, `

		INSERT INTO audit_log (id, actor_id, action, resource_type, resource_id, details, ip_address, created_at)

		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,

		log.ID, log.UserID, log.Action, log.Resource, log.ResourceID,

		detailsJSON, log.IP, log.CreatedAt)

	return err

}



func (p *PostgresStore) ListAuditLogs(ctx context.Context, limit int) ([]model.AuditLog, error) {

	if limit <= 0 || limit > 500 {

		limit = 50

	}

	rows, err := p.db.QueryContext(ctx, `

		SELECT id, actor_id, action, resource_type, resource_id, details, ip_address, created_at

		FROM audit_log ORDER BY created_at DESC LIMIT $1`, limit)

	if err != nil {

		return nil, err

	}

	defer rows.Close()

	var logs []model.AuditLog

	for rows.Next() {

		var l model.AuditLog

		var detailsJSON []byte

		if err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.Resource, &l.ResourceID, &detailsJSON, &l.IP, &l.CreatedAt); err != nil {

			return nil, err

		}

		json.Unmarshal(detailsJSON, &l.Details)

		logs = append(logs, l)

	}

	return logs, rows.Err()

}



func (p *PostgresStore) ListAuditLogsByUser(ctx context.Context, userID string, limit int) ([]model.AuditLog, error) {

	if limit <= 0 || limit > 500 {

		limit = 50

	}

	rows, err := p.db.QueryContext(ctx, `

		SELECT id, actor_id, action, resource_type, resource_id, details, ip_address, created_at

		FROM audit_log WHERE actor_id=$1 ORDER BY created_at DESC LIMIT $2`, userID, limit)

	if err != nil {

		return nil, err

	}

	defer rows.Close()

	var logs []model.AuditLog

	for rows.Next() {

		var l model.AuditLog

		var detailsJSON []byte

		if err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.Resource, &l.ResourceID, &detailsJSON, &l.IP, &l.CreatedAt); err != nil {

			return nil, err

		}

		json.Unmarshal(detailsJSON, &l.Details)

		logs = append(logs, l)

	}

	return logs, rows.Err()

}



func (p *PostgresStore) ListAuditLogsByAction(ctx context.Context, action string, limit int) ([]model.AuditLog, error) {

	if limit <= 0 || limit > 500 {

		limit = 50

	}

	rows, err := p.db.QueryContext(ctx, `

		SELECT id, actor_id, action, resource_type, resource_id, details, ip_address, created_at

		FROM audit_log WHERE action=$1 ORDER BY created_at DESC LIMIT $2`, action, limit)

	if err != nil {

		return nil, err

	}

	defer rows.Close()

	var logs []model.AuditLog

	for rows.Next() {

		var l model.AuditLog

		var detailsJSON []byte

		if err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.Resource, &l.ResourceID, &detailsJSON, &l.IP, &l.CreatedAt); err != nil {

			return nil, err

		}

		json.Unmarshal(detailsJSON, &l.Details)

		logs = append(logs, l)

	}

	return logs, rows.Err()

}

