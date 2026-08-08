package analyzer

import (
	"time"

	"eregen.dev/pipeline/internal/model"
)

// ChronicRecommendations combines chronic disease metrics and generates
// actionable recommendations for elderly care.
type ChronicRecommendations struct{}

// NewChronicRecommendations creates a recommendations generator instance.
func NewChronicRecommendations() *ChronicRecommendations {
	return &ChronicRecommendations{}
}

// ChronicMetricData carries the three core chronic disease indicators
// needed to produce combined recommendations.
type ChronicMetricData struct {
	ElderlyID   string
	Glucose     *model.AnalysisResult // nil if not available
	Uric        *model.AnalysisResult // nil if not available
	BP          *model.AnalysisResult // nil if not available
	GlucoseMode string                // test mode passed to glucose analyzer
}

// GenerateRecommendations produces a prioritized list of recommendations
// based on combined chronic disease metric analysis.
//
// Priority mapping:
//   - RiskCritical → P0（立即处理）
//   - RiskElevated → P1（重点关注）
//   - RiskNormal   → 不生成建议（或仅健康教育）
func (r *ChronicRecommendations) GenerateRecommendations(data *ChronicMetricData) []model.Recommendation {
	var recs []model.Recommendation
	now := time.Now().UTC()

	if data.Glucose != nil {
		recs = append(recs, r.recommendGlucose(data.Glucose, data.GlucoseMode, now)...)
	}
	if data.Uric != nil {
		recs = append(recs, r.recommendUric(data.Uric, now)...)
	}
	if data.BP != nil {
		recs = append(recs, r.recommendBP(data.BP, now)...)
	}

	// 多指标异常时，增加综合干预建议
	if len(recs) >= 2 {
		recs = append(recs, model.Recommendation{
			ID:          0, // assigned by store
			ElderlyID:   data.ElderlyID,
			Category:    "comprehensive",
			Severity:    "P1",
			Title:       "多指标异常，建议综合干预",
			Description: "血糖/尿酸/血压等多指标同时异常，建议尽快就医进行全面评估。",
			Action:      "尽快就诊内分泌科或老年科",
			CreatedAt:   now,
		})
	}

	return recs
}

// recommendGlucose produces glucose-specific recommendations.
func (r *ChronicRecommendations) recommendGlucose(result *model.AnalysisResult, mode string, now time.Time) []model.Recommendation {
	switch result.RiskLevel {
	case model.RiskCritical:
		return []model.Recommendation{
			{
				ElderlyID:   result.ElderlyID,
				Category:    "glucose",
				Severity:    "P0",
				Title:       "低血糖紧急处理",
				Description: "当前血糖值严重低于正常范围，存在低血糖昏迷风险。",
				Action:      "立即进食含糖食物（糖果、糖水），15分钟后复测；如意识不清请立即拨打120",
				SourceMetric: "glucose",
				Value:      result.Value,
				CreatedAt:  now,
			},
		}
	case model.RiskElevated:
		modeLabel := mode
		if modeLabel == "" {
			modeLabel = "空腹"
		}
		return []model.Recommendation{
			{
				ElderlyID:   result.ElderlyID,
				Category:    "glucose",
				Severity:    "P1",
				Title:       "血糖异常随访建议",
				Description: result.Message + "，建议近期复查空腹及餐后血糖。",
				Action:      "预约内分泌科门诊，复查空腹血糖+糖化血红蛋白（HbA1c）",
				SourceMetric: "glucose",
				Value:      result.Value,
				CreatedAt:  now,
			},
		}
	}
	return nil
}

// recommendUric produces uric acid-specific recommendations.
func (r *ChronicRecommendations) recommendUric(result *model.AnalysisResult, now time.Time) []model.Recommendation {
	if result.RiskLevel != model.RiskElevated {
		return nil
	}
	return []model.Recommendation{
		{
			ElderlyID:   result.ElderlyID,
			Category:    "uric_acid",
			Severity:    "P1",
			Title:       "高尿酸血症饮食干预",
			Description: "尿酸水平偏高，增加痛风及肾结石风险。",
			Action:      "低嘌呤饮食，避免海鲜/啤酒/动物内脏；每日饮水≥2000ml；2周后复查",
			SourceMetric: "uric_acid",
			Value:      result.Value,
			CreatedAt:  now,
		},
	}
}

// recommendBP produces blood pressure-specific recommendations.
func (r *ChronicRecommendations) recommendBP(result *model.AnalysisResult, now time.Time) []model.Recommendation {
	switch result.RiskLevel {
	case model.RiskElevated:
		if result.Message == "血压偏低" {
			return []model.Recommendation{
				{
					ElderlyID:   result.ElderlyID,
					Category:    "blood_pressure",
					Severity:    "P1",
					Title:       "低血压关注",
					Description: "血压偏低，注意防跌倒及体位性低血压。",
					Action:      "避免突然起立，增加水盐摄入；如伴头晕/晕厥请及时就医",
					SourceMetric: "blood_pressure",
					Value:      result.Value,
					CreatedAt:  now,
				},
			}
		}
		return []model.Recommendation{
			{
				ElderlyID:   result.ElderlyID,
				Category:    "blood_pressure",
				Severity:    "P1",
				Title:       "高血压管理建议",
				Description: "血压偏高，长期控制不佳可增加心脑血管事件风险。",
				Action:      "规律监测血压（早晚各一次），低盐饮食，遵医嘱服药；如收缩压≥180或舒张压≥110请立即就医",
				SourceMetric: "blood_pressure",
				Value:      result.Value,
				CreatedAt:  now,
			},
		}
	}
	return nil
}
