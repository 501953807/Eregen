package service

import (
	"context"

	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/store"

	"go.uber.org/zap"
)

// ChronicDietService manages diet/meal records for elderly patients.
type ChronicDietService struct {
	store *store.ChronicStore
	log   *zap.Logger
}

// NewChronicDietService creates a new diet service.
func NewChronicDietService(svc *store.ChronicStore, log *zap.Logger) *ChronicDietService {
	return &ChronicDietService{store: svc, log: log}
}

// CreateRecord inserts a new diet/meal log entry.
func (s *ChronicDietService) CreateRecord(ctx context.Context, elderlyID string, req *model.ChronicDietRecord) error {
	req.ElderlyID = elderlyID
	if req.MealType == "" {
		req.MealType = "snack"
	}
	if req.FoodItems == "" {
		req.FoodItems = "[]"
	}
	return s.store.SaveDietRecord(ctx, req)
}

// ListRecords returns diet records within the given day window.
func (s *ChronicDietService) ListRecords(ctx context.Context, elderlyID string, days int) ([]model.ChronicDietRecord, error) {
	return s.store.ListDietRecords(ctx, elderlyID, days)
}
