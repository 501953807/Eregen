// Package main is the entry point for the MQTT gateway service.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"eregen.dev/gateway/internal/config"
	"eregen.dev/gateway/internal/handler"
	"eregen.dev/gateway/internal/mqtt"
	"eregen.dev/gateway/internal/nats"
	"eregen.dev/gateway/internal/store"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- Database (PostgreSQL or SQLite) ---
	var dbStore *store.Store
	var dbErr error
	switch cfg.Storage.Type {
	case "sqlite":
		dbStore, dbErr = store.NewSQLite(cfg.Storage.SQLite)
		if dbErr != nil {
			log.Fatalf("Failed to connect to SQLite: %v", dbErr)
		}
		log.Println("Connected to SQLite")
	default:
		dbStore, dbErr = store.NewPostgres(ctx, cfg.Storage.DSN)
		if dbErr != nil {
			log.Fatalf("Failed to connect to PostgreSQL: %v", dbErr)
		}
		log.Println("Connected to PostgreSQL")
	}

	// --- NATS JetStream ---
	natsClient := nats.NewClient(nats.Config{
		URL:             cfg.NATS.URL,
		JetStreamDomain: cfg.NATS.JetStreamDomain,
		StreamName:      cfg.NATS.StreamName,
		GatewayID:       "gateway-1",
	})
	if err := natsClient.Connect(); err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer natsClient.Close()

	// --- Handler pipeline ---
	h := handler.New(natsClient, dbStore)

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

	router := mqtt.NewTopicRouter(mqttClient, h, dbStore)
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
	dbStore.Close()

	log.Println("Gateway stopped")
}
