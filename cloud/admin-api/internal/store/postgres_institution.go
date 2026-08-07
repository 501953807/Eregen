package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"eregen.dev/admin-api/internal/model"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"strconv"
	"strings"
)


func (s *PostgresStore) ListInstitutions(ctx context.Context, page, pageSize int, name, typ, status string) ([]model.InstitutionSummary, error) {

	query := `SELECT i.id, i.name, i.type, i.code, i.contact_name, i.contact_phone,

			  i.access_level, i.status, i.created_at, i.updated_at,

			  (SELECT COUNT(*) FROM b2b_api_keys WHERE institution_id = i.id AND active = true) as api_key_count

			  FROM b2b_institutions i`

	args := []interface{}{}

	conditions := []string{}



	if name != "" {

		conditions = append(conditions, "i.name LIKE $"+strconv.Itoa(len(args)+1))

		args = append(args, "%"+name+"%")

	}

	if typ != "" {

		conditions = append(conditions, "i.type = $"+strconv.Itoa(len(args)+1))

		args = append(args, typ)

	}

	if status != "" {

		conditions = append(conditions, "i.status = $"+strconv.Itoa(len(args)+1))

		args = append(args, status)

	}



	if len(conditions) > 0 {

		query += " WHERE " + strings.Join(conditions, " AND ")

	}

	query += " ORDER BY i.created_at DESC"



	if page > 0 && pageSize > 0 {

		offset := (page - 1) * pageSize

		query += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)

	}



	rows, err := s.db.QueryContext(ctx, query, args...)

	if err != nil {

		return nil, fmt.Errorf("list institutions: %w", err)

	}

	defer rows.Close()



	var items []model.InstitutionSummary

	for rows.Next() {

		var item model.InstitutionSummary

		err := rows.Scan(&item.ID, &item.Name, &item.Type, &item.Code,

			&item.ContactName, &item.ContactPhone, &item.AccessLevel,

			&item.Status, &item.CreatedAt, &item.UpdatedAt, &item.APIKeyCount)

		if err != nil {

			return nil, fmt.Errorf("scan institution: %w", err)

		}

		items = append(items, item)

	}

	return items, rows.Err()

}



// GetInstitution returns a single institution by ID.

func (s *PostgresStore) GetInstitution(ctx context.Context, id string) (*model.InstitutionSummary, error) {

	var item model.InstitutionSummary

	err := s.db.QueryRowContext(ctx, `

		SELECT i.id, i.name, i.type, i.code, i.contact_name, i.contact_phone,

		       i.access_level, i.status, i.created_at, i.updated_at,

		       (SELECT COUNT(*) FROM b2b_api_keys WHERE institution_id = i.id AND active = true)

		FROM b2b_institutions i WHERE i.id = $1`, id).Scan(

		&item.ID, &item.Name, &item.Type, &item.Code,

		&item.ContactName, &item.ContactPhone, &item.AccessLevel,

		&item.Status, &item.CreatedAt, &item.UpdatedAt, &item.APIKeyCount)

	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {

			return nil, nil

		}

		return nil, fmt.Errorf("get institution: %w", err)

	}

	return &item, nil

}



// CreateInstitution creates a new institution.

func (s *PostgresStore) CreateInstitution(ctx context.Context, i *model.InstitutionSummary) error {

	_, err := s.db.ExecContext(ctx, `

		INSERT INTO b2b_institutions (id, name, type, code, contact_name, contact_phone, access_level, status)

		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,

		i.ID, i.Name, i.Type, i.Code, i.ContactName, i.ContactPhone, i.AccessLevel, i.Status)

	return err

}



// UpdateInstitution updates an institution.

func (s *PostgresStore) UpdateInstitution(ctx context.Context, id string, updates map[string]any) error {

	setClauses := []string{}

	args := []interface{}{}

	i := 1

	for k, v := range updates {

		setClauses = append(setClauses, k+" = $"+strconv.Itoa(i))

		args = append(args, v)

		i++

	}

	args = append(args, id)

	_, err := s.db.ExecContext(ctx, fmt.Sprintf("UPDATE b2b_institutions SET %s WHERE id = $%d", strings.Join(setClauses, ", "), i), args...)

	return err

}



// DeleteInstitution deletes an institution.

func (s *PostgresStore) DeleteInstitution(ctx context.Context, id string) error {

	_, err := s.db.ExecContext(ctx, "DELETE FROM b2b_institutions WHERE id = $1", id)

	return err

}



// CreateInstitutionAPIKey creates a new API key for an institution.

func (s *PostgresStore) CreateInstitutionAPIKey(ctx context.Context, institutionID, name string) (string, error) {

	keyID := uuid.New().String()

	keyValue := "ek_" + uuid.New().String()[:32]

	keyHash := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(keyValue)))

	_, err := s.db.ExecContext(ctx, `

		INSERT INTO b2b_api_keys (id, institution_id, name, key_hash, key_prefix)

		VALUES ($1, $2, $3, $4, $5)`, keyID, institutionID, name, keyHash, keyValue[:8])

	if err != nil {

		return "", err

	}

	return keyValue, nil

}



// RevokeInstitutionAPIKey revokes an API key for an institution.

func (s *PostgresStore) RevokeInstitutionAPIKey(ctx context.Context, institutionID, keyID string) error {

	_, err := s.db.ExecContext(ctx, "UPDATE b2b_api_keys SET active = false WHERE id = $1 AND institution_id = $2", keyID, institutionID)

	return err

}

