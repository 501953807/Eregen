// Package store provides SQLite implementation of the Store interface.
package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"eregen.dev/admin-api/internal/model"
	"github.com/google/uuid"
)

// ========== Subscription Store ==========

func (s *SqliteStore) ListSubscriptions(ctx context.Context, page, pageSize int, status, planTier string) ([]model.SubscriptionItem, error) {
	query := `SELECT s.id, s.user_id, u.name as user_name, u.phone as user_phone,
			  s.plan_tier, s.status, s.billing_cycle, s.starts_at, s.expires_at,
			  s.cancellation_reason, COALESCE(s.total_spent, 0), s.created_at
			  FROM subscriptions s LEFT JOIN users u ON s.user_id = u.id`
	args := []interface{}{}
	conditions := []string{}

	if status != "" {
		conditions = append(conditions, "s.status = ?")
		args = append(args, status)
	}
	if planTier != "" {
		conditions = append(conditions, "s.plan_tier = ?")
		args = append(args, planTier)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY s.created_at DESC"

	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query += fmt.Sprintf(" LIMIT ? OFFSET ?")
		args = append(args, pageSize, offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}
	defer rows.Close()

	var items []model.SubscriptionItem
	for rows.Next() {
		var item model.SubscriptionItem
		var totalSpent *float64
		var nameRaw, cancelReasonRaw, startDateRaw, endDateRaw sql.NullString
		var createdAtStr string
		err := rows.Scan(&item.ID, &item.UserID, &nameRaw, &item.UserPhone,
			&item.PlanTier, &item.Status, &item.BillingCycle,
			&startDateRaw, &endDateRaw, &cancelReasonRaw,
			&totalSpent, &createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		if nameRaw.Valid {
			item.UserName = nameRaw.String
		}
		if cancelReasonRaw.Valid {
			item.CancellationReason = cancelReasonRaw.String
		}
		if startDateRaw.Valid {
			item.StartDate = startDateRaw.String
		}
		if endDateRaw.Valid {
			item.EndDate = endDateRaw.String
		}
		if totalSpent != nil {
			item.TotalSpent = *totalSpent
		}
		item.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *SqliteStore) GetSubscription(ctx context.Context, id string) (*model.SubscriptionItem, error) {
	var item model.SubscriptionItem
	var totalSpent *float64
	var nameRaw, cancelReasonRaw, startDateRaw, endDateRaw sql.NullString
	var createdAtStr string
	err := s.db.QueryRowContext(ctx, `
		SELECT s.id, s.user_id, u.name, u.phone, s.plan_tier, s.status,
		       s.billing_cycle, s.starts_at, s.expires_at, s.cancellation_reason,
		       COALESCE(s.total_spent, 0), s.created_at
		FROM subscriptions s LEFT JOIN users u ON s.user_id = u.id
		WHERE s.id = ?`, id).Scan(
		&item.ID, &item.UserID, &nameRaw, &item.UserPhone,
		&item.PlanTier, &item.Status, &item.BillingCycle,
		&startDateRaw, &endDateRaw, &cancelReasonRaw,
		&totalSpent, &createdAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get subscription: %w", err)
	}
	if nameRaw.Valid {
		item.UserName = nameRaw.String
	}
	if cancelReasonRaw.Valid {
		item.CancellationReason = cancelReasonRaw.String
	}
	if startDateRaw.Valid {
		item.StartDate = startDateRaw.String
	}
	if endDateRaw.Valid {
		item.EndDate = endDateRaw.String
	}
	if totalSpent != nil {
		item.TotalSpent = *totalSpent
	}
	item.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	return &item, nil
}

func (s *SqliteStore) CreateSubscription(ctx context.Context, sub *model.SubscriptionItem) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO subscriptions (id, user_id, plan_tier, status, billing_cycle, starts_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		sub.ID, sub.UserID, sub.PlanTier, sub.Status, sub.BillingCycle, sub.StartDate, sub.EndDate)
	return err
}

func (s *SqliteStore) UpdateSubscription(ctx context.Context, id string, updates map[string]any) error {
	setClauses := []string{}
	args := []interface{}{}
	for k, v := range updates {
		setClauses = append(setClauses, k+" = ?")
		args = append(args, v)
	}
	args = append(args, id)
	_, err := s.db.ExecContext(ctx, fmt.Sprintf("UPDATE subscriptions SET %s WHERE id = ?", strings.Join(setClauses, ", ")), args...)
	return err
}

func (s *SqliteStore) RenewSubscription(ctx context.Context, id string, endDate string) error {
	return s.UpdateSubscription(ctx, id, map[string]any{"status": "active", "expires_at": endDate})
}

// ========== Institution Store ==========

func (s *SqliteStore) ListInstitutions(ctx context.Context, page, pageSize int, name, typ, status string) ([]model.InstitutionSummary, error) {
	query := `SELECT i.id, i.name, i.type, i.code, i.contact_name, i.contact_phone,
			  i.access_level, i.status, i.created_at, i.updated_at,
			  0 as api_key_count
			  FROM b2b_institutions i`
	args := []interface{}{}
	conditions := []string{}

	if name != "" {
		conditions = append(conditions, "i.name LIKE ?")
		args = append(args, "%"+name+"%")
	}
	if typ != "" {
		conditions = append(conditions, "i.type = ?")
		args = append(args, typ)
	}
	if status != "" {
		conditions = append(conditions, "i.status = ?")
		args = append(args, status)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY i.created_at DESC"

	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query += fmt.Sprintf(" LIMIT ? OFFSET ?")
		args = append(args, pageSize, offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list institutions: %w", err)
	}
	defer rows.Close()

	var items []model.InstitutionSummary
	for rows.Next() {
		var item model.InstitutionSummary
		var codeRaw, contactNameRaw, contactPhoneRaw sql.NullString
		var createdAtStr, updatedAtStr string
		err := rows.Scan(&item.ID, &item.Name, &item.Type, &codeRaw, &contactNameRaw, &contactPhoneRaw,
			&item.AccessLevel, &item.Status, &createdAtStr, &updatedAtStr, &item.APIKeyCount)
		if err != nil {
			return nil, fmt.Errorf("scan institution: %w", err)
		}
		if codeRaw.Valid {
			item.Code = codeRaw.String
		}
		if contactNameRaw.Valid {
			item.ContactName = contactNameRaw.String
		}
		if contactPhoneRaw.Valid {
			item.ContactPhone = contactPhoneRaw.String
		}
		item.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		item.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *SqliteStore) GetInstitution(ctx context.Context, id string) (*model.InstitutionSummary, error) {
	var item model.InstitutionSummary
	var createdAtStr, updatedAtStr string
	var codeRaw, contactNameRaw, contactPhoneRaw sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT i.id, i.name, i.type, i.code, i.contact_name, i.contact_phone,
		       i.access_level, i.status, i.created_at, i.updated_at,
		       0
		FROM b2b_institutions i WHERE i.id = ?`, id).Scan(
		&item.ID, &item.Name, &item.Type, &codeRaw, &contactNameRaw, &contactPhoneRaw,
		&item.AccessLevel, &item.Status, &createdAtStr, &updatedAtStr, &item.APIKeyCount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get institution: %w", err)
	}
	if codeRaw.Valid {
		item.Code = codeRaw.String
	}
	if contactNameRaw.Valid {
		item.ContactName = contactNameRaw.String
	}
	if contactPhoneRaw.Valid {
		item.ContactPhone = contactPhoneRaw.String
	}
	item.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	item.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)
	return &item, nil
}

func (s *SqliteStore) CreateInstitution(ctx context.Context, i *model.InstitutionSummary) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO b2b_institutions (id, name, type, code, contact_name, contact_phone, access_level, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		i.ID, i.Name, i.Type, i.Code, i.ContactName, i.ContactPhone, i.AccessLevel, i.Status)
	return err
}

func (s *SqliteStore) UpdateInstitution(ctx context.Context, id string, updates map[string]any) error {
	setClauses := []string{}
	args := []interface{}{}
	for k, v := range updates {
		setClauses = append(setClauses, k+" = ?")
		args = append(args, v)
	}
	args = append(args, id)
	_, err := s.db.ExecContext(ctx, fmt.Sprintf("UPDATE b2b_institutions SET %s WHERE id = ?", strings.Join(setClauses, ", ")), args...)
	return err
}

func (s *SqliteStore) DeleteInstitution(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM b2b_institutions WHERE id = ?", id)
	return err
}

func (s *SqliteStore) CreateInstitutionAPIKey(ctx context.Context, institutionID, name string) (string, error) {
	keyID := uuid.New().String()
	keyValue := "ek_" + uuid.New().String()[:32]
	keyHash := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(keyValue)))
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO b2b_api_keys (id, institution_id, name, key_hash, key_prefix)
		VALUES (?, ?, ?, ?, ?)`, keyID, institutionID, name, keyHash, keyValue[:8])
	if err != nil {
		return "", err
	}
	return keyValue, nil
}

func (s *SqliteStore) RevokeInstitutionAPIKey(ctx context.Context, institutionID, keyID string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE b2b_api_keys SET active = 0 WHERE id = ? AND institution_id = ?", keyID, institutionID)
	return err
}
