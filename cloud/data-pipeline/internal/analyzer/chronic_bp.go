package analyzer

import (
	"time"

	"eregen.dev/pipeline/internal/model"
)

// ChronicBPAnalyzer evaluates blood pressure readings.
// Reference: 中国高血压防治指南 (2023修订版)
// 单位：mmHg
type ChronicBPAnalyzer struct{}

// NewChronicBPAnalyzer creates a blood pressure analyzer instance.
func NewChronicBPAnalyzer() *ChronicBPAnalyzer {
	return &ChronicBPAnalyzer{}
}

// AnalyzeBP classifies blood pressure from systolic and diastolic values.
//
// 临床标准（mmHg）：
//   低血压：收缩压 <90 或 舒张压 <60
//   正常：90≤收缩压<120 且 60≤舒张压<80
//   正常高值：120≤收缩压<140 或 80≤舒张压<90
//   高血压：收缩压 ≥140 或 舒张压 ≥90
func (a *ChronicBPAnalyzer) AnalyzeBP(systolic, diastolic int) *model.AnalysisResult {
	result := &model.AnalysisResult{
		Metric:    "blood_pressure",
		Value:     float64(systolic),
		Timestamp: time.Now().UTC(),
	}

	// 低血压
	if systolic < 90 || diastolic < 60 {
		result.RiskLevel = model.RiskElevated
		result.Message = "血压偏低"
		return result
	}

	// 高血压
	if systolic > 140 || diastolic > 90 {
		result.RiskLevel = model.RiskElevated
		result.Message = "血压偏高"
		return result
	}

	result.RiskLevel = model.RiskNormal
	return result
}
