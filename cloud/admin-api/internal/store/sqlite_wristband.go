package store

import (
	"context"
	"database/sql"
	"eregen.dev/admin-api/internal/model"
	"fmt"
	"time"

	"github.com/google/uuid"
)


// CreateWristband creates a new wristband device.
func (s *SqliteStore) CreateWristband(ctx context.Context, d *model.MedicalWristbandDevice) error {
	d.ID = fmt.Sprintf("mw_%s", uuid.New().String()[:8])
	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medical_wristband_devices (id, device_id, firmware_version, status, bound_patient_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		d.ID, d.DeviceID, d.FirmwareVersion, d.Status, d.BoundPatientID, d.CreatedAt, d.UpdatedAt)
	return err
}

// BindWristband binds a device to a patient.
func (s *SqliteStore) BindWristband(ctx context.Context, patientID, deviceID string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO medical_bindings (patient_id, device_id, bound_at) VALUES (?, ?, datetime('now'))
		 ON CONFLICT DO NOTHING`, patientID, deviceID)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`UPDATE medical_wristband_devices SET bound_patient_id=?, status='bound', updated_at=datetime('now') WHERE id=?`, patientID, deviceID)
	return err
}

// UnbindWristband unbinds a device from a patient.
func (s *SqliteStore) UnbindWristband(ctx context.Context, bindingID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE medical_bindings SET unbound_at=datetime('now') WHERE id=? AND unbound_at IS NULL`, bindingID)
	if err != nil {
		return err
	}
	// Also clear device binding
	_, err = s.db.ExecContext(ctx,
		`UPDATE medical_wristband_devices SET bound_patient_id=NULL, status='idle', updated_at=datetime('now') WHERE id IN (SELECT device_id FROM medical_bindings WHERE id=?)`, bindingID)
	return err
}

// ClearWristband clears all data from a wristband device.
func (s *SqliteStore) ClearWristband(ctx context.Context, deviceID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE medical_wristband_devices SET bound_patient_id=NULL, status='cleared', updated_at=datetime('now') WHERE id=?`, deviceID)
	return err
}

// ListWristbands returns all wristband devices with optional pagination/status.
func (s *SqliteStore) ListWristbands(ctx context.Context, page, pageSize int, status string) ([]model.MedicalWristbandDevice, error) {
	query := `SELECT id, device_id, firmware_version, status, bound_patient_id, created_at, updated_at
		FROM medical_wristband_devices WHERE 1=1`
	args := []interface{}{}
	idx := 1
	if status != "" {
		query += fmt.Sprintf(" AND status=?")
		args = append(args, status)
		idx++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT ? OFFSET ?")
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list wristbands: %w", err)
	}
	defer rows.Close()

	var devices []model.MedicalWristbandDevice
	for rows.Next() {
		var d model.MedicalWristbandDevice
		if err := rows.Scan(&d.ID, &d.DeviceID, &d.FirmwareVersion, &d.Status, &d.BoundPatientID, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan wristband: %w", err)
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

// GetWristbandFirmware returns firmware version for a device.
func (s *SqliteStore) GetWristbandFirmware(ctx context.Context, deviceID string) (string, error) {
	var fw string
	err := s.db.QueryRowContext(ctx, `SELECT firmware_version FROM medical_wristband_devices WHERE device_id=?`, deviceID).Scan(&fw)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("wristband not found")
		}
		return "", err
	}
	return fw, nil
}

// WriteToWristband pushes data to a wristband device (stub).
func (s *SqliteStore) WriteToWristband(ctx context.Context, deviceID, data string) error {
	// In production, this would trigger MQTT message to device
	_, err := s.db.ExecContext(ctx,
		`UPDATE medical_wristband_devices SET firmware_version=?, updated_at=datetime('now') WHERE device_id=?`,
		data, deviceID)
	return err
}
