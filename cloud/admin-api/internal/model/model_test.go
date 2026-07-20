package model

import "testing"

func TestRoleConstants(t *testing.T) {
	tests := []struct {
		name string
		got  Role
		want string
	}{
		{"admin", RoleAdmin, "admin"},
		{"operator", RoleOperator, "operator"},
		{"super_admin", RoleSuperAdmin, "super_admin"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Errorf("Role = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestDashboardStatsDefaults(t *testing.T) {
	stats := DashboardStats{}
	if stats.OnlineDevices != 0 || stats.TotalDevices != 0 || stats.ActiveAlerts != 0 {
		t.Error("zero-value DashboardStats should have 0 counts")
	}
}

func TestDeviceSummaryFields(t *testing.T) {
	dev := DeviceSummary{
		ID:            "d-1",
		DeviceID:      "BR-0001",
		Type:          "bracelet",
		Tier:          "pro",
		Status:        "online",
		FirmwareVer:   "1.2.3",
		OwnerName:     "Alice",
	}
	if dev.ID != "d-1" || dev.DeviceID != "BR-0001" || dev.Status != "online" {
		t.Error("DeviceSummary fields not set correctly")
	}
}

func TestUserSummaryFields(t *testing.T) {
	user := UserSummary{
		ID:   "u-1",
		Name: "Bob",
		Role: "family",
	}
	if user.ID != "u-1" || user.Name != "Bob" || user.Role != "family" {
		t.Error("UserSummary fields not set correctly")
	}
}

func TestAlertSummaryFields(t *testing.T) {
	alert := AlertSummary{
		ID:         "a-1",
		ElderlyID:  "e-1",
		AlertType:  "sos",
		Severity:   "P0",
		Status:     "pending",
		DeviceID:   "BR-0001",
	}
	if alert.AlertType != "sos" || alert.Severity != "P0" || alert.Status != "pending" {
		t.Error("AlertSummary fields not set correctly")
	}
}

func TestSubscriptionStatPercentage(t *testing.T) {
	sub := SubscriptionStat{Tier: "premium", Count: 100, Pct: 50.0}
	if sub.Count != 100 || sub.Pct != 50.0 {
		t.Error("SubscriptionStat fields not set correctly")
	}
}

func TestFirmwareVersionDefaults(t *testing.T) {	fw := FirmwareVersion{}
	if fw.ForceUpdate || fw.IsActive {
		t.Error("zero-value ForceUpdate and IsActive should be false")
	}
}

func TestAPIKeySummaryActiveDefault(t *testing.T) {
	key := APIKeySummary{}
	if key.Active {
		t.Error("zero-value Active should be false")
	}
}
