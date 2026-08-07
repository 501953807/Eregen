package handler

import (
	"context"
	"encoding/json"

	"eregen.dev/gateway/internal/model"
	"eregen.dev/gateway/internal/nats"
)

// HeartbeatHandler processes heartbeat messages from devices.
type HeartbeatHandler struct {
	nats *nats.Client
}

func NewHeartbeatHandler(n *nats.Client) *HeartbeatHandler {
	return &HeartbeatHandler{nats: n}
}

func (h *HeartbeatHandler) Handles() model.UpstreamMessageType {
	return model.TypeHeartbeat
}

func (h *HeartbeatHandler) Process(ctx context.Context, msg *model.DeviceMessage) error {
	var p model.HeartbeatPayload
	if err := json.Unmarshal(msg.Raw, &p); err != nil {
		return err
	}
	ev := makeNATSEvent(msg, p)
	return h.nats.Publish(ev)
}
