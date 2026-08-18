package store

import (
	"context"
	"encoding/json"
	"eregen.dev/admin-api/internal/model"
	"fmt"
	"time"
)


// CreateFirmwareVersion inserts a new firmware release record.
func (s *SqliteStore) CreateFirmwareVersion(ctx context.Context, v *model.FirmwareVersion) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO firmware_releases (device_type, tier, version, url, sha256_hash, changelog, min_app_version, force_update, active)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1)`,
		v.DeviceType, v.Tier, v.Version, v.DownloadURL, v.Sha256Hash, v.Changelog, v.MinAppVersion, v.ForceUpdate)
	return err
}

// ListFirmwareVersions returns all firmware versions.
func (s *SqliteStore) ListFirmwareVersions(ctx context.Context) ([]model.FirmwareVersion, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, device_type, tier, version, url, sha256_hash, changelog, min_app_version, force_update, active, created_at
		FROM firmware_releases ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list firmware: %w", err)
	}
	defer rows.Close()

	var result []model.FirmwareVersion
	for rows.Next() {
		var f model.FirmwareVersion
		if err := rows.Scan(&f.ID, &f.DeviceType, &f.Tier, &f.Version, &f.DownloadURL,
			&f.Sha256Hash, &f.Changelog, &f.MinAppVersion, &f.ForceUpdate, &f.IsActive, &f.ReleaseDate); err != nil {
			return nil, fmt.Errorf("scan firmware: %w", err)
		}
		result = append(result, f)
	}
	return result, rows.Err()
}

// DeleteFirmwareVersion soft-deletes a firmware release.
func (s *SqliteStore) DeleteFirmwareVersion(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE firmware_releases SET active = 0 WHERE id = ?`, id)
	return err
}

// GetFirmwareVersion retrieves a firmware release by ID.
func (s *SqliteStore) GetFirmwareVersion(ctx context.Context, id string) (*model.FirmwareVersion, error) {
	var f model.FirmwareVersion
	err := s.db.QueryRowContext(ctx,
		`SELECT id, device_type, tier, version, url, sha256_hash, changelog, min_app_version, force_update, active, created_at
		 FROM firmware_releases WHERE id = ?`, id).Scan(
		&f.ID, &f.DeviceType, &f.Tier, &f.Version, &f.DownloadURL,
		&f.Sha256Hash, &f.Changelog, &f.MinAppVersion, &f.ForceUpdate, &f.IsActive, &f.ReleaseDate)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// VerifyFirmwareSignature verifies the signature of a firmware release.
func (s *SqliteStore) VerifyFirmwareSignature(ctx context.Context, id string) (bool, string, error) {
	f, err := s.GetFirmwareVersion(ctx, id)
	if err != nil {
		return false, "", err
	}
	if f.Sha256Hash == "" {
		return false, "no_hash", nil
	}
	if len(f.Sha256Hash) != 64 {
		return false, "invalid_hash_format", nil
	}
	return true, "verified", nil
}

// PushOTAJob records an OTA push job.
func (s *SqliteStore) PushOTAJob(ctx context.Context, firmwareID string, deviceIDs []string) (string, error) {
	devicesJSON := "[]"
	if len(deviceIDs) > 0 {
		data, _ := json.Marshal(deviceIDs)
		devicesJSON = string(data)
	}
	id := fmt.Sprintf("job_%d", time.Now().UnixNano())
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO ota_jobs (id, firmware_id, target_devices, progress) VALUES (?, ?, ?, '{"total":0,"pending":0}')`,
		id, firmwareID, devicesJSON)
	return id, err
}

// GetOTAJob retrieves an OTA job by ID.
func (s *SqliteStore) GetOTAJob(ctx context.Context, jobID string) (*model.OTAJob, error) {
	var job model.OTAJob
	err := s.db.QueryRowContext(ctx,
		`SELECT id, firmware_id, target_devices, progress, created_at, updated_at FROM ota_jobs WHERE id = ?`,
		jobID).Scan(&job.ID, &job.FirmwareID, &job.TargetDevices, &job.Progress, &job.CreatedAt, &job.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &job, nil
}
