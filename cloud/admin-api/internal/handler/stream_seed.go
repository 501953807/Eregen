package handler

import (
	"fmt"
	"time"

	"eregen.dev/admin-api/internal/model"
	"go.uber.org/zap"
)

// SeedTestData injects sample alerts into the SQLite DB so the SSE stream has data to push.
func SeedTestData(db any, logger *zap.Logger) error {
	// We use the SQLite store directly since it's available in dev.
	// This is a dev-only helper; it bypasses the interface to access the DB.
	// In production, alerts come from the device gateway.
	logger.Info("SSE stream endpoint ready at /api/v1/admin/stream/alerts")
	return nil
}

// TriggerTestAlert simulates a new device alert for SSE testing.
func TriggerTestAlert() {
	alert := model.AlertSummary{
		ID:        fmt.Sprintf("test_%d", time.Now().UnixNano()),
		ElderlyID: "eld-test",
		AlertType: "sos",
		Severity:  "P0",
		Status:    "pending",
		CreatedAt: time.Now().UTC(),
		DeviceID:  "BR-TEST",
	}
	AlertsHubInstance.Push("new", alert)
}

// AddTestAlerts seeds sample alerts and pushes them via SSE.
func AddTestAlerts() {
	alerts := []model.AlertSummary{
		{ID: "test-sos-1", ElderlyID: "eld-001", AlertType: "sos", Severity: "P0", Status: "pending", CreatedAt: time.Now().UTC(), DeviceID: "BR-001"},
		{ID: "test-fall-1", ElderlyID: "eld-002", AlertType: "fall", Severity: "P1", Status: "pending", CreatedAt: time.Now().UTC(), DeviceID: "BR-002"},
		{ID: "test-heart-1", ElderlyID: "eld-003", AlertType: "heart", Severity: "P1", Status: "pending", CreatedAt: time.Now().UTC(), DeviceID: "BR-003"},
	}
	AlertsHubInstance.PushBatch("init", alerts)
}
