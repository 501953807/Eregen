package handler

import (
	"context"
	"encoding/json"
	"log"

	"eregen.dev/gateway/internal/model"
	"eregen.dev/gateway/internal/nats"
	"eregen.dev/gateway/internal/store"
)

// HealthHandler processes health data messages from devices.
type HealthHandler struct {
	nats *nats.Client
	db   *store.Store
}

func NewHealthHandler(n *nats.Client, s *store.Store) *HealthHandler {
	return &HealthHandler{nats: n, db: s}
}

func (h *HealthHandler) Handles() model.UpstreamMessageType {
	return model.TypeHealth
}

func (h *HealthHandler) Process(ctx context.Context, msg *model.DeviceMessage) error {
	var p model.HealthPayload
	if err := json.Unmarshal(msg.Raw, &p); err != nil {
		return err
	}
	validateHealth(&p)
	if err := h.db.InsertHealthRecord(ctx, msg.DeviceID, p.HeartRate, p.SpO2, p.Steps, p.Sleep, msg.Timestamp); err != nil {
		log.Printf("ERROR: persist health for %s: %v", msg.DeviceID, err)
	}
	ev := makeNATSEvent(msg, p)
	return h.nats.Publish(ev)
}
