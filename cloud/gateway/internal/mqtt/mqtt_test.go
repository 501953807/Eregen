// © 2026 Eregen (颐贞). All rights reserved.

package mqtt

import (
	"testing"
)

func TestExtractDeviceID(t *testing.T) {
	tests := []struct {
		name    string
		topic   string
		want    string
		wantErr bool
	}{
		{
			name:  "bracelet device",
			topic: "eregen/device/bracelet/BR-1234/up",
			want:  "BR-1234",
		},
		{
			name:  "pillbox device",
			topic: "eregen/device/pillbox/PX-5678/up",
			want:  "PX-5678",
		},
		{
			name:  "alphanumeric device id",
			topic: "eregen/device/bracelet/BR-AB12CD/up",
			want:  "BR-AB12CD",
		},
		{
			name:    "short topic",
			topic:   "eregen/device/up",
			want:    "",
		},
		{
			name:    "empty topic",
			topic:   "",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDeviceID(tt.topic)
			if got != tt.want {
				t.Errorf("extractDeviceID(%q) = %q, want %q", tt.topic, got, tt.want)
			}
		})
	}
}

func TestValidateDeviceID(t *testing.T) {
	validIDs := []string{
		"BR-1234",
		"PX-5678",
		"BR-ABCD",
		"PX-a1b2",
		"BR-X9Y8Z7",
	}
	invalidIDs := []string{
		"INVALID",
		"BR-",
		"PX",
		"BP-1234",
		"BR+1234",
		"BR_1234",
		"",
	}

	for _, id := range validIDs {
		if !deviceIDRegex.MatchString(id) {
			t.Errorf("expected %q to be valid", id)
		}
	}
	for _, id := range invalidIDs {
		if deviceIDRegex.MatchString(id) {
			t.Errorf("expected %q to be invalid", id)
		}
	}
}

func TestParseAndValidate(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		wantErr bool
		wantType string
		wantDev  string
	}{
		{
			name:    "valid heartbeat",
			payload: `{"type":"heartbeat","dev_id":"BR-1234","bat":85,"ts":1720000000}`,
			wantErr: false,
			wantType: "heartbeat",
			wantDev:  "BR-1234",
		},
		{
			name:    "valid location",
			payload: `{"type":"location","dev_id":"PX-5678","lat":31.23,"lon":121.47,"acc":5,"ts":1720000000}`,
			wantErr: false,
			wantType: "location",
			wantDev:  "PX-5678",
		},
		{
			name:    "missing type",
			payload: `{"dev_id":"BR-1234","ts":1720000000}`,
			wantErr: true,
		},
		{
			name:    "missing dev_id",
			payload: `{"type":"heartbeat","ts":1720000000}`,
			wantErr: true,
		},
		{
			name:    "missing ts",
			payload: `{"type":"heartbeat","dev_id":"BR-1234"}`,
			wantErr: true,
		},
		{
			name:    "invalid device format",
			payload: `{"type":"heartbeat","dev_id":"INVALID","ts":1720000000}`,
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			payload: `{not json`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := ParseAndValidate([]byte(tt.payload))
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if msg.Type != tt.wantType {
				t.Errorf("type = %q, want %q", msg.Type, tt.wantType)
			}
		})
	}
}

func TestJSONDeserialization(t *testing.T) {
	type testCase struct {
		name    string
		payload string
		wantOK  bool
	}

	testCases := []testCase{
		{
			name:    "heartbeat",
			payload: `{"type":"heartbeat","dev_id":"BR-0001","bat":90,"ts":1720000001}`,
			wantOK:  true,
		},
		{
			name:    "health",
			payload: `{"type":"health","dev_id":"BR-0002","hr":72,"spo2":98,"step":3456,"ts":1720000002}`,
			wantOK:  true,
		},
		{
			name:    "sos",
			payload: `{"type":"sos","dev_id":"BR-0003","lat":31.123,"lon":121.456,"ts":1720000003}`,
			wantOK:  true,
		},
		{
			name:    "fall",
			payload: `{"type":"fall","dev_id":"BR-0004","conf":0.95,"lat":31.123,"lon":121.456,"ts":1720000004}`,
			wantOK:  true,
		},
		{
			name:    "med_status",
			payload: `{"type":"med_status","dev_id":"PX-0001","compartment":3,"taken":true,"ts":1720000005}`,
			wantOK:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg, err := ParseAndValidate([]byte(tc.payload))
			if tc.wantOK && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tc.wantOK && err == nil {
				t.Fatal("expected error but got none")
			}
			if tc.wantOK && msg == nil {
				t.Fatal("expected valid message")
			}
		})
	}
}
