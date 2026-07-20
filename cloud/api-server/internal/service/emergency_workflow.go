package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"eregen.dev/api-server/internal/model"

	"go.uber.org/zap"
)

// EmergencyResponseWorkflow manages P0/P1 alert escalation and response tracking.
type EmergencyResponseWorkflow struct {
	store       emergencyStore
	push        *PushProvider
	tokenStore  TokenStore
	log         *zap.Logger
	mu          sync.Mutex
	activeCases map[string]*EmergencyCase
}

type emergencyStore interface {
	CreateAlert(ctx context.Context, a *model.Alert) error
	UpdateAlert(ctx context.Context, id string, status model.AlertStatus) error
	GetAlert(ctx context.Context, id string) (*model.Alert, error)
}

// EmergencyCase tracks an active emergency response.
type EmergencyCase struct {
	AlertID      string
	ElderlyID    string
	Severity     model.AlertSeverity
	Status       string // pending, dispatched, responding, resolved
	CreatedAt    time.Time
	DispatchedAt *time.Time
	ResolvedAt   *time.Time
	Notifications []NotificationRecord
}

// NotificationRecord tracks each notification sent during escalation.
type NotificationRecord struct {
	Channel   string
	SentAt    time.Time
	Success   bool
	Message   string
}

// NewEmergencyResponseWorkflow creates a new workflow manager.
func NewEmergencyResponseWorkflow(store emergencyStore, push *PushProvider, tokenStore TokenStore, log *zap.Logger) *EmergencyResponseWorkflow {
	return &EmergencyResponseWorkflow{
		store:       store,
		push:        push,
		tokenStore:  tokenStore,
		log:         log,
		activeCases: make(map[string]*EmergencyCase),
	}
}

// ProcessAlert handles incoming alerts and triggers appropriate escalation.
func (w *EmergencyResponseWorkflow) ProcessAlert(ctx context.Context, alert *model.Alert) error {
	if alert.Severity != model.AlertP0 && alert.Severity != model.AlertP1 {
		return nil // Only P0/P1 trigger emergency workflow
	}

	// Persist the alert so downstream workflows can resolve it
	if err := w.store.CreateAlert(ctx, alert); err != nil {
		w.logIf("failed to persist alert", zap.String("alert_id", alert.ID), zap.Error(err))
		return fmt.Errorf("persist alert: %w", err)
	}

	w.mu.Lock()
	w.activeCases[alert.ID] = &EmergencyCase{
		AlertID:   alert.ID,
		ElderlyID: alert.ElderlyID,
		Severity:  alert.Severity,
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	w.mu.Unlock()

	switch alert.Severity {
	case model.AlertP0:
		return w.handleP0(ctx, alert)
	case model.AlertP1:
		return w.handleP1(ctx, alert)
	default:
		return nil
	}
}

// handleP0 triggers immediate multi-channel notification for critical alerts.
func (w *EmergencyResponseWorkflow) handleP0(ctx context.Context, alert *model.Alert) error {
	w.logIf("P0 alert detected - triggering immediate escalation",
		zap.String("alert_id", alert.ID),
		zap.String("elderly_id", alert.ElderlyID),
		zap.String("type", alert.AlertType),
	)

	// Send immediate push notification
	title := w.p0Title(alert.AlertType)
	body := w.p0Body(alert.AlertType, alert.Metadata)

	if w.push != nil {
		err := w.push.SendToUser(ctx, alert.ElderlyID, title, body, w.tokenStore)
		if err != nil {
			w.logIf("p0 push notification failed", zap.Error(err))
			return fmt.Errorf("p0 push failed: %w", err)
		}
		w.recordNotification(alert.ID, "fcm", true, "P0 push sent")
	} else {
		w.recordNotification(alert.ID, "none", true, "P0 push skipped (no push provider)")
	}

	// Mark as dispatched after successful first notification
	w.markDispatched(alert.ID)

	return nil
}

// handleP1 triggers escalated notification for serious but non-critical alerts.
func (w *EmergencyResponseWorkflow) handleP1(ctx context.Context, alert *model.Alert) error {
	w.logIf("P1 alert detected - escalating",
		zap.String("alert_id", alert.ID),
		zap.String("elderly_id", alert.ElderlyID),
	)

	title := w.p1Title(alert.AlertType)
	body := w.p1Body(alert.AlertType, alert.Metadata)

	if w.push != nil {
		err := w.push.SendToUser(ctx, alert.ElderlyID, title, body, w.tokenStore)
		if err != nil {
			w.logIf("p1 push notification failed", zap.Error(err))
			return fmt.Errorf("p1 push failed: %w", err)
		}
		w.recordNotification(alert.ID, "fcm", true, "P1 push sent")
	} else {
		w.recordNotification(alert.ID, "none", true, "P1 push skipped (no push provider)")
	}

	w.markDispatched(alert.ID)

	return nil
}

// ResolveAlert marks an emergency case as resolved.
func (w *EmergencyResponseWorkflow) ResolveAlert(ctx context.Context, alertID string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	case_, ok := w.activeCases[alertID]
	if !ok {
		return fmt.Errorf("emergency case not found: %s", alertID)
	}

	now := time.Now()
	case_.Status = "resolved"
	case_.ResolvedAt = &now

	return w.store.UpdateAlert(ctx, alertID, model.AlertResolved)
}

// GetActiveCases returns all active emergency cases.
func (w *EmergencyResponseWorkflow) GetActiveCases() []*EmergencyCase {
	w.mu.Lock()
	defer w.mu.Unlock()

	var cases []*EmergencyCase
	for _, c := range w.activeCases {
		if c.Status != "resolved" {
			cases = append(cases, c)
		}
	}
	return cases
}

func (w *EmergencyResponseWorkflow) p0Title(alertType string) string {
	switch alertType {
	case "sos":
		return "SOS 紧急告警"
	case "fall":
		return "跌倒检测告警"
	default:
		return "紧急健康告警"
	}
}

func (w *EmergencyResponseWorkflow) p0Body(alertType string, metadata map[string]any) string {
	switch alertType {
	case "sos":
		return "老人触发了SOS按钮，请立即查看位置并确认安全状况"
	case "fall":
		conf := "未知"
		if c, ok := metadata["confidence"].(float64); ok {
			conf = fmt.Sprintf("%.0f%%", c*100)
		}
		return fmt.Sprintf("检测到跌倒事件，置信度%s，请立即确认老人安全", conf)
	default:
		return "检测到异常事件，请立即查看"
	}
}

func (w *EmergencyResponseWorkflow) p1Title(alertType string) string {
	switch alertType {
	case "med_missed":
		return "用药提醒"
	case "geofence_breach":
		return "电子围栏告警"
	case "device_offline":
		return "设备离线"
	default:
		return "健康告警"
	}
}

func (w *EmergencyResponseWorkflow) p1Body(alertType string, metadata map[string]any) string {
	switch alertType {
	case "med_missed":
		return "老人可能漏服药物，请提醒确认"
	case "geofence_breach":
		return "老人已离开安全区域，请确认位置"
	case "device_offline":
		return "设备已离线超过阈值，请检查设备状态"
	default:
		return "请关注老人健康状况"
	}
}

func (w *EmergencyResponseWorkflow) logIf(msg string, fields ...zap.Field) {
	if w.log != nil {
		w.log.Warn(msg, fields...)
	}
}

func (w *EmergencyResponseWorkflow) markDispatched(alertID string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if c, ok := w.activeCases[alertID]; ok {
		now := time.Now()
		c.Status = "dispatched"
		c.DispatchedAt = &now
	}
}

func (w *EmergencyResponseWorkflow) recordNotification(alertID, channel string, success bool, msg string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	c, ok := w.activeCases[alertID]
	if !ok {
		return
	}

	c.Notifications = append(c.Notifications, NotificationRecord{
		Channel: channel,
		SentAt:  time.Now(),
		Success: success,
		Message: msg,
	})
}
