// © 2026 Eregen (颐贞). All rights reserved.

package mqtt

import (
	"log"
	"strings"

	"eregen.dev/gateway/internal/handler"
	"eregen.dev/gateway/internal/store"
)

// TopicRouter subscribes to device topics and routes parsed messages through the handler.
type TopicRouter struct {
	mqttClient *Client
	handler    *handler.Handler
	db         *store.Store
}

// NewTopicRouter creates a router that bridges MQTT device topics to the handler pipeline.
func NewTopicRouter(mqttClient *Client, h *handler.Handler, db *store.Store) *TopicRouter {
	return &TopicRouter{
		mqttClient: mqttClient,
		handler:    h,
		db:         db,
	}
}

// Start subscribes to device uplink topics and begins routing.
func (r *TopicRouter) Start() error {
	topics := []string{
		"eregen/device/bracelet/+/up",
		"eregen/device/pillbox/+/up",
		"eregen/medical/wb/+/up",
		"eregen/community/wb/+/up",
	}

	for _, topic := range topics {
		if err := r.mqttClient.Subscribe(topic, OnMessage(r.handler, r.db)); err != nil {
			return err
		}
		log.Printf("Subscribed to topic: %s", topic)
	}

	return nil
}

// extractDeviceID pulls the device ID from an MQTT topic path.
// "eregen/device/bracelet/BR-1234/up" -> "BR-1234"
func extractDeviceID(topic string) string {
	parts := strings.Split(topic, "/")
	if len(parts) >= 5 {
		return parts[3]
	}
	return ""
}
