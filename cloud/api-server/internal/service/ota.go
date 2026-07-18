package service

import (
	"context"
	"fmt"
	"time"

	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/store"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// OTAService manages firmware releases and OTA push jobs.
type OTAService struct {
	pg   *store.Postgres
	nats *NatsClient
	log  *zap.Logger
}

// NewOTAService creates a new OTA service.
func NewOTAService(pg *store.Postgres, nats *NatsClient, log *zap.Logger) *OTAService {
	return &OTAService{pg: pg, nats: nats, log: log}
}

// CreateFirmwareRelease registers a new firmware version in the system.
func (s *OTAService) CreateFirmwareRelease(ctx context.Context, req *model.CreateFirmwareRequest) (*model.FirmwareRelease, error) {
	release := &model.FirmwareRelease{
		ID:            uuid.New().String(),
		DeviceType:    req.DeviceType,
		Tier:          req.Tier,
		Version:       req.Version,
		URL:           req.URL,
		Sha256Hash:    req.Sha256Hash,
		Changelog:     req.Changelog,
		MinAppVersion: req.MinAppVersion,
		ForceUpdate:   req.ForceUpdate,
		Active:        true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.pg.CreateFirmwareRelease(ctx, release); err != nil {
		return nil, fmt.Errorf("create firmware release: %w", err)
	}

	s.log.Info("new firmware release created",
		zap.String("firmware_id", release.ID),
		zap.String("device_type", release.DeviceType),
		zap.String("version", release.Version),
	)

	return release, nil
}

// ListFirmwareReleases returns all firmware releases with optional filter.
func (s *OTAService) ListFirmwareReleases(ctx context.Context, deviceType, tier string) ([]model.FirmwareRelease, error) {
	releases, err := s.pg.ListFirmwareReleases(ctx, deviceType, tier)
	if err != nil {
		return nil, fmt.Errorf("list firmware releases: %w", err)
	}
	return releases, nil
}

// GetFirmwareRelease returns a single firmware release by ID.
func (s *OTAService) GetFirmwareRelease(ctx context.Context, id string) (*model.FirmwareRelease, error) {
	release, err := s.pg.GetFirmwareRelease(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get firmware release: %w", err)
	}
	return release, nil
}

// CreateOTAJob creates a new OTA push job targeting specific or all devices.
func (s *OTAService) CreateOTAJob(ctx context.Context, firmwareID string, deviceIDs []string) (*model.OTAJob, error) {
	job := &model.OTAJob{
		ID:            uuid.New().String(),
		FirmwareID:    firmwareID,
		TargetDevices: deviceIDs,
		Progress: model.OTAJobProgress{
			Total:   len(deviceIDs),
			Pending: len(deviceIDs),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.pg.CreateOTAJob(ctx, job); err != nil {
		return nil, fmt.Errorf("create OTA job: %w", err)
	}

	s.log.Info("OTA job created",
		zap.String("job_id", job.ID),
		zap.String("firmware_id", firmwareID),
		zap.Int("target_count", len(deviceIDs)),
	)

	return job, nil
}

// PushToDevices sends OTA command to each target device via NATS.
func (s *OTAService) PushToDevices(ctx context.Context, job *model.OTAJob, firmware *model.FirmwareRelease) error {
	cmd := map[string]any{
		"type":    "ota",
		"url":     firmware.URL,
		"hash":    firmware.Sha256Hash,
		"ver":     firmware.Version,
		"force":   firmware.ForceUpdate,
	}

	for _, devID := range job.TargetDevices {
		if err := s.nats.PublishCommand(ctx, devID, cmd); err != nil {
			s.log.Error("publish OTA command",
				zap.String("device_id", devID),
				zap.Error(err),
			)
			s.pg.UpdateOTAJobProgress(ctx, job.ID, func(p *model.OTAJobProgress) {
				p.Pending--
				p.Failed++
			})
			continue
		}

		s.pg.UpdateOTAJobProgress(ctx, job.ID, func(p *model.OTAJobProgress) {
			p.Pending--
			p.Downloading++
		})

		s.log.Debug("OTA command sent",
			zap.String("device_id", devID),
			zap.String("job_id", job.ID),
		)
	}

	return nil
}

// UpdateProgress reports OTA progress for a job (called from device event handler).
func (s *OTAService) UpdateProgress(ctx context.Context, jobID, deviceID string, status string) error {
	switch status {
	case "downloading":
		return s.pg.UpdateOTAJobProgress(ctx, jobID, func(p *model.OTAJobProgress) {
			p.Pending--
			p.Downloading++
		})
	case "succeeding":
		return s.pg.UpdateOTAJobProgress(ctx, jobID, func(p *model.OTAJobProgress) {
			p.Succeeding++
		})
	case "succeeded":
		return s.pg.UpdateOTAJobProgress(ctx, jobID, func(p *model.OTAJobProgress) {
			p.Succeeding--
			p.Succeeded++
			if p.Succeeded+p.Failed == p.Total {
				p.Downloading = 0
			}
		})
	case "failed":
		return s.pg.UpdateOTAJobProgress(ctx, jobID, func(p *model.OTAJobProgress) {
			p.Failed++
			if p.Succeeded+p.Failed == p.Total {
				p.Downloading = 0
				p.Succeeding = 0
			}
		})
	default:
		return fmt.Errorf("unknown OTA status: %s", status)
	}
}

// GetOTAJob returns an OTA job by ID.
func (s *OTAService) GetOTAJob(ctx context.Context, id string) (*model.OTAJob, error) {
	job, err := s.pg.GetOTAJob(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get OTA job: %w", err)
	}
	return job, nil
}

// GetMatchingDevices returns all devices matching type and tier.
func (s *OTAService) GetMatchingDevices(ctx context.Context, deviceType, tier string) ([]model.Device, error) {
	devices, _, err := s.pg.ListDevices(ctx, "", &deviceType, 1, 1000)
	if err != nil {
		return nil, err
	}
	var matched []model.Device
	for _, d := range devices {
		if tier == "" || d.Tier == tier {
			matched = append(matched, d)
		}
	}
	return matched, nil
}
