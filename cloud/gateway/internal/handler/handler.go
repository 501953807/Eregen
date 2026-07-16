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
	return &nats.Event{
		Type:      string(msg.Type),
		DeviceID:  msg.DeviceID,
		Timestamp: msg.Timestamp,
		Payload:   data,
	}
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
