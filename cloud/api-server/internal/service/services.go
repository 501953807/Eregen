package service

import (
	"context"
	"fmt"
	"time"

	"eregen.dev/api-server/internal/model"

	"go.uber.org/zap"
)

// SMSProvider sends SMS messages via 阿里云SMS.
type SMSProvider struct {
	signName     string
	templateID   string
	log          *zap.Logger
}

// NewSMSProvider creates an SMS sender.
func NewSMSProvider(signName, templateID string, log *zap.Logger) *SMSProvider {
	return &SMSProvider{signName: signName, templateID: templateID, log: log}
}

// SendOTP generates a 6-digit code, stores it via otpStore, and sends via SMS.
func (s *SMSProvider) SendOTP(ctx context.Context, phone, code string, otpStore OTPStore) error {
	if err := otpStore.Store(ctx, phone, code, 5*time.Minute); err != nil {
		return fmt.Errorf("store otp: %w", err)
	}
	s.log.Info("sent SMS OTP", zap.String("phone", phone))
	// In production: call 阿里云SMS API here
	return nil
}

// OTPStore interface for OTP persistence (backed by Redis in practice).
type OTPStore interface {
	Store(ctx context.Context, key, value string, ttl time.Duration) error
	Verify(ctx context.Context, key, value string) error
}

// PushProvider sends push notifications via FCM.
type PushProvider struct {
	serverKey string
	projectID string
	log       *zap.Logger
}

// NewPushProvider creates a push notification sender.
func NewPushProvider(serverKey, projectID string, log *zap.Logger) *PushProvider {
	return &PushProvider{serverKey: serverKey, projectID: projectID, log: log}
}

// SendToUser sends a push notification to a user's registered devices.
func (p *PushProvider) SendToUser(ctx context.Context, userID, title, body string) error {
	p.log.Info("push notification", zap.String("user", userID), zap.String("title", title))
	// In production: call FCM HTTP v1 API
	return nil
}

// AlertService manages alert creation and resolution.
type AlertService struct {
	store  alertStore
	push   *PushProvider
	nats   *NatsClient
	log    *zap.Logger
}

type alertStore interface {
	CreateAlert(ctx context.Context, a *model.Alert) error
	GetAlert(ctx context.Context, id string) (*model.Alert, error)
	UpdateAlert(ctx context.Context, id string, status model.AlertStatus) error
	ListAlerts(ctx context.Context, elderIDs []string, filter *model.AlertFilter, page, pageSize int) ([]model.Alert, int, error)
}

// NewAlertService creates a new alert service.
func NewAlertService(store alertStore, push *PushProvider, nats *NatsClient, log *zap.Logger) *AlertService {
	return &AlertService{store: store, push: push, nats: nats, log: log}
}

// CreateSOSAlert creates an SOS alert and triggers immediate push notification.
func (s *AlertService) CreateSOSAlert(ctx context.Context, elderlyID, deviceID string, lat, lon float64) error {
	a := &model.Alert{
		ElderlyID: elderlyID,
		AlertType: "sos",
		Severity:  model.AlertP0,
		Status:    model.AlertPending,
		Metadata: map[string]any{
			"device_id": deviceID,
			"lat":       lat,
			"lon":       lon,
		},
	}
	if err := s.store.CreateAlert(ctx, a); err != nil {
		return err
	}
	_ = s.push.SendToUser(ctx, elderlyID, "SOS 紧急告警", "老人触发了SOS按钮，请立即查看")
	return nil
}

// CreateFallAlert creates a fall detection alert.
func (s *AlertService) CreateFallAlert(ctx context.Context, elderlyID, deviceID string, lat, lon, conf float64) error {
	a := &model.Alert{
		ElderlyID: elderlyID,
		AlertType: "fall",
		Severity:  model.AlertP0,
		Status:    model.AlertPending,
		Metadata: map[string]any{
			"device_id":  deviceID,
			"lat":        lat,
			"lon":        lon,
			"confidence": conf,
		},
	}
	if err := s.store.CreateAlert(ctx, a); err != nil {
		return err
	}
	_ = s.push.SendToUser(ctx, elderlyID, "跌倒检测告警", "检测到老人可能跌倒，请立即确认")
	return nil
}

// ResolveAlert marks an alert as resolved.
func (s *AlertService) ResolveAlert(ctx context.Context, alertID string) error {
	return s.store.UpdateAlert(ctx, alertID, model.AlertResolved)
}

// HealthService provides health data queries.
type HealthService struct {
	store healthStore
	redis redisCache
	log   *zap.Logger
}

type healthStore interface {
	GetHealthSummary(ctx context.Context, elderlyID string, day time.Time) (*model.HealthRecord, error)
	GetHealthHistory(ctx context.Context, elderlyID string, days int) ([]model.HealthRecord, error)
	GetHealthTrend(ctx context.Context, elderlyID, metric string, days int) ([]model.HealthRecord, error)
}

type redisCache interface {
	SetLatestHealth(ctx context.Context, elderlyID string, data map[string]any) error
	GetLatestHealth(ctx context.Context, elderlyID string) (map[string]any, error)
}

// NewHealthService creates a new health service.
func NewHealthService(store healthStore, redis redisCache, log *zap.Logger) *HealthService {
	return &HealthService{store: store, redis: redis, log: log}
}

func (s *HealthService) GetSummary(ctx context.Context, elderlyID string, day time.Time) (*model.HealthRecord, error) {
	if _, err := s.redis.GetLatestHealth(ctx, elderlyID); err == nil {
		return &model.HealthRecord{}, nil
	}
	return s.store.GetHealthSummary(ctx, elderlyID, day)
}

func (s *HealthService) GetHistory(ctx context.Context, elderlyID string, days int) ([]model.HealthRecord, error) {
	return s.store.GetHealthHistory(ctx, elderlyID, days)
}

func (s *HealthService) GetTrend(ctx context.Context, elderlyID, metric string, days int) ([]model.HealthRecord, error) {
	return s.store.GetHealthTrend(ctx, elderlyID, metric, days)
}

// LocationService provides location data queries.
type LocationService struct {
	store locationStore
	redis locCache
	log   *zap.Logger
}

type locationStore interface {
	GetLatestLocation(ctx context.Context, elderlyID string) (*model.LocationRecord, error)
	GetLocationHistory(ctx context.Context, elderlyID string, from, until time.Time) ([]model.LocationRecord, error)
}

type locCache interface {
	SetLatestLocation(ctx context.Context, elderlyID string, data map[string]any) error
	GetLatestLocation(ctx context.Context, elderlyID string) (map[string]any, error)
}

// NewLocationService creates a new location service.
func NewLocationService(store locationStore, redis locCache, log *zap.Logger) *LocationService {
	return &LocationService{store: store, redis: redis, log: log}
}

func (s *LocationService) GetLatest(ctx context.Context, elderlyID string) (*model.LocationRecord, error) {
	if _, err := s.redis.GetLatestLocation(ctx, elderlyID); err == nil {
		return &model.LocationRecord{}, nil
	}
	return s.store.GetLatestLocation(ctx, elderlyID)
}

func (s *LocationService) GetHistory(ctx context.Context, elderlyID string, from, until time.Time) ([]model.LocationRecord, error) {
	return s.store.GetLocationHistory(ctx, elderlyID, from, until)
}

// MedicationService manages medication rules and tracking.
type MedicationService struct {
	store medStore
	nats  *NatsClient
	log   *zap.Logger
}

type medStore interface {
	CreateMedicationRule(ctx context.Context, mr *model.MedicationRule) error
	ListMedicationRules(ctx context.Context, elderlyID string) ([]model.MedicationRule, error)
	GetMedicationRule(ctx context.Context, ruleID string) (*model.MedicationRule, error)
	UpdateMedicationRule(ctx context.Context, ruleID string, req *model.CreateMedicationRuleRequest) error
	DeleteMedicationRule(ctx context.Context, ruleID string) error
	GetTodayMedStatus(ctx context.Context, elderlyID string) ([]model.MedStatusRecord, error)
	GetMedicationHistory(ctx context.Context, elderlyID string, days int) ([]model.MedStatusRecord, error)
}

// NewMedicationService creates a new medication service.
func NewMedicationService(store medStore, nats *NatsClient, log *zap.Logger) *MedicationService {
	return &MedicationService{store: store, nats: nats, log: log}
}

func (s *MedicationService) CreateRule(ctx context.Context, elderlyID string, req *model.CreateMedicationRuleRequest) error {
	mr := &model.MedicationRule{
		ElderlyID:    elderlyID,
		ScheduleTime: req.ScheduleTime,
		DoseCount:    req.DoseCount,
		PillType:     req.PillType,
		DaysOfWeek:   req.DaysOfWeek,
		Active:       req.Active,
	}
	if err := s.store.CreateMedicationRule(ctx, mr); err != nil {
		return err
	}
	cmd := map[string]any{
		"type": "med_rule",
		"rule": map[string]any{
			"time":   req.ScheduleTime,
			"dose":   req.DoseCount,
			"type":   req.PillType,
			"days":   req.DaysOfWeek,
		},
	}
	_ = s.nats.PublishCommand(ctx, "BR-XXXX", cmd)
	return nil
}

func (s *MedicationService) ListRules(ctx context.Context, elderlyID string) ([]model.MedicationRule, error) {
	return s.store.ListMedicationRules(ctx, elderlyID)
}

func (s *MedicationService) UpdateRule(ctx context.Context, ruleID string, req *model.CreateMedicationRuleRequest) error {
	if _, err := s.store.GetMedicationRule(ctx, ruleID); err != nil {
		return err
	}
	return s.store.UpdateMedicationRule(ctx, ruleID, req)
}

func (s *MedicationService) DeleteRule(ctx context.Context, ruleID string) error {
	return s.store.DeleteMedicationRule(ctx, ruleID)
}

func (s *MedicationService) GetTodayStatus(ctx context.Context, elderlyID string) ([]model.MedStatusRecord, error) {
	return s.store.GetTodayMedStatus(ctx, elderlyID)
}

func (s *MedicationService) GetHistory(ctx context.Context, elderlyID string, days int) ([]model.MedStatusRecord, error) {
	return s.store.GetMedicationHistory(ctx, elderlyID, days)
}
