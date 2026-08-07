package store

import (
	"context"
	"eregen.dev/admin-api/internal/auth"
	"eregen.dev/admin-api/internal/model"
	"fmt"
	"github.com/google/uuid"
)


func (s *PostgresStore) ListUsers(ctx context.Context, page, pageSize int, role string) ([]model.UserSummary, error) {

	query := `SELECT u.id, u.name, u.role, u.created_at,

		(SELECT COUNT(*) FROM devices d WHERE d.owner_user_id = u.id)

		FROM users u WHERE 1=1`

	args := []interface{}{}

	idx := 1

	if role != "" {

		query += fmt.Sprintf(" AND u.role=$%d", idx)

		args = append(args, role)

		idx++

	}

	query += fmt.Sprintf(" ORDER BY u.created_at DESC LIMIT $%d OFFSET $%d", idx, idx+1)

	args = append(args, pageSize, (page-1)*pageSize)



	rows, err := s.db.QueryContext(ctx, query, args...)

	if err != nil {

		return nil, fmt.Errorf("list users: %w", err)

	}

	defer rows.Close()



	var users []model.UserSummary

	for rows.Next() {

		var u model.UserSummary

		if err := rows.Scan(&u.ID, &u.Name, &u.Role, &u.CreatedAt, &u.Devices); err != nil {

			return nil, fmt.Errorf("scan user: %w", err)

		}

		users = append(users, u)

	}

	return users, rows.Err()

}



// ListAlerts returns recent alerts with optional severity and status filters.

func (s *PostgresStore) SetUserRole(ctx context.Context, userID, role string) error {

	_, err := s.db.ExecContext(ctx, `UPDATE users SET role = $1 WHERE id = $2`, role, userID)

	return err

}



// CreateUser inserts a new user and returns the generated ID.

func (s *PostgresStore) CreateUser(ctx context.Context, name, email, phone, role, password string) (string, error) {

	hash, err := auth.HashPassword(password)

	if err != nil {

		return "", fmt.Errorf("hash password: %w", err)

	}

	id := uuid.New().String()

	_, err = s.db.ExecContext(ctx,

		`INSERT INTO users (id, name, email, phone, role, password_hash) VALUES ($1, $2, $3, $4, $5, $6)`,

		id, name, email, phone, role, hash)

	return id, err

}



// UpdateUser modifies an existing user's basic fields.

func (s *PostgresStore) UpdateUser(ctx context.Context, id, name, email, phone, role string) error {

	_, err := s.db.ExecContext(ctx,

		`UPDATE users SET name=$1, email=$2, phone=$3, role=$4 WHERE id=$5`,

		name, email, phone, role, id)

	return err

}



// DeleteUser removes a user by ID.

func (s *PostgresStore) DeleteUser(ctx context.Context, id string) error {

	_, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)

	return err

}



// UpdateDeviceConfig updates device settings JSONB column safely using JSON marshaling.

func (s *PostgresStore) GetUserByCredential(ctx context.Context, method, credential, secret string) (*model.UserLogin, error) {

	var u model.UserLogin

	var hash string

	switch method {

	case "email":

		err := s.db.QueryRowContext(ctx,

			`SELECT id, name, role, password_hash FROM users WHERE email = $1`, credential).Scan(

			&u.ID, &u.Name, &u.Role, &hash)

		if err != nil {

			return nil, err

		}

		if !auth.ComparePassword(secret, hash) {

			return nil, fmt.Errorf("invalid credentials")

		}

	case "phone":

		err := s.db.QueryRowContext(ctx,

			`SELECT id, name, role, password_hash FROM users WHERE phone = $1`, credential).Scan(

			&u.ID, &u.Name, &u.Role, &hash)

		if err != nil {

			return nil, err

		}

		if !auth.ComparePassword(secret, hash) {

			return nil, fmt.Errorf("invalid credentials")

		}

	default:

		return nil, fmt.Errorf("invalid method")

	}

	return &u, nil

}



// ========== Medical Wristband Methods ==========

