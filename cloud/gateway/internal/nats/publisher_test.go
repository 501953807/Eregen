package nats

import (
	"testing"
)

func TestNewClient_Defaults(t *testing.T) {
	c := NewClient(Config{})
	if c.gatewayID == "" {
		t.Error("gateway ID should not be empty")
	}
	if c.stream != "DEVICE_EVENTS" {
		t.Errorf("stream = %q, want DEVICE_EVENTS", c.stream)
	}
}

func TestNewClient_CustomGatewayID(t *testing.T) {
	c := NewClient(Config{GatewayID: "gw-1"})
	if c.gatewayID != "gw-1" {
		t.Errorf("gatewayID = %q, want gw-1", c.gatewayID)
	}
}

func TestNewClient_CustomStream(t *testing.T) {
	c := NewClient(Config{StreamName: "CUSTOM_STREAM"})
	if c.stream != "CUSTOM_STREAM" {
		t.Errorf("stream = %q, want CUSTOM_STREAM", c.stream)
	}
}

func TestCommunitySubjectPrefix(t *testing.T) {
	if communitySubjectPrefix != "eregen.community.wb." {
		t.Errorf("communitySubjectPrefix = %q, want eregen.community.wb.", communitySubjectPrefix)
	}
}

func TestMedicalSubjectPrefix(t *testing.T) {
	if medicalSubjectPrefix != "eregen.medical.wb." {
		t.Errorf("medicalSubjectPrefix = %q, want eregen.medical.wb.", medicalSubjectPrefix)
	}
}
