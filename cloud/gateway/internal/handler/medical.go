package handler

import (
	"context"
	"encoding/json"
	"log"

	"eregen.dev/gateway/internal/model"
	"eregen.dev/gateway/internal/nats"
)

// MedicalHandler processes medical wristband messages (patient register, verification, device status, alert tag, wb status).
type MedicalHandler struct {
	nats *nats.Client
}

func NewMedicalHandler(n *nats.Client) *MedicalHandler {
	return &MedicalHandler{nats: n}
}

func (h *MedicalHandler) Handles() model.UpstreamMessageType {
	// This handler processes multiple medical message types
	return model.TypePatientRegister
}

func (h *MedicalHandler) Process(ctx context.Context, msg *model.DeviceMessage) error {
	switch msg.Type {
	case model.TypePatientRegister:
		var p model.PatientRegisterPayload
		if err := json.Unmarshal(msg.Raw, &p); err != nil {
			return err
		}
		log.Printf("MEDICAL: patient register from %s -> patient_id=%s admission=%s name=%s",
			msg.DeviceID, p.PatientID, p.AdmissionNo, p.Name)
		ev := makeNATSEvent(msg, p)
		return h.nats.PublishMedical(ev)
	case model.TypeVerificationScan:
		var p model.VerificationScanPayload
		if err := json.Unmarshal(msg.Raw, &p); err != nil {
			return err
		}
		log.Printf("MEDICAL: verification scan from %s -> patient=%s result=%s type=%s",
			msg.DeviceID, p.PatientID, p.Result, p.ScanType)
		ev := makeNATSEvent(msg, p)
		return h.nats.PublishMedical(ev)
	case model.TypeDeviceStatus:
		var p model.DeviceStatusPayload
		if err := json.Unmarshal(msg.Raw, &p); err != nil {
			return err
		}
		ev := makeNATSEvent(msg, p)
		return h.nats.PublishMedical(ev)
	case model.TypeAlertTag:
		var p model.AlertTagPayload
		if err := json.Unmarshal(msg.Raw, &p); err != nil {
			return err
		}
		log.Printf("MEDICAL: alert tag [%s] from %s -> tag=%s severity=%s",
			p.Severity, msg.DeviceID, p.TagName, p.Severity)
		ev := makeNATSEvent(msg, p)
		return h.nats.PublishMedical(ev)
	case model.TypeMedicalWBStatus:
		var p model.MedicalWBStatusPayload
		if err := json.Unmarshal(msg.Raw, &p); err != nil {
			return err
		}
		log.Printf("MEDICAL: wb status from %s -> patient=%s bat=%d", msg.DeviceID, p.PatientID, p.Battery)
		ev := makeNATSEvent(msg, p)
		return h.nats.PublishMedical(ev)
	default:
		return nil
	}
}

// MedicalHandlerForType creates a handler that matches a specific medical message type.
func MedicalHandlerForType(n *nats.Client, t model.UpstreamMessageType) *MedicalHandler {
	return &MedicalHandler{nats: n}
}
