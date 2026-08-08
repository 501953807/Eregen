package service

import (
	"context"

	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/store"

	"go.uber.org/zap"
)

// ChronicExerciseService manages exercise session records for elderly patients.
type ChronicExerciseService struct {
	store *store.ChronicStore
	log   *zap.Logger
}

// NewChronicExerciseService creates a new exercise service.
func NewChronicExerciseService(svc *store.ChronicStore, log *zap.Logger) *ChronicExerciseService {
	return &ChronicExerciseService{store: svc, log: log}
}

// CreateRecord inserts a new exercise session record.
func (s *ChronicExerciseService) CreateRecord(ctx context.Context, elderlyID string, req *model.ChronicExerciseRecord) error {
	req.ElderlyID = elderlyID
	if req.Type == "" {
		req.Type = "walking"
	}
	return s.store.SaveExerciseRecord(ctx, req)
}

// ListRecords returns exercise records within the given day window.
func (s *ChronicExerciseService) ListRecords(ctx context.Context, elderlyID string, days int) ([]model.ChronicExerciseRecord, error) {
	return s.store.ListExerciseRecords(ctx, elderlyID, days)
}
