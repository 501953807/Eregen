package store

import (
	"context"
	"database/sql"
	"eregen.dev/admin-api/internal/model"
	"errors"
	"fmt"
	"strconv"
	"strings"
)


func (s *PostgresStore) ListSubscriptions(ctx context.Context, page, pageSize int, status, planTier string) ([]model.SubscriptionItem, error) {

	query := `SELECT s.id, s.user_id, u.name as user_name, u.phone as user_phone,

			  s.plan_tier, s.status, s.billing_cycle, s.starts_at, s.expires_at,

			  s.cancellation_reason, COALESCE(s.total_spent, 0), s.created_at

			  FROM subscriptions s LEFT JOIN users u ON s.user_id = u.id`

	args := []interface{}{}

	conditions := []string{}



	if status != "" {

		conditions = append(conditions, "s.status = $"+strconv.Itoa(len(args)+1))

		args = append(args, status)

	}

	if planTier != "" {

		conditions = append(conditions, "s.plan_tier = $"+strconv.Itoa(len(args)+1))

		args = append(args, planTier)

	}



	if len(conditions) > 0 {

		query += " WHERE " + strings.Join(conditions, " AND ")

	}

	query += " ORDER BY s.created_at DESC"



	if page > 0 && pageSize > 0 {

		offset := (page - 1) * pageSize

		query += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)

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

		err := rows.Scan(&item.ID, &item.UserID, &item.UserName, &item.UserPhone,

			&item.PlanTier, &item.Status, &item.BillingCycle,

			&item.StartDate, &item.EndDate, &item.CancellationReason,

			&totalSpent, &item.CreatedAt)

		if err != nil {

			return nil, fmt.Errorf("scan subscription: %w", err)

		}

		if totalSpent != nil {

			item.TotalSpent = *totalSpent

		}

		items = append(items, item)

	}

	return items, rows.Err()

}



// GetSubscription returns a single subscription by ID.

func (s *PostgresStore) GetSubscription(ctx context.Context, id string) (*model.SubscriptionItem, error) {

	var item model.SubscriptionItem

	var totalSpent *float64

	err := s.db.QueryRowContext(ctx, `

		SELECT s.id, s.user_id, u.name, u.phone, s.plan_tier, s.status,

		       s.billing_cycle, s.starts_at, s.expires_at, s.cancellation_reason,

		       COALESCE(s.total_spent, 0), s.created_at

		FROM subscriptions s LEFT JOIN users u ON s.user_id = u.id

		WHERE s.id = $1`, id).Scan(

		&item.ID, &item.UserID, &item.UserName, &item.UserPhone,

		&item.PlanTier, &item.Status, &item.BillingCycle,

		&item.StartDate, &item.EndDate, &item.CancellationReason,

		&totalSpent, &item.CreatedAt)

	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {

			return nil, nil

		}

		return nil, fmt.Errorf("get subscription: %w", err)

	}

	if totalSpent != nil {

		item.TotalSpent = *totalSpent

	}

	return &item, nil

}



// CreateSubscription creates a new subscription.

func (s *PostgresStore) CreateSubscription(ctx context.Context, sub *model.SubscriptionItem) error {

	_, err := s.db.ExecContext(ctx, `

		INSERT INTO subscriptions (id, user_id, plan_tier, status, billing_cycle, starts_at, expires_at)

		VALUES ($1, $2, $3, $4, $5, $6, $7)`,

		sub.ID, sub.UserID, sub.PlanTier, sub.Status, sub.BillingCycle, sub.StartDate, sub.EndDate)

	return err

}



// UpdateSubscription updates a subscription.

func (s *PostgresStore) UpdateSubscription(ctx context.Context, id string, updates map[string]any) error {

	setClauses := []string{}

	args := []interface{}{}

	i := 1

	for k, v := range updates {

		setClauses = append(setClauses, k+" = $"+strconv.Itoa(i))

		args = append(args, v)

		i++

	}

	args = append(args, id)

	_, err := s.db.ExecContext(ctx, fmt.Sprintf("UPDATE subscriptions SET %s WHERE id = $%d", strings.Join(setClauses, ", "), i), args...)

	return err

}



// RenewSubscription renews a subscription.

func (s *PostgresStore) RenewSubscription(ctx context.Context, id string, endDate string) error {

	return s.UpdateSubscription(ctx, id, map[string]any{"status": "active", "expires_at": endDate})

}



// ListInstitutions returns a paginated list of institutions.