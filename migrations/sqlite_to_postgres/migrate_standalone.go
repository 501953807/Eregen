package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s <sqlite_db_path> <postgres_dsn> <output_dir>\n", os.Args[0])
		os.Exit(1)
	}

	sqlitePath := os.Args[1]
	pgDSN := os.Args[2]
	outputDir := os.Args[3]

	fmt.Printf("SQLite: %s\n", sqlitePath)
	fmt.Printf("PostgreSQL: %s\n", pgDSN)
	fmt.Printf("Output: %s\n", outputDir)
	fmt.Println("Migration script placeholder - run: go run main.go " + sqlitePath + " " + pgDSN + " " + outputDir)
}
