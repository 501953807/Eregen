package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	_ "modernc.org/sqlite"

	smsChannel "eregen.dev/push/internal/channel/sms"
	wechatChannel "eregen.dev/push/internal/channel/wechat"
	"eregen.dev/push/internal/config"
	"eregen.dev/push/internal/fcm"
	"eregen.dev/push/internal/publisher"
	"eregen.dev/push/internal/router"
	"eregen.dev/push/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load: %v", err)
	}

	log.Printf("push-service starting (port=%d, db=%s)", cfg.Port, cfg.StorageType)

	// Database connection
	var db *sql.DB
	switch cfg.StorageType {
	case "postgres":
		db, err = sql.Open("postgres", cfg.PostgresDSN)
		if err != nil {
			log.Fatalf("postgres connect: %v", err)
		}
	default:
		// Expand ~ in path
		sqlitePath := cfg.SQLitePath
		if len(sqlitePath) > 1 && sqlitePath[0] == '~' {
			home, _ := os.UserHomeDir()
			sqlitePath = filepath.Join(home, sqlitePath[2:])
		}
		// Ensure directory exists
		if dir := filepath.Dir(sqlitePath); dir != "" {
			if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
				log.Fatalf("mkdir sqlite path: %v", mkErr)
			}
		}
		db, err = sql.Open("sqlite", sqlitePath)
		if err != nil {
			log.Fatalf("sqlite connect: %v", err)
		}
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err = db.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}
	defer db.Close()

	pgStore, err := store.NewStore(cfg.StorageType, cfg.PostgresDSN, cfg.SQLitePath)

	// NATS subscriber for device events from gateway — passes pgStore for DB lookups
	natsSub, err := publisher.NewSubscriber(cfg.NATSURL, pgStore)
	if err != nil {
		log.Fatalf("nats connect: %v", err)
	}
	defer natsSub.Close()

	// Channel clients
	fcmClient := fcm.NewClient()
	wechatClient := wechatChannel.NewWeChatClient(cfg.WeChatAppID, cfg.WeChatAppSecret)
	smsClient := smsChannel.NewSMSClient(cfg.SMSAccessKey, cfg.SMSAccessSecret, cfg.SSignName)

	// Channel router — fan-out to all channels
	rtr := router.NewRouter(fcmClient, wechatClient, smsClient)

	// HTTP health endpoint (standard format for unified launch system)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"status":"ok"}}`))
	})

	httpServer := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// Start NATS subscriber in background
	go func() {
		natsSub.Start(rtr)
	}()

	// Graceful shutdown
	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
		<-sigch
		log.Println("shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		httpServer.Shutdown(ctx)
	}()

	log.Fatal(httpServer.ListenAndServe())
}
