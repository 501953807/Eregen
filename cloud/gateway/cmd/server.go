// © 2026 Eregen (颐贞). All rights reserved.

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"eregen/cloud/gateway/config"
	"eregen/cloud/gateway/internal/mqtt"
	"eregen/cloud/gateway/internal/nats"
)

func main() {
	cfg := config.Load()

	log.Println("Starting Eregen MQTT Gateway...")

	// Connect to EMQX
	mqttCfg := mqtt.MQTTConfig{
		Broker:    cfg.MQTT.Broker,
		ClientID:  cfg.MQTT.ClientID,
		Username:  cfg.MQTT.Username,
		Password:  cfg.MQTT.Password,
		TLS:       mqtt.TLSConfig(cfg.MQTT.TLS),
		KeepAlive: cfg.MQTT.KeepAlive,
	}
	mqttClient := mqtt.NewClient(mqttCfg)
	if err := mqttClient.Connect(); err != nil {
		log.Fatalf("Failed to connect to EMQX: %v", err)
	}
	defer mqttClient.Disconnect()

	// Connect to NATS
	natsClient := nats.NewClient(nats.Config{
		URL:       cfg.NATS.URL,
		GatewayID: "gateway-1",
	})
	if err := natsClient.Connect(); err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer natsClient.Close()

	// Start message router
	router := mqtt.NewTopicRouter(mqttClient, natsClient)
	if err := router.Start(); err != nil {
		log.Fatalf("Failed to start topic router: %v", err)
	}

	log.Println("MQTT Gateway started successfully")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down MQTT Gateway...")
}
