package store

import (
	"database/sql"
	"errors"
)

// PostgresStore wraps database access for admin operations.
var ErrNotImplemented = errors.New("feature not yet implemented — backend schema required")

type PostgresStore struct {
	db *sql.DB
}

// NewPostgres opens a connection pool to PostgreSQL.
func NewPostgres(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		panic(err)
	}
	return db
}

// NewPostgresStore creates a new PostgresStore from an existing db connection.
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}
