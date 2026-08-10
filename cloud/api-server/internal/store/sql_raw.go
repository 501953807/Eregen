package store

import (
	"context"
	"database/sql"
)

// sqlRaw wraps *sql.DB to implement RawDB interface.
type sqlRaw struct {
	db *sql.DB
}

func (r *sqlRaw) QueryRow(ctx context.Context, query string, args ...any) RawRow {
	return &sqlRawRow{row: r.db.QueryRowContext(ctx, query, args...)}
}

func (r *sqlRaw) Query(ctx context.Context, query string, args ...any) (RawRows, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &sqlRawRows{rows: rows}, nil
}

func (r *sqlRaw) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

type sqlRawRow struct{ row interface{ Scan(...any) error } }

func (r *sqlRawRow) Scan(dest ...any) error { return r.row.Scan(dest...) }

type sqlRawRows struct{ rows *sql.Rows }

func (r *sqlRawRows) Next() bool     { return r.rows.Next() }
func (r *sqlRawRows) Scan(dest ...any) error { return r.rows.Scan(dest...) }
func (r *sqlRawRows) Close() error   { return r.rows.Close() }
func (r *sqlRawRows) Err() error     { return r.rows.Err() }
