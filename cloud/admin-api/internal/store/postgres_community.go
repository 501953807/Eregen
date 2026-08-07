package store

import (
	"context"
	"eregen.dev/admin-api/internal/model"
	"fmt"
	"time"
)


func (p *PostgresStore) CreateCommunityElder(ctx context.Context, e *model.CommunityElder) error {

	now := time.Now().UTC(); e.CreatedAt = now; e.UpdatedAt = now

	if e.ID == "" { e.ID = fmt.Sprintf("ce_%d", now.UnixNano()) }

	_, err := p.db.ExecContext(ctx, `INSERT INTO community_elders (id,name,id_card,gender,age,address,emergency_contact,bank_account,hospital_id,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, e.ID, e.Name, e.IDCard, e.Gender, e.Age, e.Address, e.EmergencyContact, e.BankAccount, e.HospitalID, e.Status, e.CreatedAt, e.UpdatedAt)

	return err

}

func (p *PostgresStore) GetCommunityElder(ctx context.Context, id string) (*model.CommunityElder, error) {

	var e model.CommunityElder

	err := p.db.QueryRowContext(ctx, `SELECT id,name,id_card,gender,age,address,emergency_contact,bank_account,hospital_id,status,created_at,updated_at,deactivated_at,deactivated_reason FROM community_elders WHERE id=$1`, id).Scan(&e.ID, &e.Name, &e.IDCard, &e.Gender, &e.Age, &e.Address, &e.EmergencyContact, &e.BankAccount, &e.HospitalID, &e.Status, &e.CreatedAt, &e.UpdatedAt, &e.DeactivatedAt, &e.DeactivatedReason)

	return &e, err

}

func (p *PostgresStore) ListCommunityElders(ctx context.Context, page, pageSize int, status string) ([]model.CommunityElder, error) {

	q := `SELECT id,name,id_card,gender,age,address,emergency_contact,bank_account,hospital_id,status,created_at,updated_at,deactivated_at,deactivated_reason FROM community_elders WHERE 1=1`

	args := []interface{}{}

	if status != "" { q += " AND status=$" + fmt.Sprintf("%d", len(args)+1); args = append(args, status) }

	q += " ORDER BY created_at DESC LIMIT $" + fmt.Sprintf("%d", len(args)+1) + " OFFSET $" + fmt.Sprintf("%d", len(args)+2)

	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := p.db.QueryContext(ctx, q, args...)

	if err != nil { return nil, err }

	defer rows.Close()

	var items []model.CommunityElder

	for rows.Next() {

		var e model.CommunityElder

		if err := rows.Scan(&e.ID, &e.Name, &e.IDCard, &e.Gender, &e.Age, &e.Address, &e.EmergencyContact, &e.BankAccount, &e.HospitalID, &e.Status, &e.CreatedAt, &e.UpdatedAt, &e.DeactivatedAt, &e.DeactivatedReason); err != nil { return nil, err }

		items = append(items, e)

	}

	return items, rows.Err()

}

func (p *PostgresStore) UpdateCommunityElder(ctx context.Context, e *model.CommunityElder) error {

	e.UpdatedAt = time.Now().UTC()

	_, err := p.db.ExecContext(ctx, `UPDATE community_elders SET name=$1,id_card=$2,gender=$3,age=$4,address=$5,emergency_contact=$6,bank_account=$7,hospital_id=$8,status=$9,updated_at=$10,deactivated_at=$11,deactivated_reason=$12 WHERE id=$13`, e.Name, e.IDCard, e.Gender, e.Age, e.Address, e.EmergencyContact, e.BankAccount, e.HospitalID, e.Status, e.UpdatedAt, e.DeactivatedAt, e.DeactivatedReason, e.ID)

	return err

}

func (p *PostgresStore) DeleteCommunityElder(ctx context.Context, id string) error {

	_, err := p.db.ExecContext(ctx, `UPDATE community_elders SET status='deactivated',deactivated_at=NOW(),deactivated_reason='deleted' WHERE id=$1`, id)

	return err

}

func (p *PostgresStore) BulkUpsertCommunityElders(ctx context.Context, elders []model.CommunityElder) error {

	tx, err := p.db.BeginTx(ctx, nil)

	if err != nil { return err }

	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO community_elders (id,name,id_card,gender,age,address,emergency_contact,bank_account,hospital_id,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,now(),now()) ON CONFLICT(id_card) DO UPDATE SET name=excluded.name,status=excluded.status`)

	if err != nil { return err }

	defer stmt.Close()

	for _, e := range elders {

		if e.ID == "" { e.ID = fmt.Sprintf("ce_%d", time.Now().UnixNano()) }

		if _, err := stmt.ExecContext(ctx, e.ID, e.Name, e.IDCard, e.Gender, e.Age, e.Address, e.EmergencyContact, e.BankAccount, e.HospitalID, e.Status); err != nil { return err }

	}

	return tx.Commit()

}

func (p *PostgresStore) GetCommunityElderStats(ctx context.Context) (*model.CommunityElderStats, error) {

	stats := &model.CommunityElderStats{}

	p.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM community_elders`).Scan(&stats.TotalElders)

	p.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM community_elders WHERE status='active'`).Scan(&stats.ActiveElders)

	today := time.Now().Format("2006-01-02")

	p.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM community_signin_records WHERE date(signin_time)=$1 AND is_welfare_signin=1`, today).Scan(&stats.TodaySignins)

	p.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM community_pharmacy_logs WHERE date(dispense_time)=$1`, today).Scan(&stats.TodayDispenses)

	p.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM community_elder_welfare WHERE revoked_at IS NULL`).Scan(&stats.ActiveWelfareTags)

	return stats, nil

}

func (p *PostgresStore) CreateCommunityDevice(ctx context.Context, d *model.CommunityWristbandDevice) error {

	now := time.Now().UTC(); d.CreatedAt = now; d.UpdatedAt = now

	if d.ID == "" { d.ID = fmt.Sprintf("cd_%d", now.UnixNano()) }

	_, err := p.db.ExecContext(ctx, `INSERT INTO community_wristband_devices (id,device_id,firmware_version,mode,status,last_seen,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, d.ID, d.DeviceID, d.FirmwareVersion, d.Mode, d.Status, d.LastSeen, d.CreatedAt, d.UpdatedAt)

	return err

}

func (p *PostgresStore) GetCommunityDevice(ctx context.Context, deviceID string) (*model.CommunityWristbandDevice, error) {

	var d model.CommunityWristbandDevice

	err := p.db.QueryRowContext(ctx, `SELECT id,device_id,firmware_version,mode,status,last_seen,created_at,updated_at FROM community_wristband_devices WHERE device_id=$1`, deviceID).Scan(&d.ID, &d.DeviceID, &d.FirmwareVersion, &d.Mode, &d.Status, &d.LastSeen, &d.CreatedAt, &d.UpdatedAt)

	return &d, err

}

func (p *PostgresStore) ListCommunityDevices(ctx context.Context, page, pageSize int, status string) ([]model.CommunityWristbandDevice, error) {

	q := `SELECT id,device_id,firmware_version,mode,status,last_seen,created_at,updated_at FROM community_wristband_devices WHERE 1=1`

	args := []interface{}{}

	if status != "" { q += " AND status=$" + fmt.Sprintf("%d", len(args)+1); args = append(args, status) }

	q += " ORDER BY created_at DESC LIMIT $" + fmt.Sprintf("%d", len(args)+1) + " OFFSET $" + fmt.Sprintf("%d", len(args)+2)

	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := p.db.QueryContext(ctx, q, args...)

	if err != nil { return nil, err }

	defer rows.Close()

	var items []model.CommunityWristbandDevice

	for rows.Next() {

		var d model.CommunityWristbandDevice

		if err := rows.Scan(&d.ID, &d.DeviceID, &d.FirmwareVersion, &d.Mode, &d.Status, &d.LastSeen, &d.CreatedAt, &d.UpdatedAt); err != nil { return nil, err }

		items = append(items, d)

	}

	return items, rows.Err()

}

func (p *PostgresStore) UpdateCommunityDevice(ctx context.Context, d *model.CommunityWristbandDevice) error {

	d.UpdatedAt = time.Now().UTC()

	_, err := p.db.ExecContext(ctx, `UPDATE community_wristband_devices SET firmware_version=$1,status=$2,last_seen=$3,updated_at=$4 WHERE id=$5`, d.FirmwareVersion, d.Status, d.LastSeen, d.UpdatedAt, d.ID)

	return err

}

func (p *PostgresStore) BindCommunityElderDevice(ctx context.Context, elderID, deviceID string) error {

	id := fmt.Sprintf("cb_%d", time.Now().UnixNano())

	_, err := p.db.ExecContext(ctx, `INSERT INTO community_elder_bindings (id,elder_id,device_id,bound_at) VALUES ($1,$2,$3,NOW())`, id, elderID, deviceID)

	return err

}

func (p *PostgresStore) UnbindCommunityElderDevice(ctx context.Context, bindingID string) error {

	_, err := p.db.ExecContext(ctx, `UPDATE community_elder_bindings SET unbound_at=NOW() WHERE id=$1`, bindingID)

	return err

}

func (p *PostgresStore) CreateWelfareTagConfig(ctx context.Context, c *model.CommunityWelfareTagConfig) error {

	now := time.Now().UTC(); c.CreatedAt = now; c.UpdatedAt = now

	if c.ID == "" { c.ID = fmt.Sprintf("wtc_%d", now.UnixNano()) }

	_, err := p.db.ExecContext(ctx, `INSERT INTO community_welfare_tag_config (id,tag_code,tag_name,issuer,renewal_period_days,benefit_amount,enabled) VALUES ($1,$2,$3,$4,$5,$6,$7)`, c.ID, c.TagCode, c.TagName, c.Issuer, c.RenewalPeriodDays, c.BenefitAmount, c.Enabled)

	return err

}

func (p *PostgresStore) UpdateWelfareTagConfig(ctx context.Context, c *model.CommunityWelfareTagConfig) error {

	c.UpdatedAt = time.Now().UTC()

	_, err := p.db.ExecContext(ctx, `UPDATE community_welfare_tag_config SET tag_name=$1,issuer=$2,renewal_period_days=$3,benefit_amount=$4,enabled=$5,updated_at=$6 WHERE tag_code=$7`, c.TagName, c.Issuer, c.RenewalPeriodDays, c.BenefitAmount, c.Enabled, c.UpdatedAt, c.TagCode)

	return err

}

func (p *PostgresStore) ListWelfareTagConfigs(ctx context.Context) ([]model.CommunityWelfareTagConfig, error) {

	rows, err := p.db.QueryContext(ctx, `SELECT id,tag_code,tag_name,issuer,renewal_period_days,benefit_amount,enabled,created_at,updated_at FROM community_welfare_tag_config ORDER BY tag_code`)

	if err != nil { return nil, err }

	defer rows.Close()

	var items []model.CommunityWelfareTagConfig

	for rows.Next() {

		var c model.CommunityWelfareTagConfig

		rows.Scan(&c.ID, &c.TagCode, &c.TagName, &c.Issuer, &c.RenewalPeriodDays, &c.BenefitAmount, &c.Enabled, &c.CreatedAt, &c.UpdatedAt)

		items = append(items, c)

	}

	return items, rows.Err()

}

func (p *PostgresStore) GetWelfareTagConfig(ctx context.Context, tagCode string) (*model.CommunityWelfareTagConfig, error) {

	var c model.CommunityWelfareTagConfig

	err := p.db.QueryRowContext(ctx, `SELECT id,tag_code,tag_name,issuer,renewal_period_days,benefit_amount,enabled,created_at,updated_at FROM community_welfare_tag_config WHERE tag_code=$1`, tagCode).Scan(&c.ID, &c.TagCode, &c.TagName, &c.Issuer, &c.RenewalPeriodDays, &c.BenefitAmount, &c.Enabled, &c.CreatedAt, &c.UpdatedAt)

	return &c, err

}

func (p *PostgresStore) AssignWelfareTag(ctx context.Context, welfare *model.CommunityElderWelfare) error {

	now := time.Now().UTC(); welfare.EffectiveAt = now

	if welfare.ID == "" { welfare.ID = fmt.Sprintf("ewf_%d", now.UnixNano()) }

	_, err := p.db.ExecContext(ctx, `INSERT INTO community_elder_welfare (id,elder_id,tag_code,valid_from,valid_to,certified_by,certification_doc,effective_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, welfare.ID, welfare.ElderID, welfare.TagCode, welfare.ValidFrom, welfare.ValidTo, welfare.CertifiedBy, welfare.CertificationDoc, welfare.EffectiveAt)

	return err

}

func (p *PostgresStore) RevokeWelfareTag(ctx context.Context, elderID, tagCode string) error {

	_, err := p.db.ExecContext(ctx, `UPDATE community_elder_welfare SET revoked_at=NOW() WHERE elder_id=$1 AND tag_code=$2 AND revoked_at IS NULL`, elderID, tagCode)

	return err

}

func (p *PostgresStore) ListElderWelfareTags(ctx context.Context, elderID string) ([]model.CommunityElderWelfare, error) {

	rows, err := p.db.QueryContext(ctx, `SELECT id,elder_id,tag_code,valid_from,valid_to,certified_by,certification_doc,effective_at,revoked_at FROM community_elder_welfare WHERE elder_id=$1 AND revoked_at IS NULL`, elderID)

	if err != nil { return nil, err }

	defer rows.Close()

	var items []model.CommunityElderWelfare

	for rows.Next() {

		var t model.CommunityElderWelfare

		if err := rows.Scan(&t.ID, &t.ElderID, &t.TagCode, &t.ValidFrom, &t.ValidTo, &t.CertifiedBy, &t.CertificationDoc, &t.EffectiveAt, &t.RevokedAt); err != nil { return nil, err }

		items = append(items, t)

	}

	return items, rows.Err()

}

func (p *PostgresStore) CreateSigninRecord(ctx context.Context, sRec *model.CommunitySigninRecord) error {

	sRec.SigninTime = time.Now().UTC()

	if sRec.ID == "" { sRec.ID = fmt.Sprintf("sr_%d", sRec.SigninTime.UnixNano()) }

	_, err := p.db.ExecContext(ctx, `INSERT INTO community_signin_records (id,elder_id,device_id,hospital_id,pharmacist_id,signin_time,period,activated_tags,is_medical_signin,is_welfare_signin,notes) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, sRec.ID, sRec.ElderID, sRec.DeviceID, sRec.HospitalID, sRec.PharmacistID, sRec.SigninTime, sRec.Period, sRec.ActivatedTags, sRec.IsMedicalSignin, sRec.IsWelfareSignin, sRec.Notes)

	return err

}

func (p *PostgresStore) ListSigninRecords(ctx context.Context, elderID, period, hospitalID string, page, pageSize int) ([]model.CommunitySigninRecord, error) {

	q := `SELECT id,elder_id,device_id,hospital_id,pharmacist_id,signin_time,period,activated_tags,is_medical_signin,is_welfare_signin,notes FROM community_signin_records WHERE 1=1`

	args := []interface{}{}

	if elderID != "" { q += " AND elder_id=$" + fmt.Sprintf("%d", len(args)+1); args = append(args, elderID) }

	if period != "" { q += " AND period=$" + fmt.Sprintf("%d", len(args)+1); args = append(args, period) }

	if hospitalID != "" { q += " AND hospital_id=$" + fmt.Sprintf("%d", len(args)+1); args = append(args, hospitalID) }

	q += " ORDER BY signin_time DESC LIMIT $" + fmt.Sprintf("%d", len(args)+1) + " OFFSET $" + fmt.Sprintf("%d", len(args)+2)

	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := p.db.QueryContext(ctx, q, args...)

	if err != nil { return nil, err }

	defer rows.Close()

	var items []model.CommunitySigninRecord

	for rows.Next() {

		var r model.CommunitySigninRecord

		if err := rows.Scan(&r.ID, &r.ElderID, &r.DeviceID, &r.HospitalID, &r.PharmacistID, &r.SigninTime, &r.Period, &r.ActivatedTags, &r.IsMedicalSignin, &r.IsWelfareSignin, &r.Notes); err != nil { return nil, err }

		items = append(items, r)

	}

	return items, rows.Err()

}

func (p *PostgresStore) GetSigninSummary(ctx context.Context, elderID, period string) (*model.CommunitySigninRecord, error) {

	var r model.CommunitySigninRecord

	err := p.db.QueryRowContext(ctx, `SELECT id,elder_id,device_id,hospital_id,pharmacist_id,signin_time,period,activated_tags,is_medical_signin,is_welfare_signin,notes FROM community_signin_records WHERE elder_id=$1 AND period=$2 ORDER BY signin_time DESC LIMIT 1`, elderID, period).Scan(&r.ID, &r.ElderID, &r.DeviceID, &r.HospitalID, &r.PharmacistID, &r.SigninTime, &r.Period, &r.ActivatedTags, &r.IsMedicalSignin, &r.IsWelfareSignin, &r.Notes)

	return &r, err

}

func (p *PostgresStore) CreatePharmacyLog(ctx context.Context, pLog *model.CommunityPharmacyLog) error {

	pLog.DispenseTime = time.Now().UTC()

	if pLog.ID == "" { pLog.ID = fmt.Sprintf("pl_%d", pLog.DispenseTime.UnixNano()) }

	_, err := p.db.ExecContext(ctx, `INSERT INTO community_pharmacy_logs (id,elder_id,device_id,hospital_id,pharmacist_id,dispense_time,period,items,total_cost,insurance_covered,self_pay,notes) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, pLog.ID, pLog.ElderID, pLog.DeviceID, pLog.HospitalID, pLog.PharmacistID, pLog.DispenseTime, pLog.Period, pLog.Items, pLog.TotalCost, pLog.InsuranceCovered, pLog.SelfPay, pLog.Notes)

	return err

}

func (p *PostgresStore) ListPharmacyLogs(ctx context.Context, elderID, period string, page, pageSize int) ([]model.CommunityPharmacyLog, error) {

	q := `SELECT id,elder_id,device_id,hospital_id,pharmacist_id,dispense_time,period,items,total_cost,insurance_covered,self_pay,notes FROM community_pharmacy_logs WHERE 1=1`

	args := []interface{}{}

	if elderID != "" { q += " AND elder_id=$" + fmt.Sprintf("%d", len(args)+1); args = append(args, elderID) }

	if period != "" { q += " AND period=$" + fmt.Sprintf("%d", len(args)+1); args = append(args, period) }

	q += " ORDER BY dispense_time DESC LIMIT $" + fmt.Sprintf("%d", len(args)+1) + " OFFSET $" + fmt.Sprintf("%d", len(args)+2)

	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := p.db.QueryContext(ctx, q, args...)

	if err != nil { return nil, err }

	defer rows.Close()

	var items []model.CommunityPharmacyLog

	for rows.Next() {

		var l model.CommunityPharmacyLog

		if err := rows.Scan(&l.ID, &l.ElderID, &l.DeviceID, &l.HospitalID, &l.PharmacistID, &l.DispenseTime, &l.Period, &l.Items, &l.TotalCost, &l.InsuranceCovered, &l.SelfPay, &l.Notes); err != nil { return nil, err }

		items = append(items, l)

	}

	return items, rows.Err()

}

func (p *PostgresStore) CreateMinzhengSync(ctx context.Context, m *model.CommunityMinzhengSync) error {

	now := time.Now().UTC(); m.CreatedAt = now

	if m.ID == "" { m.ID = fmt.Sprintf("ms_%d", now.UnixNano()) }

	_, err := p.db.ExecContext(ctx, `INSERT INTO community_minzheng_sync (id,source,filename,imported_count,matched_count,pending_review_count,error_count,status,created_at,completed_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, m.ID, m.Source, m.Filename, m.ImportedCount, m.MatchedCount, m.PendingReviewCount, m.ErrorCount, m.Status, m.CreatedAt, m.CompletedAt)

	return err

}

func (p *PostgresStore) ListMinzhengSync(ctx context.Context, page, pageSize int) ([]model.CommunityMinzhengSync, error) {

	rows, err := p.db.QueryContext(ctx, `SELECT id,source,filename,imported_count,matched_count,pending_review_count,error_count,status,created_at,completed_at FROM community_minzheng_sync ORDER BY created_at DESC LIMIT $1 OFFSET $2`, pageSize, (page-1)*pageSize)

	if err != nil { return nil, err }

	defer rows.Close()

	var items []model.CommunityMinzhengSync

	for rows.Next() {

		var m model.CommunityMinzhengSync

		if err := rows.Scan(&m.ID, &m.Source, &m.Filename, &m.ImportedCount, &m.MatchedCount, &m.PendingReviewCount, &m.ErrorCount, &m.Status, &m.CreatedAt, &m.CompletedAt); err != nil { return nil, err }

		items = append(items, m)

	}

	return items, rows.Err()

}

func (p *PostgresStore) GetLatestMinzhengSync(ctx context.Context) (*model.CommunityMinzhengSync, error) {

	var m model.CommunityMinzhengSync

	err := p.db.QueryRowContext(ctx, `SELECT id,source,filename,imported_count,matched_count,pending_review_count,error_count,status,created_at,completed_at FROM community_minzheng_sync ORDER BY created_at DESC LIMIT 1`).Scan(&m.ID, &m.Source, &m.Filename, &m.ImportedCount, &m.MatchedCount, &m.PendingReviewCount, &m.ErrorCount, &m.Status, &m.CreatedAt, &m.CompletedAt)

	return &m, err

}

func (p *PostgresStore) CreateBatchPayment(ctx context.Context, pmt *model.CommunityBatchPayment) error {

	now := time.Now().UTC(); pmt.CreatedAt = now

	if pmt.ID == "" { pmt.ID = fmt.Sprintf("bp_%d", now.UnixNano()) }

	_, err := p.db.ExecContext(ctx, `INSERT INTO community_batch_payments (id,batch_id,period,pay_type,elder_id,amount,bank_account,status,failure_reason,executed_at,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, pmt.ID, pmt.BatchID, pmt.Period, pmt.PayType, pmt.ElderID, pmt.Amount, pmt.BankAccount, pmt.Status, pmt.FailureReason, pmt.ExecutedAt, pmt.CreatedAt)

	return err

}

func (p *PostgresStore) BulkCreateBatchPayments(ctx context.Context, payments []model.CommunityBatchPayment) error {

	tx, err := p.db.BeginTx(ctx, nil)

	if err != nil { return err }

	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO community_batch_payments (id,batch_id,period,pay_type,elder_id,amount,bank_account,status,failure_reason,executed_at,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`)

	if err != nil { return err }

	defer stmt.Close()

	now := time.Now().UTC()

	for _, pmt := range payments {

		if pmt.ID == "" { pmt.ID = fmt.Sprintf("bp_%d", now.UnixNano()) }

		if _, err := stmt.ExecContext(ctx, pmt.ID, pmt.BatchID, pmt.Period, pmt.PayType, pmt.ElderID, pmt.Amount, pmt.BankAccount, pmt.Status, pmt.FailureReason, pmt.ExecutedAt, pmt.CreatedAt); err != nil { return err }

	}

	return tx.Commit()

}

func (p *PostgresStore) ListBatchPayments(ctx context.Context, batchID string, page, pageSize int) ([]model.CommunityBatchPayment, error) {

	q := `SELECT id,batch_id,period,pay_type,elder_id,amount,bank_account,status,failure_reason,executed_at,created_at FROM community_batch_payments WHERE 1=1`

	args := []interface{}{}

	if batchID != "" { q += " AND batch_id=$" + fmt.Sprintf("%d", len(args)+1); args = append(args, batchID) }

	q += " ORDER BY created_at DESC LIMIT $" + fmt.Sprintf("%d", len(args)+1) + " OFFSET $" + fmt.Sprintf("%d", len(args)+2)

	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := p.db.QueryContext(ctx, q, args...)

	if err != nil { return nil, err }

	defer rows.Close()

	var items []model.CommunityBatchPayment

	for rows.Next() {

		var pmt model.CommunityBatchPayment

		if err := rows.Scan(&pmt.ID, &pmt.BatchID, &pmt.Period, &pmt.PayType, &pmt.ElderID, &pmt.Amount, &pmt.BankAccount, &pmt.Status, &pmt.FailureReason, &pmt.ExecutedAt, &pmt.CreatedAt); err != nil { return nil, err }

		items = append(items, pmt)

	}

	return items, rows.Err()

}

func (p *PostgresStore) UpdateBatchPaymentStatus(ctx context.Context, id, status string, failureReason string) error {

	_, err := p.db.ExecContext(ctx, `UPDATE community_batch_payments SET status=$1,failure_reason=$2,executed_at=NOW() WHERE id=$3`, status, failureReason, id)

	return err

}

func (p *PostgresStore) CountPendingPayments(ctx context.Context) (int64, error) {

	var count int64

	err := p.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM community_batch_payments WHERE status='pending'`).Scan(&count)

	return count, err

}



// ========== Clinical Workflow Methods ==========

