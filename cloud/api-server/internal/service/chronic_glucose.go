package service

import (
	"context"

	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/store"

	"go.uber.org/zap"
)

// ChronicGlucoseService manages blood glucose records for elderly patients.
type ChronicGlucoseService struct {
	store *store.ChronicStore
	log   *zap.Logger
}

// NewChronicGlucoseService creates a new glucose service.
func NewChronicGlucoseService(svc *store.ChronicStore, log *zap.Logger) *ChronicGlucoseService {
	return &ChronicGlucoseService{store: svc, log: log}
}

// CreateRecord inserts a new manual or device glucose reading.
func (s *ChronicGlucoseService) CreateRecord(ctx context.Context, elderlyID string, req *model.ChronicGlucoseRecord) error {
	req.ElderlyID = elderlyID
	if req.Unit == "" {
		req.Unit = "mmol/L"
	}
	if req.TestMode == "" {
		req.TestMode = "random"
	}
	if req.Source == "" {
		req.Source = "manual"
	}
	return s.store.SaveGlucoseRecord(ctx, req)
}

// ListRecords returns glucose readings within the given day window.
func (s *ChronicGlucoseService) ListRecords(ctx context.Context, elderlyID string, days int) ([]model.ChronicGlucoseRecord, error) {
	return s.store.ListGlucoseRecords(ctx, elderlyID, days)
}

// GetTrend returns aggregated glucose trend data for visualization.
func (s *ChronicGlucoseService) GetTrend(ctx context.Context, elderlyID string, days int) (*store.GlucoseTrendData, error) {
	return s.store.GetGlucoseTrend(ctx, elderlyID, days)
}

// CreateTestStripRead records a glucose test-strip reading from the bracelet device.
func (s *ChronicGlucoseService) CreateTestStripRead(ctx context.Context, elderlyID string, req *model.ChronicGlucoseRecord) error {
	req.ElderlyID = elderlyID
	if req.Unit == "" {
		req.Unit = "mmol/L"
	}
	req.Source = "test_strip"
	if req.TestMode == "" {
		req.TestMode = "random"
	}
	return s.store.SaveGlucoseRecord(ctx, req)
}
