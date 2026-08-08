package service

import (
	"context"

	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/store"

	"go.uber.org/zap"
)

// ChronicBPService manages blood pressure records for elderly patients.
type ChronicBPService struct {
	store *store.ChronicStore
	log   *zap.Logger
}

// NewChronicBPService creates a new blood pressure service.
func NewChronicBPService(svc *store.ChronicStore, log *zap.Logger) *ChronicBPService {
	return &ChronicBPService{store: svc, log: log}
}

// CreateRecord inserts a new manual blood pressure reading.
func (s *ChronicBPService) CreateRecord(ctx context.Context, elderlyID string, req *model.ChronicBPRecord) error {
	req.ElderlyID = elderlyID
	return s.store.SaveBPRecord(ctx, req)
}

// ListRecords returns BP readings within the given day window.
func (s *ChronicBPService) ListRecords(ctx context.Context, elderlyID string, days int) ([]model.ChronicBPRecord, error) {
	return s.store.ListBPRecords(ctx, elderlyID, days)
}

// SyncFromDevice inserts a BP reading from a BLE/medical device.
func (s *ChronicBPService) SyncFromDevice(ctx context.Context, elderlyID string, req *model.ChronicBPRecord) error {
	req.ElderlyID = elderlyID
	// Mark source as device for tracking
	if req.Pulse == nil {
		// pulse is optional for device sync
	}
	return s.store.SaveBPRecord(ctx, req)
}
