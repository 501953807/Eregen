package handler

import (
	"context"
	"encoding/json"
	"log"

	"eregen.dev/gateway/internal/model"
	"eregen.dev/gateway/internal/nats"
	"eregen.dev/gateway/internal/store"
)

// MedStatusHandler processes medication status messages from pillboxes.
type MedStatusHandler struct {
	nats *nats.Client
	db   *store.Store
}

func NewMedStatusHandler(n *nats.Client, s *store.Store) *MedStatusHandler {
	return &MedStatusHandler{nats: n, db: s}
}

func (h *MedStatusHandler) Handles() model.UpstreamMessageType {
	return model.TypeMedStatus
}

func (h *MedStatusHandler) Process(ctx context.Context, msg *model.DeviceMessage) error {
	var p model.MedStatusPayload
	if err := json.Unmarshal(msg.Raw, &p); err != nil {
		return err
	}
	if err := h.db.InsertMedStatusRecord(ctx, msg.DeviceID, p.Compartment, p.Taken, msg.Timestamp); err != nil {
		log.Printf("ERROR: persist med_status for %s: %v", msg.DeviceID, err)
	}
	ev := makeNATSEvent(msg, p)
	return h.nats.Publish(ev)
}
