package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"eregen.dev/admin-api/internal/model"
	"fmt"
	"time"
)


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
