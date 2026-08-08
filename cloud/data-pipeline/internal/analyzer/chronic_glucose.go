package analyzer

import (
	"time"

	"eregen.dev/pipeline/internal/model"
)

// ChronicGlucoseAnalyzer evaluates blood glucose readings against clinical thresholds.
// Reference: 中国2型糖尿病防治指南 (2020版)
type ChronicGlucoseAnalyzer struct{}

// NewChronicGlucoseAnalyzer creates a glucose analyzer instance.
func NewChronicGlucoseAnalyzer() *ChronicGlucoseAnalyzer {
	return &ChronicGlucoseAnalyzer{}
}

// AnalyzeGlucose classifies a single glucose value based on test context.
//
//   testMode values:
//     "fasting"    — 空腹血糖 (FPG)
//     "postprandial" — 餐后2小时血糖 (2hPG)
//     "random"     — 随机血糖
//     ""           — 默认按空腹处理
//
// 诊断阈值（mmol/L）：
//   - 空腹：低血糖 <3.9，正常 3.9~7.0，空腹受损 7.0~7.8，糖尿病 ≥7.8
//   - 餐后2h：正常 <7.8，糖耐量受损 7.8~11.1，糖尿病 ≥11.1
//   - 随机：异常 ≥11.1
func (a *ChronicGlucoseAnalyzer) AnalyzeGlucose(value float64, testMode string) *model.AnalysisResult {
	result := &model.AnalysisResult{
		Metric:    "glucose",
		Value:     value,
		Timestamp: time.Now().UTC(),
	}

	mode := testMode
	if mode == "" {
		mode = "fasting"
	}

	switch mode {
	case "fasting":
		// 空腹血糖
		if value < 3.9 {
			result.RiskLevel = model.RiskCritical
			result.Message = "低血糖"
		} else if value <= 7.0 {
			result.RiskLevel = model.RiskNormal
		} else if value <= 7.8 {
			result.RiskLevel = model.RiskElevated
			result.Message = "空腹血糖偏高（空腹受损）"
		} else {
			result.RiskLevel = model.RiskElevated
			result.Message = "空腹血糖偏高"
		}

	case "postprandial":
		// 餐后2小时血糖
		if value < 3.9 {
			result.RiskLevel = model.RiskCritical
			result.Message = "低血糖"
		} else if value <= 7.8 {
			result.RiskLevel = model.RiskNormal
		} else if value <= 10.0 {
			result.RiskLevel = model.RiskElevated
			result.Message = "血糖偏高"
		} else {
			result.RiskLevel = model.RiskElevated
			result.Message = "餐后血糖显著偏高"
		}

	default:
		// 随机血糖或其他模式
		if value < 3.9 {
			result.RiskLevel = model.RiskCritical
			result.Message = "低血糖"
		} else if value <= 7.8 {
			result.RiskLevel = model.RiskNormal
		} else if value <= 10.0 {
			result.RiskLevel = model.RiskElevated
			result.Message = "血糖偏高"
		} else {
			result.RiskLevel = model.RiskElevated
			result.Message = "血糖显著偏高"
		}
	}

	return result
}
