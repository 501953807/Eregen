package service

import (
	"context"
	"testing"
)

func TestCheckMedications_DetectsInteractions(t *testing.T) {
	checker := NewMedicationInteractionChecker(nil)
	ctx := context.Background()

	interactions, err := checker.CheckMedications(ctx, []string{"warfarin", "aspirin"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(interactions) != 1 {
		t.Fatalf("expected 1 interaction, got %d", len(interactions))
	}
	if interactions[0].Severity != "severe" {
		t.Errorf("expected severe severity, got %s", interactions[0].Severity)
	}
}

func TestCheckMedications_NoInteractions(t *testing.T) {
	checker := NewMedicationInteractionChecker(nil)
	ctx := context.Background()

	interactions, err := checker.CheckMedications(ctx, []string{"metformin", "amlodipine"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(interactions) != 0 {
		t.Errorf("expected no interactions, got %d", len(interactions))
	}
}

func TestCheckMedications_SingleMedication(t *testing.T) {
	checker := NewMedicationInteractionChecker(nil)
	ctx := context.Background()

	interactions, err := checker.CheckMedications(ctx, []string{"warfarin"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if interactions != nil {
		t.Errorf("expected nil for single medication, got %v", interactions)
	}
}

func TestCheckConditions_DetectsInteractions(t *testing.T) {
	checker := NewMedicationInteractionChecker(nil)
	ctx := context.Background()

	interactions, err := checker.CheckConditions(ctx, []string{"ibuprofen"}, []string{"chronic_kidney_disease"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(interactions) != 1 {
		t.Fatalf("expected 1 interaction, got %d", len(interactions))
	}
	if interactions[0].Condition != "chronic_kidney_disease" {
		t.Errorf("expected condition chronic_kidney_disease, got %s", interactions[0].Condition)
	}
}

func TestCheckConditions_NoInteractions(t *testing.T) {
	checker := NewMedicationInteractionChecker(nil)
	ctx := context.Background()

	interactions, err := checker.CheckConditions(ctx, []string{"metformin"}, []string{"diabetes"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(interactions) != 0 {
		t.Errorf("expected no interactions, got %d", len(interactions))
	}
}

func TestAddInteraction(t *testing.T) {
	checker := NewMedicationInteractionChecker(nil)
	initialCount := checker.GetInteractionCount()

	checker.AddInteraction(DrugInteraction{
		DrugA:        "test_a",
		DrugB:        "test_b",
		Severity:     "mild",
		Description:  "Test interaction",
		Recommendation: "Monitor",
	})

	if checker.GetInteractionCount() != initialCount+1 {
		t.Errorf("expected count to increase by 1")
	}
}

func TestAddConditionInteraction(t *testing.T) {
	checker := NewMedicationInteractionChecker(nil)
	initialCount := checker.GetInteractionCount()

	checker.AddConditionInteraction(ConditionInteraction{
		Drug:           "test_drug",
		Condition:      "test_condition",
		Severity:       "moderate",
		Description:    "Test condition interaction",
		Recommendation: "Avoid",
	})

	if checker.GetInteractionCount() != initialCount+1 {
		t.Errorf("expected count to increase by 1")
	}
}

func TestCheckMedications_CaseInsensitive(t *testing.T) {
	checker := NewMedicationInteractionChecker(nil)
	ctx := context.Background()

	interactions, err := checker.CheckMedications(ctx, []string{"WARFARIN", "ASPIRIN"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(interactions) != 1 {
		t.Fatalf("expected 1 interaction (case-insensitive), got %d", len(interactions))
	}
}
