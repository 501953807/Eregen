package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"eregen.dev/api-server/internal/model"

	"go.uber.org/zap"
)

// mockOTPStore implements OTPStore for testing
type mockOTPStore struct {
	stored map[string]string
}

func newMockOTPStore() *mockOTPStore {
	return &mockOTPStore{stored: make(map[string]string)}
}

func (m *mockOTPStore) Store(ctx context.Context, key, value string, ttl time.Duration) error {
	m.stored[key] = value
	return nil
}

func (m *mockOTPStore) Verify(ctx context.Context, key, value string) error {
	if stored, ok := m.stored[key]; !ok || stored != value {
		return context.Canceled
	}
	delete(m.stored, key)
	return nil
}

func TestSMSProvider_SendOTPSkipsWhenNotConfigured(t *testing.T) {
	log := zap.NewNop()
	sms := NewSMSProvider("", "", "", "", log)
	ctx := context.Background()
	store := newMockOTPStore()

	err := sms.SendOTP(ctx, "13800138000", "123456", store)
	if err != nil {
		t.Fatalf("SendOTP failed: %v", err)
	}

	if _, ok := store.stored["13800138000"]; !ok {
		t.Error("OTP should be stored even when SMS not configured")
	}
}

func TestSMSProvider_RateLimit(t *testing.T) {
	log := zap.NewNop()
	sms := NewSMSProvider("key", "secret", "sign", "tmpl", log)
	ctx := context.Background()
	store := newMockOTPStore()

	// Fill up the rate limit
	for i := 0; i < 10; i++ {
		_ = sms.allowSend(10) // bypass internal counter
	}

	err := sms.SendOTP(ctx, "13800138000", "123456", store)
	if err == nil {
		t.Error("expected rate limit error after exceeding limit")
	}
}

func TestPushProvider_SendToUserNoTokens(t *testing.T) {
	log := zap.NewNop()
	push := NewPushProvider("", "", log)
	ctx := context.Background()

	// Mock token store that returns no tokens
	mockStore := &mockTokenStore{}
	err := push.SendToUser(ctx, "user-1", "title", "body", mockStore)
	if err != nil {
		t.Fatalf("SendToUser failed: %v", err)
	}
}

type mockTokenStore struct{}

func (m *mockTokenStore) GetDeviceTokens(ctx context.Context, userID string) ([]string, error) {
	return nil, nil
}

func TestHealthService_GetSummaryFromCache(t *testing.T) {
	log := zap.NewNop()

	// Mock health store that returns a valid record
	mockHealthStore := &mockHealthStore{
		summaryErr: nil,
		summaryRec: &model.HealthRecord{
			ElderlyID: "elderly-1",
			Timestamp: time.Now(),
		},
	}

	svc := NewHealthService(mockHealthStore, &mockRedisCache{}, log)
	ctx := context.Background()

	rec, err := svc.GetSummary(ctx, "elderly-1", time.Now())
	if err != nil {
		t.Fatalf("GetSummary failed: %v", err)
	}
	if rec.ElderlyID != "elderly-1" {
		t.Errorf("ElderlyID = %q, want elderly-1", rec.ElderlyID)
	}
}

type mockHealthStore struct {
	summaryErr error
	summaryRec *model.HealthRecord
}

func (m *mockHealthStore) GetHealthSummary(ctx context.Context, elderlyID string, day time.Time) (*model.HealthRecord, error) {
	return m.summaryRec, m.summaryErr
}

func (m *mockHealthStore) GetHealthHistory(ctx context.Context, elderlyID string, days int) ([]model.HealthRecord, error) {
	return nil, nil
}

func (m *mockHealthStore) GetHealthTrend(ctx context.Context, elderlyID, metric string, days int) ([]model.HealthRecord, error) {
	return nil, nil
}

type mockRedisCache struct{}

func (m *mockRedisCache) SetLatestHealth(ctx context.Context, elderlyID string, data map[string]any) error {
	return nil
}

func (m *mockRedisCache) GetLatestHealth(ctx context.Context, elderlyID string) (map[string]any, error) {
	return nil, fmt.Errorf("cache miss")
}

func TestLocationService_GetLatestFromCache(t *testing.T) {
	log := zap.NewNop()

	// Mock location store that returns a valid record
	mockLocStore := &mockLocationStore{
		latestErr: nil,
		latestRec: &model.LocationRecord{
			ElderlyID: "elderly-1",
			Lat:       31.23,
			Lon:       121.47,
		},
	}

	svc := NewLocationService(mockLocStore, &mockLocCache{}, log)
	ctx := context.Background()

	loc, err := svc.GetLatest(ctx, "elderly-1")
	if err != nil {
		t.Fatalf("GetLatest failed: %v", err)
	}
	if loc.ElderlyID != "elderly-1" {
		t.Errorf("ElderlyID = %q, want elderly-1", loc.ElderlyID)
	}
}

type mockLocationStore struct {
	latestErr error
	latestRec *model.LocationRecord
}

func (m *mockLocationStore) GetLatestLocation(ctx context.Context, elderlyID string) (*model.LocationRecord, error) {
	return m.latestRec, m.latestErr
}

func (m *mockLocationStore) GetLocationHistory(ctx context.Context, elderlyID string, from, until time.Time) ([]model.LocationRecord, error) {
	return nil, nil
}

func (m *mockLocationStore) CreateGeofence(ctx context.Context, gf *model.Geofence) error {
	return nil
}

func (m *mockLocationStore) ListGeofences(ctx context.Context, elderlyID string) ([]model.Geofence, error) {
	return nil, nil
}

type mockLocCache struct{}

func (m *mockLocCache) SetLatestLocation(ctx context.Context, elderlyID string, data map[string]any) error {
	return nil
}

func (m *mockLocCache) GetLatestLocation(ctx context.Context, elderlyID string) (map[string]any, error) {
	return nil, fmt.Errorf("cache miss")
}

func TestMedicationService_CreateRulePublishesCommand(t *testing.T) {
	log := zap.NewNop()
	nats := &mockNatsClient{}

	// Mock medication store
	mockMedStore := &mockMedStore{
		deviceID: "PX-TEST01",
	}

	svc := NewMedicationService(mockMedStore, nats, log)
	ctx := context.Background()

	req := &model.CreateMedicationRuleRequest{
		ScheduleTime: "08:00",
		DoseCount:    1,
		PillType:     "capsule",
		DaysOfWeek:   []int{1, 2, 3, 4, 5},
		Active:       true,
	}

	err := svc.CreateRule(ctx, "elderly-1", req)
	if err != nil {
		t.Fatalf("CreateRule failed: %v", err)
	}

	if nats.publishedCmd == nil {
		t.Error("expected NATS command to be published")
	}
	if nats.publishedCmd["type"] != "med_rule" {
		t.Errorf("published command type = %v, want med_rule", nats.publishedCmd["type"])
	}
}

type mockNatsClient struct {
	publishedCmd map[string]any
}

func (m *mockNatsClient) PublishCommand(ctx context.Context, deviceID string, cmd any) error {
	m.publishedCmd = cmd.(map[string]any)
	return nil
}

type mockMedStore struct {
	deviceID string
}

func (m *mockMedStore) CreateMedicationRule(ctx context.Context, mr *model.MedicationRule) error {
	return nil
}

func (m *mockMedStore) ListMedicationRules(ctx context.Context, elderlyID string) ([]model.MedicationRule, error) {
	return nil, nil
}

func (m *mockMedStore) GetMedicationRule(ctx context.Context, ruleID string) (*model.MedicationRule, error) {
	return &model.MedicationRule{}, nil
}

func (m *mockMedStore) UpdateMedicationRule(ctx context.Context, ruleID string, req *model.CreateMedicationRuleRequest) error {
	return nil
}

func (m *mockMedStore) DeleteMedicationRule(ctx context.Context, ruleID string) error {
	return nil
}

func (m *mockMedStore) GetTodayMedStatus(ctx context.Context, elderlyID string) ([]model.MedStatusRecord, error) {
	return nil, nil
}

func (m *mockMedStore) GetMedicationHistory(ctx context.Context, elderlyID string, days int) ([]model.MedStatusRecord, error) {
	return nil, nil
}

func (m *mockMedStore) GetDeviceByElderlyID(ctx context.Context, elderlyID string) (string, error) {
	return m.deviceID, nil
}
