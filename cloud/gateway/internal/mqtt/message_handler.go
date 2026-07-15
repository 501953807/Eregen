// © 2026 Eregen (颐贞). All rights reserved.

package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sync/atomic"

	"eregen/cloud/gateway/internal/nats"
)

var deviceIDRegex = regexp.MustCompile(`^(BR|PX)-[A-Za-z0-9]+$`)

// ParsedMessage holds a validated incoming MQTT message.
type ParsedMessage struct {
	Type string          `json:"type"`
	Raw  json.RawMessage `json:"-"`
}

// eventHandler handles a parsed event for a specific device.
type eventHandler func(nc *nats.Client, deviceID string, raw json.RawMessage)

// messageCounts tracks per-type message counts for metrics.
var messageCounts atomic.Int64

// ParseAndValidate parses an MQTT payload and validates required fields.
// Returns a ParsedMessage if valid, or an error if the message should be dropped.
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

	if !deviceIDRegex.MatchString(devID) {
		return nil, fmt.Errorf("invalid device ID format: %s (must match BR-XXXX or PX-XXXX)", devID)
	}

	if _, ok := generic["ts"]; !ok {
		return nil, fmt.Errorf("missing 'ts' (timestamp) field")
	}

	return &ParsedMessage{
		Type: typ,
		Raw:  payload,
	}, nil
}

// ForwardToNATS publishes a validated message to NATS.
func ForwardToNATS(nc *nats.Client, deviceID string, msg *ParsedMessage) {
	messageCounts.Add(1)

	eventJSON, err := marshalEvent(msg.Type, deviceID, msg.Raw)
	if err != nil {
		log.Printf("ERROR: failed to marshal event %s for %s: %v", msg.Type, deviceID, err)
		return
	}

	if err := nc.Publish(msg.Type, eventJSON); err != nil {
		log.Printf("ERROR: failed to publish event %s for %s: %v", msg.Type, deviceID, err)
	}
}

// marshalEvent adds metadata to the raw payload.
func marshalEvent(eventType, deviceID string, raw json.RawMessage) ([]byte, error) {
	var fields map[string]interface{}
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, fmt.Errorf("re-parse raw: %w", err)
	}
	fields["dev_id"] = deviceID

	out, err := json.Marshal(fields)
	if err != nil {
		return nil, fmt.Errorf("marshal enriched event: %w", err)
	}
	return out, nil
}

// GetMessageCount returns the total number of messages processed.
func GetMessageCount() int64 {
	return messageCounts.Load()
}

// --- Type-specific handlers ---

func handleHeartbeat(nc *nats.Client, deviceID string, raw json.RawMessage) {
	msg, _ := ParseAndValidate(raw)
	if msg == nil {
		return
	}
	ForwardToNATS(nc, deviceID, msg)
}

func handleLocation(nc *nats.Client, deviceID string, raw json.RawMessage) {
	msg, _ := ParseAndValidate(raw)
	if msg == nil {
		return
	}
	ForwardToNATS(nc, deviceID, msg)
}

func handleHealth(nc *nats.Client, deviceID string, raw json.RawMessage) {
	msg, _ := ParseAndValidate(raw)
	if msg == nil {
		return
	}
	ForwardToNATS(nc, deviceID, msg)
}

func handleSOS(nc *nats.Client, deviceID string, raw json.RawMessage) {
	msg, _ := ParseAndValidate(raw)
	if msg == nil {
		return
	}
	ForwardToNATS(nc, deviceID, msg)
}

func handleFall(nc *nats.Client, deviceID string, raw json.RawMessage) {
	msg, _ := ParseAndValidate(raw)
	if msg == nil {
		return
	}
	ForwardToNATS(nc, deviceID, msg)
}

func handleMedStatus(nc *nats.Client, deviceID string, raw json.RawMessage) {
	msg, _ := ParseAndValidate(raw)
	if msg == nil {
		return
	}
	ForwardToNATS(nc, deviceID, msg)
}

func handleFenceAlert(nc *nats.Client, deviceID string, raw json.RawMessage) {
	msg, _ := ParseAndValidate(raw)
	if msg == nil {
		return
	}
	ForwardToNATS(nc, deviceID, msg)
}

func handleInventoryWarning(nc *nats.Client, deviceID string, raw json.RawMessage) {
	msg, _ := ParseAndValidate(raw)
	if msg == nil {
		return
	}
	ForwardToNATS(nc, deviceID, msg)
}
