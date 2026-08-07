package store

import (
	"context"
	"encoding/json"
	"eregen.dev/admin-api/internal/model"
	"fmt"
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

// PushOTAJob records an OTA push job.
func (s *SqliteStore) PushOTAJob(ctx context.Context, firmwareID string, deviceIDs []string) error {
	devicesJSON := "[]"
	if len(deviceIDs) > 0 {
		data, _ := json.Marshal(deviceIDs)
		devicesJSON = string(data)
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO ota_jobs (firmware_id, target_devices, progress) VALUES (?, ?, '{"total":0,"pending":0}')`,
		firmwareID, devicesJSON)
	return err
}
