// Package store provides a shared SQLite migration helper.
// Services embed schema/sqlite.sql via go:embed and execute it at startup.
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// LoadSchemaPath returns the absolute path to schema/sqlite.sql relative
// to the project root. It is resolved from the binary's working directory
// or falling back to the module root.
func LoadSchemaPath() string {
	candidates := []string{
		"schema/sqlite.sql",
		"../../schema/sqlite.sql",
		"../../../schema/sqlite.sql",
	}
	for _, c := range candidates {
		abs, err := filepath.Abs(c)
		if err != nil {
			continue
		}
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}
	return "schema/sqlite.sql"
}

// RunMigrations reads the unified schema SQL and executes each statement
// individually so that CREATE TABLE IF NOT EXISTS / INSERT OR IGNORE are
// naturally idempotent. Errors other than "already exists" / "duplicate"
// are returned immediately.
func RunMigrations(db *sql.DB) error {
	schemaPath := LoadSchemaPath()
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}

	stmts := splitStatements(string(data))
	for i, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			if isIgnorable(err) {
				continue
			}
			return fmt.Errorf("migration %d failed: %w\nSQL: %s", i, err, truncate(stmt, 200))
		}
	}
	return nil
}

// splitStatements splits raw SQL by semicolons, respecting single-quoted
// strings and parenthesized expressions. Line-start comments (-- ...) are
// skipped entirely so they don't interfere with statement parsing.
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

		// Treat line-start comments as statement boundaries — skip them
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

func isIgnorable(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "already exists") ||
		strings.Contains(msg, "duplicate column") ||
		strings.Contains(msg, "table") && strings.Contains(msg, "exists") ||
		strings.Contains(msg, "duplicate record")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...(truncated)"
}
