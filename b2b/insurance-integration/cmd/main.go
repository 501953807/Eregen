package main

import (
	"os"
	"path/filepath"

	"eregen.dev/b2b-insurance-integration/internal/router"
	"eregen.dev/b2b-insurance-integration/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	log, _ := zap.NewProduction()
	defer log.Sync()

	dbType := os.Getenv("DATABASE_TYPE")
	if dbType == "" {
		dbType = "sqlite"
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = getEnvFallback("POSTGRES_DSN", "postgres://eregen:eregen@localhost:5432/eregen?sslmode=disable")
	}
	sqlitePath := os.Getenv("SQLITE_PATH")
	if sqlitePath == "" {
		sqlitePath = "./data/eregen.db"
	}

	var st store.Store
	var closer func() error
	switch dbType {
	case "postgres":
		pool, err := pgxpool.New(nil, dbURL)
		if err != nil {
			log.Fatal("failed to connect to postgres", zap.Error(err))
		}
		st = store.NewPostgresStore(pool, log)
		closer = func() error { pool.Close(); return nil }
	default:
		if err := os.MkdirAll(filepath.Dir(sqlitePath), 0755); err != nil {
			log.Fatal("failed to create sqlite dir", zap.Error(err))
		}
		db, err := store.NewSqlite(sqlitePath)
		if err != nil {
			log.Fatal("failed to init sqlite", zap.Error(err))
		}
		st = db
		closer = func() error { return db.Close() }
	}
	defer closer()

	engine := gin.Default()
	router.Register(engine, st, log)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}
	log.Info("starting b2b insurance integration API", zap.String("port", port), zap.String("db_type", dbType))
	if err := engine.Run(":" + port); err != nil {
		log.Fatal("server failed", zap.Error(err))
	}
}

func getEnvFallback(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
