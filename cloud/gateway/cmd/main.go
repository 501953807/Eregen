// © 2026 Eregen (颐贞). All rights reserved.

// Package main is the entry point for the MQTT gateway service.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"eregen.dev/gateway/internal/config"
	"eregen.dev/gateway/internal/handler"
	"eregen.dev/gateway/internal/mqtt"
	"eregen.dev/gateway/internal/nats"
	"eregen.dev/gateway/internal/store"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/redis/go-redis/v9"
)

const banner = `
  ___   ___  ___  ___  _______  ___      ___  ___  ___  ___   ___
 / _ \ / _ \/ __ \/ _ \/ __/ _ \/ _ \    / _ \/ _ \/ _ \/ _ \ / _ \
/ // // // /\ \/ // // //_/ // // , \  / // // // // // // // // /
\__/ \\___/ /___/ \___/\____/ /_/ \_\ /____/\___/\___/\___/\___/
  Eregen Cloud Gateway — MQTT → NATS JetStream
`

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	log.Print(banner)

	cfg := config.Load()

	// Logger level
	if cfg.LogLevel != "" {
		// glog-level control via env; default info is fine for now.
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- PostgreSQL ---
	dbStore, err := store.New(ctx, cfg.Postgres.DSN)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	// --- Redis ---
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to ping Redis: %v", err)
	}
	log.Println("Connected to Redis")

	// --- InfluxDB ---
	influxClient := influxdb2.NewClient(cfg.InfluxDB.URL, cfg.InfluxDB.Token)
	if _, err := influxClient.Ping(ctx); err != nil {
		log.Printf("WARN: InfluxDB unavailable: %v (will retry on write)", err)
	}
	defer influxClient.Close()
	log.Println("Connected to InfluxDB")

	// --- NATS JetStream ---
	natsClient := nats.NewClient(nats.Config{
		URL:           cfg.NATS.URL,
		JetStreamDomain: cfg.NATS.JetStreamDomain,
		StreamName:    cfg.NATS.StreamName,
		GatewayID:     "gateway-1",
	})
	if err := natsClient.Connect(); err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer natsClient.Close()

	// --- Handler pipeline ---
	h := handler.New(natsClient, dbStore, redisClient, influxClient, cfg.InfluxDB.Bucket, cfg.InfluxDB.Org)

	// --- MQTT ---
	mqttCfg := &mqtt.Config{
		Broker:    cfg.MQTT.Broker,
		ClientID:  cfg.MQTT.ClientID,
		Username:  cfg.MQTT.Username,
		Password:  cfg.MQTT.Password,
		TLS: mqtt.TLSConfig{
				Enabled: cfg.MQTT.TLS.Enabled,
				CACert:  cfg.MQTT.TLS.CACert,
				Cert:    cfg.MQTT.TLS.Cert,
				Key:     cfg.MQTT.TLS.Key,
			},
		KeepAlive: cfg.MQTT.KeepAlive,
	}
	mqttClient := mqtt.NewClient(mqttCfg)
	if err := mqttClient.Connect(); err != nil {
		log.Fatalf("Failed to connect to EMQX: %v", err)
	}
	defer mqttClient.Disconnect()

	router := mqtt.NewTopicRouter(mqttClient, h)
	if err := router.Start(); err != nil {
		log.Fatalf("Failed to start topic router: %v", err)
	}

	log.Println("MQTT Gateway started successfully")

	// --- Graceful shutdown ---
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
	cancel()

	// Give in-flight operations time to complete.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	dbStore.Close()
	redisClient.Shutdown(shutdownCtx).Err()

	log.Println("Gateway stopped")
}
