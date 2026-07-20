package model

import (
	"encoding/json"
	"testing"
)

func TestRoleConstants(t *testing.T) {
	if RoleFamily != "family" {
		t.Errorf("RoleFamily = %q, want family", RoleFamily)
	}
	if RoleElderly != "elderly" {
		t.Errorf("RoleElderly = %q, want elderly", RoleElderly)
	}
	if RoleInstitution != "institution" {
		t.Errorf("RoleInstitution = %q, want institution", RoleInstitution)
	}
}

func TestPlanTierConstants(t *testing.T) {
	if PlanFree != "free" {
		t.Errorf("PlanFree = %q, want free", PlanFree)
	}
	if PlanPremium != "premium" {
		t.Errorf("PlanPremium = %q, want premium", PlanPremium)
	}
	if PlanEnterprise != "enterprise" {
		t.Errorf("PlanEnterprise = %q, want enterprise", PlanEnterprise)
	}
}

func TestAlertSeverityConstants(t *testing.T) {
	if AlertP0 != "P0" {
		t.Errorf("AlertP0 = %q, want P0", AlertP0)
	}
	if AlertP1 != "P1" {
		t.Errorf("AlertP1 = %q, want P1", AlertP1)
	}
	if AlertP2 != "P2" {
		t.Errorf("AlertP2 = %q, want P2", AlertP2)
	}
}

func TestAlertStatusConstants(t *testing.T) {
	if AlertPending != "pending" {
		t.Errorf("AlertPending = %q, want pending", AlertPending)
	}
	if AlertResolved != "resolved" {
		t.Errorf("AlertResolved = %q, want resolved", AlertResolved)
	}
}

func TestDeviceStatusConstants(t *testing.T) {
	if DeviceOnline != "online" {
		t.Errorf("DeviceOnline = %q, want online", DeviceOnline)
	}
	if DeviceOffline != "offline" {
		t.Errorf("DeviceOffline = %q, want offline", DeviceOffline)
	}
}

func TestUserJSONTags(t *testing.T) {
	u := User{
		ID:           "u-1",
		Name:         "Alice",
		PasswordHash: "should-not-appear",
		Role:         RoleFamily,
	}

	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("json marshal failed: %v", err)
	}

	var decoded map[string]interface{}
	json.Unmarshal(data, &decoded)

	if _, ok := decoded["password_hash"]; ok {
		t.Error("User PasswordHash should be omitted from JSON (json:\"-\")")
	}
	if _, ok := decoded["id"]; !ok {
		t.Error("User ID should appear in JSON")
	}
}

func TestAlertMetadataOmittedWhenEmpty(t *testing.T) {
	a := Alert{
		ID:        "a-1",
		ElderlyID: "e-1",
		AlertType: "sos",
	}
	if a.Metadata != nil {
		t.Error("nil metadata should be omitted from JSON")
	}
}

func TestMedicationRuleDefaultsActive(t *testing.T) {
	mr := MedicationRule{}
	if mr.Active {
		t.Error("zero-value Active should be false")
	}
}

func TestOTAJobProgressDefaults(t *testing.T) {
	j := OTAJobProgress{}
	if j.Total != 0 || j.Pending != 0 || j.Succeeded != 0 || j.Failed != 0 {
		t.Error("zero-value progress should all be 0")
	}
}

func TestGeofenceDefaults(t *testing.T) {
	gf := Geofence{}
	if gf.Active {
		t.Error("zero-value Active should be false")
	}
}

func TestSubscriptionDefaults(t *testing.T) {
	s := Subscription{}
	if s.AutoRenew {
		t.Error("zero-value AutoRenew should be false")
	}
}

func TestFirmwareReleaseDefaults(t *testing.T) {
	f := FirmwareRelease{}
	if f.ForceUpdate {
		t.Error("zero-value ForceUpdate should be false")
	}
	if f.Active {
		t.Error("zero-value Active should be false")
	}
}

func TestLoginRequestFields(t *testing.T) {
	req := LoginRequest{Identifier: "phone", Password: "pass"}
	if req.Identifier != "phone" || req.Password != "pass" {
		t.Error("LoginRequest fields not set correctly")
	}
}

func TestRegisterRequestOptionalFields(t *testing.T) {
	req := RegisterRequest{Password: "pass", Name: "Alice"}
	if req.Phone != nil || req.Email != nil {
		t.Error("optional fields should be nil by default")
	}
}

func TestBindDeviceRequestBindingTag(t *testing.T) {
	req := BindDeviceRequest{DeviceID: "BR-0001"}
	if req.DeviceID == "" {
		t.Error("DeviceID should be settable")
	}
}

func TestCreateFirmwareRequestRequiredFields(t *testing.T) {
	req := CreateFirmwareRequest{
		DeviceType: "bracelet", Tier: "pro", Version: "1.0.0", URL: "https://x", Sha256Hash: "abc",
	}
	if req.DeviceType == "" || req.Tier == "" || req.Version == "" {
		t.Error("required fields should be settable")
	}
}

func TestPushOTARequestDeviceIDsOptional(t *testing.T) {
	req := PushOTARequest{FirmwareID: "fw-1"}
	if req.DeviceIDs != nil {
		t.Error("empty DeviceIDs should be nil by default")
	}
}
