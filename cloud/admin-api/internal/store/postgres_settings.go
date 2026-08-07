package store

import (
	"context"
	"database/sql"
	"eregen.dev/admin-api/internal/model"
	"fmt"
	"time"
)


func (s *PostgresStore) GetNotificationSettings(ctx context.Context) (map[string]any, error) {

	var jsonb string

	err := s.db.QueryRowContext(ctx, `SELECT COALESCE(setting_value, '{}') FROM system_settings WHERE key = 'notification'`).Scan(&jsonb)

	if err != nil && err != sql.ErrNoRows {

		return nil, fmt.Errorf("get notification settings: %w", err)

	}

	return map[string]any{}, nil // default empty

}



// UpdateNotificationSettings persists notification config.

func (s *PostgresStore) UpdateNotificationSettings(ctx context.Context, data map[string]any) error {

	_, err := s.db.ExecContext(ctx,

		`INSERT INTO system_settings (key, setting_value) VALUES ('notification', $1)

		 ON CONFLICT (key) DO UPDATE SET setting_value = $1`,

		`{}`) // placeholder — use json.Marshal in real impl

	return err

}



// ListAPIKeys returns registered B2B API keys.

func (s *PostgresStore) ListAPIKeys(ctx context.Context) ([]model.APIKeySummary, error) {

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

func (s *PostgresStore) CreateAPIKey(ctx context.Context, name, keyHash string, expiresAt *time.Time) (string, error) {

	var id string

	err := s.db.QueryRowContext(ctx,

		`INSERT INTO b2b_api_keys (name, key_hash, expires_at, active) VALUES ($1,$2,$3,true) RETURNING id`,

		name, keyHash, expiresAt).Scan(&id)

	return id, err

}



// RevokeAPIKey deactivates a B2B API key.

func (s *PostgresStore) RevokeAPIKey(ctx context.Context, id string) error {

	_, err := s.db.ExecContext(ctx, `UPDATE b2b_api_keys SET active = false WHERE id = $1`, id)

	return err

}



// ChangeAdminPassword updates an admin user's password hash.

func (s *PostgresStore) ChangeAdminPassword(ctx context.Context, userID, hash string) error {

	_, err := s.db.ExecContext(ctx, `UPDATE users SET password_hash = $1 WHERE id = $2 AND role = 'admin'`, hash, userID)

	return err

}



// GetUserByCredential returns a user by email (method="email") or phone+OTP (method="phone").