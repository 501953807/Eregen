package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"eregen.dev/admin-api/internal/model"
	"fmt"
	"time"

	"github.com/google/uuid"
)


// CreateDevice inserts a new device record.
func (s *SqliteStore) CreateDevice(ctx context.Context, d *model.DeviceSummary) error {
	d.ID = fmt.Sprintf("dev_%s", uuid.New().String()[:8])
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO devices (id, device_id, device_type, tier, status, last_seen, owner_user_id, settings, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
		d.ID, d.DeviceID, d.Type, d.Tier, d.Status,
		d.LastSeen.Format("2006-01-02 15:04:05"), "", "{}")
	return err
}

// ListDevices returns a paginated list of devices with optional filters.
func (s *SqliteStore) ListDevices(ctx context.Context, page, pageSize int, status, devType, tier string) ([]model.DeviceSummary, error) {
	query := `SELECT id, device_id, device_type, tier, status, COALESCE(last_seen, '0001-01-01'),
		COALESCE((SELECT u.name FROM users u JOIN devices d ON d.owner_user_id = u.id WHERE d.id = devices.id LIMIT 1), ''),
		COALESCE(json_extract(settings, '$.fw_version'),'v0.1')
		FROM devices WHERE 1=1`
	args := []interface{}{}
	idx := 1
	if status != "" {
		query += " AND status=?"
		args = append(args, status)
		idx++
	}
	if devType != "" {
		query += " AND device_type=?"
		args = append(args, devType)
		idx++
	}
	if tier != "" {
		query += " AND tier=?"
		args = append(args, tier)
		idx++
	}
	query += fmt.Sprintf(" ORDER BY last_seen DESC LIMIT ? OFFSET ?")
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	defer rows.Close()

	var devices []model.DeviceSummary
	for rows.Next() {
		var d model.DeviceSummary
		var lastSeenStr string
		if err := rows.Scan(&d.ID, &d.DeviceID, &d.Type, &d.Tier, &d.Status, &lastSeenStr, &d.OwnerName, &d.FirmwareVer); err != nil {
			return nil, fmt.Errorf("scan device: %w", err)
		}
		d.LastSeen = parseTimeOrDefault(lastSeenStr, time.Time{})
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

// UpdateDeviceConfig updates device settings JSON column.
func (s *SqliteStore) UpdateDeviceConfig(ctx context.Context, deviceID string, config map[string]interface{}) error {
	settingsJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `UPDATE devices SET settings = json_patch(COALESCE(settings, '{}'), ?) WHERE device_id = ?`, string(settingsJSON), deviceID)
	return err
}

// TriggerOTA schedules an OTA update for a device.
func (s *SqliteStore) TriggerOTA(ctx context.Context, deviceID, firmwareURL, sha256Hash string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE devices SET ota_url = ?, ota_hash = ?, ota_status = 'pending' WHERE device_id = ?`,
		firmwareURL, sha256Hash, deviceID)
	return err
}

// GetDeviceByID returns a single device by its database ID.
func (s *SqliteStore) GetDeviceByID(ctx context.Context, id string) (*model.DeviceDetail, error) {
	var d model.DeviceDetail
	var lastSeenStr, settingsStr, elderlyName string
	err := s.db.QueryRowContext(ctx, `
		SELECT d.id, d.device_id, d.device_type, d.tier, d.status, COALESCE(d.last_seen, '0001-01-01'),
		       u.name, COALESCE(json_extract(d.settings, '$.fw_version'),'v0.1'),
		       d.settings,
		       COALESCE(e.name, '')
			FROM devices d LEFT JOIN users u ON d.owner_user_id = u.id
			LEFT JOIN elderly_devices ed ON d.id = ed.device_id
			LEFT JOIN elderly_profiles e ON ed.elderly_id = e.id
			WHERE d.id = ?`, id).Scan(
		&d.ID, &d.DeviceID, &d.Type, &d.Tier, &d.Status, &lastSeenStr,
		&d.OwnerName, &d.FirmwareVer, &settingsStr, &elderlyName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found")
		}
		return nil, fmt.Errorf("get device: %w", err)
	}
	d.LastSeen = parseTimeOrDefault(lastSeenStr, time.Time{})
	d.ElderlyName = elderlyName
	if settingsStr != "" {
		json.Unmarshal([]byte(settingsStr), &d.SettingsJSON)
	}
	return &d, nil
}

// UnbindDevice removes a device from its owner and all elderly links.
func (s *SqliteStore) UnbindDevice(ctx context.Context, deviceID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM elderly_devices WHERE device_id = ?;
		 UPDATE devices SET owner_user_id = NULL WHERE device_id = ?`,
		deviceID, deviceID)
	return err
}

// BatchTriggerOTA schedules OTA updates for multiple devices.
func (s *SqliteStore) BatchTriggerOTA(ctx context.Context, deviceIDs, firmwareURL, sha256Hash []string) error {
	for i, id := range deviceIDs {
		url := firmwareURL[i%len(firmwareURL)]
		hash := sha256Hash[i%len(sha256Hash)]
		if _, err := s.db.ExecContext(ctx,
			`UPDATE devices SET ota_url = ?, ota_hash = ?, ota_status = 'pending' WHERE device_id = ?`,
			url, hash, id); err != nil {
			return fmt.Errorf("batch OTA device %s: %w", id, err)
		}
	}
	return nil
}
