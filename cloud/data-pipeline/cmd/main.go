package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"eregen.dev/pipeline/internal/analyzer"
	"eregen.dev/pipeline/internal/config"
	"eregen.dev/pipeline/internal/store"
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
	_ = healthAnalyzer
	_ = riskCalc

	// NATS subscriber for device events
	// Connect to NATS and start processing
	log.Println("listening for device events...")

	// HTTP health endpoint
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"pipeline"}`))
	})

	httpServer := &http.Server{
		Addr:         ":8087",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

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
