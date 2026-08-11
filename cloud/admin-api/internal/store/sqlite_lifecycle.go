package store

import (
	"context"
	"database/sql"
	"fmt"

	"eregen.dev/admin-api/internal/model"
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

	_, err = s.db.ExecContext(ctx,
		`UPDATE person_profiles SET status = ?, reason = ?, updated_at = datetime('now')
		 WHERE person_id = ? AND business_chain = ?`,
		newStatus, reason, personID, chain)
	return err
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
