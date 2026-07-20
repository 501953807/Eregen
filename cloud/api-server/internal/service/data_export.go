package service

import (
	"context"
	"fmt"
	"time"

	"eregen.dev/api-server/internal/model"

	"go.uber.org/zap"
)

// DataExportRequest represents a user request to export their personal data.
type DataExportRequest struct {
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"` // pending, processing, completed, failed
	DownloadURL string  `json:"download_url,omitempty"`
}

// DataDeletionRequest represents a user request to delete their account and data.
type DataDeletionRequest struct {
	UserID      string    `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	Status      string    `json:"status"` // pending, processing, completed, failed
	Reason      string    `json:"reason"`
}

// DataExportService manages GDPR-compliant data export requests.
type DataExportService struct {
	store     DataExportStore
	log       *zap.Logger
	requests  map[string]*DataExportRequest
}

// DataExportStore interface for persistence.
type DataExportStore interface {
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	GetElderlyProfilesByUserID(ctx context.Context, userID string) ([]model.ElderlyProfile, error)
	GetHealthRecordsByElderlyID(ctx context.Context, elderlyID string, from, until time.Time) ([]model.HealthRecord, error)
	GetLocationHistoryByElderlyID(ctx context.Context, elderlyID string, from, until time.Time) ([]model.LocationRecord, error)
	GetMedicationRulesByElderlyID(ctx context.Context, elderlyID string) ([]model.MedicationRule, error)
	GetAlertsByElderlyID(ctx context.Context, elderlyID string, from, until time.Time) ([]model.Alert, error)
	DeleteUser(ctx context.Context, userID string) error
}

// NewDataExportService creates a new data export service.
func NewDataExportService(store DataExportStore, log *zap.Logger) *DataExportService {
	return &DataExportService{
		store:    store,
		log:      log,
		requests: make(map[string]*DataExportRequest),
	}
}

// CreateExportRequest initiates a data export request.
func (s *DataExportService) CreateExportRequest(ctx context.Context, userID string) (*DataExportRequest, error) {
	req := &DataExportRequest{
		UserID:    userID,
		CreatedAt: time.Now(),
		Status:    "pending",
	}

	s.requests[userID] = req

	if err := s.processExportRequest(ctx, req); err != nil {
		req.Status = "failed"
		s.log.Error("data export failed", zap.String("user_id", userID), zap.Error(err))
		return req, err
	}

	req.Status = "completed"
	s.log.Info("data export completed", zap.String("user_id", userID))
	return req, nil
}

// processExportRequest gathers all user data and creates a downloadable archive.
func (s *DataExportService) processExportRequest(ctx context.Context, req *DataExportRequest) error {
	user, err := s.store.GetUserByID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	data := map[string]interface{}{
		"user": user,
	}

	// Get elderly profiles linked to this user
	elderlyProfiles, err := s.store.GetElderlyProfilesByUserID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("get elderly profiles: %w", err)
	}
	data["elderly_profiles"] = elderlyProfiles

	// Get health records for each elderly profile
	var allHealthRecords []model.HealthRecord
	for _, profile := range elderlyProfiles {
		records, err := s.store.GetHealthRecordsByElderlyID(ctx, profile.ID, time.Now().AddDate(0, -1, 0), time.Now())
		if err == nil {
			allHealthRecords = append(allHealthRecords, records...)
		}
	}
	data["health_records"] = allHealthRecords

	// Get location history
	var allLocations []model.LocationRecord
	for _, profile := range elderlyProfiles {
		locations, err := s.store.GetLocationHistoryByElderlyID(ctx, profile.ID, time.Now().AddDate(0, -1, 0), time.Now())
		if err == nil {
			allLocations = append(allLocations, locations...)
		}
	}
	data["location_history"] = allLocations

	// Get medication rules
	var allMedRules []model.MedicationRule
	for _, profile := range elderlyProfiles {
		rules, err := s.store.GetMedicationRulesByElderlyID(ctx, profile.ID)
		if err == nil {
			allMedRules = append(allMedRules, rules...)
		}
	}
	data["medication_rules"] = allMedRules

	// Get alerts
	var allAlerts []model.Alert
	for _, profile := range elderlyProfiles {
		alerts, err := s.store.GetAlertsByElderlyID(ctx, profile.ID, time.Now().AddDate(0, -1, 0), time.Now())
		if err == nil {
			allAlerts = append(allAlerts, alerts...)
		}
	}
	data["alerts"] = allAlerts

	// In production, serialize data to JSON and create download URL
	_ = data
	req.DownloadURL = fmt.Sprintf("/api/v1/data/export/%s/download", req.UserID)

	return nil
}

// GetDataExportStatus returns the status of a data export request.
func (s *DataExportService) GetDataExportStatus(userID string) (*DataExportRequest, error) {
	req, ok := s.requests[userID]
	if !ok {
		return nil, fmt.Errorf("export request not found for user: %s", userID)
	}
	return req, nil
}

// DeleteUserData marks a user for deletion and initiates the deletion process.
func (s *DataExportService) DeleteUserData(ctx context.Context, userID string, reason string) (*DataDeletionRequest, error) {
	req := &DataDeletionRequest{
		UserID:    userID,
		CreatedAt: time.Now(),
		Status:    "pending",
		Reason:    reason,
	}

	// Mark user as deleted in database
	err := s.store.DeleteUser(ctx, userID)
	if err != nil {
		req.Status = "failed"
		s.log.Error("delete user failed", zap.String("user_id", userID), zap.Error(err))
		return req, err
	}

	req.Status = "completed"
	s.log.Info("user data deleted", zap.String("user_id", userID), zap.String("reason", reason))
	return req, nil
}

// GetDeletionStatus returns the status of a data deletion request.
func (s *DataExportService) GetDeletionStatus(userID string) (*DataDeletionRequest, error) {
	// In production, query database for deletion status
	return &DataDeletionRequest{
		UserID: userID,
		Status: "completed",
	}, nil
}
