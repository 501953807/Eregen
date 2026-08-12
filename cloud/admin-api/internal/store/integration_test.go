package store

import (
	"context"
	"testing"
	"time"

	"eregen.dev/admin-api/internal/model"
)

func setupTestStore(t *testing.T) (*SqliteStore, func()) {
	t.Helper()
	db, err := NewSqlite(":memory:")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	s := NewSqliteStore(db)
	return s, func() {
		if db != nil {
			db.Close()
		}
	}
}

func intPtr(v int) *int { return &v }
func int64Ptr(v int64) *int64 { return &v }
func float64Ptr(v float64) *float64 { return &v }

// ========== User Store Tests ==========

func TestUserStore_CreateUser(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	id, err := s.CreateUser(context.Background(), "张三", "zhangsan@example.com", "13800138001", "admin", "password123")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	if id == "" {
		t.Error("CreateUser should return a non-empty ID")
	}
}

func TestUserStore_GetUserByCredential_Email(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	id, err := s.CreateUser(context.Background(), "张三", "zhangsan@example.com", "13800138001", "admin", "password123")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	user, err := s.GetUserByCredential(context.Background(), "email", "zhangsan@example.com", "password123")
	if err != nil {
		t.Fatalf("GetUserByCredential failed: %v", err)
	}
	if user.ID != id {
		t.Errorf("expected user ID %s, got %s", id, user.ID)
	}
	if user.Name != "张三" {
		t.Errorf("expected name '张三', got %s", user.Name)
	}
}

func TestUserStore_GetUserByCredential_InvalidPassword(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	s.CreateUser(context.Background(), "张三", "test@example.com", "13800138001", "admin", "password123")
	_, err := s.GetUserByCredential(context.Background(), "email", "test@example.com", "wrongpassword")
	if err == nil {
		t.Error("expected error for invalid password")
	}
}

func TestUserStore_UpdateUser(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	id, _ := s.CreateUser(context.Background(), "张三", "zhangsan@example.com", "13800138001", "admin", "password123")
	if err := s.UpdateUser(context.Background(), id, "李四", "lisi@example.com", "13900139001", "family"); err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}
	user, err := s.GetUserByCredential(context.Background(), "email", "lisi@example.com", "password123")
	if err != nil {
		t.Fatalf("GetUserByCredential after update failed: %v", err)
	}
	if user.Name != "李四" {
		t.Errorf("expected name '李四', got %s", user.Name)
	}
}

func TestUserStore_SetUserRole(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	id, _ := s.CreateUser(context.Background(), "张三", "zhangsan@example.com", "13800138001", "admin", "password123")
	if err := s.SetUserRole(context.Background(), id, "family"); err != nil {
		t.Fatalf("SetUserRole failed: %v", err)
	}
	user, _ := s.GetUserByCredential(context.Background(), "email", "zhangsan@example.com", "password123")
	if user.Role != "family" {
		t.Errorf("expected role 'family', got %s", user.Role)
	}
}

func TestUserStore_ListUsers(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	for i := 0; i < 5; i++ {
		email := "user" + string(rune('0'+i)) + "@example.com"
		s.CreateUser(context.Background(), "User"+string(rune('0'+i)), email, "1380013800"+string(rune('0'+i)), "admin", "password123")
	}
	users, err := s.ListUsers(context.Background(), 1, 10, "")
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if len(users) != 5 {
		t.Errorf("expected 5 users, got %d", len(users))
	}
}

func TestUserStore_ListUsers_RoleFilter(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	s.CreateUser(context.Background(), "Admin", "admin@example.com", "13800138001", "admin", "password123")
	s.CreateUser(context.Background(), "Operator", "operator@example.com", "13800138002", "operator", "password123")

	users, err := s.ListUsers(context.Background(), 1, 10, "admin")
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("expected 1 admin user, got %d", len(users))
	}
}

func TestUserStore_DeleteUser(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	id, _ := s.CreateUser(context.Background(), "张三", "delete@example.com", "13800138001", "admin", "password123")
	s.DeleteUser(context.Background(), id)
	_, err := s.GetUserByCredential(context.Background(), "email", "delete@example.com", "password123")
	if err == nil {
		t.Error("expected error after delete")
	}
}

// ========== Alert Store Tests ==========

func TestAlertStore_CreateAlert(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	alert := &model.AlertSummary{
		ElderlyID: "elderly-001",
		AlertType: "sos",
		Severity:  "high",
		Status:    "pending",
		DeviceID:  "device-001",
	}
	if err := s.CreateAlert(context.Background(), alert); err != nil {
		t.Fatalf("CreateAlert failed: %v", err)
	}
	if alert.ID == "" {
		t.Error("CreateAlert should set ID")
	}
}

func TestAlertStore_ListAlerts(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	for i := 0; i < 5; i++ {
		s.CreateAlert(context.Background(), &model.AlertSummary{
			ElderlyID: "elderly-001", AlertType: "fall", Severity: "high", Status: "pending", DeviceID: "device-001",
		})
	}
	alerts, err := s.ListAlerts(context.Background(), "", "", 10)
	if err != nil {
		t.Fatalf("ListAlerts failed: %v", err)
	}
	if len(alerts) != 5 {
		t.Errorf("expected 5 alerts, got %d", len(alerts))
	}
}

func TestAlertStore_ListAlerts_SeverityFilter(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	s.CreateAlert(context.Background(), &model.AlertSummary{ElderlyID: "e1", AlertType: "sos", Severity: "high", Status: "pending", DeviceID: "d1"})
	s.CreateAlert(context.Background(), &model.AlertSummary{ElderlyID: "e1", AlertType: "fall", Severity: "medium", Status: "pending", DeviceID: "d1"})
	s.CreateAlert(context.Background(), &model.AlertSummary{ElderlyID: "e1", AlertType: "low_blood", Severity: "low", Status: "pending", DeviceID: "d1"})

	alerts, err := s.ListAlerts(context.Background(), "high", "", 10)
	if err != nil {
		t.Fatalf("ListAlerts failed: %v", err)
	}
	if len(alerts) != 1 {
		t.Errorf("expected 1 high severity alert, got %d", len(alerts))
	}
}

func TestAlertStore_ResolveAlert(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	alert := &model.AlertSummary{ElderlyID: "e1", AlertType: "sos", Severity: "high", Status: "pending", DeviceID: "d1"}
	s.CreateAlert(context.Background(), alert)
	if err := s.ResolveAlert(context.Background(), alert.ID); err != nil {
		t.Fatalf("ResolveAlert failed: %v", err)
	}
	alerts, _ := s.ListAlerts(context.Background(), "", "resolved", 10)
	if len(alerts) != 1 {
		t.Errorf("expected 1 resolved alert, got %d", len(alerts))
	}
}

func TestAlertStore_UpdateAlertStatus(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	alert := &model.AlertSummary{ElderlyID: "e1", AlertType: "sos", Severity: "high", Status: "pending", DeviceID: "d1"}
	s.CreateAlert(context.Background(), alert)
	if err := s.UpdateAlertStatus(context.Background(), alert.ID, "acknowledged"); err != nil {
		t.Fatalf("UpdateAlertStatus failed: %v", err)
	}
}

// ========== ElderlyProfile Store Tests ==========

func TestElderlyStore_CreateElderly(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	ep, err := s.CreateElderly(context.Background(), "张大爷", "1950-01-01", "user-001", []string{"cardiovascular", "diabetes"}, "https://example.com/avatar.jpg")
	if err != nil {
		t.Fatalf("CreateElderly failed: %v", err)
	}
	if ep.ID == "" {
		t.Error("CreateElderly should return a non-empty ID")
	}
	if ep.Name != "张大爷" {
		t.Errorf("expected name '张大爷', got %s", ep.Name)
	}
}

func TestElderlyStore_GetElderly(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	ep, _ := s.CreateElderly(context.Background(), "张大爷", "1950-01-01", "user-001", nil, "")
	found, err := s.GetElderly(context.Background(), ep.ID)
	if err != nil {
		t.Fatalf("GetElderly failed: %v", err)
	}
	if found.Name != "张大爷" {
		t.Errorf("expected name '张大爷', got %s", found.Name)
	}
}

func TestElderlyStore_ListElderly(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	for i := 0; i < 5; i++ {
		s.CreateElderly(context.Background(), "老人"+string(rune('0'+i)), "1950-01-01", "user-001", nil, "")
	}
	elders, err := s.ListElderly(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("ListElderly failed: %v", err)
	}
	if len(elders) != 5 {
		t.Errorf("expected 5 elders, got %d", len(elders))
	}
}

func TestElderlyStore_UpdateElderly(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	ep, _ := s.CreateElderly(context.Background(), "张大爷", "1950-01-01", "user-001", nil, "")
	updated, err := s.UpdateElderly(context.Background(), ep.ID, "李大爷", "1948-05-15", "user-001", []string{"diabetes"}, "")
	if err != nil {
		t.Fatalf("UpdateElderly failed: %v", err)
	}
	if updated.Name != "李大爷" {
		t.Errorf("expected name '李大爷', got %s", updated.Name)
	}
}

func TestElderlyStore_DeleteElderly(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	ep, _ := s.CreateElderly(context.Background(), "张大爷", "1950-01-01", "user-001", nil, "")
	s.DeleteElderly(context.Background(), ep.ID)
	_, err := s.GetElderly(context.Background(), ep.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

// ========== Device Store Tests ==========

func TestDeviceStore_ListDevices(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	s.db.Exec("INSERT INTO devices (id, device_id, device_type, tier, status, owner_user_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))", "dev-001", "BR-001", "bracelet", "plus", "online", "user-001")
	s.db.Exec("INSERT INTO devices (id, device_id, device_type, tier, status, owner_user_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))", "dev-002", "PX-001", "pillbox", "pro", "offline", "user-001")

	devices, err := s.ListDevices(context.Background(), 1, 10, "", "", "")
	if err != nil {
		t.Fatalf("ListDevices failed: %v", err)
	}
	if len(devices) != 2 {
		t.Errorf("expected 2 devices, got %d", len(devices))
	}
}

func TestDeviceStore_ListDevices_StatusFilter(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	s.db.Exec("INSERT INTO devices (id, device_id, device_type, tier, status, owner_user_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))", "dev-001", "BR-001", "bracelet", "plus", "online", "user-001")
	s.db.Exec("INSERT INTO devices (id, device_id, device_type, tier, status, owner_user_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))", "dev-002", "PX-001", "pillbox", "pro", "offline", "user-001")

	devices, err := s.ListDevices(context.Background(), 1, 10, "online", "", "")
	if err != nil {
		t.Fatalf("ListDevices failed: %v", err)
	}
	if len(devices) != 1 {
		t.Errorf("expected 1 online device, got %d", len(devices))
	}
}

func TestDeviceStore_GetDeviceByID(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	// Create user first so LEFT JOIN returns non-NULL name
	userID, _ := s.CreateUser(context.Background(), "测试用户", "test@example.com", "13800138001", "admin", "password123")
	s.db.Exec("INSERT INTO devices (id, device_id, device_type, tier, status, owner_user_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))", "dev-001", "BR-001", "bracelet", "plus", "online", userID)

	device, err := s.GetDeviceByID(context.Background(), "dev-001")
	if err != nil {
		t.Fatalf("GetDeviceByID failed: %v", err)
	}
	if device.ID != "dev-001" {
		t.Errorf("expected ID 'dev-001', got %s", device.ID)
	}
}

func TestDeviceStore_UpdateDeviceConfig(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	sqlDB, _ := NewSqlite(":memory:")
	_ = migrate(sqlDB)
	ds := NewSqliteStore(sqlDB)
	defer sqlDB.Close()
	ds.db.Exec("INSERT INTO devices (id, device_id, device_type, tier, status, owner_user_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))", "dev-001", "BR-001", "bracelet", "plus", "online", "user-001")

	if err := s.UpdateDeviceConfig(context.Background(), "dev-001", map[string]interface{}{"volume": 80, "brightness": 50}); err != nil {
		t.Fatalf("UpdateDeviceConfig failed: %v", err)
	}
}

// ========== Health Record Tests ==========

func TestElderlyStore_CreateHealthRecord(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	ep, _ := s.CreateElderly(context.Background(), "张大爷", "1950-01-01", "user-001", nil, "")
	record := &model.HealthRecordRow{
		ElderlyID: ep.ID,
		HR:        intPtr(72),
		SpO2:      intPtr(98),
		Steps:     int64Ptr(3456),
		Timestamp: time.Now(),
	}
	if err := s.CreateHealthRecord(context.Background(), record); err != nil {
		t.Fatalf("CreateHealthRecord failed: %v", err)
	}
}

func TestElderlyStore_GetElderlyHealthRecords(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	ep, _ := s.CreateElderly(context.Background(), "张大爷", "1950-01-01", "user-001", nil, "")
	for i := 0; i < 5; i++ {
		h := 70 + i
		spo2 := 98 - i
		step := int64(3000 + i*100)
		record := &model.HealthRecordRow{
			ElderlyID: ep.ID,
			HR:        &h,
			SpO2:      &spo2,
			Steps:     &step,
			Timestamp: time.Now().Add(time.Duration(-i) * time.Hour),
		}
		s.CreateHealthRecord(context.Background(), record)
	}
	records, err := s.GetElderlyHealthRecords(context.Background(), ep.ID, 5)
	if err != nil {
		t.Fatalf("GetElderlyHealthRecords failed: %v", err)
	}
	if len(records) != 5 {
		t.Errorf("expected 5 health records, got %d", len(records))
	}
}

// ========== Medication Rule Tests ==========

func TestElderlyStore_CreateMedicationRule(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	ep, _ := s.CreateElderly(context.Background(), "张大爷", "1950-01-01", "user-001", nil, "")
	rule := &model.MedicationRuleRow{
		ElderlyID:    ep.ID,
		ScheduleTime: "08:00",
		PillType:     "capsule",
		DoseCount:    1,
		Active:       true,
	}
	if err := s.CreateMedicationRule(context.Background(), ep.ID, rule); err != nil {
		t.Fatalf("CreateMedicationRule failed: %v", err)
	}
}

func TestElderlyStore_GetElderlyMedicationRules(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	ep, _ := s.CreateElderly(context.Background(), "张大爷", "1950-01-01", "user-001", nil, "")
	rule := &model.MedicationRuleRow{ElderlyID: ep.ID, ScheduleTime: "08:00", PillType: "capsule", DoseCount: 1, Active: true}
	s.CreateMedicationRule(context.Background(), ep.ID, rule)

	rules, err := s.GetElderlyMedicationRules(context.Background(), ep.ID)
	if err != nil {
		t.Fatalf("GetElderlyMedicationRules failed: %v", err)
	}
	if len(rules) != 1 {
		t.Errorf("expected 1 medication rule, got %d", len(rules))
	}
}

func TestElderlyStore_UpdateMedicationRule(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	ep, _ := s.CreateElderly(context.Background(), "张大爷", "1950-01-01", "user-001", nil, "")
	rule := &model.MedicationRuleRow{ElderlyID: ep.ID, ScheduleTime: "08:00", PillType: "capsule", DoseCount: 1, Active: true}
	s.CreateMedicationRule(context.Background(), ep.ID, rule)

	if err := s.UpdateMedicationRule(context.Background(), ep.ID, rule.ID, map[string]interface{}{"dose_count": 2}); err != nil {
		t.Fatalf("UpdateMedicationRule failed: %v", err)
	}
}

func TestElderlyStore_DeleteMedicationRule(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	ep, _ := s.CreateElderly(context.Background(), "张大爷", "1950-01-01", "user-001", nil, "")
	rule := &model.MedicationRuleRow{ElderlyID: ep.ID, ScheduleTime: "08:00", PillType: "capsule", DoseCount: 1, Active: true}
	s.CreateMedicationRule(context.Background(), ep.ID, rule)

	if err := s.DeleteMedicationRule(context.Background(), ep.ID, rule.ID); err != nil {
		t.Fatalf("DeleteMedicationRule failed: %v", err)
	}
}

// ========== Location History Tests ==========

func TestElderlyStore_CreateLocation(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	ep, _ := s.CreateElderly(context.Background(), "张大爷", "1950-01-01", "user-001", nil, "")
	loc := &model.LocationPoint{
		ElderlyID: ep.ID,
		Lat:       31.2304,
		Lon:       121.4737,
		Accuracy:  float64Ptr(10),
		Timestamp: time.Now(),
	}
	if err := s.CreateLocation(context.Background(), loc); err != nil {
		t.Fatalf("CreateLocation failed: %v", err)
	}
}

func TestElderlyStore_GetElderlyLocationHistory(t *testing.T) {
	s, cleanup := setupTestStore(t)
	defer cleanup()

	ep, _ := s.CreateElderly(context.Background(), "张大爷", "1950-01-01", "user-001", nil, "")
	for i := 0; i < 5; i++ {
		acc := float64(10 + i)
		loc := &model.LocationPoint{
			ElderlyID: ep.ID,
			Lat:       31.2304 + float64(i)*0.001,
			Lon:       121.4737 + float64(i)*0.001,
			Accuracy:  &acc,
			Timestamp: time.Now().Add(time.Duration(-i) * time.Hour),
		}
		s.CreateLocation(context.Background(), loc)
	}
	locations, err := s.GetElderlyLocationHistory(context.Background(), ep.ID, 5)
	if err != nil {
		t.Fatalf("GetElderlyLocationHistory failed: %v", err)
	}
	if len(locations) != 5 {
		t.Errorf("expected 5 locations, got %d", len(locations))
	}
}
