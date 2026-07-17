package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

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

	log.Printf("push-service starting (port=%d)", cfg.Port)

	// PostgreSQL connection for family member lookup
	db, err := sql.Open("postgres", cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("postgres connect: %v", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err = db.Ping(); err != nil {
		log.Fatalf("postgres ping: %v", err)
	}
	defer db.Close()

	pgStore := store.NewPostgres(db)

	// NATS subscriber for device events from gateway
	natsSub, err := publisher.NewSubscriber(cfg.NATSURL)
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

	// HTTP health endpoint
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"push"}`))
	})

	httpServer := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// Start NATS subscriber in background with DB store
	go func() {
		natsSub.Start(rtr, pgStore)
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
