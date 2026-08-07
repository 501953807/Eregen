package handler

import (
	"context"
	"encoding/json"
	"log"

	"eregen.dev/gateway/internal/model"
	"eregen.dev/gateway/internal/nats"
)

// CommunityHandler processes community wristband messages (signin, welfare update, dispense).
type CommunityHandler struct {
	nats *nats.Client
}

func NewCommunityHandler(n *nats.Client) *CommunityHandler {
	return &CommunityHandler{nats: n}
}

func (h *CommunityHandler) Handles() model.UpstreamMessageType {
	return model.TypeCommunitySignin
}

func (h *CommunityHandler) Process(ctx context.Context, msg *model.DeviceMessage) error {
	switch msg.Type {
	case model.TypeCommunitySignin:
		var p model.CommunitySigninPayload
		if err := json.Unmarshal(msg.Raw, &p); err != nil {
			return err
		}
		log.Printf("COMMUNITY: signin from %s -> elder=%s period=%s medical=%v welfare=%v",
			msg.DeviceID, p.ElderID, p.Period, p.IsMedical, p.IsWelfare)
		ev := makeNATSEvent(msg, p)
		return h.nats.PublishCommunity(ev)
	case model.TypeCommunityWelfareUpdate:
		var p model.CommunityWelfareUpdatePayload
		if err := json.Unmarshal(msg.Raw, &p); err != nil {
			return err
		}
		log.Printf("COMMUNITY: welfare update from %s -> elder=%s tag=%s action=%s",
			msg.DeviceID, p.ElderID, p.TagCode, p.Action)
		ev := makeNATSEvent(msg, p)
		return h.nats.PublishCommunity(ev)
	case model.TypeCommunityDispense:
		var p model.CommunityDispensePayload
		if err := json.Unmarshal(msg.Raw, &p); err != nil {
			return err
		}
		log.Printf("COMMUNITY: pharmacy dispense from %s -> elder=%s period=%s cost=%.2f",
			msg.DeviceID, p.ElderID, p.Period, p.TotalCost)
		ev := makeNATSEvent(msg, p)
		return h.nats.PublishCommunity(ev)
	default:
		return nil
	}
}
