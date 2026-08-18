package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"eregen.dev/admin-api/internal/model"

	"github.com/google/uuid"
)

// LifecycleStore implementation for SqliteStore.

func (s *SqliteStore) TransitionStatus(ctx context.Context, personID string, chain model.BusinessChain, newStatus, reason string) error {
	var currentStatus string
	err := s.db.QueryRowContext(ctx,
		`SELECT status FROM person_profiles WHERE person_id = ? AND business_chain = ?`,
		personID, chain).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		return fmt.Errorf("profile not found for person %s in chain %s", personID, chain)
	}
	if err != nil {
		return fmt.Errorf("get profile status: %w", err)
	}

	// Validate transition
	allowed := allowedTransitions[string(chain)][currentStatus]
	if !allowed[newStatus] {
		return fmt.Errorf("invalid transition from %s to %s for chain %s", currentStatus, newStatus, chain)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`UPDATE person_profiles SET status = ?, reason = ?, updated_at = datetime('now')
		 WHERE person_id = ? AND business_chain = ?`,
		newStatus, reason, personID, chain)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	// Write audit log
	logID := uuid.New().String()
	_, err = tx.ExecContext(ctx,
		`INSERT INTO status_transition_logs (id, person_id, business_chain, from_status, to_status, reason, performed_by, metadata, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		logID, personID, string(chain), currentStatus, newStatus, reason, "", "{}", time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("log transition: %w", err)
	}

	return tx.Commit()
}

func (s *SqliteStore) GetPersonStatus(ctx context.Context, personID string, chain model.BusinessChain) (string, error) {
	var status string
	err := s.db.QueryRowContext(ctx,
		`SELECT status FROM person_profiles WHERE person_id = ? AND business_chain = ?`,
		personID, chain).Scan(&status)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return status, err
}

func (s *SqliteStore) LinkPersons(ctx context.Context, personID1, personID2 string, chain1, chain2 model.BusinessChain) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE person_profiles SET linked_person_id = ?, updated_at = datetime('now')
		 WHERE person_id = ? AND business_chain = ?`,
		personID2, personID1, chain1)
	if err != nil {
		return fmt.Errorf("link person 1: %w", err)
	}
	_, err = s.db.ExecContext(ctx,
		`UPDATE person_profiles SET linked_person_id = ?, updated_at = datetime('now')
		 WHERE person_id = ? AND business_chain = ?`,
		personID1, personID2, chain2)
	return err
}

func (s *SqliteStore) GetStatusHistory(ctx context.Context, personID string, chain model.BusinessChain, limit int) ([]model.StatusTransition, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, person_id, business_chain, from_status, to_status, reason, performed_by, metadata, created_at
		 FROM status_transition_logs
		 WHERE person_id = ? AND business_chain = ?
		 ORDER BY created_at DESC
		 LIMIT ?`,
		personID, string(chain), limit)
	if err != nil {
		return nil, fmt.Errorf("query status history: %w", err)
	}
	defer rows.Close()

	var results []model.StatusTransition
	for rows.Next() {
		var t model.StatusTransition
		var metadata, performedBy sql.NullString
		err := rows.Scan(&t.ID, &t.PersonID, &t.BusinessChain, &t.FromStatus, &t.ToStatus, &t.Reason, &performedBy, &metadata, &t.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan status transition: %w", err)
		}
		if performedBy.Valid {
			t.PerformedBy = performedBy.String
		}
		if metadata.Valid {
			t.Metadata = metadata.String
		}
		results = append(results, t)
	}
	return results, rows.Err()
}

// allowedTransitions defines allowed status transitions per business chain.
var allowedTransitions = map[string]map[string]map[string]bool{
	"self": {
		"pending":   {"active": true},
		"active":    {"suspended": true, "cancelled": true},
		"suspended": {"active": true, "cancelled": true},
		"cancelled": {},
	},
	"hospital": {
		"pending":      {"admitted": true},
		"admitted":     {"in_treatment": true},
		"in_treatment": {"discharged": true},
		"discharged":   {"archived": true},
		"archived":     {},
	},
	"community": {
		"pending":     {"certified": true},
		"certified":   {"active": true},
		"active":      {"suspended": true, "deactivated": true},
		"suspended":   {"active": true},
		"deactivated": {},
	},
}
