package analyzer

import (
	"time"

	"eregen.dev/pipeline/internal/model"
)

// ChronicUricAnalyzer evaluates blood uric acid levels.
// Reference: 中国高尿酸血症与痛风诊疗指南 (2019)
// 单位：μmol/L
type ChronicUricAnalyzer struct{}

// NewChronicUricAnalyzer creates a uric acid analyzer instance.
func NewChronicUricAnalyzer() *ChronicUricAnalyzer {
	return &ChronicUricAnalyzer{}
}

// AnalyzeUricAcid classifies a single uric acid value.
//
// 正常参考范围（成人，μmol/L）：
//   男性：143~420
//   女性：89~357（绝经后接近男性）
//
// 高尿酸血症诊断标准：>420 μmol/L（任意性别）
func (a *ChronicUricAnalyzer) AnalyzeUricAcid(value float64) *model.AnalysisResult {
	result := &model.AnalysisResult{
		Metric:    "uric_acid",
		Value:     value,
		Timestamp: time.Now().UTC(),
	}

	if value < 143 {
		result.RiskLevel = model.RiskNormal
		result.Message = "尿酸偏低"
	} else if value <= 420 {
		result.RiskLevel = model.RiskNormal
	} else {
		result.RiskLevel = model.RiskElevated
		result.Message = "尿酸偏高"
	}

	return result
}
