package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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

	// Seed demo data if database is empty (for development/demo purposes)
	if err := seedDatabase(db); err != nil {
		log.Printf("failed to seed database: %v", err)
	}

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

// seedDatabase inserts demo data if the tables are empty.
// This is for development/demo purposes only.
func seedDatabase(db *sql.DB) error {
	// Check if there are any users - if yes, skip seeding
	var count int
	err := db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return fmt.Errorf("check user count: %w", err)
	}
	if count > 0 {
		log.Printf("Database already has %d users, skipping seeding", count)
		return nil
	}

	log.Printf("Seeding demo database...")

	// Insert admin user
	adminID := "admin-user-001"
	_, err = db.ExecContext(context.Background(),
		`INSERT INTO users (id, name, email, role, password_hash, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		adminID, "系统管理员", "admin@eregen.com", "admin", "$2a$10$sYd8gEoGUA0O9fBB/jlfEeQv9CuyHoKgaH.qDWgOfBpSoT1Kh8Yba",
	)
	if err != nil && !isUniqueConstraintError(err) {
		return fmt.Errorf("insert admin user: %w", err)
	}

	// Insert elderly user and profile
	elderly1ID := "elderly-001"
	user1ID := "user-001"
	_, err = db.ExecContext(context.Background(),
		`INSERT INTO users (id, name, email, role, password_hash, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		user1ID, "张大爷", "zhang@example.com", "user", "$2a$10$JVMdHOp3Ect5e6WY7m3wpeJMDIM/iUjXvt7OAYYM9U6dJe0qvFkHe",
	)
	if err != nil && !isUniqueConstraintError(err) {
		return fmt.Errorf("insert user 1: %w", err)
	}

	_, err = db.ExecContext(context.Background(),
		`INSERT INTO elderly_profiles (id, name, user_id, birth_date, health_tiers, created_at, updated_at)
		 VALUES (?, ?, ?, datetime('1950-01-01'), '["cardiovascular","diabetes"]', datetime('now'), datetime('now'))`,
		elderly1ID, "张建国", user1ID,
	)
	if err != nil && !isUniqueConstraintError(err) {
		return fmt.Errorf("insert elderly profile 1: %w", err)
	}

	// Insert devices for elderly1 - using simplified version that matches CREATE TABLE
	device1ID := "device-001" // bracelet
	device2ID := "device-002" // pillbox
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = db.ExecContext(context.Background(),
		`INSERT INTO devices (id, device_id, device_type, tier, status, last_seen, owner_user_id, settings, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		device1ID, "BR-ZHANG001", "bracelet", "plus", "online", now, user1ID, "{\"fw_version\":\"v1.2.3\"}",
	)
	if err != nil && !isUniqueConstraintError(err) {
		return fmt.Errorf("insert device 1: %w", err)
	}

	_, err = db.ExecContext(context.Background(),
		`INSERT INTO devices (id, device_id, device_type, tier, status, last_seen, owner_user_id, settings, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		device2ID, "PX-ZHANG001", "pillbox", "pro", "online", now, user1ID, "{\"fw_version\":\"v2.1.0\"}",
	)
	if err != nil && !isUniqueConstraintError(err) {
		return fmt.Errorf("insert device 2: %w", err)
	}

	// Link elderly to devices via elderly_devices table
	_, err = db.ExecContext(context.Background(),
		`INSERT INTO elderly_devices (id, elderly_id, device_id, created_at)
		 VALUES (?, ?, ?, datetime('now'))`,
		"eld-dev-001", elderly1ID, device1ID,
	)
	if err != nil && !isUniqueConstraintError(err) {
		return fmt.Errorf("link device 1: %w", err)
	}
	_, err = db.ExecContext(context.Background(),
		`INSERT INTO elderly_devices (id, elderly_id, device_id, created_at)
		 VALUES (?, ?, ?, datetime('now'))`,
		"eld-dev-002", elderly1ID, device2ID,
	)
	if err != nil && !isUniqueConstraintError(err) {
		return fmt.Errorf("link device 2: %w", err)
	}

	// Insert a few alerts
	alert1ID := "alert-001"
	_, err = db.ExecContext(context.Background(),
		`INSERT INTO alerts (id, elderly_id, alert_type, severity, status, message, device_id, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, datetime('now'))`,
		alert1ID, elderly1ID, "sos", "high", "pending", "老人按下SOS按钮", device1ID,
	)
	if err != nil && !isUniqueConstraintError(err) {
		return fmt.Errorf("insert alert 1: %w", err)
	}

	_, err = db.ExecContext(context.Background(),
		`INSERT INTO alerts (id, elderly_id, alert_type, severity, status, message, device_id, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, datetime('now'))`,
		"alert-002", elderly1ID, "fall", "medium", "resolved", "检测到跌倒，已处理", device1ID,
	)
	if err != nil && !isUniqueConstraintError(err) {
		return fmt.Errorf("insert alert 2: %w", err)
	}

	// Insert health records
	nowTime := time.Now()
	for i := 0; i < 5; i++ {
		ts := nowTime.Add(time.Duration(-i)*time.Hour)
		_, err = db.ExecContext(context.Background(),
			`INSERT INTO health_records (id, elderly_id, timestamp, hr, spo2, steps, sleep_hours)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			"hr-"+fmt.Sprintf("%d", i), elderly1ID, ts.Format("2006-01-02 15:04:05"), 72+i, 98-i, randomSteps(float64(i)), nil)
		if err != nil && !isUniqueConstraintError(err) {
			// Don't fail if one record fails, continue with others
		}
	}

	log.Println("Demo seeding completed successfully")
	return nil
}

// isUniqueConstraintError checks if the error is about a unique constraint violation.
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToUpper(err.Error()), "UNIQUE") || strings.Contains(err.Error(), "constraint")
}

// randomSteps generates varying step counts for demo health records.
func randomSteps(offset float64) *int {
	steps := 3000 + int(offset*1000)
	return &steps
}