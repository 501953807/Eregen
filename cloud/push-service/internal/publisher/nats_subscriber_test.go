package publisher

import (
	"testing"

	"github.com/nats-io/nats.go"
)

func TestBuildAlertMsg_SOS(t *testing.T) {
	msg := buildAlertMsg("sos", map[string]interface{}{
		"lat": float64(31.23),
		"lon": float64(121.47),
	})
	if msg != "紧急告警：设备触发SOS按钮(31.2300, 121.4700)" {
		t.Errorf("unexpected message: %q", msg)
	}
}

func TestBuildAlertMsg_Fall(t *testing.T) {
	msg := buildAlertMsg("fall", map[string]interface{}{
		"conf": float64(0.95),
		"lat":  float64(31.0),
		"lon":  float64(121.0),
	})
	expected := "跌倒检测告警，置信度=0.95(31.0000, 121.0000)"
	if msg != expected {
		t.Errorf("message = %q, want %q", msg, expected)
	}
}

func TestBuildAlertMsg_Default(t *testing.T) {
	msg := buildAlertMsg("unknown_type", nil)
	if msg != "设备告警：unknown_type" {
		t.Errorf("message = %q, want generic alert", msg)
	}
}

func TestBuildAlertMsg_NoLocation(t *testing.T) {
	msg := buildAlertMsg("sos", map[string]interface{}{})
	if msg != "紧急告警：设备触发SOS按钮" {
		t.Errorf("message = %q, want without location", msg)
	}
}

type fakeNatsMsg struct {
	*nats.Msg
	data []byte
	acked bool
}

func (f *fakeNatsMsg) Ack() { f.acked = true }
func (f *fakeNatsMsg) Data() []byte { return f.data }
