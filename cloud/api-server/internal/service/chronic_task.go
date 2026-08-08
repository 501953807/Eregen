package service

import (
	"context"
	"time"

	"eregen.dev/api-server/internal/model"

	"go.uber.org/zap"
)

// chronicTaskStore defines the data access contract for daily tasks.
type chronicTaskStore interface {
	ListDailyTasks(ctx context.Context, elderlyID, taskDate string) ([]model.ChronicDailyTask, error)
	UpdateDailyTaskComplete(ctx context.Context, taskID string) error
}

// ChronicTaskService manages chronic daily health tasks.
type ChronicTaskService struct {
	store chronicTaskStore
	log   *zap.Logger
}

// NewChronicTaskService creates a new chronic task service.
func NewChronicTaskService(store chronicTaskStore, log *zap.Logger) *ChronicTaskService {
	return &ChronicTaskService{store: store, log: log}
}

// ListTasks returns daily tasks for an elderly person on the given date.
// If date is empty, defaults to today. Results are sorted by scheduled_time.
func (s *ChronicTaskService) ListTasks(ctx context.Context, elderlyID, taskDate string) ([]model.ChronicDailyTask, error) {
	if taskDate == "" {
		taskDate = time.Now().Format("2006-01-02")
	}
	tasks, err := s.store.ListDailyTasks(ctx, elderlyID, taskDate)
	if err != nil {
		s.log.Error("list daily tasks", zap.String("elderly_id", elderlyID), zap.String("task_date", taskDate), zap.Error(err))
		return nil, err
	}
	return tasks, nil
}

// MarkComplete marks a daily task as completed.
func (s *ChronicTaskService) MarkComplete(ctx context.Context, taskID string) error {
	if err := s.store.UpdateDailyTaskComplete(ctx, taskID); err != nil {
		s.log.Error("mark task complete", zap.String("task_id", taskID), zap.Error(err))
		return err
	}
	return nil
}
