package service

import (
	"context"

	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/store"

	"go.uber.org/zap"
)

// ChronicUricAcidService manages uric acid records for elderly patients.
type ChronicUricAcidService struct {
	store *store.ChronicStore
	log   *zap.Logger
}

// NewChronicUricAcidService creates a new uric acid service.
func NewChronicUricAcidService(svc *store.ChronicStore, log *zap.Logger) *ChronicUricAcidService {
	return &ChronicUricAcidService{store: svc, log: log}
}

// CreateRecord inserts a new manual uric acid reading.
func (s *ChronicUricAcidService) CreateRecord(ctx context.Context, elderlyID string, req *model.ChronicUricAcidRecord) error {
	req.ElderlyID = elderlyID
	if req.Unit == "" {
		req.Unit = "μmol/L"
	}
	if req.Source == "" {
		req.Source = "manual"
	}
	return s.store.SaveUricAcidRecord(ctx, req)
}

// ListRecords returns uric acid readings within the given day window.
func (s *ChronicUricAcidService) ListRecords(ctx context.Context, elderlyID string, days int) ([]model.ChronicUricAcidRecord, error) {
	return s.store.ListUricAcidRecords(ctx, elderlyID, days)
}
