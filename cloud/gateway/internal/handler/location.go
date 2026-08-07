package handler

import (
	"context"
	"encoding/json"
	"log"

	"eregen.dev/gateway/internal/model"
	"eregen.dev/gateway/internal/nats"
	"eregen.dev/gateway/internal/store"
)

// LocationHandler processes location messages from devices.
type LocationHandler struct {
	nats *nats.Client
	db   *store.Store
}

func NewLocationHandler(n *nats.Client, s *store.Store) *LocationHandler {
	return &LocationHandler{nats: n, db: s}
}

func (h *LocationHandler) Handles() model.UpstreamMessageType {
	return model.TypeLocation
}

func (h *LocationHandler) Process(ctx context.Context, msg *model.DeviceMessage) error {
	var p model.LocationPayload
	if err := json.Unmarshal(msg.Raw, &p); err != nil {
		return err
	}
	if !validGPS(p.Lat, p.Lon) {
		log.Printf("WARN: invalid GPS coords from %s: (%.4f, %.4f)", msg.DeviceID, p.Lat, p.Lon)
		return nil
	}
	if err := h.db.InsertLocationRecord(ctx, msg.DeviceID, p.Lat, p.Lon, p.Accuracy, msg.Timestamp); err != nil {
		log.Printf("WARN: persist location for %s: %v", msg.DeviceID, err)
	}
	ev := makeNATSEvent(msg, p)
	return h.nats.Publish(ev)
}
