package service

import (
	"context"
	"fmt"
	"sync"
	"time"

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

// AuditEntry represents a single audit log record.
type AuditEntry struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Action    AuditAction `json:"action"`
	Resource  string    `json:"resource"`
	ResourceID string   `json:"resource_id,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Timestamp time.Time `json:"timestamp"`
}

// AuditLogger provides audit logging for sensitive operations.
type AuditLogger struct {
	mu       sync.Mutex
	entries  []AuditEntry
	maxSize  int
	log      *zap.Logger
}

// NewAuditLogger creates a new audit logger with configurable max entries.
func NewAuditLogger(maxSize int, log *zap.Logger) *AuditLogger {
	if maxSize <= 0 {
		maxSize = 10000 // default
	}
	return &AuditLogger{
		entries: make([]AuditEntry, 0, maxSize),
		maxSize: maxSize,
		log:     log,
	}
}

// Log records an audit entry.
func (l *AuditLogger) Log(ctx context.Context, userID string, action AuditAction, resource string, resourceID string, details map[string]any, ip string, userAgent string) {
	entry := AuditEntry{
		ID:        fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		ResourceID: resourceID,
		Details:   details,
		IP:        ip,
		UserAgent: userAgent,
		Timestamp: time.Now(),
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.entries = append(l.entries, entry)

	// Trim old entries if over max size
	if len(l.entries) > l.maxSize {
		l.entries = l.entries[len(l.entries)-l.maxSize:]
	}

	if l.log != nil {
		l.log.Info("audit entry recorded",
			zap.String("user_id", userID),
			zap.String("action", string(action)),
			zap.String("resource", resource),
			zap.String("resource_id", resourceID),
		)
	}
}

// GetEntries returns recent audit entries, optionally filtered by action.
func (l *AuditLogger) GetEntries(limit int) []AuditEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	if limit <= 0 || limit > len(l.entries) {
		limit = len(l.entries)
	}

	// Return last N entries in chronological order
	start := len(l.entries) - limit
	result := make([]AuditEntry, limit)
	copy(result, l.entries[start:])
	return result
}

// GetEntriesByUser returns audit entries for a specific user.
func (l *AuditLogger) GetEntriesByUser(userID string, limit int) []AuditEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	var filtered []AuditEntry
	for _, e := range l.entries {
		if e.UserID == userID {
			filtered = append(filtered, e)
		}
	}

	if limit > 0 && limit < len(filtered) {
		filtered = filtered[len(filtered)-limit:]
	}

	return filtered
}

// GetEntriesByAction returns audit entries for a specific action.
func (l *AuditLogger) GetEntriesByAction(action AuditAction, limit int) []AuditEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	var filtered []AuditEntry
	for _, e := range l.entries {
		if e.Action == action {
			filtered = append(filtered, e)
		}
	}

	if limit > 0 && limit < len(filtered) {
		filtered = filtered[len(filtered)-limit:]
	}

	return filtered
}
