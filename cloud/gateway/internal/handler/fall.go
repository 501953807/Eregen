package handler

import (
	"context"
	"encoding/json"
	"log"

	"eregen.dev/gateway/internal/model"
	"eregen.dev/gateway/internal/nats"
)

// FallHandler processes fall detection messages from devices.
type FallHandler struct {
	nats *nats.Client
}

func NewFallHandler(n *nats.Client) *FallHandler {
	return &FallHandler{nats: n}
}

func (h *FallHandler) Handles() model.UpstreamMessageType {
	return model.TypeFall
}

func (h *FallHandler) Process(ctx context.Context, msg *model.DeviceMessage) error {
	var p model.FallPayload
	if err := json.Unmarshal(msg.Raw, &p); err != nil {
		return err
	}
	log.Printf("ALERT: fall detected from %s (conf=%.2f) at (%.4f, %.4f)",
		msg.DeviceID, p.Confidence, p.Lat, p.Lon)
	ev := makeNATSEvent(msg, p)
	return h.nats.Publish(ev)
}
