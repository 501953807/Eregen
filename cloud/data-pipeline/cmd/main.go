package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"eregen.dev/pipeline/internal/analyzer"
	"eregen.dev/pipeline/internal/config"
	"eregen.dev/pipeline/internal/store"
	"eregen.dev/pipeline/internal/subscriber"

	"github.com/nats-io/nats.go"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load: %v", err)
	}

	log.Println("data-pipeline starting...")

	// Database connections
	db, err := store.NewStore(cfg)
	if err != nil {
		log.Fatalf("store init: %v", err)
	}
	defer db.Close()

	// AI analyzers
	healthAnalyzer := analyzer.NewHealthAnalyzer(cfg.BaselineDays)
	riskCalc := analyzer.NewRiskScoreCalculator(
		cfg.RiskVitalsWeight,
		cfg.RiskMedWeight,
		cfg.RiskActivityWt,
		cfg.RiskSleepWt,
	)

	// NATS connection for JetStream event processing
	nc, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		log.Fatalf("nats connect: %v", err)
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("nats jetstream: %v", err)
	}

	hAnalyzer := subscriber.NewHandler(js, healthAnalyzer, riskCalc, db)
	if err := hAnalyzer.Start(); err != nil {
		log.Fatalf("nats subscriber start: %v", err)
	}
	log.Println("NATS subscriber started on eregen.event.>")

	// HTTP health endpoint
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"pipeline"}`))
	})

	httpServer := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// Graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		<-ctx.Done()
		log.Println("shutting down...")
		shutdownCtx, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		httpServer.Shutdown(shutdownCtx)
	}()

	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("http server: %v", err)
	}

	log.Println("stopped")
}
