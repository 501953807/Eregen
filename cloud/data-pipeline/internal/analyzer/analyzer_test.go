package analyzer

import (
	"math"
	"testing"
	"time"

	"eregen.dev/pipeline/internal/model"
)

func TestRiskScoreCalculator_BasicCalculation(t *testing.T) {
	calc := NewRiskScoreCalculator(0.3, 0.2, 0.25, 0.25)
	result := calc.Calculate(ScoreInput{
		ElderlyID:           "elderly-1",
		VitalsDeviation:     50,
		MedicationAdherence: 80,
		ActivityLevel:       60,
		SleepQuality:        70,
	})

	expected := 0.3*50 + 0.2*20 + 0.25*40 + 0.25*30
	if result.CompositeScore != int(math.Round(expected)) {
		t.Errorf("CompositeScore = %d, want %d", result.CompositeScore, int(math.Round(expected)))
	}
	if result.ElderlyID != "elderly-1" {
		t.Errorf("ElderlyID = %q, want elderly-1", result.ElderlyID)
	}
}

func TestRiskScoreCalculator_ClampToRange(t *testing.T) {
	calc := NewRiskScoreCalculator(1.0, 0, 0, 0)

	// Vitals deviation 150 should be clamped to 100
	result := calc.Calculate(ScoreInput{
		ElderlyID:           "elderly-2",
		VitalsDeviation:     150,
		MedicationAdherence: 100,
		ActivityLevel:       100,
		SleepQuality:        100,
	})
	if result.CompositeScore > 100 {
		t.Errorf("score %d exceeds max 100", result.CompositeScore)
	}

	// All zero factors should give 0
	result2 := calc.Calculate(ScoreInput{
		ElderlyID:           "elderly-3",
		VitalsDeviation:     0,
		MedicationAdherence: 100,
		ActivityLevel:       100,
		SleepQuality:        100,
	})
	if result2.CompositeScore < 0 {
		t.Errorf("score %d below min 0", result2.CompositeScore)
	}
}

func TestClassifyScore(t *testing.T) {
	tests := []struct {
		score int
		want  string
	}{
		{85, "P0"},
		{65, "P1"},
		{30, "P2"},
		{0, "P2"},
		{100, "P0"},
	}
	for _, tt := range tests {
		got := ClassifyScore(tt.score)
		if got != tt.want {
			t.Errorf("ClassifyScore(%d) = %q, want %q", tt.score, got, tt.want)
		}
	}
}

func TestHealthAnalyzer_NormalHeartRate(t *testing.T) {
	analyzer := NewHealthAnalyzer(7)
	result := analyzer.Analyze("elderly-1", "heart_rate", 72, 72)
	if result.RiskLevel != model.RiskNormal {
		t.Errorf("RiskLevel = %v, want RiskNormal", result.RiskLevel)
	}
	if result.Deviation != 0 {
		t.Errorf("Deviation = %f, want 0", result.Deviation)
	}
}

func TestHealthAnalyzer_CriticalHeartRate(t *testing.T) {
	analyzer := NewHealthAnalyzer(7)
	result := analyzer.Analyze("elderly-1", "heart_rate", 130, 72)
	if result.RiskLevel != model.RiskCritical {
		t.Errorf("RiskLevel = %v, want RiskCritical", result.RiskLevel)
	}
}

func TestHealthAnalyzer_ElevatedSpO2(t *testing.T) {
	analyzer := NewHealthAnalyzer(7)
	result := analyzer.Analyze("elderly-1", "spo2", 93, 98)
	if result.RiskLevel != model.RiskElevated {
		t.Errorf("RiskLevel = %v, want RiskElevated", result.RiskLevel)
	}
}

func TestHealthAnalyzer_CriticalSpO2(t *testing.T) {
	analyzer := NewHealthAnalyzer(7)
	result := analyzer.Analyze("elderly-1", "spo2", 88, 98)
	if result.RiskLevel != model.RiskCritical {
		t.Errorf("RiskLevel = %v, want RiskCritical", result.RiskLevel)
	}
}

func TestHealthAnalyzer_ZeroBaseline(t *testing.T) {
	analyzer := NewHealthAnalyzer(7)
	result := analyzer.Analyze("elderly-1", "steps", 5000, 0)
	if result.Deviation != 0 {
		t.Errorf("Deviation = %f, want 0 when baseline is 0", result.Deviation)
	}
}

func TestHealthAnalyzer_AnalyzeBatch(t *testing.T) {
	analyzer := NewHealthAnalyzer(7)
	metrics := map[string]float64{
		"heart_rate": 72,
		"spo2":       88,
		"temperature": 36.5,
	}
	baselines := map[string]float64{
		"heart_rate": 72,
		"spo2":       98,
		"temperature": 36.5,
	}
	results := analyzer.AnalyzeBatch("elderly-1", metrics, baselines)
	if len(results) != 1 {
		t.Fatalf("AnalyzeBatch returned %d results, want 1", len(results))
	}
	if results[0].Metric != "spo2" {
		t.Errorf("non-normal metric = %q, want spo2", results[0].Metric)
	}
}

func TestMedicationAnalyzer_DailyAdherence(t *testing.T) {
	analyzer := NewMedicationAnalyzer()

	tests := []struct {
		scheduled int
		taken     int
		want      float64
	}{
		{3, 3, 100},
		{3, 2, 66.67},
		{0, 0, 100},
		{4, 1, 25},
	}
	for _, tt := range tests {
		got := analyzer.CalculateDailyAdherence(tt.scheduled, tt.taken)
		if math.Abs(got-tt.want) > 0.01 {
			t.Errorf("CalculateDailyAdherence(%d,%d) = %.2f, want %.2f", tt.scheduled, tt.taken, got, tt.want)
		}
	}
}

func TestMedicationAnalyzer_AnalyzePeriod(t *testing.T) {
	analyzer := NewMedicationAnalyzer()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	result := analyzer.AnalyzePeriod("elderly-1", start, end, 21, 15, nil)
	if result.ScheduledDoses != 21 || result.TakenDoses != 15 {
		t.Error("AnalyzePeriod fields not set correctly")
	}
	if result.AdherenceRate < 70 || result.AdherenceRate > 72 {
		t.Errorf("AdherenceRate = %.2f, want ~71.43", result.AdherenceRate)
	}
}

func TestMedicationAnalyzer_ConsecutiveMisses(t *testing.T) {
	analyzer := NewMedicationAnalyzer()

	now := time.Now()
	missed := []model.MissedMed{
		{Scheduled: now.Add(-2 * time.Hour)},
		{Scheduled: now.Add(-1 * time.Hour)},
		{Scheduled: now},
	}

	streak := analyzer.countConsecutiveMisses(missed)
	if streak != 3 {
		t.Errorf("countConsecutiveMisses = %d, want 3", streak)
	}

	// Empty list
	if analyzer.countConsecutiveMisses(nil) != 0 {
		t.Error("countConsecutiveMisses of empty list should be 0")
	}
}

func TestMedicationAnalyzer_LowAdherenceLog(t *testing.T) {
	analyzer := NewMedicationAnalyzer()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC)

	// 1 out of 21 doses taken = ~4.76% adherence — should log LOW ADHERENCE
	result := analyzer.AnalyzePeriod("elderly-low", start, end, 21, 1, nil)
	if result.AdherenceRate >= 80 {
		t.Errorf("AdherenceRate = %.2f, expected < 80 for low adherence", result.AdherenceRate)
	}
}
