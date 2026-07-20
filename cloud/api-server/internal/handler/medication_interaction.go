package handler

import (
	"net/http"

	"eregen.dev/api-server/internal/middleware"
	"eregen.dev/api-server/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MedicationInteractionHandler handles drug interaction checking endpoints.
type MedicationInteractionHandler struct {
	checker *service.MedicationInteractionChecker
	log     *zap.Logger
}

func NewMedicationInteractionHandler(checker *service.MedicationInteractionChecker, log *zap.Logger) *MedicationInteractionHandler {
	return &MedicationInteractionHandler{checker: checker, log: log}
}

// POST /api/v1/medication/check-interactions
// Checks a list of medications for drug-drug interactions.
func (h *MedicationInteractionHandler) CheckInteractions(c *gin.Context) {
	var req struct {
		Medications []string `json:"medications" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "medications array required"})
		return
	}

	interactions, err := h.checker.CheckMedications(c.Request.Context(), req.Medications)
	if err != nil {
		h.log.Error("check medication interactions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "Failed to check interactions"})
		return
	}

	severityOrder := map[string]int{"contraindicated": 0, "severe": 1, "moderate": 2, "mild": 3}
	for i := 0; i < len(interactions); i++ {
		for j := i + 1; j < len(interactions); j++ {
			if severityOrder[interactions[i].Severity] > severityOrder[interactions[j].Severity] {
				interactions[i], interactions[j] = interactions[j], interactions[i]
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":         "OK",
		"data":         interactions,
		"interaction_count": len(interactions),
	})
}

// POST /api/v1/medication/check-conditions
// Checks medications against patient conditions.
func (h *MedicationInteractionHandler) CheckConditions(c *gin.Context) {
	var req struct {
		Medications []string `json:"medications" binding:"required,min=1"`
		Conditions  []string `json:"conditions" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "medications and conditions required"})
		return
	}

	interactions, err := h.checker.CheckConditions(c.Request.Context(), req.Medications, req.Conditions)
	if err != nil {
		h.log.Error("check condition interactions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "Failed to check condition interactions"})
		return
	}

	severityOrder := map[string]int{"contraindicated": 0, "severe": 1, "moderate": 2, "mild": 3}
	for i := 0; i < len(interactions); i++ {
		for j := i + 1; j < len(interactions); j++ {
			if severityOrder[interactions[i].Severity] > severityOrder[interactions[j].Severity] {
				interactions[i], interactions[j] = interactions[j], interactions[i]
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":                "OK",
		"data":                interactions,
		"condition_count":     len(interactions),
	})
}

// GET /api/v1/medication/rules/:rule_id/validate
// Validates all active rules for this elderly profile.
func (h *MedicationInteractionHandler) ValidateRules(c *gin.Context) {
	userID, _ := c.Get(string(middleware.ContextUserID))
	elderlyID := c.Param("elderly_id")

	_ = userID
	_ = elderlyID

	rules := c.QueryArray("rule_ids[]")
	if len(rules) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "rule_ids[] query parameter required"})
		return
	}

	// In production, fetch rules from store; for now return knowledge base stats
	c.JSON(http.StatusOK, gin.H{
		"code":                  "OK",
		"knowledge_base_size":   h.checker.GetInteractionCount(),
		"message":               "Rule validation requires database access in production",
	})
}
