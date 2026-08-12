#!/bin/bash
# SQLite to PostgreSQL Migration Script
# Usage: ./migrate.sh <sqlite_db_path> <postgres_dsn> [output_dir]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SQLITE_DB="${1:?Usage: $0 <sqlite_db_path> <postgres_dsn> [output_dir]}"
PG_DSN="${2:?Usage: $0 <sqlite_db_path> <postgres_dsn> [output_dir]}"
OUTPUT_DIR="${3:-${SCRIPT_DIR}/output}"

echo "=== Eregen SQLite to PostgreSQL Migration ==="
echo "SQLite database: ${SQLITE_DB}"
echo "PostgreSQL DSN: ${PG_DSN}"
echo "Output directory: ${OUTPUT_DIR}"
echo ""

# Check if SQLite database exists
if [ ! -f "${SQLITE_DB}" ]; then
    echo "Error: SQLite database not found: ${SQLITE_DB}"
    exit 1
fi

# Create output directory
mkdir -p "${OUTPUT_DIR}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go to run the migration."
    exit 1
fi

# Build the migration tool
echo "Building migration tool..."
cd "${SCRIPT_DIR}"
go build -o migrate_tool main.go
if [ $? -ne 0 ]; then
    echo "Error: Failed to build migration tool"
    exit 1
fi

# Run the migration
echo "Running migration..."
./migrate_tool "${SQLITE_DB}" "${PG_DSN}" "${OUTPUT_DIR}"
MIGRATE_EXIT_CODE=$?

# Cleanup
rm -f migrate_tool

if [ ${MIGRATE_EXIT_CODE} -eq 0 ]; then
    echo ""
    echo "=== Migration completed successfully ==="
    echo "Output files saved to: ${OUTPUT_DIR}"
    echo ""
    echo "Generated files:"
    ls -la "${OUTPUT_DIR}"/*.sql 2>/dev/null || echo "No SQL files generated"
else
    echo ""
    echo "=== Migration completed with warnings ==="
    echo "Check the output directory for partial results"
fi

exit ${MIGRATE_EXIT_CODE}
