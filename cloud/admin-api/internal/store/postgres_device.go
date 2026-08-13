package store

import (
	"github.com/google/uuid"
	"context"
	"database/sql"
	"encoding/json"
	"eregen.dev/admin-api/internal/model"
	"fmt"
)


func (s *PostgresStore) ListDevices(ctx context.Context, page, pageSize int, status, devType, tier string) ([]model.DeviceSummary, error) {

	query := `SELECT id, device_id, device_type, tier, status, COALESCE(last_seen, '0001-01-01'),

		(SELECT u.name FROM users u JOIN devices d ON d.owner_user_id = u.id WHERE d.id = devices.id LIMIT 1),

		COALESCE(settings->>'fw_version','v0.1')

		FROM devices WHERE 1=1`

	args := []interface{}{}

	idx := 1

	if status != "" {

		query += fmt.Sprintf(" AND status=$%d", idx)

		args = append(args, status)

		idx++

	}

	if devType != "" {

		query += fmt.Sprintf(" AND device_type=$%d", idx)

		args = append(args, devType)

		idx++

	}

	if tier != "" {

		query += fmt.Sprintf(" AND tier=$%d", idx)

		args = append(args, tier)

		idx++

	}

	query += fmt.Sprintf(" ORDER BY last_seen DESC NULLS LAST LIMIT $%d OFFSET $%d", idx, idx+1)

	args = append(args, pageSize, (page-1)*pageSize)



	rows, err := s.db.QueryContext(ctx, query, args...)

	if err != nil {

		return nil, fmt.Errorf("list devices: %w", err)

	}

	defer rows.Close()



	var devices []model.DeviceSummary

	for rows.Next() {

		var d model.DeviceSummary

		if err := rows.Scan(&d.ID, &d.DeviceID, &d.Type, &d.Tier, &d.Status, &d.LastSeen, &d.OwnerName, &d.FirmwareVer); err != nil {

			return nil, fmt.Errorf("scan device: %w", err)

		}

		devices = append(devices, d)

	}

	return devices, rows.Err()

}



// ListUsers returns a paginated list of users with optional role filter.

func (s *PostgresStore) GetDeviceByID(ctx context.Context, id string) (*model.DeviceDetail, error) {

	var d model.DeviceDetail

	err := s.db.QueryRowContext(ctx, `

		SELECT d.id, d.device_id, d.device_type, d.tier, d.status, COALESCE(d.last_seen, '0001-01-01'),

		       u.name, COALESCE(d.settings->>'fw_version','v0.1'),

		       d.settings,

		       e.name AS elderly_name

		FROM devices d LEFT JOIN users u ON d.owner_user_id = u.id

		LEFT JOIN elderly_profiles e ON d.id = ANY((SELECT ed.device_id FROM elderly_devices ed WHERE ed.elderly_id = e.id LIMIT 1))

		WHERE d.id = $1`, id).Scan(

		&d.ID, &d.DeviceID, &d.Type, &d.Tier, &d.Status, &d.LastSeen,

		&d.OwnerName, &d.FirmwareVer, &d.SettingsJSON, &d.ElderlyName,

	)

	if err != nil {

		if err == sql.ErrNoRows {

			return nil, fmt.Errorf("device not found")

		}

		return nil, fmt.Errorf("get device: %w", err)

	}

	return &d, nil

}



// UnbindDevice removes a device from its owner and all elderly links.

func (s *PostgresStore) UpdateDeviceConfig(ctx context.Context, deviceID string, config map[string]interface{}) error {

	// SAFETY: Use json.Marshal to properly encode the configuration as JSON,

	// preventing SQL injection that could occur with manual string concatenation.

	data, err := json.Marshal(config)

	if err != nil {

		return fmt.Errorf("failed to marshal config: %w", err)

	}

	_, err = s.db.ExecContext(ctx, `UPDATE devices SET settings = settings || $1::jsonb WHERE device_id = $2`, data, deviceID)

	return err

}



// TriggerOTA schedules an OTA update for a device.

func (s *PostgresStore) TriggerOTA(ctx context.Context, deviceID, firmwareURL, sha256Hash string) error {

	_, err := s.db.ExecContext(ctx,

		`UPDATE devices SET ota_url = $1, ota_hash = $2, ota_status = 'pending' WHERE device_id = $3`,

		firmwareURL, sha256Hash, deviceID)

	return err

}



// ResolveAlert marks an alert as resolved.

func (s *PostgresStore) UnbindDevice(ctx context.Context, deviceID string) error {

	_, err := s.db.ExecContext(ctx,

		`DELETE FROM elderly_devices WHERE device_id = $1;

		 UPDATE devices SET owner_user_id = NULL WHERE device_id = $2`,

		deviceID, deviceID)

	return err

}



// BatchTriggerOTA schedules OTA updates for multiple devices.

func (s *PostgresStore) BatchTriggerOTA(ctx context.Context, deviceIDs, firmwareURL, sha256Hash []string) error {

	for i, id := range deviceIDs {

		url := firmwareURL[i % len(firmwareURL)]

		hash := sha256Hash[i % len(sha256Hash)]

		if _, err := s.db.ExecContext(ctx,

			`UPDATE devices SET ota_url = $1, ota_hash = $2, ota_status = 'pending' WHERE device_id = $3`,

			url, hash, id); err != nil {

			return fmt.Errorf("batch OTA device %s: %w", id, err)

		}

	}

	return nil

}



// CreateDevice inserts a new device record.
func (s *PostgresStore) CreateDevice(ctx context.Context, d *model.DeviceSummary) error {
	d.ID = fmt.Sprintf("dev_%s", uuid.New().String()[:8])
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO devices (id, device_id, device_type, tier, status, last_seen, owner_user_id, settings, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())`,
		d.ID, d.DeviceID, d.Type, d.Tier, d.Status,
		d.LastSeen.Format("2006-01-02 15:04:05"), "", "{}")
	return err
}
// CreateFirmwareVersion inserts a new firmware release record.