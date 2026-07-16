package analyzer

import (
	"log"
	"math"
	"time"

	"eregen.dev/pipeline/internal/model"
)

// MedicationAnalyzer tracks adherence to medication schedules.
type MedicationAnalyzer struct{}

// NewMedicationAnalyzer creates a new medication adherence analyzer.
func NewMedicationAnalyzer() *MedicationAnalyzer {
	return &MedicationAnalyzer{}
}

// AnalyzePeriod calculates adherence over a date range.
func (a *MedicationAnalyzer) AnalyzePeriod(elderlyID string, periodStart, periodEnd time.Time,
	scheduledDoses, takenDoses int, missedMeds []model.MissedMed,
) *model.MedicationAdherence {
	adherence := &model.MedicationAdherence{
		ElderlyID:       elderlyID,
		PeriodStart:     periodStart,
		PeriodEnd:       periodEnd,
		ScheduledDoses:  scheduledDoses,
		TakenDoses:      takenDoses,
		CreatedAt:       time.Now().UTC(),
		MissedMedications: missedMeds,
	}

	if scheduledDoses > 0 {
		adherence.AdherenceRate = math.Round(float64(takenDoses)/float64(scheduledDoses)*10000) / 100
	}

	if adherence.AdherenceRate < 80 {
		log.Printf("[med-analyzer] LOW ADHERENCE: %s %.2f%% (%d/%d)",
			elderlyID, adherence.AdherenceRate, takenDoses, scheduledDoses)
	}

	consecutiveMisses := a.countConsecutiveMisses(missedMeds)
	if consecutiveMisses >= 2 {
		log.Printf("[med-analyzer] CONSECUTIVE MISSES: %s %d in a row",
			elderlyID, consecutiveMisses)
	}

	return adherence
}

// countConsecutiveMisses finds the longest streak of consecutive missed doses.
func (a *MedicationAnalyzer) countConsecutiveMisses(missed []model.MissedMed) int {
	if len(missed) == 0 {
		return 0
	}

	maxStreak := 1
	currentStreak := 1
	for i := 1; i < len(missed); i++ {
		diff := missed[i].Scheduled.Sub(missed[i-1].Scheduled).Hours()
		if diff <= 4 {
			currentStreak++
			if currentStreak > maxStreak {
				maxStreak = currentStreak
			}
		} else {
			currentStreak = 1
		}
	}
	return maxStreak
}

// CalculateDailyAdherence computes today's adherence rate.
func (a *MedicationAnalyzer) CalculateDailyAdherence(scheduled, taken int) float64 {
	if scheduled == 0 {
		return 100
	}
	return math.Round(float64(taken)/float64(scheduled)*10000) / 100
}
