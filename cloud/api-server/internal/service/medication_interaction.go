package service

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"eregen.dev/api-server/internal/model"

	"go.uber.org/zap"
)

// DrugInteraction represents a known drug-drug or drug-condition interaction.
type DrugInteraction struct {
	DrugA          string `json:"drug_a"`
	DrugB          string `json:"drug_b"`
	Severity       string `json:"severity"` // mild, moderate, severe, contraindicated
	Description    string `json:"description"`
	Recommendation string `json:"recommendation"`
}

// ConditionInteraction represents a drug-condition interaction.
type ConditionInteraction struct {
	Drug           string `json:"drug"`
	Condition      string `json:"condition"`
	Severity       string `json:"severity"`
	Description    string `json:"description"`
	Recommendation string `json:"recommendation"`
}

// MedicationInteractionChecker detects interactions between medications and conditions.
type MedicationInteractionChecker struct {
	mu                    sync.RWMutex
	interactions          []DrugInteraction
	conditionInteractions []ConditionInteraction
	log                   *zap.Logger
}

// NewMedicationInteractionChecker creates a new checker with built-in knowledge base.
func NewMedicationInteractionChecker(log *zap.Logger) *MedicationInteractionChecker {
	c := &MedicationInteractionChecker{log: log}
	c.initDefaultKnowledgeBase()
	return c
}

// initDefaultKnowledgeBase populates the checker with common elderly medication interactions.
func (c *MedicationInteractionChecker) initDefaultKnowledgeBase() {
	c.interactions = []DrugInteraction{
		{DrugA: "warfarin", DrugB: "aspirin", Severity: "severe", Description: "Increased bleeding risk when combining anticoagulant with antiplatelet agent", Recommendation: "Monitor INR closely; consider gastroprotection if combination necessary"},
		{DrugA: "metformin", DrugB: "contrast_dye", Severity: "severe", Description: "Risk of lactic acidosis with iodinated contrast media", Recommendation: "Hold metformin 48h before and after contrast procedure"},
		{DrugA: "lisinopril", DrugB: "potassium_chloride", Severity: "moderate", Description: "Hyperkalemia risk with ACE inhibitor plus potassium supplement", Recommendation: "Monitor serum potassium regularly"},
		{DrugA: "digoxin", DrugB: "amiodarone", Severity: "severe", Description: "Amiodarone increases digoxin levels by 70-100%, risk of toxicity", Recommendation: "Reduce digoxin dose by 50% when starting amiodarone"},
		{DrugA: "simvastatin", DrugB: "clarithromycin", Severity: "severe", Description: "CYP3A4 inhibition increases statin levels, risk of rhabdomyolysis", Recommendation: "Suspend simvastatin during macrolide therapy or switch to pravastatin"},
		{DrugA: "metformin", DrugB: "furosemide", Severity: "moderate", Description: "Furosemide may impair renal function, increasing metformin accumulation risk", Recommendation: "Monitor renal function; ensure adequate hydration"},
		{DrugA: "amlodipine", DrugB: "simvastatin", Severity: "moderate", Description: "Amlodipine increases simvastatin exposure, limit simvastatin to 20mg/day", Recommendation: "Do not exceed simvastatin 20mg with amlodipine"},
		{DrugA: "warfarin", DrugB: "ibuprofen", Severity: "severe", Description: "NSAIDs increase bleeding risk through antiplatelet effects and GI irritation", Recommendation: "Avoid NSAIDs; use acetaminophen for pain relief"},
		{DrugA: "sertraline", DrugB: "tramadol", Severity: "severe", Description: "Serotonin syndrome risk with SSRI plus tramadol combination", Recommendation: "Avoid combination; consider alternative analgesic"},
		{DrugA: "ciprofloxacin", DrugB: "theophylline", Severity: "moderate", Description: "Ciprofloxacin inhibits theophylline metabolism, risk of toxicity", Recommendation: "Monitor theophylline levels or use alternative antibiotic"},
	}

	c.conditionInteractions = []ConditionInteraction{
		{Drug: "ibuprofen", Condition: "chronic_kidney_disease", Severity: "severe", Description: "NSAIDs reduce renal blood flow, worsening kidney function", Recommendation: "Avoid NSAIDs in CKD; use acetaminophen or topical alternatives"},
		{Drug: "metformin", Condition: "renal_impairment", Severity: "severe", Description: "Reduced renal clearance increases lactic acidosis risk", Recommendation: "Contraindicated if eGFR <30 mL/min; dose adjust if 30-45"},
		{Drug: "lisinopril", Condition: "hyperkalemia", Severity: "moderate", Description: "ACE inhibitors reduce aldosterone, worsening hyperkalemia", Recommendation: "Monitor potassium; consider alternative antihypertensive"},
		{Drug: "prednisone", Condition: "diabetes", Severity: "moderate", Description: "Corticosteroids increase blood glucose levels", Recommendation: "Monitor blood glucose closely; adjust diabetes medications"},
		{Drug: "oxybutynin", Condition: "dementia", Severity: "moderate", Description: "Anticholinergic effects may worsen cognitive function", Recommendation: "Avoid in dementia patients; consider behavioral interventions"},
		{Drug: "zolpidem", Condition: "fall_risk", Severity: "severe", Description: "Sedative-hypnotics increase fall risk in elderly", Recommendation: "Use lowest effective dose for shortest duration; consider CBT-I"},
		{Drug: "flunitrazepam", Condition: "fall_risk", Severity: "severe", Description: "Benzodiazepines cause sedation, confusion, and falls", Recommendation: "Avoid in fall-risk patients per Beers criteria"},
		{Drug: "atropine", Condition: "glaucoma", Severity: "severe", Description: "Anticholinergics can increase intraocular pressure", Recommendation: "Contraindicated in narrow-angle glaucoma"},
	}
}

// CheckMedications validates a list of medications for drug-drug interactions.
func (c *MedicationInteractionChecker) CheckMedications(ctx context.Context, medications []string) ([]DrugInteraction, error) {
	if len(medications) < 2 {
		return nil, nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	var found []DrugInteraction
	seen := make(map[string]bool)

	for i := 0; i < len(medications); i++ {
		for j := i + 1; j < len(medications); j++ {
			key := interactionKey(medications[i], medications[j])
			if seen[key] {
				continue
			}
			for _, inter := range c.interactions {
				if (strings.EqualFold(inter.DrugA, medications[i]) && strings.EqualFold(inter.DrugB, medications[j])) ||
					(strings.EqualFold(inter.DrugA, medications[j]) && strings.EqualFold(inter.DrugB, medications[i])) {
					found = append(found, inter)
					seen[key] = true
					break
				}
			}
		}
	}

	return found, nil
}

// CheckConditions validates medications against patient conditions.
func (c *MedicationInteractionChecker) CheckConditions(ctx context.Context, medications []string, conditions []string) ([]ConditionInteraction, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var found []ConditionInteraction
	for _, med := range medications {
		for _, cond := range conditions {
			for _, inter := range c.conditionInteractions {
				if strings.EqualFold(inter.Drug, med) && strings.EqualFold(inter.Condition, cond) {
					found = append(found, inter)
					break
				}
			}
		}
	}

	return found, nil
}

// AddInteraction adds a custom interaction to the knowledge base.
func (c *MedicationInteractionChecker) AddInteraction(inter DrugInteraction) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.interactions = append(c.interactions, inter)
}

// AddConditionInteraction adds a custom condition interaction.
func (c *MedicationInteractionChecker) AddConditionInteraction(inter ConditionInteraction) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conditionInteractions = append(c.conditionInteractions, inter)
}

// GetInteractionCount returns the number of interactions in the knowledge base.
func (c *MedicationInteractionChecker) GetInteractionCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.interactions) + len(c.conditionInteractions)
}

// ValidateMedicationRules checks all active medication rules for interactions.
func (c *MedicationInteractionChecker) ValidateMedicationRules(ctx context.Context, rules []model.MedicationRule, conditions []string) ([]DrugInteraction, []ConditionInteraction, error) {
	var medications []string
	for _, rule := range rules {
		if rule.Active && rule.PillType != "" {
			medications = append(medications, rule.PillType)
		}
	}

	interactions, err := c.CheckMedications(ctx, medications)
	if err != nil {
		return nil, nil, fmt.Errorf("check medications: %w", err)
	}

	conditionInteractions, err := c.CheckConditions(ctx, medications, conditions)
	if err != nil {
		return nil, nil, fmt.Errorf("check conditions: %w", err)
	}

	return interactions, conditionInteractions, nil
}

func interactionKey(a, b string) string {
	if a > b {
		a, b = b, a
	}
	return a + "|" + b
}
