// Package service provides business logic for admin-api.
package service

import (
	"context"
	"fmt"
	"time"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"

	"go.uber.org/zap"
)

// AuditAction represents the type of operation being audited.
type AuditAction string

const (
	ActionUserLogin       AuditAction = "user.login"
	ActionUserLogout      AuditAction = "user.logout"
	ActionUserRegister    AuditAction = "user.register"
	ActionUserUpdate      AuditAction = "user.update"
	ActionElderlyCreate   AuditAction = "elderly.create"
	ActionElderlyUpdate   AuditAction = "elderly.update"
	ActionDeviceBind      AuditAction = "device.bind"
	ActionDeviceUnbind    AuditAction = "device.unbind"
	ActionMedicationRule  AuditAction = "medication.rule"
	ActionAlertResolve    AuditAction = "alert.resolve"
	ActionOTAUpdate       AuditAction = "ota.update"
	ActionAdminAction     AuditAction = "admin.action"
)

// AuditLogger provides audit logging for sensitive operations.
type AuditLogger struct {
	store store.Store
	log   *zap.Logger
}

// NewAuditLogger creates a new audit logger with database persistence.
func NewAuditLogger(s store.Store, logger *zap.Logger) *AuditLogger {
	return &AuditLogger{
		store: s,
		log:   logger,
	}
}

// Log records an audit entry to the database.
func (l *AuditLogger) Log(ctx context.Context, userID string, action AuditAction, resource string, resourceID string, details map[string]interface{}, ip string, userAgent string) {
	entry := &model.AuditLog{
		ID:         fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		UserID:     userID,
		Action:     string(action),
		Resource:   resource,
		ResourceID: resourceID,
		Details:    details,
		IP:         ip,
		UserAgent:  userAgent,
		CreatedAt:  time.Now(),
	}

	if err := l.store.CreateAuditLog(ctx, entry); err != nil {
		l.log.Error("failed to save audit log", zap.Error(err))
	}

	l.log.Info("audit entry recorded",
		zap.String("user_id", userID),
		zap.String("action", string(action)),
		zap.String("resource", resource),
		zap.String("resource_id", resourceID),
	)
}

// GetEntries returns recent audit entries from database.
func (l *AuditLogger) GetEntries(limit int) []model.AuditLog {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	entries, err := l.store.ListAuditLogs(context.Background(), limit)
	if err != nil {
		l.log.Error("failed to get audit logs", zap.Error(err))
		return nil
	}
	return entries
}

// GetEntriesByUser returns audit entries for a specific user.
func (l *AuditLogger) GetEntriesByUser(userID string, limit int) []model.AuditLog {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	entries, err := l.store.ListAuditLogsByUser(context.Background(), userID, limit)
	if err != nil {
		l.log.Error("failed to get user audit logs", zap.Error(err))
		return nil
	}
	return entries
}

// GetEntriesByAction returns audit entries for a specific action.
func (l *AuditLogger) GetEntriesByAction(action AuditAction, limit int) []model.AuditLog {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	entries, err := l.store.ListAuditLogsByAction(context.Background(), string(action), limit)
	if err != nil {
		l.log.Error("failed to get action audit logs", zap.Error(err))
		return nil
	}
	return entries
}
