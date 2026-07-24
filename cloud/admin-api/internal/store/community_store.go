package store

import (
	"context"
	"fmt"
	"time"

	"eregen.dev/admin-api/internal/model"
)

// ========== Community Store Methods (SqliteStore) ==========

func (s *SqliteStore) CreateCommunityElder(ctx context.Context, e *model.CommunityElder) error {
	now := time.Now().UTC()
	e.CreatedAt = now
	e.UpdatedAt = now
	if e.ID == "" {
		e.ID = fmt.Sprintf("ce_%d", now.UnixNano())
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO community_elders (id, name, id_card, gender, age, address, emergency_contact, bank_account, hospital_id, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.Name, e.IDCard, e.Gender, e.Age, e.Address, e.EmergencyContact, e.BankAccount, e.HospitalID, e.Status, e.CreatedAt, e.UpdatedAt)
	return err
}

func (s *SqliteStore) GetCommunityElder(ctx context.Context, id string) (*model.CommunityElder, error) {
	var e model.CommunityElder
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, id_card, gender, age, address, emergency_contact, bank_account, hospital_id,
			 status, created_at, updated_at, deactivated_at, deactivated_reason
		 FROM community_elders WHERE id=?`, id).Scan(
		&e.ID, &e.Name, &e.IDCard, &e.Gender, &e.Age, &e.Address, &e.EmergencyContact,
		&e.BankAccount, &e.HospitalID, &e.Status, &e.CreatedAt, &e.UpdatedAt,
		&e.DeactivatedAt, &e.DeactivatedReason)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *SqliteStore) ListCommunityElders(ctx context.Context, page, pageSize int, status string) ([]model.CommunityElder, error) {
	q := `SELECT id, name, id_card, gender, age, address, emergency_contact, bank_account, hospital_id,
			 status, created_at, updated_at, deactivated_at, deactivated_reason
		 FROM community_elders WHERE 1=1`
	args := []interface{}{}
	if status != "" {
		q += " AND status=?"
		args = append(args, status)
	}
	q += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var elders []model.CommunityElder
	for rows.Next() {
		var e model.CommunityElder
		if err := rows.Scan(&e.ID, &e.Name, &e.IDCard, &e.Gender, &e.Age, &e.Address,
			&e.EmergencyContact, &e.BankAccount, &e.HospitalID, &e.Status, &e.CreatedAt, &e.UpdatedAt,
			&e.DeactivatedAt, &e.DeactivatedReason); err != nil {
			return nil, err
		}
		elders = append(elders, e)
	}
	return elders, rows.Err()
}

func (s *SqliteStore) UpdateCommunityElder(ctx context.Context, e *model.CommunityElder) error {
	e.UpdatedAt = time.Now().UTC()
	_, err := s.db.ExecContext(ctx,
		`UPDATE community_elders SET name=?, id_card=?, gender=?, age=?, address=?,
		 emergency_contact=?, bank_account=?, hospital_id=?, status=?, updated_at=?,
		 deactivated_at=?, deactivated_reason=?
		 WHERE id=?`,
		e.Name, e.IDCard, e.Gender, e.Age, e.Address, e.EmergencyContact, e.BankAccount,
		e.HospitalID, e.Status, e.UpdatedAt, e.DeactivatedAt, e.DeactivatedReason, e.ID)
	return err
}

func (s *SqliteStore) DeleteCommunityElder(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE community_elders SET status='deactivated', deactivated_at=datetime('now'), deactivated_reason='deleted' WHERE id=?`, id)
	return err
}

func (s *SqliteStore) BulkUpsertCommunityElders(ctx context.Context, elders []model.CommunityElder) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO community_elders (id, name, id_card, gender, age, address, emergency_contact, bank_account, hospital_id, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))
		 ON CONFLICT(id_card) DO UPDATE SET name=excluded.name, status=excluded.status`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range elders {
		if e.ID == "" {
			e.ID = fmt.Sprintf("ce_%d", time.Now().UnixNano())
		}
		_, err := stmt.ExecContext(ctx, e.ID, e.Name, e.IDCard, e.Gender, e.Age, e.Address,
			e.EmergencyContact, e.BankAccount, e.HospitalID, e.Status)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SqliteStore) GetCommunityElderStats(ctx context.Context) (*model.CommunityElderStats, error) {
	stats := &model.CommunityElderStats{}
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM community_elders`).Scan(&stats.TotalElders)
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM community_elders WHERE status='active'`).Scan(&stats.ActiveElders)
	today := time.Now().Format("2006-01-02")
	s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM community_signin_records WHERE date(signin_time)=? AND is_welfare_signin=1`, today).Scan(&stats.TodaySignins)
	s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM community_pharmacy_logs WHERE date(dispense_time)=?`, today).Scan(&stats.TodayDispenses)
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM community_elder_welfare WHERE revoked_at IS NULL`).Scan(&stats.ActiveWelfareTags)
	return stats, nil
}

// Device management
func (s *SqliteStore) CreateCommunityDevice(ctx context.Context, d *model.CommunityWristbandDevice) error {
	now := time.Now().UTC()
	d.CreatedAt = now
	d.UpdatedAt = now
	if d.ID == "" {
		d.ID = fmt.Sprintf("cd_%d", now.UnixNano())
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO community_wristband_devices (id, device_id, firmware_version, mode, status, last_seen, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		d.ID, d.DeviceID, d.FirmwareVersion, d.Mode, d.Status, d.LastSeen, d.CreatedAt, d.UpdatedAt)
	return err
}

func (s *SqliteStore) GetCommunityDevice(ctx context.Context, deviceID string) (*model.CommunityWristbandDevice, error) {
	var d model.CommunityWristbandDevice
	err := s.db.QueryRowContext(ctx,
		`SELECT id, device_id, firmware_version, mode, status, last_seen, created_at, updated_at
		 FROM community_wristband_devices WHERE device_id=?`, deviceID).Scan(
		&d.ID, &d.DeviceID, &d.FirmwareVersion, &d.Mode, &d.Status, &d.LastSeen, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *SqliteStore) ListCommunityDevices(ctx context.Context, page, pageSize int, status string) ([]model.CommunityWristbandDevice, error) {
	q := `SELECT id, device_id, firmware_version, mode, status, last_seen, created_at, updated_at
		 FROM community_wristband_devices WHERE 1=1`
	args := []interface{}{}
	if status != "" {
		q += " AND status=?"
		args = append(args, status)
	}
	q += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []model.CommunityWristbandDevice
	for rows.Next() {
		var d model.CommunityWristbandDevice
		if err := rows.Scan(&d.ID, &d.DeviceID, &d.FirmwareVersion, &d.Mode, &d.Status,
			&d.LastSeen, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

func (s *SqliteStore) UpdateCommunityDevice(ctx context.Context, d *model.CommunityWristbandDevice) error {
	d.UpdatedAt = time.Now().UTC()
	_, err := s.db.ExecContext(ctx,
		`UPDATE community_wristband_devices SET firmware_version=?, status=?, last_seen=?, updated_at=? WHERE id=?`,
		d.FirmwareVersion, d.Status, d.LastSeen, d.UpdatedAt, d.ID)
	return err
}

// Bindings
func (s *SqliteStore) BindCommunityElderDevice(ctx context.Context, elderID, deviceID string) error {
	id := fmt.Sprintf("cb_%d", time.Now().UnixNano())
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO community_elder_bindings (id, elder_id, device_id, bound_at) VALUES (?, ?, ?, datetime('now'))`,
		id, elderID, deviceID)
	return err
}

func (s *SqliteStore) UnbindCommunityElderDevice(ctx context.Context, bindingID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE community_elder_bindings SET unbound_at=datetime('now') WHERE id=?`, bindingID)
	return err
}

// Welfare tags
func (s *SqliteStore) CreateWelfareTagConfig(ctx context.Context, c *model.CommunityWelfareTagConfig) error {
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now
	if c.ID == "" {
		c.ID = fmt.Sprintf("wtc_%d", now.UnixNano())
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO community_welfare_tag_config (id, tag_code, tag_name, issuer, renewal_period_days, benefit_amount, enabled)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.TagCode, c.TagName, c.Issuer, c.RenewalPeriodDays, c.BenefitAmount, c.Enabled)
	return err
}

func (s *SqliteStore) UpdateWelfareTagConfig(ctx context.Context, c *model.CommunityWelfareTagConfig) error {
	c.UpdatedAt = time.Now().UTC()
	_, err := s.db.ExecContext(ctx,
		`UPDATE community_welfare_tag_config SET tag_name=?, issuer=?, renewal_period_days=?, benefit_amount=?, enabled=?, updated_at=? WHERE tag_code=?`,
		c.TagName, c.Issuer, c.RenewalPeriodDays, c.BenefitAmount, c.Enabled, c.UpdatedAt, c.TagCode)
	return err
}

func (s *SqliteStore) ListWelfareTagConfigs(ctx context.Context) ([]model.CommunityWelfareTagConfig, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, tag_code, tag_name, issuer, renewal_period_days, benefit_amount, enabled, created_at, updated_at
		 FROM community_welfare_tag_config ORDER BY tag_code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []model.CommunityWelfareTagConfig
	for rows.Next() {
		var c model.CommunityWelfareTagConfig
		if err := rows.Scan(&c.ID, &c.TagCode, &c.TagName, &c.Issuer, &c.RenewalPeriodDays,
			&c.BenefitAmount, &c.Enabled, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		configs = append(configs, c)
	}
	return configs, rows.Err()
}

func (s *SqliteStore) GetWelfareTagConfig(ctx context.Context, tagCode string) (*model.CommunityWelfareTagConfig, error) {
	var c model.CommunityWelfareTagConfig
	err := s.db.QueryRowContext(ctx,
		`SELECT id, tag_code, tag_name, issuer, renewal_period_days, benefit_amount, enabled, created_at, updated_at
		 FROM community_welfare_tag_config WHERE tag_code=?`, tagCode).Scan(
		&c.ID, &c.TagCode, &c.TagName, &c.Issuer, &c.RenewalPeriodDays,
		&c.BenefitAmount, &c.Enabled, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *SqliteStore) AssignWelfareTag(ctx context.Context, a *model.CommunityElderWelfare) error {
	now := time.Now().UTC()
	a.EffectiveAt = now
	if a.ID == "" {
		a.ID = fmt.Sprintf("ewf_%d", now.UnixNano())
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO community_elder_welfare (id, elder_id, tag_code, valid_from, valid_to, certified_by, certification_doc, effective_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.ElderID, a.TagCode, a.ValidFrom, a.ValidTo, a.CertifiedBy, a.CertificationDoc, a.EffectiveAt)
	return err
}

func (s *SqliteStore) RevokeWelfareTag(ctx context.Context, elderID, tagCode string) error {
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx,
		`UPDATE community_elder_welfare SET revoked_at=? WHERE elder_id=? AND tag_code=? AND revoked_at IS NULL`,
		now, elderID, tagCode)
	return err
}

func (s *SqliteStore) ListElderWelfareTags(ctx context.Context, elderID string) ([]model.CommunityElderWelfare, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, elder_id, tag_code, valid_from, valid_to, certified_by, certification_doc, effective_at, revoked_at
		 FROM community_elder_welfare WHERE elder_id=? AND revoked_at IS NULL`, elderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.CommunityElderWelfare
	for rows.Next() {
		var t model.CommunityElderWelfare
		if err := rows.Scan(&t.ID, &t.ElderID, &t.TagCode, &t.ValidFrom, &t.ValidTo,
			&t.CertifiedBy, &t.CertificationDoc, &t.EffectiveAt, &t.RevokedAt); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

// Sign-in
func (s *SqliteStore) CreateSigninRecord(ctx context.Context, sRec *model.CommunitySigninRecord) error {
	now := time.Now().UTC()
	sRec.SigninTime = now
	if sRec.ID == "" {
		sRec.ID = fmt.Sprintf("sr_%d", now.UnixNano())
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO community_signin_records (id, elder_id, device_id, hospital_id, pharmacist_id, signin_time, period, activated_tags, is_medical_signin, is_welfare_signin, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sRec.ID, sRec.ElderID, sRec.DeviceID, sRec.HospitalID, sRec.PharmacistID,
		sRec.SigninTime, sRec.Period, sRec.ActivatedTags, sRec.IsMedicalSignin, sRec.IsWelfareSignin, sRec.Notes)
	if err != nil {
		// UNIQUE(elder_id, device_id, period) violation — return a clear error for the handler
		return fmt.Errorf("duplicate signin: elder %s already signed in this period on device %s", sRec.ElderID, sRec.DeviceID)
	}
	return nil
}

func (s *SqliteStore) ListSigninRecords(ctx context.Context, elderID, period, hospitalID string, page, pageSize int) ([]model.CommunitySigninRecord, error) {
	q := `SELECT id, elder_id, device_id, hospital_id, pharmacist_id, signin_time, period, activated_tags, is_medical_signin, is_welfare_signin, notes
		 FROM community_signin_records WHERE 1=1`
	args := []interface{}{}
	if elderID != "" {
		q += " AND elder_id=?"
		args = append(args, elderID)
	}
	if period != "" {
		q += " AND period=?"
		args = append(args, period)
	}
	if hospitalID != "" {
		q += " AND hospital_id=?"
		args = append(args, hospitalID)
	}
	q += " ORDER BY signin_time DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.CommunitySigninRecord
	for rows.Next() {
		var r model.CommunitySigninRecord
		if err := rows.Scan(&r.ID, &r.ElderID, &r.DeviceID, &r.HospitalID, &r.PharmacistID,
			&r.SigninTime, &r.Period, &r.ActivatedTags, &r.IsMedicalSignin, &r.IsWelfareSignin, &r.Notes); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

func (s *SqliteStore) GetSigninSummary(ctx context.Context, elderID, period string) (*model.CommunitySigninRecord, error) {
	var r model.CommunitySigninRecord
	err := s.db.QueryRowContext(ctx,
		`SELECT id, elder_id, device_id, hospital_id, pharmacist_id, signin_time, period, activated_tags, is_medical_signin, is_welfare_signin, notes
		 FROM community_signin_records WHERE elder_id=? AND period=? ORDER BY signin_time DESC LIMIT 1`,
		elderID, period).Scan(
		&r.ID, &r.ElderID, &r.DeviceID, &r.HospitalID, &r.PharmacistID, &r.SigninTime,
		&r.Period, &r.ActivatedTags, &r.IsMedicalSignin, &r.IsWelfareSignin, &r.Notes)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// Pharmacy
func (s *SqliteStore) CreatePharmacyLog(ctx context.Context, p *model.CommunityPharmacyLog) error {
	now := time.Now().UTC()
	p.DispenseTime = now
	if p.ID == "" {
		p.ID = fmt.Sprintf("pl_%d", now.UnixNano())
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO community_pharmacy_logs (id, elder_id, device_id, hospital_id, pharmacist_id, dispense_time, period, items, total_cost, insurance_covered, self_pay, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.ElderID, p.DeviceID, p.HospitalID, p.PharmacistID, p.DispenseTime,
		p.Period, p.Items, p.TotalCost, p.InsuranceCovered, p.SelfPay, p.Notes)
	return err
}

func (s *SqliteStore) ListPharmacyLogs(ctx context.Context, elderID, period string, page, pageSize int) ([]model.CommunityPharmacyLog, error) {
	q := `SELECT id, elder_id, device_id, hospital_id, pharmacist_id, dispense_time, period, items, total_cost, insurance_covered, self_pay, notes
		 FROM community_pharmacy_logs WHERE 1=1`
	args := []interface{}{}
	if elderID != "" {
		q += " AND elder_id=?"
		args = append(args, elderID)
	}
	if period != "" {
		q += " AND period=?"
		args = append(args, period)
	}
	q += " ORDER BY dispense_time DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.CommunityPharmacyLog
	for rows.Next() {
		var l model.CommunityPharmacyLog
		if err := rows.Scan(&l.ID, &l.ElderID, &l.DeviceID, &l.HospitalID, &l.PharmacistID,
			&l.DispenseTime, &l.Period, &l.Items, &l.TotalCost, &l.InsuranceCovered, &l.SelfPay, &l.Notes); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

// Minzheng sync
func (s *SqliteStore) CreateMinzhengSync(ctx context.Context, m *model.CommunityMinzhengSync) error {
	now := time.Now().UTC()
	m.CreatedAt = now
	if m.ID == "" {
		m.ID = fmt.Sprintf("ms_%d", now.UnixNano())
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO community_minzheng_sync (id, source, filename, imported_count, matched_count, pending_review_count, error_count, status, created_at, completed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.Source, m.Filename, m.ImportedCount, m.MatchedCount, m.PendingReviewCount,
		m.ErrorCount, m.Status, m.CreatedAt, m.CompletedAt)
	return err
}

func (s *SqliteStore) ListMinzhengSync(ctx context.Context, page, pageSize int) ([]model.CommunityMinzhengSync, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, source, filename, imported_count, matched_count, pending_review_count, error_count, status, created_at, completed_at
		 FROM community_minzheng_sync ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var syncs []model.CommunityMinzhengSync
	for rows.Next() {
		var s model.CommunityMinzhengSync
		if err := rows.Scan(&s.ID, &s.Source, &s.Filename, &s.ImportedCount, &s.MatchedCount,
			&s.PendingReviewCount, &s.ErrorCount, &s.Status, &s.CreatedAt, &s.CompletedAt); err != nil {
			return nil, err
		}
		syncs = append(syncs, s)
	}
	return syncs, rows.Err()
}

func (s *SqliteStore) GetLatestMinzhengSync(ctx context.Context) (*model.CommunityMinzhengSync, error) {
	var m model.CommunityMinzhengSync
	err := s.db.QueryRowContext(ctx,
		`SELECT id, source, filename, imported_count, matched_count, pending_review_count, error_count, status, created_at, completed_at
		 FROM community_minzheng_sync ORDER BY created_at DESC LIMIT 1`).Scan(
		&m.ID, &m.Source, &m.Filename, &m.ImportedCount, &m.MatchedCount,
		&m.PendingReviewCount, &m.ErrorCount, &m.Status, &m.CreatedAt, &m.CompletedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// Batch payments
func (s *SqliteStore) CreateBatchPayment(ctx context.Context, p *model.CommunityBatchPayment) error {
	now := time.Now().UTC()
	p.CreatedAt = now
	if p.ID == "" {
		p.ID = fmt.Sprintf("bp_%d", now.UnixNano())
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO community_batch_payments (id, batch_id, period, pay_type, elder_id, amount, bank_account, status, failure_reason, executed_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.BatchID, p.Period, p.PayType, p.ElderID, p.Amount, p.BankAccount,
		p.Status, p.FailureReason, p.ExecutedAt, p.CreatedAt)
	return err
}

func (s *SqliteStore) BulkCreateBatchPayments(ctx context.Context, payments []model.CommunityBatchPayment) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO community_batch_payments (id, batch_id, period, pay_type, elder_id, amount, bank_account, status, failure_reason, executed_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC()
	for _, p := range payments {
		if p.ID == "" {
			p.ID = fmt.Sprintf("bp_%d", now.UnixNano())
		}
		_, err := stmt.ExecContext(ctx, p.ID, p.BatchID, p.Period, p.PayType, p.ElderID,
			p.Amount, p.BankAccount, p.Status, p.FailureReason, p.ExecutedAt, p.CreatedAt)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SqliteStore) ListBatchPayments(ctx context.Context, batchID string, page, pageSize int) ([]model.CommunityBatchPayment, error) {
	q := `SELECT id, batch_id, period, pay_type, elder_id, amount, bank_account, status, failure_reason, executed_at, created_at
		 FROM community_batch_payments WHERE 1=1`
	args := []interface{}{}
	if batchID != "" {
		q += " AND batch_id=?"
		args = append(args, batchID)
	}
	q += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []model.CommunityBatchPayment
	for rows.Next() {
		var p model.CommunityBatchPayment
		if err := rows.Scan(&p.ID, &p.BatchID, &p.Period, &p.PayType, &p.ElderID, &p.Amount,
			&p.BankAccount, &p.Status, &p.FailureReason, &p.ExecutedAt, &p.CreatedAt); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, rows.Err()
}

func (s *SqliteStore) UpdateBatchPaymentStatus(ctx context.Context, id, status string, failureReason string) error {
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx,
		`UPDATE community_batch_payments SET status=?, failure_reason=?, executed_at=? WHERE id=?`,
		status, failureReason, now, id)
	return err
}

func (s *SqliteStore) CountPendingPayments(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM community_batch_payments WHERE status='pending'`).Scan(&count)
	return count, err
}
