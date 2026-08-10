// Package store defines the unified database access interface for api-server.
// Both Postgres and SQLite backends implement this interface.
package store

import "context"

// Database is the unified interface for all database backends in api-server.
// Use Raw() for dynamic SQL queries with ? placeholders.
type Database interface {
	Raw() RawDB
}

// Compile-time checks
var _ Database = (*Postgres)(nil)
var _ Database = (*SqliteStore)(nil)

// RawDB provides raw query access. Queries use ? placeholders.
type RawDB interface {
	QueryRow(ctx context.Context, query string, args ...any) RawRow
	Query(ctx context.Context, query string, args ...any) (RawRows, error)
	Ping(ctx context.Context) error
}

// RawRow is a single-row query result.
type RawRow interface {
	Scan(dest ...any) error
}

// RawRows is a multi-row query result.
type RawRows interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
	Err() error
}
