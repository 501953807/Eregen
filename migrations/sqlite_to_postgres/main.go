package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s <sqlite_db_path> <postgres_dsn> <output_dir>\n", os.Args[0])
		os.Exit(1)
	}

	sqlitePath := os.Args[1]
	pgDSN := os.Args[2]
	outputDir := os.Args[3]

	// Open SQLite database
	sqliteDB, err := sql.Open("sqlite3", sqlitePath)
	if err != nil {
		log.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer sqliteDB.Close()

	// Connect to PostgreSQL (using pgx)
	pgDB, err := sql.Open("postgres", pgDSN)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pgDB.Close()

	if err := pgDB.Ping(); err != nil {
		log.Fatalf("Failed to ping PostgreSQL: %v", err)
	}

	fmt.Println("Connected to both databases successfully")

	// Get all tables from SQLite
	rows, err := sqliteDB.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		log.Fatalf("Failed to list tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatalf("Failed to scan table name: %v", err)
		}
		tables = append(tables, name)
	}

	fmt.Printf("Found %d tables to migrate\n", len(tables))

	// Migrate each table
	for _, tableName := range tables {
		fmt.Printf("Migrating table: %s\n", tableName)
		if err := migrateTable(sqliteDB, pgDB, tableName, outputDir); err != nil {
			log.Printf("Warning: Failed to migrate table %s: %v\n", tableName, err)
			continue
		}
	}

	fmt.Println("Migration completed!")
}

func migrateTable(sqliteDB, pgDB *sql.DB, tableName, outputDir string) error {
	// Get SQLite table schema
	schemaRows, err := sqliteDB.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, tableName))
	if err != nil {
		return fmt.Errorf("failed to get table info: %v", err)
	}
	defer schemaRows.Close()

	type Column struct {
		CID        int
		Name       string
		Type       string
		NotNul     int
		DefaultVal sql.NullString
		PrimaryKey int
	}

	var columns []Column
	for schemaRows.Next() {
		var c Column
		if err := schemaRows.Scan(&c.CID, &c.Name, &c.Type, &c.NotNul, &c.DefaultVal, &c.PrimaryKey); err != nil {
			return fmt.Errorf("failed to scan column: %v", err)
		}
		columns = append(columns, c)
	}

	// Build PostgreSQL CREATE TABLE statement
	var pgCols []string
	for _, col := range columns {
		pgType := convertType(col.Type)
		def := ""
		if col.DefaultVal.Valid {
			def = " DEFAULT " + convertDefault(col.DefaultVal.String)
		}
		notNull := ""
		if col.NotNul != 0 {
			notNull = " NOT NULL"
		}
		pk := ""
		if col.PrimaryKey != 0 {
			pk = " PRIMARY KEY"
		}
		pgCols = append(pgCols, fmt.Sprintf("    %s %s%s%s%s", col.Name, pgType, notNull, def, pk))
	}

	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n%s\n);",
		convertTableName(tableName),
		strings.Join(pgCols, ",\n"))

	// Write CREATE TABLE to file
	createFile := fmt.Sprintf("%s/%s_create.sql", outputDir, convertTableName(tableName))
	if err := os.WriteFile(createFile, []byte(createSQL+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to write create file: %v", err)
	}

	// Execute CREATE TABLE in PostgreSQL
	if _, err := pgDB.Exec(createSQL); err != nil {
		return fmt.Errorf("failed to create table in PostgreSQL: %v", err)
	}

	// Read data from SQLite
	dataRows, err := sqliteDB.Query(fmt.Sprintf("SELECT * FROM %s", tableName))
	if err != nil {
		return fmt.Errorf("failed to query data: %v", err)
	}
	defer dataRows.Close()

	columnNames, err := dataRows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %v", err)
	}

	// Build INSERT statement
	var insertBuilder strings.Builder
	insertBuilder.WriteString(fmt.Sprintf("INSERT INTO %s (", convertTableName(tableName)))
	insertBuilder.WriteString(strings.Join(columnNames, ", "))
	insertBuilder.WriteString(") VALUES (")

	// Create placeholders
	placeholders := make([]string, len(columnNames))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	insertBuilder.WriteString(strings.Join(placeholders, ", "))
	insertBuilder.WriteString(")")

	insertSQL := insertBuilder.String()

	// Read and insert rows
	rowCount := 0
	for dataRows.Next() {
		values := make([]interface{}, len(columnNames))
		valuePtrs := make([]interface{}, len(columnNames))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := dataRows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("failed to scan row: %v", err)
		}

		// Convert values
		for i, col := range columnNames {
			values[i] = convertValue(col, values[i])
		}

		if _, err := pgDB.Exec(insertSQL, values...); err != nil {
			return fmt.Errorf("failed to insert row: %v", err)
		}
		rowCount++
	}

	// Write INSERT script to file
	insertFile := fmt.Sprintf("%s/%s_insert.sql", outputDir, convertTableName(tableName))
	if err := os.WriteFile(insertFile, []byte(insertSQL), 0644); err != nil {
		return fmt.Errorf("failed to write insert file: %v", err)
	}

	fmt.Printf("  Migrated %d rows from %s\n", rowCount, tableName)
	return nil
}

func convertType(sqliteType string) string {
	switch strings.ToUpper(sqliteType) {
	case "INTEGER":
		return "SERIAL"
	case "TEXT":
		return "TEXT"
	case "REAL":
		return "NUMERIC"
	case "DATETIME", "TIMESTAMP", "TIMESTAMPTZ":
		return "TIMESTAMPTZ"
	case "DATE":
		return "DATE"
	case "BOOLEAN", "BOOL":
		return "BOOLEAN"
	case "BLOB":
		return "BYTEA"
	default:
		return "TEXT"
	}
}

func convertDefault(value string) string {
	switch strings.ToLower(value) {
	case "current_timestamp":
		return "now()"
	case "datetime('now')":
		return "now()"
	default:
		return value
	}
}

func convertTableName(name string) string {
	return strings.ToLower(name)
}

func convertValue(columnName string, value interface{}) interface{} {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []byte:
		if strings.EqualFold(columnName, "id") || strings.HasSuffix(columnName, "_id") {
			if len(v) == 32 {
				return fmt.Sprintf("%s-%s-%s-%s-%s", v[0:8], v[8:12], v[12:16], v[16:20], v[20:32])
			}
		}
		return string(v)
	case string:
		if strings.EqualFold(columnName, "id") || strings.HasSuffix(columnName, "_id") {
			if len(v) == 36 && v[8] == '-' && v[13] == '-' && v[18] == '-' && v[23] == '-' {
				return v
			}
			if len(v) == 32 {
				return fmt.Sprintf("%s-%s-%s-%s-%s", v[0:8], v[8:12], v[12:16], v[16:20], v[20:32])
			}
		}
		return v
	case int64:
		if strings.EqualFold(columnName, "active") || strings.EqualFold(columnName, "enabled") ||
			strings.EqualFold(columnName, "matched") || strings.EqualFold(columnName, "sent") ||
			strings.EqualFold(columnName, "violated") || strings.EqualFold(columnName, "force_update") {
			return v != 0
		}
		return int(v)
	case float64:
		return v
	default:
		return v
	}
}
