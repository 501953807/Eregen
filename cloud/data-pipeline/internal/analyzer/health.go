package analyzer

import (
	"log"
	"math"
	"time"

	"eregen.dev/pipeline/internal/model"
)

// HealthAnalyzer evaluates incoming health metrics against baselines and thresholds.
type HealthAnalyzer struct {
	baselineDays int
}

// NewHealthAnalyzer creates an analyzer with configurable baseline window.
func NewHealthAnalyzer(baselineDays int) *HealthAnalyzer {
	return &HealthAnalyzer{baselineDays: baselineDays}
}

// Analyze checks a single health metric and returns the analysis result.
func (a *HealthAnalyzer) Analyze(elderlyID, metric string, value float64, baseline float64) *model.AnalysisResult {
	result := &model.AnalysisResult{
		ElderlyID: elderlyID,
		Metric:    metric,
		Value:     value,
		Baseline:  baseline,
		Timestamp: time.Now().UTC(),
	}

	if baseline <= 0 {
		result.Deviation = 0
		result.RiskLevel = classifyAbsolute(metric, value)
		return result
	}

	deviation := math.Abs(value-baseline) / baseline * 100
	result.Deviation = math.Round(deviation*100) / 100

	if deviation > 50 || classifyAbsolute(metric, value) == model.RiskCritical {
		result.RiskLevel = model.RiskCritical
	} else if deviation > 30 || classifyAbsolute(metric, value) == model.RiskElevated {
		result.RiskLevel = model.RiskElevated
	} else {
		result.RiskLevel = model.RiskNormal
	}

	return result
}

// classifyAbsolute uses clinical thresholds to classify a metric value.
func classifyAbsolute(metric string, value float64) model.RiskLevel {
	switch metric {
	case "heart_rate":
		if value < 40 || value > 120 {
			return model.RiskCritical
		}
		if value < 50 || value > 110 {
			return model.RiskElevated
		}
		return model.RiskNormal

	case "spo2":
		if value < 90 {
			return model.RiskCritical
		}
		if value < 95 {
			return model.RiskElevated
		}
		return model.RiskNormal

	case "bp_systolic":
		if value > 180 || value < 85 {
			return model.RiskCritical
		}
		if value > 160 || value < 90 {
			return model.RiskElevated
		}
		return model.RiskNormal

	case "bp_diastolic":
		if value > 110 || value < 55 {
			return model.RiskCritical
		}
		if value > 100 || value < 60 {
			return model.RiskElevated
		}
		return model.RiskNormal

	case "temperature":
		if value > 39 || value < 35 {
			return model.RiskCritical
		}
		if value > 38 || value < 35.5 {
			return model.RiskElevated
		}
		return model.RiskNormal

	case "steps":
		return model.RiskNormal

	default:
		return model.RiskNormal
	}
}

// AnalyzeBatch processes multiple metrics at once.
func (a *HealthAnalyzer) AnalyzeBatch(elderlyID string, metrics map[string]float64, baselines map[string]float64) []*model.AnalysisResult {
	var results []*model.AnalysisResult
	for metric, value := range metrics {
		baseline := baselines[metric]
		result := a.Analyze(elderlyID, metric, value, baseline)
		if result.RiskLevel != model.RiskNormal {
			results = append(results, result)
			log.Printf("[analyzer] %s: %s=%.1f baseline=%.1f risk=%s",
				elderlyID, metric, value, baseline, result.RiskLevel)
		}
	}
	return results
}
