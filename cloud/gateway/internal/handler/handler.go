package handler

import (
	"context"
	"log"

	"eregen.dev/gateway/internal/model"
	"eregen.dev/gateway/internal/nats"
	"eregen.dev/gateway/internal/store"
)

// Handler dispatches parsed device messages to the appropriate subsystem.
type Handler struct {
	handlers []MessageHandler
}

// New creates a new Handler with all message handlers registered.
func New(n *nats.Client, s *store.Store) *Handler {
	h := &Handler{}

	// Register all message handlers
	h.handlers = append(h.handlers, NewHeartbeatHandler(n))
	h.handlers = append(h.handlers, NewLocationHandler(n, s))
	h.handlers = append(h.handlers, NewHealthHandler(n, s))
	h.handlers = append(h.handlers, NewSOSHandler(n))
	h.handlers = append(h.handlers, NewFallHandler(n))
	h.handlers = append(h.handlers, NewMedStatusHandler(n, s))

	// Register medical handlers (one per medical message type)
	medicalTypes := []model.UpstreamMessageType{
		model.TypePatientRegister,
		model.TypeVerificationScan,
		model.TypeDeviceStatus,
		model.TypeAlertTag,
		model.TypeMedicalWBStatus,
	}
	for _, t := range medicalTypes {
		h.handlers = append(h.handlers, MedicalHandlerForType(n, t))
	}

	// Register community handlers
	communityTypes := []model.UpstreamMessageType{
		model.TypeCommunitySignin,
		model.TypeCommunityWelfareUpdate,
		model.TypeCommunityDispense,
	}
	for _, t := range communityTypes {
		_ = t
		h.handlers = append(h.handlers, &CommunityHandler{nats: n})
	}

	return h
}

// Handle dispatches a parsed device message to the appropriate handler.
func (h *Handler) Handle(ctx context.Context, msg *model.DeviceMessage) error {
	for _, handler := range h.handlers {
		if handler.Handles() == msg.Type {
			return handler.Process(ctx, msg)
		}
	}
	log.Printf("WARN: unknown event type %q for device %s", msg.Type, msg.DeviceID)
	return nil
}
