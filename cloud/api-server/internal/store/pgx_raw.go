package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// pgxRaw wraps pgxpool.Pool to implement RawDB interface.
// It converts ? placeholders to $1, $2, etc. for PostgreSQL.
type pgxRaw struct {
	pool *pgxpool.Pool
}

func (r *pgxRaw) QueryRow(ctx context.Context, query string, args ...any) RawRow {
	q, a := toPGParams(query, args)
	return &pgxRawRow{row: r.pool.QueryRow(ctx, q, a...)}
}

func (r *pgxRaw) Query(ctx context.Context, query string, args ...any) (RawRows, error) {
	q, a := toPGParams(query, args)
	rows, err := r.pool.Query(ctx, q, a...)
	if err != nil {
		return nil, err
	}
	return &pgxRawRows{rows: rows}, nil
}

func (r *pgxRaw) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

// toPGParams converts ? placeholders to $1, $2, ... for pgx.
func toPGParams(query string, args []any) (string, []any) {
	if !strings.Contains(query, "?") {
		return query, args
	}
	var sb strings.Builder
	i := 1
	for _, ch := range query {
		if ch == '?' {
			fmt.Fprintf(&sb, "$%d", i)
			i++
		} else {
			sb.WriteRune(ch)
		}
	}
	return sb.String(), args
}

// pgx bridge types for RawDB interface

type pgxRawRow struct{ row interface{ Scan(...any) error } }

func (r *pgxRawRow) Scan(dest ...any) error { return r.row.Scan(dest...) }

type pgxRawRows struct{ rows interface{ Next() bool; Scan(...any) error; Close(); Err() error } }

func (r *pgxRawRows) Next() bool     { return r.rows.Next() }
func (r *pgxRawRows) Scan(dest ...any) error { return r.rows.Scan(dest...) }
func (r *pgxRawRows) Close() error   { r.rows.Close(); return nil }
func (r *pgxRawRows) Err() error     { return r.rows.Err() }
