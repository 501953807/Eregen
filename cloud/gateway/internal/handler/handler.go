// © 2026 Eregen (颐贞). All rights reserved.

// Package handler dispatches parsed device messages to NATS, InfluxDB, and Redis.
package handler

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"eregen.dev/gateway/internal/model"
	"eregen.dev/gateway/internal/nats"
	"eregen.dev/gateway/internal/store"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/redis/go-redis/v9"
)

// Handler routes validated device messages to downstream systems.
type Handler struct {
	nats   *nats.Client
	db     *store.Store
	redis  *redis.Client
	influx influxdb2.Client
	bucket string
	org    string
}

// New creates a new Handler.
func New(n *nats.Client, s *store.Store, r *redis.Client, i influxdb2.Client, bucket, org string) *Handler {
	return &Handler{
		nats:   n,
		db:     s,
		redis:  r,
		influx: i,
		bucket: bucket,
		org:    org,
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
	key := "device:status:" + msg.DeviceID
	val := map[string]any{
		"battery":    p.Battery,
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	}
	data, _ := json.Marshal(val)
	h.redis.Set(ctx, key, data, 5*time.Minute)
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
	writeInfluxPoint(h.influx, h.org, h.bucket, "location", map[string]string{
		"dev_id": msg.DeviceID,
	}, map[string]interface{}{
		"lat":      p.Lat,
		"lon":      p.Lon,
		"accuracy": float64(p.Accuracy),
	}, time.Unix(msg.Timestamp, 0))
	ev := makeNATSEvent(msg, p)
	return h.nats.Publish(ev)
}

func (h *Handler) handleHealth(ctx context.Context, msg *model.DeviceMessage) error {
	var p model.HealthPayload
	if err := json.Unmarshal(msg.Raw, &p); err != nil {
		return err
	}
	validateHealth(&p)
	writeInfluxPoint(h.influx, h.org, h.bucket, "health", map[string]string{
		"dev_id": msg.DeviceID,
	}, map[string]interface{}{
		"hr":    float64(p.HeartRate),
		"spo2":  float64(p.SpO2),
		"steps": float64(p.Steps),
	}, time.Unix(msg.Timestamp, 0))
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
	key := "alert:sos:" + msg.DeviceID
	h.redis.Set(ctx, key, msg.Raw, 1*time.Hour)
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
	key := "alert:fall:" + msg.DeviceID
	h.redis.Set(ctx, key, msg.Raw, 1*time.Hour)
	ev := makeNATSEvent(msg, p)
	return h.nats.Publish(ev)
}

func (h *Handler) handleMedStatus(ctx context.Context, msg *model.DeviceMessage) error {
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

func (h *Handler) handlePatientRegister(ctx context.Context, msg *model.DeviceMessage) error {
	var p model.PatientRegisterPayload
	if err := json.Unmarshal(msg.Raw, &p); err != nil {
		return err
	}
	log.Printf("MEDICAL: patient register from %s -> patient_id=%s admission=%s name=%s",
		msg.DeviceID, p.PatientID, p.AdmissionNo, p.Name)
	key := "medical:patient:" + p.PatientID
	val := map[string]any{
		"admission_no": p.AdmissionNo,
		"name":         p.Name,
		"department":   p.Department,
		"bed_number":   p.BedNumber,
		"updated_at":   time.Now().UTC().Format(time.RFC3339),
	}
	data, _ := json.Marshal(val)
	h.redis.Set(ctx, key, data, 7*24*time.Hour)
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
	if p.Result == "unmatched" || p.Result == "not_found" {
		redisKey := "medical:alert:unmatched:" + msg.DeviceID
		h.redis.Set(ctx, redisKey, msg.Raw, 1*time.Hour)
	}
	ev := makeNATSEvent(msg, p)
	return h.nats.PublishMedical(ev)
}

func (h *Handler) handleDeviceStatus(ctx context.Context, msg *model.DeviceMessage) error {
	var p model.DeviceStatusPayload
	if err := json.Unmarshal(msg.Raw, &p); err != nil {
		return err
	}
	key := "medical:device_status:" + msg.DeviceID
	val := map[string]any{
		"battery":    p.Battery,
		"firmware":   p.FirmwareVer,
		"status":     p.Status,
		"bind_count": p.BindCount,
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	}
	data, _ := json.Marshal(val)
	h.redis.Set(ctx, key, data, 5*time.Minute)
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
	redisKey := "medical:alert_tag:" + p.TagID
	if p.Severity == "critical" {
		h.redis.Set(ctx, redisKey, msg.Raw, 24*time.Hour)
	}
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
	key := "community:signin:" + p.ElderID + ":" + p.Period
	val := map[string]any{
		"elder_id":      p.ElderID,
		"hospital_id":   p.HospitalID,
		"period":        p.Period,
		"is_medical":    p.IsMedical,
		"is_welfare":    p.IsWelfare,
		"tags":          p.ActivatedTags,
		"updated_at":    time.Now().UTC().Format(time.RFC3339),
	}
	data, _ := json.Marshal(val)
	h.redis.Set(ctx, key, data, 30*24*time.Hour)
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
	key := "community:dispense:" + p.ElderID + ":" + p.Period
	val := map[string]any{
		"elder_id":         p.ElderID,
		"hospital_id":      p.HospitalID,
		"period":           p.Period,
		"items":            p.Items,
		"total_cost":       p.TotalCost,
		"insurance_covered": p.Insurance,
		"self_pay":         p.SelfPay,
		"updated_at":       time.Now().UTC().Format(time.RFC3339),
	}
	data, _ := json.Marshal(val)
	h.redis.Set(ctx, key, data, 30*24*time.Hour)
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

func writeInfluxPoint(client influxdb2.Client, org, bucket, measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) {
	p := influxdb2.NewPointWithMeasurement(measurement)
	for k, v := range tags {
		p.AddTag(k, v)
	}
	for k, v := range fields {
		p.AddField(k, v)
	}
	p.SetTime(ts)
	api := client.WriteAPI(org, bucket)
	api.WritePoint(p)
	api.Flush()
}
