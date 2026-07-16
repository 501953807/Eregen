// © 2026 Eregen (颐贞). All rights reserved.

package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"eregen.dev/gateway/internal/handler"
	"eregen.dev/gateway/internal/model"
)

// ParsedMessage holds a validated incoming MQTT message.
type ParsedMessage struct {
	Type      model.UpstreamMessageType `json:"type"`
	DeviceID  string                    `json:"dev_id"`
	Timestamp int64                     `json:"ts"`
	Raw       json.RawMessage           `json:"-"`
}

// ParseAndValidate parses an MQTT payload and validates required fields.
func ParseAndValidate(payload []byte) (*ParsedMessage, error) {
	var generic map[string]interface{}
	if err := json.Unmarshal(payload, &generic); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	typ, ok := generic["type"].(string)
	if !ok || typ == "" {
		return nil, fmt.Errorf("missing or invalid 'type' field")
	}

	devID, ok := generic["dev_id"].(string)
	if !ok || devID == "" {
		return nil, fmt.Errorf("missing or invalid 'dev_id' field")
	}

	if !ValidateDeviceID(devID) {
		return nil, fmt.Errorf("invalid device ID format: %s", devID)
	}

	ts, ok := generic["ts"]
	if !ok {
		return nil, fmt.Errorf("missing 'ts' (timestamp) field")
	}
	tsInt, ok := toInt64(ts)
	if !ok {
		return nil, fmt.Errorf("invalid 'ts' type")
	}

	return &ParsedMessage{
		Type:      model.UpstreamMessageType(typ),
		DeviceID:  devID,
		Timestamp: tsInt,
		Raw:       payload,
	}, nil
}

func toInt64(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case int:
		return int64(n), true
	case int64:
		return n, true
	default:
		return 0, false
	}
}

// ToDeviceMessage converts a ParsedMessage to the model type used by the handler pipeline.
func (p *ParsedMessage) ToDeviceMessage() *model.DeviceMessage {
	return &model.DeviceMessage{
		Type:      p.Type,
		DeviceID:  p.DeviceID,
		Timestamp: p.Timestamp,
		Raw:       p.Raw,
	}
}

// OnMessage is the callback invoked by the MQTT client for each received packet.
func OnMessage(h *handler.Handler) func(topic string, payload []byte) {
	return func(topic string, payload []byte) {
		deviceID := DeviceIDFromTopic(topic)
		if deviceID == "" {
			log.Printf("WARN: could not extract device ID from topic: %s", topic)
			return
		}

		msg, err := ParseAndValidate(payload)
		if err != nil {
			log.Printf("WARN: invalid message from %s: %v", deviceID, err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := h.Handle(ctx, msg.ToDeviceMessage()); err != nil {
			log.Printf("ERROR: handling message from %s: %v", msg.DeviceID, err)
		}
	}
}
