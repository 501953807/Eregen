// Package store provides SQLite implementation of the Store interface.
package store

import (
	"database/sql"
	_ "embed"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	_ "modernc.org/sqlite"
)

// parseTimeOrDefault parses a time string in various formats, returning a default value on error.
func parseTimeOrDefault(s string, defaultVal time.Time) time.Time {
	if s == "" {
		return defaultVal
	}
	return parseTimeStrict(s)
}

// parseTimeStrict parses a time string in various formats.
func parseTimeStrict(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	for _, layout := range []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05Z",
		"2006-01-02T15:04:05Z07:00",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// scanTimePtr scans a string column into a *time.Time, returning nil for empty values.
func scanTimePtr(s string) *time.Time {
	if s == "" {
		return nil
	}
	t := parseTimeStrict(s)
	return &t
}

// SqliteStore wraps database access for admin operations using SQLite.
type SqliteStore struct {
	db *sql.DB
}

// NewSqlite opens a connection to a SQLite database and runs migrations.
func NewSqlite(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite: %w", err)
	}
	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate sqlite: %w", err)
	}
	return db, nil
}

// NewSqliteStore creates a SqliteStore from an existing *sql.DB.
func NewSqliteStore(db *sql.DB) *SqliteStore {
	return &SqliteStore{db: db}
}

//go:embed sqlite.sql
var schemaSQL string

func migrate(db *sql.DB) error {
	stmts := splitStatements(schemaSQL)
	for i, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			msg := err.Error()
			if strings.Contains(msg, "already exists") ||
				strings.Contains(msg, "duplicate column") ||
				strings.Contains(msg, "duplicate record") {
				continue
			}
			return fmt.Errorf("migration failed: %w\nSQL: %s", err, truncateSQL(stmt, 200))
		}
		if i < 3 {
			fmt.Printf("[MIGRATE] [%d] %s\n", i, stmt[:min(len(stmt), 60)])
		}
	}
	return nil
}

func min(a, b int) int { if a < b { return a }; return b }

// debugSchema prints the first 500 chars of the embedded schema for debugging.
func debugSchema() {
	fmt.Printf("[DEBUG] schemaSQL length: %d\n", len(schemaSQL))
	fmt.Printf("[DEBUG] First 200: %q\n", truncateSQL(schemaSQL, 200))
}

func splitStatements(sql string) []string {
	var parts []string
	var buf strings.Builder
	inQuote := false
	depth := 0
	b := []byte(sql)
	i := 0
	for i < len(b) {
		r, size := utf8.DecodeRune(b[i:])
		if r == utf8.RuneError && size == 1 {
			buf.WriteByte(b[i])
			i++
			continue
		}

		// Treat line-start comments as statement boundaries — skip them entirely
		if r == '-' && i+1 < len(b) && b[i+1] == '-' {
			if i == 0 || b[i-1] == '\n' {
				// Flush current buffer
				if s := strings.TrimSpace(buf.String()); s != "" {
					parts = append(parts, s)
				}
				buf.Reset()
				// Skip the comment line
				for i < len(b) && b[i] != '\n' {
					i++
				}
				if i < len(b) {
					i++ // skip newline
				}
				continue
			}
		}

		switch r {
		case '\'':
			inQuote = !inQuote
			buf.WriteRune(r)
		case '(':
			if !inQuote {
				depth++
			}
			buf.WriteRune(r)
		case ')':
			if !inQuote && depth > 0 {
				depth--
			}
			buf.WriteRune(r)
		case ';':
			if !inQuote && depth == 0 {
				parts = append(parts, buf.String())
				buf.Reset()
			} else {
				buf.WriteRune(r)
			}
		default:
			buf.WriteRune(r)
		}
		i += size
	}
	if s := strings.TrimSpace(buf.String()); s != "" {
		parts = append(parts, s)
	}
	return parts
}

func truncateSQL(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...(truncated)"
}
