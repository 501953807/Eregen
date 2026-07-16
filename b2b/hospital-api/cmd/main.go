package main

import (
	"context"
	"os"
	"time"

	"eregen.dev/b2b-hospital-api/internal/handler"
	"eregen.dev/b2b-hospital-api/internal/middleware"
	"eregen.dev/b2b-hospital-api/internal/router"
	"eregen.dev/b2b-hospital-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	log, _ := zap.NewProduction()
	defer log.Sync()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5432/eregen_b2b?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal("failed to create connection pool", zap.Error(err))
	}
	defer pool.Close()

	engine := gin.Default()
	pgStore := store.NewPostgres(pool, log)
	router.Register(engine, pgStore, log)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	log.Info("starting b2b hospital API", zap.String("port", port))
	engine.Run(":" + port)
}
