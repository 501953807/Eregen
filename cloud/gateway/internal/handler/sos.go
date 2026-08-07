package handler

import (
	"context"
	"encoding/json"
	"log"

	"eregen.dev/gateway/internal/model"
	"eregen.dev/gateway/internal/nats"
)

// SOSHandler processes SOS alert messages from devices.
type SOSHandler struct {
	nats *nats.Client
}

func NewSOSHandler(n *nats.Client) *SOSHandler {
	return &SOSHandler{nats: n}
}

func (h *SOSHandler) Handles() model.UpstreamMessageType {
	return model.TypeSOS
}

func (h *SOSHandler) Process(ctx context.Context, msg *model.DeviceMessage) error {
	var p model.SOSPayload
	if err := json.Unmarshal(msg.Raw, &p); err != nil {
		return err
	}
	log.Printf("ALERT: SOS from %s at (%.4f, %.4f)", msg.DeviceID, p.Lat, p.Lon)
	ev := makeNATSEvent(msg, p)
	return h.nats.Publish(ev)
}
