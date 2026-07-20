package audit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Action represents the type of operation being audited.
type Action string

const (
	ActionInstitutionCreate  Action = "institution.create"
	ActionInstitutionUpdate  Action = "institution.update"
	ActionInstitutionDelete  Action = "institution.delete"
	ActionAPIKeyCreate       Action = "api_key.create"
	ActionHealthDataReceive  Action = "health_data.receive"
	ActionLinkCreate         Action = "link.create"
	ActionLinkDelete         Action = "link.delete"
	ActionEventCreate        Action = "event.create"
	ActionEventUpdate        Action = "event.update"
	ActionEventDelete        Action = "event.delete"
	ActionHealthCheckCreate  Action = "health_check.create"
	ActionCarePlanCreate     Action = "care_plan.create"
	ActionClaimCreate        Action = "claim.create"
	ActionClaimUpdate        Action = "claim.update"
	ActionEvidenceUpload     Action = "evidence.upload"
	ActionExportCreate       Action = "export.create"
	ActionPolicyCreate       Action = "policy.create"
	ActionReminderSend       Action = "reminder.send"
)

// Entry represents a single audit log record.
type Entry struct {
	ID         string                 `json:"id"`
	UserID     string                 `json:"user_id"`
	Action     Action                 `json:"action"`
	Resource   string                 `json:"resource"`
	ResourceID string                 `json:"resource_id,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	IP         string                 `json:"ip"`
	UserAgent  string                 `json:"user_agent"`
	Timestamp  time.Time              `json:"timestamp"`
}

// Logger provides audit logging for B2B services.
type Logger struct {
	mu      sync.Mutex
	entries []Entry
	maxSize int
	log     *zap.Logger
}

// NewLogger creates a new audit logger with configurable max entries.
func NewLogger(maxSize int, log *zap.Logger) *Logger {
	if maxSize <= 0 {
		maxSize = 10000
	}
	return &Logger{
		entries: make([]Entry, 0, maxSize),
		maxSize: maxSize,
		log:     log,
	}
}

// Log records an audit entry.
func (l *Logger) Log(ctx context.Context, userID string, action Action, resource string, resourceID string, details map[string]interface{}, ip string, userAgent string) {
	entry := Entry{
		ID:         fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    details,
		IP:         ip,
		UserAgent:  userAgent,
		Timestamp:  time.Now(),
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

// GetEntries returns recent audit entries.
func (l *Logger) GetEntries(limit int) []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()

	if limit <= 0 || limit > len(l.entries) {
		limit = len(l.entries)
	}

	start := len(l.entries) - limit
	result := make([]Entry, limit)
	copy(result, l.entries[start:])
	return result
}

// GetEntriesByAction returns audit entries for a specific action.
func (l *Logger) GetEntriesByAction(action Action, limit int) []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()

	var filtered []Entry
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
