package service

import (
	"context"
	"testing"
	"time"
)

func TestAuditLogger_LogEntry(t *testing.T) {
	logger := NewAuditLogger(100, nil)

	logger.Log(context.Background(), "user-001", ActionUserLogin, "user", "", map[string]any{"method": "password"}, "127.0.0.1", "test-agent")

	entries := logger.GetEntries(10)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Action != ActionUserLogin {
		t.Errorf("expected action user.login, got %s", entries[0].Action)
	}
	if entries[0].UserID != "user-001" {
		t.Errorf("expected user_id user-001, got %s", entries[0].UserID)
	}
}

func TestAuditLogger_GetEntriesByUser(t *testing.T) {
	logger := NewAuditLogger(100, nil)

	logger.Log(context.Background(), "user-001", ActionUserLogin, "user", "", nil, "127.0.0.1", "")
	logger.Log(context.Background(), "user-002", ActionUserLogin, "user", "", nil, "127.0.0.1", "")
	logger.Log(context.Background(), "user-001", ActionUserLogout, "user", "", nil, "127.0.0.1", "")

	entries := logger.GetEntriesByUser("user-001", 10)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries for user-001, got %d", len(entries))
	}
}

func TestAuditLogger_GetEntriesByAction(t *testing.T) {
	logger := NewAuditLogger(100, nil)

	logger.Log(context.Background(), "user-001", ActionUserLogin, "user", "", nil, "127.0.0.1", "")
	logger.Log(context.Background(), "user-001", ActionDeviceBind, "device", "dev-001", nil, "127.0.0.1", "")

	entries := logger.GetEntriesByAction(ActionDeviceBind, 10)
	if len(entries) != 1 {
		t.Fatalf("expected 1 device.bind entry, got %d", len(entries))
	}
}

func TestAuditLogger_MaxSizeTrimming(t *testing.T) {
	logger := NewAuditLogger(3, nil)

	for i := 0; i < 5; i++ {
		logger.Log(context.Background(), "user-001", ActionUserLogin, "user", "", nil, "127.0.0.1", "")
	}

	entries := logger.GetEntries(100)
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries after trimming, got %d", len(entries))
	}
}

func TestAuditLogger_TimestampSet(t *testing.T) {
	logger := NewAuditLogger(100, nil)

	before := time.Now()
	logger.Log(context.Background(), "user-001", ActionUserLogin, "user", "", nil, "127.0.0.1", "")
	after := time.Now()

	entries := logger.GetEntries(1)
	if len(entries) != 1 {
		t.Fatal("expected 1 entry")
	}
	if entries[0].Timestamp.Before(before) || entries[0].Timestamp.After(after) {
		t.Errorf("timestamp out of range: %v", entries[0].Timestamp)
	}
}
