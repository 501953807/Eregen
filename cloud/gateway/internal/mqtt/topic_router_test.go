package mqtt

import (
	"testing"
)

func TestExtractDeviceID_Bracelet(t *testing.T) {
	got := extractDeviceID("eregen/device/bracelet/BR-0001/up")
	if got != "BR-0001" {
		t.Errorf("extractDeviceID() = %q, want BR-0001", got)
	}
}

func TestExtractDeviceID_Pillbox(t *testing.T) {
	got := extractDeviceID("eregen/device/pillbox/PX-0001/up")
	if got != "PX-0001" {
		t.Errorf("extractDeviceID() = %q, want PX-0001", got)
	}
}

func TestExtractDeviceID_ShortTopic(t *testing.T) {
	got := extractDeviceID("short/topic")
	if got != "" {
		t.Errorf("extractDeviceID() = %q, want empty", got)
	}
}

func TestExtractDeviceID_EmptyDeviceID(t *testing.T) {
	got := extractDeviceID("eregen/device/bracelet//up")
	if got != "" {
		t.Errorf("extractDeviceID() = %q, want empty", got)
	}
}
