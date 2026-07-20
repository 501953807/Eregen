package service

import (
	"context"
	"fmt"
	"testing"

	"eregen.dev/api-server/internal/model"
)

type mockEmergencyStore struct {
	alerts map[string]*model.Alert
}

func newMockEmergencyStore() *mockEmergencyStore {
	return &mockEmergencyStore{alerts: make(map[string]*model.Alert)}
}

func (m *mockEmergencyStore) CreateAlert(ctx context.Context, a *model.Alert) error {
	m.alerts[a.ID] = a
	return nil
}

func (m *mockEmergencyStore) UpdateAlert(ctx context.Context, id string, status model.AlertStatus) error {
	if a, ok := m.alerts[id]; ok {
		a.Status = status
		return nil
	}
	return fmt.Errorf("alert %s not found", id)
}

func (m *mockEmergencyStore) GetAlert(ctx context.Context, id string) (*model.Alert, error) {
	a, ok := m.alerts[id]
	if !ok {
		return nil, fmt.Errorf("alert %s not found", id)
	}
	return a, nil
}

func TestProcessAlert_P0_SOS(t *testing.T) {
	store := newMockEmergencyStore()
	wf := NewEmergencyResponseWorkflow(store, nil, nil, nil)

	alert := &model.Alert{
		ID:        "alert-sos-001",
		ElderlyID: "elderly-001",
		AlertType: "sos",
		Severity:  model.AlertP0,
		Status:    model.AlertPending,
		Metadata: map[string]any{
			"lat": 31.23,
			"lon": 121.47,
		},
	}

	err := wf.ProcessAlert(context.Background(), alert)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cases := wf.GetActiveCases()
	if len(cases) != 1 {
		t.Fatalf("expected 1 active case, got %d", len(cases))
	}
	if cases[0].Status != "dispatched" {
		t.Errorf("expected status dispatched, got %s", cases[0].Status)
	}
	if len(cases[0].Notifications) != 1 {
		t.Errorf("expected 1 notification, got %d", len(cases[0].Notifications))
	}
}

func TestProcessAlert_P0_Fall(t *testing.T) {
	store := newMockEmergencyStore()
	wf := NewEmergencyResponseWorkflow(store, nil, nil, nil)

	alert := &model.Alert{
		ID:        "alert-fall-001",
		ElderlyID: "elderly-002",
		AlertType: "fall",
		Severity:  model.AlertP0,
		Status:    model.AlertPending,
		Metadata: map[string]any{
			"confidence": 0.95,
		},
	}

	err := wf.ProcessAlert(context.Background(), alert)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cases := wf.GetActiveCases()
	if len(cases) != 1 {
		t.Fatalf("expected 1 active case, got %d", len(cases))
	}
}

func TestProcessAlert_P1_MedMissed(t *testing.T) {
	store := newMockEmergencyStore()
	wf := NewEmergencyResponseWorkflow(store, nil, nil, nil)

	alert := &model.Alert{
		ID:        "alert-med-001",
		ElderlyID: "elderly-003",
		AlertType: "med_missed",
		Severity:  model.AlertP1,
		Status:    model.AlertPending,
		Metadata:  map[string]any{},
	}

	err := wf.ProcessAlert(context.Background(), alert)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cases := wf.GetActiveCases()
	if len(cases) != 1 {
		t.Fatalf("expected 1 active case, got %d", len(cases))
	}
}

func TestProcessAlert_P2_Ignores(t *testing.T) {
	store := newMockEmergencyStore()
	wf := NewEmergencyResponseWorkflow(store, nil, nil, nil)

	alert := &model.Alert{
		ID:        "alert-p2-001",
		ElderlyID: "elderly-004",
		AlertType: "info",
		Severity:  model.AlertP2,
		Status:    model.AlertPending,
	}

	err := wf.ProcessAlert(context.Background(), alert)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cases := wf.GetActiveCases()
	if len(cases) != 0 {
		t.Fatalf("expected 0 active cases for P2, got %d", len(cases))
	}
}

func TestResolveAlert_Success(t *testing.T) {
	store := newMockEmergencyStore()
	wf := NewEmergencyResponseWorkflow(store, nil, nil, nil)

	alert := &model.Alert{
		ID:        "alert-resolve-001",
		ElderlyID: "elderly-005",
		AlertType: "sos",
		Severity:  model.AlertP0,
		Status:    model.AlertPending,
		Metadata:  map[string]any{},
	}

	_ = wf.ProcessAlert(context.Background(), alert)
	err := wf.ResolveAlert(context.Background(), "alert-resolve-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cases := wf.GetActiveCases()
	if len(cases) != 0 {
		t.Fatalf("expected 0 active cases after resolve, got %d", len(cases))
	}
}

func TestResolveAlert_NotFound(t *testing.T) {
	store := newMockEmergencyStore()
	wf := NewEmergencyResponseWorkflow(store, nil, nil, nil)

	err := wf.ResolveAlert(context.Background(), "nonexistent-alert")
	if err == nil {
		t.Fatal("expected error for nonexistent alert")
	}
}
