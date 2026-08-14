package store

import (
	"context"
	"database/sql"
	"eregen.dev/admin-api/internal/model"
	"fmt"
)


func (s *PostgresStore) BindWristband(ctx context.Context, patientID, deviceID string) error {

	_, err := s.db.ExecContext(ctx,

		`INSERT INTO medical_bindings (patient_id, device_id, bound_at) VALUES ($1,$2,NOW())

		 ON CONFLICT DO NOTHING`, patientID, deviceID)

	if err != nil {

		return err

	}

	_, err = s.db.ExecContext(ctx,

		`UPDATE medical_wristband_devices SET bound_patient_id=$1, status='bound', updated_at=NOW() WHERE id=$2`, patientID, deviceID)

	return err

}



func (s *PostgresStore) UnbindWristband(ctx context.Context, bindingID string) error {

	_, err := s.db.ExecContext(ctx,

		`UPDATE medical_bindings SET unbound_at=NOW() WHERE id=$1 AND unbound_at IS NULL`, bindingID)

	if err != nil {

		return err

	}

	_, err = s.db.ExecContext(ctx,

		`UPDATE medical_wristband_devices SET bound_patient_id=NULL, status='idle', updated_at=NOW() WHERE id IN (SELECT device_id FROM medical_bindings WHERE id=$1)`, bindingID)

	return err

}



func (s *PostgresStore) ClearWristband(ctx context.Context, deviceID string) error {

	_, err := s.db.ExecContext(ctx,

		`UPDATE medical_wristband_devices SET bound_patient_id=NULL, status='cleared', updated_at=NOW() WHERE id=$1`, deviceID)

	return err

}



func (s *PostgresStore) ListWristbands(ctx context.Context, page, pageSize int, status string) ([]model.MedicalWristbandDevice, error) {

	query := `SELECT id, device_id, firmware_version, status, bound_patient_id, created_at, updated_at

		FROM medical_wristband_devices WHERE 1=1`

	var args []interface{}

	idx := 1

	if status != "" {

		query += fmt.Sprintf(" AND status=$%d", idx)

		args = append(args, status)

		idx++

	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", idx, idx+1)

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



func (s *PostgresStore) GetWristbandFirmware(ctx context.Context, deviceID string) (string, error) {

	var fw string

	err := s.db.QueryRowContext(ctx, `SELECT firmware_version FROM medical_wristband_devices WHERE device_id=$1`, deviceID).Scan(&fw)

	if err != nil {

		if err == sql.ErrNoRows {

			return "", fmt.Errorf("wristband not found")

		}

		return "", err

	}

	return fw, nil

}



func (s *PostgresStore) WriteToWristband(ctx context.Context, deviceID, data string) error {

	_, err := s.db.ExecContext(ctx,

		`UPDATE medical_wristband_devices SET firmware_version=$1, updated_at=NOW() WHERE device_id=$2`,

		data, deviceID)

	return err

}


// CreateWristband creates a new wristband device.
func (s *PostgresStore) CreateWristband(ctx context.Context, d *model.MedicalWristbandDevice) error {
	return ErrNotImplemented
}
