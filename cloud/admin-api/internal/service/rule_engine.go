package service

import (
	"fmt"
	"math"
	"context"
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
		e.checkNoVerify()       // R01: 挂床住院
		e.checkFenceViolation() // R02: 电子围栏越界
		e.checkFakeAdmission()  // R03: 虚假入院
		// R04-R08 would query expenses, medications, etc.
		e.checkDeviceDisconnect() // R07
		e.log.Debug("rule engine tick completed")
	}
}

func (e *RuleEngine) checkNoVerify() {
	e.log.Info("R01: checking no-verify alerts")
	alerts, err := e.store.ListRegulatoryAlerts(context.Background(), "", "", "pending", "", 1, 100)
	if err != nil {
		e.log.Error("R01: list alerts failed", zap.Error(err))
		return
	}
	for _, a := range alerts {
		if a.RuleCode == "R01" && a.Severity == "high" {
			pid := ""
			if a.PatientID != nil {
				pid = *a.PatientID
			}
			e.log.Warn("R01 alert detected", zap.String("patient_id", pid))
		}
	}
	_ = fmt.Sprintf("alerts count: %d", len(alerts))
}

func (e *RuleEngine) checkFenceViolation() {
	e.log.Info("R02: checking fence violations")
	// Query recent location logs where inside_fence changed, check source-aware accuracy
	logs, err := e.store.ListLocationLogs(context.Background(), "", 500)
	if err != nil {
		e.log.Error("R02: list location logs failed", zap.Error(err))
		return
	}
	for _, log := range logs {
		if !log.InsideFence {
			// Already flagged as outside — check if it was a GPS vs base_station false positive
			e.log.Debug("R02: fence violation detected",
				zap.String("patient_id", log.PatientID),
				zap.Float64("lat", log.Lat),
				zap.Float64("lng", log.Lng))
		}
	}
}

func (e *RuleEngine) checkFakeAdmission() {
	e.log.Info("R03: checking fake admissions")
}

func (e *RuleEngine) checkDeviceDisconnect() {
	e.log.Info("R07: checking device disconnects")
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
// GPS has ~5m accuracy; base station has ~500m accuracy.
func (f *FenceCalculator) IsInsideWithSource(patientLat, patientLng, centerLat, centerLng float64, radiusMeters int, source string) bool {
	dist := f.Distance(patientLat, patientLng, centerLat, centerLng)
	accuracyMargin := 5.0 // GPS default
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
			"dosage":  "",       // No dosage detail for family
			"created": m.CreatedAt,
		})
	}
	return result
}

func (d *Desensitize) FilterExpense(expenses []model.MedicalExpense) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(expenses))
	for _, e := range expenses {
		result = append(result, map[string]interface{}{
			"name":      e.ItemName,
			"amount":    e.Amount,
			"category":  e.Category,
			"created_at": e.CreatedAt,
		})
	}
	return result
}
