package analyzer

import (
	"testing"

	"eregen.dev/pipeline/internal/model"
)

// --- ChronicGlucoseAnalyzer 测试 ---

func TestChronicGlucoseAnalyzer_Fasting_Normal(t *testing.T) {
	a := NewChronicGlucoseAnalyzer()
	result := a.AnalyzeGlucose(5.5, "fasting")
	if result.RiskLevel != model.RiskNormal {
		t.Errorf("RiskLevel = %v, want RiskNormal", result.RiskLevel)
	}
}

func TestChronicGlucoseAnalyzer_Fasting_Hypoglycemia(t *testing.T) {
	a := NewChronicGlucoseAnalyzer()
	result := a.AnalyzeGlucose(3.5, "fasting")
	if result.RiskLevel != model.RiskCritical {
		t.Errorf("RiskLevel = %v, want RiskCritical", result.RiskLevel)
	}
	if result.Message != "低血糖" {
		t.Errorf("Message = %q, want 低血糖", result.Message)
	}
}

func TestChronicGlucoseAnalyzer_Fasting_High(t *testing.T) {
	a := NewChronicGlucoseAnalyzer()
	result := a.AnalyzeGlucose(8.5, "fasting")
	if result.RiskLevel != model.RiskElevated {
		t.Errorf("RiskLevel = %v, want RiskElevated", result.RiskLevel)
	}
}

func TestChronicGlucoseAnalyzer_Postprandial_Normal(t *testing.T) {
	a := NewChronicGlucoseAnalyzer()
	result := a.AnalyzeGlucose(6.5, "postprandial")
	if result.RiskLevel != model.RiskNormal {
		t.Errorf("RiskLevel = %v, want RiskNormal", result.RiskLevel)
	}
}

func TestChronicGlucoseAnalyzer_Postprandial_Borderline(t *testing.T) {
	a := NewChronicGlucoseAnalyzer()
	// 7.8 < value <= 10.0
	result := a.AnalyzeGlucose(8.5, "postprandial")
	if result.RiskLevel != model.RiskElevated {
		t.Errorf("RiskLevel = %v, want RiskElevated", result.RiskLevel)
	}
	if result.Message != "血糖偏高" {
		t.Errorf("Message = %q, want 血糖偏高", result.Message)
	}
}

func TestChronicGlucoseAnalyzer_Postprandial_High(t *testing.T) {
	a := NewChronicGlucoseAnalyzer()
	// value > 10.0
	result := a.AnalyzeGlucose(12.0, "postprandial")
	if result.RiskLevel != model.RiskElevated {
		t.Errorf("RiskLevel = %v, want RiskElevated", result.RiskLevel)
	}
	if result.Message != "餐后血糖显著偏高" {
		t.Errorf("Message = %q, want 餐后血糖显著偏高", result.Message)
	}
}

func TestChronicGlucoseAnalyzer_Random_Elevated(t *testing.T) {
	a := NewChronicGlucoseAnalyzer()
	result := a.AnalyzeGlucose(9.0, "random")
	if result.RiskLevel != model.RiskElevated {
		t.Errorf("RiskLevel = %v, want RiskElevated", result.RiskLevel)
	}
}

func TestChronicGlucoseAnalyzer_EmptyMode_FallsBackToFasting(t *testing.T) {
	a := NewChronicGlucoseAnalyzer()
	// 空 mode 应等同于 fasting
	result := a.AnalyzeGlucose(6.0, "")
	if result.RiskLevel != model.RiskNormal {
		t.Errorf("RiskLevel = %v, want RiskNormal (empty mode defaults to fasting)", result.RiskLevel)
	}
}

// --- ChronicUricAnalyzer 测试 ---

func TestChronicUricAnalyzer_Normal(t *testing.T) {
	a := NewChronicUricAnalyzer()
	result := a.AnalyzeUricAcid(350)
	if result.RiskLevel != model.RiskNormal {
		t.Errorf("RiskLevel = %v, want RiskNormal", result.RiskLevel)
	}
}

func TestChronicUricAnalyzer_High(t *testing.T) {
	a := NewChronicUricAnalyzer()
	result := a.AnalyzeUricAcid(500)
	if result.RiskLevel != model.RiskElevated {
		t.Errorf("RiskLevel = %v, want RiskElevated", result.RiskLevel)
	}
	if result.Message != "尿酸偏高" {
		t.Errorf("Message = %q, want 尿酸偏高", result.Message)
	}
}

func TestChronicUricAnalyzer_Low(t *testing.T) {
	a := NewChronicUricAnalyzer()
	result := a.AnalyzeUricAcid(100)
	if result.RiskLevel != model.RiskNormal {
		t.Errorf("RiskLevel = %v, want RiskNormal", result.RiskLevel)
	}
}

func TestChronicUricAnalyzer_Boundary_Normal(t *testing.T) {
	a := NewChronicUricAnalyzer()
	// 边界值 143 和 420 应为 normal
	for _, v := range []float64{143, 420} {
		result := a.AnalyzeUricAcid(v)
		if result.RiskLevel != model.RiskNormal {
			t.Errorf("uric=%v RiskLevel = %v, want RiskNormal", v, result.RiskLevel)
		}
	}
}

func TestChronicUricAnalyzer_Boundary_High(t *testing.T) {
	a := NewChronicUricAnalyzer()
	// >420 应为 elevated
	result := a.AnalyzeUricAcid(421)
	if result.RiskLevel != model.RiskElevated {
		t.Errorf("RiskLevel = %v, want RiskElevated", result.RiskLevel)
	}
}

// --- ChronicBPAnalyzer 测试 ---

func TestChronicBPAnalyzer_Normal(t *testing.T) {
	a := NewChronicBPAnalyzer()
	result := a.AnalyzeBP(120, 80)
	if result.RiskLevel != model.RiskNormal {
		t.Errorf("RiskLevel = %v, want RiskNormal", result.RiskLevel)
	}
}

func TestChronicBPAnalyzer_High(t *testing.T) {
	a := NewChronicBPAnalyzer()
	result := a.AnalyzeBP(150, 95)
	if result.RiskLevel != model.RiskElevated {
		t.Errorf("RiskLevel = %v, want RiskElevated", result.RiskLevel)
	}
}

func TestChronicBPAnalyzer_Low(t *testing.T) {
	a := NewChronicBPAnalyzer()
	result := a.AnalyzeBP(85, 55)
	if result.RiskLevel != model.RiskElevated {
		t.Errorf("RiskLevel = %v, want RiskElevated (低血压也标记 elevated)", result.RiskLevel)
	}
}

func TestChronicBPAnalyzer_SysHigh_DiaNormal(t *testing.T) {
	a := NewChronicBPAnalyzer()
	result := a.AnalyzeBP(145, 78)
	if result.RiskLevel != model.RiskElevated {
		t.Errorf("RiskLevel = %v, want RiskElevated", result.RiskLevel)
	}
}

func TestChronicBPAnalyzer_SysNormal_DiaHigh(t *testing.T) {
	a := NewChronicBPAnalyzer()
	result := a.AnalyzeBP(118, 92)
	if result.RiskLevel != model.RiskElevated {
		t.Errorf("RiskLevel = %v, want RiskElevated", result.RiskLevel)
	}
}

func TestChronicBPAnalyzer_Boundary_Normal(t *testing.T) {
	a := NewChronicBPAnalyzer()
	// 边界值 90/60 和 140/90 应为 normal
	tests := [][2]int{{90, 60}, {140, 90}, {119, 79}, {90, 80}}
	for _, tc := range tests {
		result := a.AnalyzeBP(tc[0], tc[1])
		if result.RiskLevel != model.RiskNormal {
			t.Errorf("BP=(%d,%d) RiskLevel = %v, want RiskNormal", tc[0], tc[1], result.RiskLevel)
		}
	}
}

// --- ChronicRecommendations 测试 ---

func TestChronicRecommendations_NoMetrics_NoRecs(t *testing.T) {
	a := NewChronicRecommendations()
	recs := a.GenerateRecommendations(&ChronicMetricData{ElderlyID: "e1"})
	if len(recs) != 0 {
		t.Errorf("len(recs) = %d, want 0", len(recs))
	}
}

func TestChronicRecommendations_GlucoseCritical(t *testing.T) {
	a := NewChronicRecommendations()
	data := &ChronicMetricData{
		ElderlyID: "e1",
		Glucose: &model.AnalysisResult{
			ElderlyID: "e1",
			Metric:    "glucose",
			Value:     3.2,
			RiskLevel: model.RiskCritical,
			Message:   "低血糖",
		},
		GlucoseMode: "fasting",
	}
	recs := a.GenerateRecommendations(data)
	if len(recs) != 1 {
		t.Fatalf("len(recs) = %d, want 1", len(recs))
	}
	if recs[0].Severity != "P0" {
		t.Errorf("Severity = %q, want P0", recs[0].Severity)
	}
	if recs[0].Category != "glucose" {
		t.Errorf("Category = %q, want glucose", recs[0].Category)
	}
}

func TestChronicRecommendations_GlucoseElevated(t *testing.T) {
	a := NewChronicRecommendations()
	data := &ChronicMetricData{
		ElderlyID: "e1",
		Glucose: &model.AnalysisResult{
			ElderlyID: "e1",
			Metric:    "glucose",
			Value:     8.5,
			RiskLevel: model.RiskElevated,
			Message:   "空腹血糖偏高",
		},
		GlucoseMode: "fasting",
	}
	recs := a.GenerateRecommendations(data)
	if len(recs) != 1 {
		t.Fatalf("len(recs) = %d, want 1", len(recs))
	}
	if recs[0].Severity != "P1" {
		t.Errorf("Severity = %q, want P1", recs[0].Severity)
	}
}

func TestChronicRecommendations_UricElevated(t *testing.T) {
	a := NewChronicRecommendations()
	data := &ChronicMetricData{
		ElderlyID: "e1",
		Uric: &model.AnalysisResult{
			ElderlyID: "e1",
			Metric:    "uric_acid",
			Value:     500,
			RiskLevel: model.RiskElevated,
			Message:   "尿酸偏高",
		},
	}
	recs := a.GenerateRecommendations(data)
	if len(recs) != 1 {
		t.Fatalf("len(recs) = %d, want 1", len(recs))
	}
	if recs[0].Category != "uric_acid" {
		t.Errorf("Category = %q, want uric_acid", recs[0].Category)
	}
	if recs[0].Severity != "P1" {
		t.Errorf("Severity = %q, want P1", recs[0].Severity)
	}
}

func TestChronicRecommendations_BPElevated(t *testing.T) {
	a := NewChronicRecommendations()
	data := &ChronicMetricData{
		ElderlyID: "e1",
		BP: &model.AnalysisResult{
			ElderlyID: "e1",
			Metric:    "blood_pressure",
			Value:     150,
			RiskLevel: model.RiskElevated,
			Message:   "血压偏高",
		},
	}
	recs := a.GenerateRecommendations(data)
	if len(recs) != 1 {
		t.Fatalf("len(recs) = %d, want 1", len(recs))
	}
	if recs[0].Category != "blood_pressure" {
		t.Errorf("Category = %q, want blood_pressure", recs[0].Category)
	}
}

func TestChronicRecommendations_BPLow(t *testing.T) {
	a := NewChronicRecommendations()
	data := &ChronicMetricData{
		ElderlyID: "e1",
		BP: &model.AnalysisResult{
			ElderlyID: "e1",
			Metric:    "blood_pressure",
			Value:     85,
			RiskLevel: model.RiskElevated,
			Message:   "血压偏低",
		},
	}
	recs := a.GenerateRecommendations(data)
	if len(recs) != 1 {
		t.Fatalf("len(recs) = %d, want 1", len(recs))
	}
	if recs[0].Category != "blood_pressure" {
		t.Errorf("Category = %q, want blood_pressure", recs[0].Category)
	}
}

func TestChronicRecommendations_NormalMetrics_NoRecs(t *testing.T) {
	a := NewChronicRecommendations()
	data := &ChronicMetricData{
		ElderlyID: "e1",
		Glucose: &model.AnalysisResult{
			ElderlyID: "e1",
			Metric:    "glucose",
			Value:     5.5,
			RiskLevel: model.RiskNormal,
		},
		Uric: &model.AnalysisResult{
			ElderlyID: "e1",
			Metric:    "uric_acid",
			Value:     350,
			RiskLevel: model.RiskNormal,
		},
		BP: &model.AnalysisResult{
			ElderlyID: "e1",
			Metric:    "blood_pressure",
			Value:     120,
			RiskLevel: model.RiskNormal,
		},
	}
	recs := a.GenerateRecommendations(data)
	if len(recs) != 0 {
		t.Errorf("len(recs) = %d, want 0", len(recs))
	}
}

func TestChronicRecommendations_MultiAbnormal_Comprehensive(t *testing.T) {
	a := NewChronicRecommendations()
	data := &ChronicMetricData{
		ElderlyID: "e1",
		Glucose: &model.AnalysisResult{
			ElderlyID: "e1",
			Metric:    "glucose",
			Value:     9.0,
			RiskLevel: model.RiskElevated,
			Message:   "血糖偏高",
		},
		Uric: &model.AnalysisResult{
			ElderlyID: "e1",
			Metric:    "uric_acid",
			Value:     500,
			RiskLevel: model.RiskElevated,
			Message:   "尿酸偏高",
		},
		BP: &model.AnalysisResult{
			ElderlyID: "e1",
			Metric:    "blood_pressure",
			Value:     130,
			RiskLevel: model.RiskElevated,
			Message:   "血压偏高",
		},
	}
	recs := a.GenerateRecommendations(data)
	// 3条单指标 + 1条综合建议
	if len(recs) != 4 {
		t.Errorf("len(recs) = %d, want 4", len(recs))
	}
	// 检查综合建议
	var foundComprehensive bool
	for _, rec := range recs {
		if rec.Category == "comprehensive" {
			foundComprehensive = true
			break
		}
	}
	if !foundComprehensive {
		t.Error("missing comprehensive recommendation when multiple metrics abnormal")
	}
}
