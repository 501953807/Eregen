package subscriber

import (
	"encoding/json"
	"log"
	"math"

	"eregen.dev/pipeline/internal/analyzer"
	"eregen.dev/pipeline/internal/model"
	"eregen.dev/pipeline/internal/store"

	"github.com/nats-io/nats.go"
)

// Handler processes device events from NATS and runs AI analysis.
type Handler struct {
	js            nats.JetStreamContext
	healthAnalyzer  *analyzer.HealthAnalyzer
	riskCalculator  *analyzer.RiskScoreCalculator
	store           *store.Store
}

// NewHandler creates a NATS event handler with analyzer and store.
func NewHandler(js nats.JetStreamContext, hAnalyzer *analyzer.HealthAnalyzer,
	rCalc *analyzer.RiskScoreCalculator, st *store.Store,
) *Handler {
	return &Handler{
		js:             js,
		healthAnalyzer: hAnalyzer,
		riskCalculator: rCalc,
		store:          st,
	}
}

// Start subscribes to the DEVICE_EVENTS stream and processes health data.
func (h *Handler) Start() error {
	_, err := h.js.Subscribe("device.events", h.onMessage,
		nats.Durable("pipeline-service"),
	)
	return err
}

func (h *Handler) onMessage(msg *nats.Msg) {
	var envelope struct {
		Type    string                 `json:"type"`
		DevID   string                 `json:"dev_id"`
		Payload map[string]interface{} `json:"payload"`
	}
	if err := json.Unmarshal(msg.Data, &envelope); err != nil {
		log.Printf("[pipeline] unmarshal: %v", err)
		msg.Ack()
		return
	}

	switch envelope.Type {
	case "health":
		h.processHealth(envelope.DevID, envelope.Payload)
	case "location":
		h.processLocation(envelope.DevID, envelope.Payload)
	case "med_status":
		h.processMedStatus(envelope.DevID, envelope.Payload)
	case "heartbeat":
		// Only updates device status — no AI analysis needed
	}

	msg.Ack()
}

// processHealth runs anomaly detection on vital signs.
func (h *Handler) processHealth(elderlyID string, payload map[string]interface{}) {
	metrics := make(map[string]float64)
	baselines := make(map[string]float64)

	// Extract metrics from payload
	if v, ok := payload["hr"].(float64); ok {
		metrics["heart_rate"] = v
	}
	if v, ok := payload["spo2"].(float64); ok {
		metrics["spo2"] = v
	}
	if v, ok := payload["step"].(float64); ok {
		metrics["steps"] = v
	}
	if bp, ok := payload["bp"].(map[string]interface{}); ok {
		if sys, ok := bp["systolic"].(float64); ok {
			metrics["bp_systolic"] = sys
		}
		if dia, ok := bp["diastolic"].(float64); ok {
			metrics["bp_diastolic"] = dia
		}
	}

	// Fetch 7-day baselines from InfluxDB
	for metric := range metrics {
		baseline, err := h.store.QueryBaseline(elderlyID, metric, 7)
		if err != nil {
			log.Printf("[pipeline] baseline query %s: %v", metric, err)
			continue
		}
		baselines[metric] = baseline
	}

	// Run analysis
	results := h.healthAnalyzer.AnalyzeBatch(elderlyID, metrics, baselines)

	// Store results and trigger alerts if needed
	for _, result := range results {
		if err := h.store.InsertAnalysisResult(result); err != nil {
			log.Printf("[pipeline] store result: %v", err)
		}

		// Update risk score
		h.updateRiskScore(elderlyID)
	}
}

// processLocation stores location data and checks geofence breaches.
func (h *Handler) processLocation(elderlyID string, payload map[string]interface{}) {
	lat, _ := payload["lat"].(float64)
	lon, _ := payload["lon"].(float64)

	if lat == 0 && lon == 0 {
		return // invalid coordinates
	}

	// Store location
	if err := h.store.InsertLocation(elderlyID, lat, lon); err != nil {
		log.Printf("[pipeline] store location: %v", err)
	}

	// TODO: Check geofence boundaries
	// geofences := h.store.GetGeofences(elderlyID)
	// for _, gf := range geofences {
	//     dist := haversine(lat, lon, gf.Lat, gf.Lon)
	//     if dist > gf.Radius {
	//         log.Printf("[pipeline] GEOFENCE BREACH: elderly=%s near (%.4f, %.4f)",
	//             elderlyID, lat, lon)
	//     }
	// }
}

// processMedStatus updates medication adherence tracking.
func (h *Handler) processMedStatus(pillboxID string, payload map[string]interface{}) {
	// Medication status comes from pillbox auto tier
	// Track taken vs scheduled doses for adherence calculation
	// This runs hourly aggregation in production
	_ = pillboxID
	_ = payload
}

// updateRiskScore recalculates and stores the composite risk score.
func (h *Handler) updateRiskScore(elderlyID string) {
	// Fetch latest components from store
	vitalsDev, _ := h.store.GetLatestVitalsDeviation(elderlyID, 7)
	medAdherence, _ := h.store.GetLatestMedAdherence(elderlyID, 7)
	activityLevel, _ := h.store.GetLatestActivityLevel(elderlyID, 7)
	sleepQuality, _ := h.store.GetLatestSleepQuality(elderlyID, 7)

	input := analyzer.ScoreInput{
		ElderlyID:           elderlyID,
		VitalsDeviation:     vitalsDev,
		MedicationAdherence: medAdherence,
		ActivityLevel:       activityLevel,
		SleepQuality:        sleepQuality,
	}

	result := h.riskCalculator.Calculate(input)

	if err := h.store.InsertRiskScore(&model.RiskScore{
		ElderlyID:           input.ElderlyID,
		CompositeScore:      result.CompositeScore,
		VitalsDeviation:     result.VitalsDeviation,
		MedicationAdherence: result.MedicationAdherence,
		ActivityLevel:       result.ActivityLevel,
		SleepQuality:        result.SleepQuality,
		RecordedAt:          result.CalculatedAt,
	}); err != nil {
		log.Printf("[pipeline] store risk score: %v", err)
		return
	}

	// If score crosses threshold, trigger push notification
	alertLevel := analyzer.ClassifyScore(result.CompositeScore)
	if alertLevel == "P0" || alertLevel == "P1" {
		log.Printf("[pipeline] RISK ALERT: elderly_id=%s score=%d level=%s",
			elderlyID, result.CompositeScore, alertLevel)
		// TODO: Publish alert event to NATS for push-service consumption
	}
}

// haversine calculates distance between two GPS points in meters.
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Earth radius in meters
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
