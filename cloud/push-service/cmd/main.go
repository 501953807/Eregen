package main

import (
	"context"
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
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load: %v", err)
	}

	log.Printf("push-service starting (port=%d)", cfg.Port)

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
