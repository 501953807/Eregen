// Package handler dispatches parsed device messages to NATS and the database.
package handler

import (
	"context"
	"encoding/json"
	"log"

	"eregen.dev/gateway/internal/model"
	"eregen.dev/gateway/internal/nats"
	"eregen.dev/gateway/internal/store"
)

// Handler routes validated device messages to NATS and persists to database.
type Handler struct {
	nats *nats.Client
	db   *store.Store
}

// New creates a new Handler.
func New(n *nats.Client, s *store.Store) *Handler {
	return &Handler{
		nats: n,
		db:   s,
	}
}

// Handle dispatches a parsed device message to the appropriate subsystem.
func (h *Handler) Handle(ctx context.Context, msg *model.DeviceMessage) error {
	switch msg.Type {
	case model.TypeHeartbeat:
		return h.handleHeartbeat(ctx, msg)
	case model.TypeLocation:
		return h.handleLocation(ctx, msg)
	case model.TypeHealth:
		return h.handleHealth(ctx, msg)
	case model.TypeSOS:
		return h.handleSOS(ctx, msg)
	case model.TypeFall:
		return h.handleFall(ctx, msg)
	case model.TypeMedStatus:
		return h.handleMedStatus(ctx, msg)
	case model.TypePatientRegister:
		return h.handlePatientRegister(ctx, msg)
	case model.TypeVerificationScan:
		return h.handleVerificationScan(ctx, msg)
	case model.TypeDeviceStatus:
		return h.handleDeviceStatus(ctx, msg)
	case model.TypeAlertTag:
		return h.handleAlertTag(ctx, msg)
	case model.TypeCommunitySignin:
		return h.handleCommunitySignin(ctx, msg)
	case model.TypeCommunityWelfareUpdate:
		return h.handleCommunityWelfareUpdate(ctx, msg)
	case model.TypeCommunityDispense:
		return h.handleCommunityDispense(ctx, msg)
	default:
		log.Printf("WARN: unknown event type %q for device %s", msg.Type, msg.DeviceID)
		return nil
	}
}

func (h *Handler) handleHeartbeat(ctx context.Context, msg *model.DeviceMessage) error {
	var p model.HeartbeatPayload
	if err := json.Unmarshal(msg.Raw, &p); err != nil {
		return err
	}
	ev := makeNATSEvent(msg, p)
	return h.nats.Publish(ev)
}

func (h *Handler) handleLocation(ctx context.Context, msg *model.DeviceMessage) error {
	var p model.LocationPayload
	if err := json.Unmarshal(msg.Raw, &p); err != nil {
		return err
	}
	if !validGPS(p.Lat, p.Lon) {
		log.Printf("WARN: invalid GPS coords from %s: (%.4f, %.4f)", msg.DeviceID, p.Lat, p.Lon)
		return nil
	}
	// Persist location to database
	if err := h.db.InsertLocationRecord(ctx, msg.DeviceID, p.Lat, p.Lon, p.Accuracy, msg.Timestamp); err != nil {
		log.Printf("WARN: persist location for %s: %v", msg.DeviceID, err)
	}
	ev := makeNATSEvent(msg, p)
	return h.nats.Publish(ev)
}

func (h *Handler) handleHealth(ctx context.Context, msg *model.DeviceMessage) error {
	var p model.HealthPayload
	if err := json.Unmarshal(msg.Raw, &p); err != nil {
		return err
	}
	validateHealth(&p)
	// Persist health data to database
	if err := h.db.InsertHealthRecord(ctx, msg.DeviceID, p.HeartRate, p.SpO2, p.Steps, p.Sleep, msg.Timestamp); err != nil {
		log.Printf("ERROR: persist health for %s: %v", msg.DeviceID, err)
	}
	ev := makeNATSEvent(msg, p)
	return h.nats.Publish(ev)
}

func (h *Handler) handleSOS(ctx context.Context, msg *model.DeviceMessage) error {
	var p model.SOSPayload
	if err := json.Unmarshal(msg.Raw, &p); err != nil {
		return err
	}
	log.Printf("ALERT: SOS from %s at (%.4f, %.4f)", msg.DeviceID, p.Lat, p.Lon)
	ev := makeNATSEvent(msg, p)
	return h.nats.Publish(ev)
}

func (h *Handler) handleFall(ctx context.Context, msg *model.DeviceMessage) error {
	var p model.FallPayload
	if err := json.Unmarshal(msg.Raw, &p); err != nil {
		return err
	}
	log.Printf("ALERT: fall detected from %s (conf=%.2f) at (%.4f, %.4f)",
		msg.DeviceID, p.Confidence, p.Lat, p.Lon)
	ev := makeNATSEvent(msg, p)
	return h.nats.Publish(ev)
}

func (h *Handler) handleMedStatus(ctx context.Context, msg *model.DeviceMessage) error {
	var p model.MedStatusPayload
	if err := json.Unmarshal(msg.Raw, &p); err != nil {
		return err
	}
	// Persist medication status to database
	if err := h.db.InsertMedStatusRecord(ctx, msg.DeviceID, p.Compartment, p.Taken, msg.Timestamp); err != nil {
		log.Printf("ERROR: persist med_status for %s: %v", msg.DeviceID, err)
	}
	ev := makeNATSEvent(msg, p)
	return h.nats.Publish(ev)
}

func (h *Handler) handlePatientRegister(ctx context.Context, msg *model.DeviceMessage) error {
	var p model.PatientRegisterPayload
	if err := json.Unmarshal(msg.Raw, &p); err != nil {
		return err
	}
	log.Printf("MEDICAL: patient register from %s -> patient_id=%s admission=%s name=%s",
		msg.DeviceID, p.PatientID, p.AdmissionNo, p.Name)
	ev := makeNATSEvent(msg, p)
	return h.nats.PublishMedical(ev)
}

func (h *Handler) handleVerificationScan(ctx context.Context, msg *model.DeviceMessage) error {
	var p model.VerificationScanPayload
	if err := json.Unmarshal(msg.Raw, &p); err != nil {
		return err
	}
	log.Printf("MEDICAL: verification scan from %s -> patient=%s result=%s type=%s",
		msg.DeviceID, p.PatientID, p.Result, p.ScanType)
	ev := makeNATSEvent(msg, p)
	return h.nats.PublishMedical(ev)
}

func (h *Handler) handleDeviceStatus(ctx context.Context, msg *model.DeviceMessage) error {
	var p model.DeviceStatusPayload
	if err := json.Unmarshal(msg.Raw, &p); err != nil {
		return err
	}
	ev := makeNATSEvent(msg, p)
	return h.nats.PublishMedical(ev)
}

func (h *Handler) handleAlertTag(ctx context.Context, msg *model.DeviceMessage) error {
	var p model.AlertTagPayload
	if err := json.Unmarshal(msg.Raw, &p); err != nil {
		return err
	}
	log.Printf("MEDICAL: alert tag [%s] from %s -> tag=%s severity=%s",
		p.Severity, msg.DeviceID, p.TagName, p.Severity)
	ev := makeNATSEvent(msg, p)
	return h.nats.PublishMedical(ev)
}

func (h *Handler) handleMedicalWBStatus(ctx context.Context, msg *model.DeviceMessage) error {
	var p model.MedicalWBStatusPayload
	if err := json.Unmarshal(msg.Raw, &p); err != nil {
		return err
	}
	log.Printf("MEDICAL: wb status from %s -> patient=%s bat=%d", msg.DeviceID, p.PatientID, p.Battery)
	ev := makeNATSEvent(msg, p)
	return h.nats.PublishMedical(ev)
}

func (h *Handler) handleCommunitySignin(ctx context.Context, msg *model.DeviceMessage) error {
	var p model.CommunitySigninPayload
	if err := json.Unmarshal(msg.Raw, &p); err != nil {
		return err
	}
	log.Printf("COMMUNITY: signin from %s -> elder=%s period=%s medical=%v welfare=%v",
		msg.DeviceID, p.ElderID, p.Period, p.IsMedical, p.IsWelfare)
	ev := makeNATSEvent(msg, p)
	return h.nats.PublishCommunity(ev)
}

func (h *Handler) handleCommunityWelfareUpdate(ctx context.Context, msg *model.DeviceMessage) error {
	var p model.CommunityWelfareUpdatePayload
	if err := json.Unmarshal(msg.Raw, &p); err != nil {
		return err
	}
	log.Printf("COMMUNITY: welfare update from %s -> elder=%s tag=%s action=%s",
		msg.DeviceID, p.ElderID, p.TagCode, p.Action)
	ev := makeNATSEvent(msg, p)
	return h.nats.PublishCommunity(ev)
}

func (h *Handler) handleCommunityDispense(ctx context.Context, msg *model.DeviceMessage) error {
	var p model.CommunityDispensePayload
	if err := json.Unmarshal(msg.Raw, &p); err != nil {
		return err
	}
	log.Printf("COMMUNITY: pharmacy dispense from %s -> elder=%s period=%s cost=%.2f",
		msg.DeviceID, p.ElderID, p.Period, p.TotalCost)
	ev := makeNATSEvent(msg, p)
	return h.nats.PublishCommunity(ev)
}

func validGPS(lat, lon float64) bool {
	return lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180
}

func validateHealth(h *model.HealthPayload) {
	if h.HeartRate < 0 || h.HeartRate > 300 {
		h.HeartRate = 0
	}
	if h.SpO2 < 0 || h.SpO2 > 100 {
		h.SpO2 = 0
	}
	if h.Steps < 0 {
		h.Steps = 0
	}
}

func makeNATSEvent(msg *model.DeviceMessage, payload any) *nats.Event {
	data, _ := json.Marshal(payload)
	ev := &nats.Event{
		Type:      string(msg.Type),
		DeviceID:  msg.DeviceID,
		Timestamp: msg.Timestamp,
		Payload:   data,
	}
	// Extract hospital_id from community signin payloads for cross-hospital tracking
	switch p := payload.(type) {
	case model.CommunitySigninPayload:
		if p.HospitalID != "" {
			ev.HospitalID = p.HospitalID
		}
	}
	return ev
}
