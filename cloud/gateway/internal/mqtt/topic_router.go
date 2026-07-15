// © 2026 Eregen (颐贞). All rights reserved.

package mqtt

import (
	"fmt"
	"log"
	"strings"

	"eregen/cloud/gateway/internal/nats"
)

// TopicRouter subscribes to device topics and routes parsed messages to NATS.
type TopicRouter struct {
	mqttClient *Client
	natsClient *nats.Client
	handlers   map[string]eventHandler
}

// NewTopicRouter creates a router that bridges MQTT device topics to NATS events.
func NewTopicRouter(mqttClient *Client, natsClient *nats.Client) *TopicRouter {
	return &TopicRouter{
		mqttClient: mqttClient,
		natsClient: natsClient,
		handlers: map[string]eventHandler{
			"heartbeat":         handleHeartbeat,
			"location":          handleLocation,
			"health":            handleHealth,
			"sos":               handleSOS,
			"fall":              handleFall,
			"med_status":        handleMedStatus,
			"fence_alert":       handleFenceAlert,
			"inventory_warning": handleInventoryWarning,
		},
	}
}

// Start subscribes to device uplink topics and begins routing.
func (r *TopicRouter) Start() error {
	topics := []string{
		"eregen/device/bracelet/+/up",
		"eregen/device/pillbox/+/up",
	}

	for _, topic := range topics {
		if err := r.mqttClient.Subscribe(topic, r.onMessage); err != nil {
			return fmt.Errorf("subscribe to %s: %w", topic, err)
		}
		log.Printf("Subscribed to topic: %s", topic)
	}

	return nil
}

// onMessage dispatches incoming MQTT messages after extracting device ID and parsing.
func (r *TopicRouter) onMessage(rawTopic string, payload []byte) {
	deviceID := extractDeviceID(rawTopic)
	if deviceID == "" {
		log.Printf("WARN: could not extract device ID from topic: %s", rawTopic)
		return
	}

	msg, err := ParseAndValidate(payload)
	if err != nil {
		log.Printf("WARN: invalid message from %s: %v", deviceID, err)
		return
	}

	handler, ok := r.handlers[msg.Type]
	if !ok {
		log.Printf("WARN: unknown message type %q from %s", msg.Type, deviceID)
		return
	}

	handler(r.natsClient, deviceID, msg.Raw)
}

// extractDeviceID pulls the device ID from an MQTT topic path.
// "eregen/device/bracelet/BR-1234/up" -> "BR-1234"
func extractDeviceID(topic string) string {
	parts := strings.Split(topic, "/")
	// Expected: eregen / device / {bracelet,pillbox} / {dev_id} / up
	if len(parts) >= 5 {
		return parts[3]
	}
	return ""
}
