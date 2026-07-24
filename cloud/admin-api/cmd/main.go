package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"eregen.dev/admin-api/internal/config"
	"eregen.dev/admin-api/internal/router"
	"eregen.dev/admin-api/internal/store"

	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	var db *sql.DB
	var err error
	switch cfg.DatabaseType {
	case "postgres":
		db = store.NewPostgres(cfg.DatabaseURL)
	default: // sqlite (default)
		db, err = store.NewSqlite(cfg.SQLitePath)
		if err != nil {
			log.Fatalf("sqlite init failed: %v", err)
		}
	}
	defer db.Close()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	r := router.Setup(db, logger, cfg.DatabaseType)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("admin-api starting on :%s (db=%s)", cfg.Port, cfg.DatabaseType)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced shutdown: %v", err)
	}
	log.Println("server exited")
}
