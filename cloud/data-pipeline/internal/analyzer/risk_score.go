package analyzer

import (
	"log"
	"math"
	"time"
)

// RiskScoreCalculator computes a composite health risk score (0-100).
type RiskScoreCalculator struct {
	vitalsWeight   float64
	medWeight      float64
	activityWeight float64
	sleepWeight    float64
}

// NewRiskScoreCalculator creates a calculator with configured weights.
func NewRiskScoreCalculator(vitalsW, medW, activityW, sleepW float64) *RiskScoreCalculator {
	return &RiskScoreCalculator{
		vitalsWeight:   vitalsW,
		medWeight:      medW,
		activityWeight: activityW,
		sleepWeight:    sleepW,
	}
}

// ScoreInput holds the inputs for risk score calculation.
type ScoreInput struct {
	ElderlyID           string
	VitalsDeviation     float64 // 0-100, higher = more deviation from normal
	MedicationAdherence float64 // 0-100, higher = better adherence
	ActivityLevel       float64 // 0-100, lower = less active = higher risk
	SleepQuality        float64 // 0-100, lower = worse sleep = higher risk
}

// ScoreResult holds the computed risk score and its components.
type ScoreResult struct {
	ElderlyID           string
	CompositeScore      int
	VitalsDeviation     float64
	MedicationAdherence float64
	ActivityLevel       float64
	SleepQuality        float64
	CalculatedAt        time.Time
}

// Calculate computes the composite risk score.
func (c *RiskScoreCalculator) Calculate(input ScoreInput) *ScoreResult {
	vitalsFactor := input.VitalsDeviation
	medFactor := 100 - input.MedicationAdherence
	activityFactor := 100 - input.ActivityLevel
	sleepFactor := 100 - input.SleepQuality

	composite := c.vitalsWeight*vitalsFactor +
		c.medWeight*medFactor +
		c.activityWeight*activityFactor +
		c.sleepWeight*sleepFactor

	composite = math.Max(0, math.Min(100, composite))
	score := int(math.Round(composite))

	result := &ScoreResult{
		ElderlyID:           input.ElderlyID,
		CompositeScore:      score,
		VitalsDeviation:     math.Round(vitalsFactor*100) / 100,
		MedicationAdherence: math.Round(medFactor*100) / 100,
		ActivityLevel:       math.Round(activityFactor*100) / 100,
		SleepQuality:        math.Round(sleepFactor*100) / 100,
		CalculatedAt:        time.Now().UTC(),
	}

	if score > 80 {
		log.Printf("[risk] CRITICAL score=%d elderly_id=%s", score, input.ElderlyID)
	} else if score > 60 {
		log.Printf("[risk] ELEVATED score=%d elderly_id=%s", score, input.ElderlyID)
	}

	return result
}

// Classify returns the alert level for a composite score.
func ClassifyScore(score int) string {
	switch {
	case score > 80:
		return "P0"
	case score > 60:
		return "P1"
	default:
		return "P2"
	}
}
