// Package store defines the unified data access interface for both SQLite and
// PostgreSQL backends in the api-server. This interface is shared between
// the adapter layers to provide a consistent facade for routing handlers.
package store

import (
	"context"

	"eregen.dev/api-server/internal/model"
)

// Store provides a minimal, API-focused contract suitable for frontend verification.
// It mirrors the key methods needed by /api/v1/* endpoints while abstracting
// away the underlying SQL dialect.
type Store interface {
	// Health checks that the database connection is alive.
	Health(ctx context.Context) error

	// ListElderly returns paginated elderly profiles.
	ListElderly(ctx context.Context, page, pageSize int) ([]model.ElderlyProfile, error)

	// ListUsers returns paginated user summaries with optional role filter.
	ListUsers(ctx context.Context, page, pageSize int, role string) ([]model.UserSummary, error)

	// ListDevices returns device summaries filtered by status (empty = all).
	ListDevices(ctx context.Context, status string) ([]model.DeviceSummary, error)

	// GetActiveAlerts returns pending/resolved alerts summary.
	GetActiveAlerts(ctx context.Context) ([]model.AlertSummary, error)

	// ValidateToken validates an API token/key and returns associated user ID.
	ValidateToken(ctx context.Context, token string) (string, error)
}

var _ Store = (*SqliteStore)(nil)
