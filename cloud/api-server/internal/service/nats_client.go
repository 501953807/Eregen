package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"eregen.dev/api-server/internal/model"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

const (
	natsDeviceSubject = "device.events.>"
	natsTopicFormat   = "device.command.%s"
)

// NatsClient manages NATS JetStream connections for device event processing.
type NatsClient struct {
	nc  *nats.Conn
	js  nats.JetStreamContext
	log *zap.Logger
}

// NewNatsClient creates a NATS client connected to the given URL.
func NewNatsClient(url string, log *zap.Logger) (*NatsClient, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("connect nats: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("jetstream: %w", err)
	}

	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "DEVICE_EVENTS",
		Subjects: []string{"device.events.>"},
		Storage:  nats.FileStorage,
	})
	if err != nil && err != nats.ErrStreamNameAlreadyInUse {
		return nil, fmt.Errorf("add stream: %w", err)
	}

	return &NatsClient{nc: nc, js: js, log: log}, nil
}

// Close shuts down the NATS connection.
func (n *NatsClient) Close() {
	if n.nc != nil {
		n.nc.Close()
	}
}

// SubscribeDeviceEvents starts consuming device events from JetStream.
func (n *NatsClient) SubscribeDeviceEvents(ctx context.Context, handler *EventHandler) error {
	sub, err := n.js.Subscribe(natsDeviceSubject, func(msg *nats.Msg) {
		var event DeviceEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			n.log.Warn("unmarshal device event", zap.Error(err))
			msg.Nak()
			return
		}

		switch event.Type {
		case "health":
			handler.handleHealth(ctx, &event)
		case "location":
			handler.handleLocation(ctx, &event)
		case "sos":
			handler.handleSOS(ctx, &event)
		case "fall":
			handler.handleFall(ctx, &event)
		case "heartbeat":
			handler.handleHeartbeat(ctx, &event)
		case "med_status":
			handler.handleMedStatus(ctx, &event)
		default:
			n.log.Debug("unknown device event type", zap.String("type", event.Type))
		}

		msg.Ack()
	}, nats.Durable("api-server-events"))

	if err != nil {
		return fmt.Errorf("subscribe device events: %w", err)
	}

	<-ctx.Done()
	sub.Unsubscribe()
	return nil
}

// PublishCommand publishes a device command (config update, TTS, OTA).
func (n *NatsClient) PublishCommand(ctx context.Context, deviceID string, cmd any) error {
	topic := fmt.Sprintf(natsTopicFormat, deviceID)
	data, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("marshal command: %w", err)
	}
	_, err = n.js.Publish(topic, data)
	if err != nil {
		return fmt.Errorf("publish command to %s: %w", topic, err)
	}
	return nil
}

// DeviceEvent represents a raw event from a hardware device.
type DeviceEvent struct {
	Type          string    `json:"type"`
	DevID         string    `json:"dev_id"`
	ElderlyID     string    `json:"elderly_id,omitempty"`
	Timestamp     int64     `json:"ts"`
	Lat           float64   `json:"lat,omitempty"`
	Lon           float64   `json:"lon,omitempty"`
	Acc           *float64  `json:"acc,omitempty"`
	Bat           *int      `json:"bat,omitempty"`
	HR            *int      `json:"hr,omitempty"`
	SPO2          *int      `json:"spo2,omitempty"`
	Steps         *int64    `json:"step,omitempty"`
	Conf          *float64  `json:"conf,omitempty"`
	Compartment   *int      `json:"compartment,omitempty"`
	Taken         *bool     `json:"taken,omitempty"`
}

// EventHandler routes processed events to storage layers.
type EventHandler struct {
	influx  influxDBClient
	log     *zap.Logger
	alertCb func(ctx context.Context, a *model.Alert) error
	healthCb func(ctx context.Context, r *model.HealthRecord) error
	locCb   func(ctx context.Context, r *model.LocationRecord) error
	medCb   func(ctx context.Context, r *model.MedStatusRecord) error
}

type influxDBClient interface {
	WritePoint(points ...any)
}

// NewEventHandler creates an event handler wired to storage callbacks.
func NewEventHandler(influx influxDBClient, log *zap.Logger) *EventHandler {
	return &EventHandler{influx: influx, log: log}
}

// SetAlertCallback sets the callback for creating alerts.
func (h *EventHandler) SetAlertCallback(fn func(ctx context.Context, a *model.Alert) error) {
	h.alertCb = fn
}

// SetHealthCallback sets the callback for storing health records.
func (h *EventHandler) SetHealthCallback(fn func(ctx context.Context, r *model.HealthRecord) error) {
	h.healthCb = fn
}

// SetLocationCallback sets the callback for storing location records.
func (h *EventHandler) SetLocationCallback(fn func(ctx context.Context, r *model.LocationRecord) error) {
	h.locCb = fn
}

// SetMedStatusCallback sets the callback for storing medication status.
func (h *EventHandler) SetMedStatusCallback(fn func(ctx context.Context, r *model.MedStatusRecord) error) {
	h.medCb = fn
}

func (h *EventHandler) handleHealth(ctx context.Context, ev *DeviceEvent) {
	if h.healthCb == nil {
		return
	}
	r := &model.HealthRecord{
		ID:        uuid.New().String(),
		ElderlyID: ev.ElderlyID,
		Timestamp: time.Unix(ev.Timestamp, 0),
		HR:        ev.HR,
		SPO2:      ev.SPO2,
		Steps:     ev.Steps,
	}
	if err := h.healthCb(ctx, r); err != nil {
		h.log.Error("store health record", zap.Error(err))
	}
}

func (h *EventHandler) handleLocation(ctx context.Context, ev *DeviceEvent) {
	if h.locCb == nil {
		return
	}
	r := &model.LocationRecord{
		ID:        uuid.New().String(),
		ElderlyID: ev.ElderlyID,
		Timestamp: time.Unix(ev.Timestamp, 0),
		Lat:       ev.Lat,
		Lon:       ev.Lon,
		Accuracy:  ev.Acc,
	}
	if err := h.locCb(ctx, r); err != nil {
		h.log.Error("store location record", zap.Error(err))
	}
}

func (h *EventHandler) handleSOS(ctx context.Context, ev *DeviceEvent) {
	if h.alertCb == nil {
		return
	}
	a := &model.Alert{
		ElderlyID: ev.ElderlyID,
		AlertType: "sos",
		Severity:  model.AlertP0,
		Status:    model.AlertPending,
		Metadata: map[string]any{
			"device_id": ev.DevID,
			"lat":       ev.Lat,
			"lon":       ev.Lon,
		},
	}
	if err := h.alertCb(ctx, a); err != nil {
		h.log.Error("create SOS alert", zap.Error(err))
	}
}

func (h *EventHandler) handleFall(ctx context.Context, ev *DeviceEvent) {
	if h.alertCb == nil {
		return
	}
	conf := 0.0
	if ev.Conf != nil {
		conf = *ev.Conf
	}
	a := &model.Alert{
		ElderlyID: ev.ElderlyID,
		AlertType: "fall",
		Severity:  model.AlertP0,
		Status:    model.AlertPending,
		Metadata: map[string]any{
			"device_id":  ev.DevID,
			"confidence": conf,
			"lat":        ev.Lat,
			"lon":        ev.Lon,
		},
	}
	if err := h.alertCb(ctx, a); err != nil {
		h.log.Error("create fall alert", zap.Error(err))
	}
}

func (h *EventHandler) handleHeartbeat(ctx context.Context, ev *DeviceEvent) {
	h.log.Debug("device heartbeat",
		zap.String("device", ev.DevID),
		zap.Int("battery", getInt(ev.Bat)),
	)
}

func (h *EventHandler) handleMedStatus(ctx context.Context, ev *DeviceEvent) {
	if h.medCb == nil {
		return
	}
	taken := false
	if ev.Taken != nil {
		taken = *ev.Taken
	}
	now := time.Now()
	r := &model.MedStatusRecord{
		ID:        uuid.New().String(),
		ElderlyID: ev.ElderlyID,
		TakenAt:   &now,
		Taken:     taken,
	}
	if err := h.medCb(ctx, r); err != nil {
		h.log.Error("store med status", zap.Error(err))
	}
}

func getInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
