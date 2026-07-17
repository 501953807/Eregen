package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"eregen.dev/push/internal/model"
	"eregen.dev/push/internal/router"

	"github.com/nats-io/nats.go"
)

// Subscriber connects to NATS and dispatches device events to handlers.
type Subscriber struct {
	nc    *nats.Conn
	js    nats.JetStreamContext
	alert chan model.AlertPushEvent
}

// NewSubscriber creates a NATS JetStream consumer for push events.
func NewSubscriber(natsURL string) (*Subscriber, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("connect nats: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("jetstream: %w", err)
	}

	s := &Subscriber{
		nc:    nc,
		js:    js,
		alert: make(chan model.AlertPushEvent, 128),
	}

	_, err = js.Subscribe("eregen.event.>", s.onMessage,
		nats.Durable("push-service"),
	)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("subscribe: %w", err)
	}

	return s, nil
}

func (s *Subscriber) onMessage(msg *nats.Msg) {
	var envelope struct {
		Type    string                 `json:"type"`
		DevID   string                 `json:"dev_id"`
		Ts      int64                  `json:"ts"`
		Payload map[string]interface{} `json:"payload"`
	}
	if err := json.Unmarshal(msg.Data, &envelope); err != nil {
		log.Printf("[nats] unmarshal: %v", err)
		msg.Ack()
		return
	}

	switch envelope.Type {
	case "sos", "fall":
		ev := model.AlertPushEvent{
			AlertID:   envelope.DevID,
			ElderlyID: envelope.DevID,
			Severity:  model.SeverityP0,
			AlertType: envelope.Type,
			Message:   buildAlertMsg(envelope.Type, envelope.Payload),
			Timestamp: time.Unix(envelope.Ts, 0),
			RawData:   envelope.Payload,
		}
		select {
		case s.alert <- ev:
		default:
			log.Println("[nats] alert channel full, dropping")
		}

	case "med_missed":
		ev := model.AlertPushEvent{
			AlertID:   envelope.DevID,
			ElderlyID: envelope.DevID,
			Severity:  model.SeverityP1,
			AlertType: "med_missed",
			Message:   "用药提醒：有药物未按时服用",
			Timestamp: time.Unix(envelope.Ts, 0),
		}
		select {
		case s.alert <- ev:
		default:
			log.Println("[nats] alert channel full, dropping")
		}
	}

	msg.Ack()
}

func buildAlertMsg(alertType string, payload map[string]interface{}) string {
	loc := ""
	if lat, ok := payload["lat"]; ok {
		lon := payload["lon"]
		loc = fmt.Sprintf("(%.4f, %.4f)", lat, lon)
	}
	switch alertType {
	case "sos":
		return "紧急告警：设备触发SOS按钮" + loc
	case "fall":
		conf := payload["conf"]
		return "跌倒检测告警，置信度=" + fmt.Sprintf("%v", conf) + loc
	default:
		return "设备告警：" + alertType
	}
}

// Start begins dispatching events to the router. Blocks until Close is called.
func (s *Subscriber) Start(rtr *router.Router) {
	for {
		select {
		case ev := <-s.alert:
			log.Printf("[alert] id=%s type=%s severity=%s", ev.AlertID, ev.AlertType, ev.Severity)
			members := []router.Member{
				{UserID: ev.ElderlyID}, // TODO: fetch from DB
			}
			rtr.DeliverAlert(context.Background(), ev, members)
		}
	}
}

// Close shuts down the NATS connection.
func (s *Subscriber) Close() {
	s.nc.Close()
}
