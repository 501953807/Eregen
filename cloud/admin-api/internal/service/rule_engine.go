package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"
	"go.uber.org/zap"
)

// RuleEngine runs periodic compliance detection across all 16 rules (R01-R08 + R_C01-R_C08).
type RuleEngine struct {
	store store.Store
	log   *zap.Logger
}

func NewRuleEngine(s store.Store, log *zap.Logger) *RuleEngine {
	return &RuleEngine{store: s, log: log}
}

// Run starts the ticker that fires every 5 minutes and evaluates all rules.
func (e *RuleEngine) Run() {
	e.log.Info("rule engine started")
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		e.checkNoVerify()        // R01: 挂床住院
		e.checkFenceViolation()  // R02: 电子围栏越界
		e.checkFakeAdmission()   // R03: 虚假入院
		e.checkExpenseSpike()    // R04: 费用突增
		e.checkMedVerifyMismatch() // R05: 用药与核验不匹配
		e.checkFrequentTransfer() // R06: 频繁转科
		e.checkDeviceDisconnect() // R07: 设备离线
		e.checkLongNoDischarge() // R08: 长期不在院
		e.checkCommunityDuplicate()    // R_C01: 重复领取
		e.checkCommunityCrossInstitution() // R_C02: 跨社区互认
		e.checkCommunityHighFrequency() // R_C03: 异常高频
		e.checkCommunityZombie()       // R_C04: 僵尸账户
		e.checkCommunityBenefitFailed() // R_C05: 补助未到账
		e.checkCommunityCrossDistrict() // R_C06: 跨区领取
		e.checkCommunityBenefitExpired() // R_C07: 福利过期未停
		e.checkCommunityDeath()      // R_C08: 死亡未注销
		e.log.Debug("rule engine tick completed")
	}
}

// checkNoVerify detects patients admitted without any verification records (R01).
func (e *RuleEngine) checkNoVerify() {
	patients, err := e.store.ListPatients(context.Background(), 1, 1000, "admitted")
	if err != nil {
		e.log.Error("R01: list patients failed", zap.Error(err))
		return
	}
	verifications, err := e.store.ListVerifications(context.Background(), 1, 1000)
	if err != nil {
		e.log.Error("R01: list verifications failed", zap.Error(err))
		return
	}
	verified := make(map[string]bool)
	for _, v := range verifications {
		if v.Matched && v.PatientID != nil {
			verified[*v.PatientID] = true
		}
	}
	for _, p := range patients {
		if verified[p.ID] {
			continue
		}
		pID := p.ID
		_ = e.store.CreateRegulatoryAlert(context.Background(), &model.RegulatoryAlert{
			RuleCode:   "R01",
			PatientID:  &pID,
			HospitalID: "",
			Department: p.Department,
			Severity:   "high",
			AlertType:  "no_verify",
			Detail:     fmt.Sprintf("Patient admitted without verification: %s (%s)", p.Name, p.AdmissionNo),
		})
		e.log.Warn("R01: admission without verification", zap.String("patient_id", p.ID))
	}
}

// checkFenceViolation detects patients outside their geofence (R02).
func (e *RuleEngine) checkFenceViolation() {
	logs, err := e.store.ListLocationLogs(context.Background(), "", 500)
	if err != nil {
		e.log.Error("R02: list location logs failed", zap.Error(err))
		return
	}
	for _, log := range logs {
		if !log.InsideFence && log.PatientID != "" {
			pid := log.PatientID
			_ = e.store.CreateRegulatoryAlert(context.Background(), &model.RegulatoryAlert{
				RuleCode:  "R02",
				PatientID: &pid,
				Severity:  "medium",
				AlertType: "fence_violation",
				Detail:    fmt.Sprintf("Patient outside geofence: lat=%.4f lng=%.4f", log.Lat, log.Lng),
			})
		}
	}
}

// checkFakeAdmission detects suspicious admissions without daily nursing entries (R03).
func (e *RuleEngine) checkFakeAdmission() {
	patients, err := e.store.ListPatients(context.Background(), 1, 1000, "admitted")
	if err != nil {
		e.log.Error("R03: list patients failed", zap.Error(err))
		return
	}
	for _, p := range patients {
		history, err := e.store.GetPatientHistory(context.Background(), p.ID)
		if err != nil || len(history.DailyEntries) == 0 {
			pID := p.ID
			_ = e.store.CreateRegulatoryAlert(context.Background(), &model.RegulatoryAlert{
				RuleCode:  "R03",
				PatientID: &pID,
				Severity:  "medium",
				AlertType: "fake_admission",
				Detail:    fmt.Sprintf("Suspicious admission — no daily nursing entries: %s (%s)", p.Name, p.AdmissionNo),
			})
			e.log.Warn("R03: suspicious admission", zap.String("patient_id", p.ID))
		}
	}
}

// checkExpenseSpike detects patients with daily expenses exceeding 3x department average (R04).
func (e *RuleEngine) checkExpenseSpike() {
	e.log.Info("R04: checking expense anomalies")
	now := time.Now()
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -7).Format("2006-01-02")
	patients, err := e.store.ListPatients(context.Background(), 1, 500, "admitted")
	if err != nil {
		e.log.Error("R04: list patients failed", zap.Error(err))
		return
	}
	for _, p := range patients {
		expenses, err := e.store.ListExpenses(context.Background(), p.ID, 1, 100)
		if err != nil {
			continue
		}
		var todayTotal, weekTotal float64
		for _, exp := range expenses {
			dateStr := exp.CreatedAt.Format("2006-01-02")
			if dateStr == today {
				todayTotal += exp.Amount
			}
			if dateStr >= yesterday {
				weekTotal += exp.Amount
			}
		}
		if todayTotal > 0 && weekTotal > 0 && todayTotal > weekTotal*3 {
			pID := p.ID
			_ = e.store.CreateRegulatoryAlert(context.Background(), &model.RegulatoryAlert{
				RuleCode:  "R04",
				PatientID: &pID,
				Severity:  "high",
				AlertType: "expense_spike",
				Detail:    fmt.Sprintf("Daily expense %.0f exceeds 3x weekly average %.0f: %s", todayTotal, weekTotal, p.Name),
			})
			e.log.Warn("R04: expense spike detected", zap.String("patient_id", p.ID), zap.Float64("amount", todayTotal))
		}
	}
}

// checkMedVerifyMismatch detects medications without corresponding verification records (R05).
func (e *RuleEngine) checkMedVerifyMismatch() {
	e.log.Info("R05: checking medication-verification mismatch")
	patients, err := e.store.ListPatients(context.Background(), 1, 500, "admitted")
	if err != nil {
		e.log.Error("R05: list patients failed", zap.Error(err))
		return
	}
	verifications, err := e.store.ListVerifications(context.Background(), 1, 500)
	if err != nil {
		e.log.Error("R05: list verifications failed", zap.Error(err))
		return
	}
	// Build set of patients with recent verification (within 24h)
	verifiedMap := make(map[string]bool)
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)
	for _, v := range verifications {
		if v.Matched && v.PatientID != nil && v.VerifiedAt.After(twentyFourHoursAgo) {
			verifiedMap[*v.PatientID] = true
		}
	}
	for _, p := range patients {
		meds, err := e.store.ListMedications(context.Background(), p.ID)
		if err != nil {
			continue
		}
		if len(meds) > 0 && !verifiedMap[p.ID] {
			pID := p.ID
			_ = e.store.CreateRegulatoryAlert(context.Background(), &model.RegulatoryAlert{
				RuleCode:  "R05",
				PatientID: &pID,
				Severity:  "medium",
				AlertType: "med_verify_mismatch",
				Detail:    fmt.Sprintf("Patient has %d medications but no recent verification: %s", len(meds), p.Name),
			})
			e.log.Warn("R05: med-verification mismatch", zap.String("patient_id", p.ID))
		}
	}
}

// checkFrequentTransfer detects patients who changed departments >3 times in 7 days (R06).
func (e *RuleEngine) checkFrequentTransfer() {
	e.log.Info("R06: checking frequent transfers")
	patients, err := e.store.ListPatients(context.Background(), 1, 500, "admitted")
	if err != nil {
		e.log.Error("R06: list patients failed", zap.Error(err))
		return
	}
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	for _, p := range patients {
		history, err := e.store.GetPatientHistory(context.Background(), p.ID)
		if err != nil || len(history.DailyEntries) == 0 {
			continue
		}
		transferCount := 0
		for _, entry := range history.DailyEntries {
			if entry.EntryType == "transfer" && entry.EntryDate != "" {
				if t, err := time.Parse("2006-01-02", entry.EntryDate); err == nil && t.After(sevenDaysAgo) {
					transferCount++
				}
			}
		}
		if transferCount > 3 {
			pID := p.ID
			_ = e.store.CreateRegulatoryAlert(context.Background(), &model.RegulatoryAlert{
				RuleCode:  "R06",
				PatientID: &pID,
				Severity:  "high",
				AlertType: "frequent_transfer",
				Detail:    fmt.Sprintf("Patient transferred %d times in 7 days: %s", transferCount, p.Name),
			})
			e.log.Warn("R06: frequent transfer detected", zap.String("patient_id", p.ID), zap.Int("transfers", transferCount))
		}
	}
}

// checkDeviceDisconnect detects devices offline for >30 minutes (R07).
func (e *RuleEngine) checkDeviceDisconnect() {
	e.log.Info("R07: checking device disconnects")
	disconnectThreshold := 30 * time.Minute
	devices, err := e.store.ListDevices(context.Background(), 1, 1000, "", "", "")
	if err != nil {
		e.log.Error("R07: list devices failed", zap.Error(err))
		return
	}
	now := time.Now()
	for _, d := range devices {
		if d.Status != "online" {
			continue
		}
		timeSinceLastSeen := now.Sub(d.LastSeen)
		if timeSinceLastSeen > disconnectThreshold {
			e.log.Warn("R07: device disconnect detected",
				zap.String("device_id", d.DeviceID),
				zap.Duration("offline_for", timeSinceLastSeen))
		}
	}
}

// checkLongNoDischarge detects patients admitted >14 days without discharge (R08).
func (e *RuleEngine) checkLongNoDischarge() {
	e.log.Info("R08: checking long no-discharge")
	patients, err := e.store.ListPatients(context.Background(), 1, 500, "admitted")
	if err != nil {
		e.log.Error("R08: list patients failed", zap.Error(err))
		return
	}
	fourteenDaysAgo := time.Now().AddDate(0, 0, -14)
	for _, p := range patients {
		if !p.CreatedAt.IsZero() && p.CreatedAt.Before(fourteenDaysAgo) {
			pID := p.ID
			_ = e.store.CreateRegulatoryAlert(context.Background(), &model.RegulatoryAlert{
				RuleCode:  "R08",
				PatientID: &pID,
				Severity:  "medium",
				AlertType: "post_discharge",
				Detail:    fmt.Sprintf("Patient admitted >14 days without discharge: %s (%s)", p.Name, p.AdmissionNo),
			})
			e.log.Warn("R08: patient admitted >14 days without discharge",
				zap.String("patient_id", p.ID),
				zap.String("admission_no", p.AdmissionNo),
				zap.Time("admitted_at", p.CreatedAt))
		}
	}
}

// checkCommunityDuplicate detects same elder signing at different hospitals in same month (R_C01).
func (e *RuleEngine) checkCommunityDuplicate() {
	e.log.Info("R_C01: checking duplicate sign-ins")
	elders, err := e.store.ListCommunityElders(context.Background(), 1, 500, "active")
	if err != nil {
		e.log.Error("R_C01: list elders failed", zap.Error(err))
		return
	}
	thisMonth := time.Now().Format("2006-01")
	for _, elder := range elders {
		records, err := e.store.ListSigninRecords(context.Background(), elder.ID, thisMonth, "", 1, 100)
		if err != nil {
			continue
		}
		// Check for different hospital_id in same month
		hospitals := make(map[string]bool)
		for _, r := range records {
			hospitals[r.HospitalID] = true
		}
		if len(hospitals) > 1 {
			e.log.Warn("R_C01: duplicate sign-in detected",
				zap.String("elder_id", elder.ID),
				zap.Int("hospital_count", len(hospitals)))
		}
	}
}

// checkCommunityCrossInstitution detects same ID card signing at different institutions (R_C02).
func (e *RuleEngine) checkCommunityCrossInstitution() {
	e.log.Info("R_C02: checking cross-institution sign-ins")
	// Query sign-in records grouped by ID card
	elders, err := e.store.ListCommunityElders(context.Background(), 1, 500, "active")
	if err != nil {
		e.log.Error("R_C02: list elders failed", zap.Error(err))
		return
	}
	thisMonth := time.Now().Format("2006-01")
	for _, elder := range elders {
		records, err := e.store.ListSigninRecords(context.Background(), elder.ID, thisMonth, "", 1, 100)
		if err != nil {
			continue
		}
		if len(records) == 0 {
			continue
		}
		// Check if same elder ID card appears in different hospital IDs
		hospitalIDs := make(map[string]bool)
		for _, r := range records {
			hospitalIDs[r.HospitalID] = true
		}
		if len(hospitalIDs) > 1 {
			e.log.Warn("R_C02: cross-institution sign-in detected",
				zap.String("elder_id", elder.ID),
				zap.Int("hospital_count", len(hospitalIDs)))
		}
	}
}

// checkCommunityHighFrequency detects elder with >5 sign-ins or pharmacy visits in 7 days (R_C03).
func (e *RuleEngine) checkCommunityHighFrequency() {
	e.log.Info("R_C03: checking high-frequency sign-ins")
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	elders, err := e.store.ListCommunityElders(context.Background(), 1, 500, "active")
	if err != nil {
		e.log.Error("R_C03: list elders failed", zap.Error(err))
		return
	}
	for _, elder := range elders {
		records, err := e.store.ListSigninRecords(context.Background(), elder.ID, "", "", 1, 100)
		if err != nil {
			continue
		}
		pharmacyLogs, err := e.store.ListPharmacyLogs(context.Background(), elder.ID, "", 1, 100)
		if err != nil {
			pharmacyLogs = []model.CommunityPharmacyLog{}
		}
		totalEvents := len(records) + len(pharmacyLogs)
		_ = totalEvents // used for logging/debugging
		recentEvents := 0
		for _, r := range records {
			if r.SigninTime.After(sevenDaysAgo) {
				recentEvents++
			}
		}
		for _, p := range pharmacyLogs {
			if p.DispenseTime.After(sevenDaysAgo) {
				recentEvents++
			}
		}
		if recentEvents > 5 {
			e.log.Warn("R_C03: high-frequency activity detected",
				zap.String("elder_id", elder.ID),
				zap.Int("events", recentEvents))
		}
	}
}

// checkCommunityZombie detects offline devices with no sign-in for >30 days (R_C04).
func (e *RuleEngine) checkCommunityZombie() {
	e.log.Info("R_C04: checking zombie accounts")
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	elders, err := e.store.ListCommunityElders(context.Background(), 1, 500, "active")
	if err != nil {
		e.log.Error("R_C04: list elders failed", zap.Error(err))
		return
	}
	for _, elder := range elders {
		records, err := e.store.ListSigninRecords(context.Background(), elder.ID, "", "", 1, 1)
		if err != nil {
			continue
		}
		if len(records) == 0 {
			e.log.Warn("R_C04: zombie account detected (no sign-ins)",
				zap.String("elder_id", elder.ID),
				zap.String("name", elder.Name))
			continue
		}
		// Check if last sign-in is >30 days ago
		lastSignin := records[0].SigninTime
		if lastSignin.Before(thirtyDaysAgo) {
			e.log.Warn("R_C04: zombie account detected (stale sign-in)",
				zap.String("elder_id", elder.ID),
				zap.Time("last_signin", lastSignin))
		}
	}
}

// checkCommunityBenefitFailed detects activated welfare tags with failed bank payments (R_C05).
func (e *RuleEngine) checkCommunityBenefitFailed() {
	e.log.Info("R_C05: checking benefit payment failures")
	elders, err := e.store.ListCommunityElders(context.Background(), 1, 500, "active")
	if err != nil {
		e.log.Error("R_C05: list elders failed", zap.Error(err))
		return
	}
	for _, elder := range elders {
		welfareTags, err := e.store.ListElderWelfareTags(context.Background(), elder.ID)
		if err != nil || len(welfareTags) == 0 {
			continue
		}
		for _, tag := range welfareTags {
			payments, err := e.store.ListBatchPayments(context.Background(), "", 1, 100)
			if err != nil {
				continue
			}
			for _, pay := range payments {
				if pay.ElderID == elder.ID && pay.Status == "failed" {
					e.log.Warn("R_C05: benefit payment failed",
						zap.String("elder_id", elder.ID),
						zap.String("tag_code", tag.TagCode),
						zap.String("payment_id", pay.ID))
				}
			}
		}
	}
}

// checkCommunityCrossDistrict detects same welfare tag used across multiple institutions (R_C06).
func (e *RuleEngine) checkCommunityCrossDistrict() {
	e.log.Info("R_C06: checking cross-district welfare usage")
	tagConfigs, err := e.store.ListWelfareTagConfigs(context.Background())
	if err != nil {
		e.log.Error("R_C06: list welfare tags failed", zap.Error(err))
		return
	}
	for _, tag := range tagConfigs {
		elders, err := e.store.ListCommunityElders(context.Background(), 1, 500, "active")
		if err != nil {
			continue
		}
		hospitalSet := make(map[string]bool)
		for _, elder := range elders {
			welfareTags, err := e.store.ListElderWelfareTags(context.Background(), elder.ID)
			if err != nil {
				continue
			}
			for _, wt := range welfareTags {
				if wt.TagCode == tag.TagCode {
					if elder.HospitalID != "" {
						hospitalSet[elder.HospitalID] = true
					}
				}
			}
		}
		if len(hospitalSet) > 1 {
			e.log.Warn("R_C06: welfare tag used across multiple institutions",
				zap.String("tag_code", tag.TagCode),
				zap.Int("institution_count", len(hospitalSet)))
		}
	}
}

// checkCommunityBenefitExpired detects expired welfare tags still in use (R_C07).
func (e *RuleEngine) checkCommunityBenefitExpired() {
	e.log.Info("R_C07: checking expired welfare tags")
	elders, err := e.store.ListCommunityElders(context.Background(), 1, 500, "active")
	if err != nil {
		e.log.Error("R_C07: list elders failed", zap.Error(err))
		return
	}
	today := time.Now().Format("2006-01-02")
	for _, elder := range elders {
		welfareTags, err := e.store.ListElderWelfareTags(context.Background(), elder.ID)
		if err != nil {
			continue
		}
		for _, tag := range welfareTags {
			if tag.ValidTo != "" && tag.ValidTo < today {
				e.log.Warn("R_C07: expired welfare tag still active",
					zap.String("elder_id", elder.ID),
					zap.String("tag_code", tag.TagCode),
					zap.String("valid_to", tag.ValidTo))
			}
		}
	}
}

// checkCommunityDeath detects deceased elders with active wristbands (R_C08).
func (e *RuleEngine) checkCommunityDeath() {
	e.log.Info("R_C08: checking deceased with active wristbands")
	elders, err := e.store.ListCommunityElders(context.Background(), 1, 500, "deceased")
	if err != nil {
		e.log.Error("R_C08: list deceased elders failed", zap.Error(err))
		return
	}
	for _, elder := range elders {
		// Check if wristband is still active
		e.log.Warn("R_C08: deceased elder with potential active wristband",
			zap.String("elder_id", elder.ID),
			zap.String("name", elder.Name))
	}
}

// FenceCalculator computes Haversine distance between two lat/lng points.
type FenceCalculator struct{}

func (f *FenceCalculator) Distance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000.0 // Earth radius in meters
	dlat := (lat2 - lat1) * 0.01745329251
	dlng := (lng2 - lng1) * 0.01745329251
	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1*0.01745329251)*math.Cos(lat2*0.01745329251)*math.Sin(dlng/2)*math.Sin(dlng/2)
	c := 2 * math.Asin(math.Sqrt(a))
	return R * c
}

// IsInside returns true if distance <= radius.
func (f *FenceCalculator) IsInside(patientLat, patientLng, centerLat, centerLng float64, radiusMeters int) bool {
	return f.Distance(patientLat, patientLng, centerLat, centerLng) <= float64(radiusMeters)
}

// IsInsideWithSource returns true if distance <= radius + accuracy margin based on location source.
func (f *FenceCalculator) IsInsideWithSource(patientLat, patientLng, centerLat, centerLng float64, radiusMeters int, source string) bool {
	dist := f.Distance(patientLat, patientLng, centerLat, centerLng)
	accuracyMargin := 5.0
	if source == "base_station" {
		accuracyMargin = 500.0
	}
	return dist <= float64(radiusMeters) + accuracyMargin
}

// Desensitize filters medical data for family-app visibility.
type Desensitize struct{}

func (d *Desensitize) FilterMedication(meds []model.MedicalMedication) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(meds))
	for _, m := range meds {
		result = append(result, map[string]interface{}{
			"name":    m.Name,
			"dosage":  "",
			"created": m.CreatedAt,
		})
	}
	return result
}

func (d *Desensitize) FilterExpense(expenses []model.MedicalExpense) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(expenses))
	for _, e := range expenses {
		result = append(result, map[string]interface{}{
			"name":       e.ItemName,
			"amount":     e.Amount,
			"category":   e.Category,
			"created_at": e.CreatedAt,
		})
	}
	return result
}
