package handler

import (
	"net/http"
	"strconv"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"
	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

// MedicationHandler manages medication rules and executions per person + chain.
type MedicationHandler struct {
	store store.MedicationRuleStore
}

func NewMedicationHandler(s store.MedicationRuleStore) *MedicationHandler {
	return &MedicationHandler{store: s}
}

// ListRules returns medication rules for a person in a given chain.
func (h *MedicationHandler) ListRules(c *gin.Context) {
	personID := c.Param("personId")
	chain := c.Query("chain")
	rules, err := h.store.ListMedicationRules(c.Request.Context(), personID, model.BusinessChain(chain))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": rules})
}

// CreateRule adds a new medication rule.
func (h *MedicationHandler) CreateRule(c *gin.Context) {
	var r model.MedicationRuleRow
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	r.ID = uuid.New().String()
	r.CreatedAt = ""
	if err := h.store.CreateMedicationRuleV2(c.Request.Context(), &r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": r})
}

// UpdateRule modifies an existing medication rule.
func (h *MedicationHandler) UpdateRule(c *gin.Context) {
	id := c.Param("ruleId")
	var updates map[string]any
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.UpdateMedicationRuleV2(c.Request.Context(), id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK"})
}

// DeleteRule removes a medication rule.
func (h *MedicationHandler) DeleteRule(c *gin.Context) {
	id := c.Param("ruleId")
	if err := h.store.DeleteMedicationRuleV2(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK"})
}

// CreateExecution records a medication taking event.
func (h *MedicationHandler) CreateExecution(c *gin.Context) {
	var e model.MedicationExecution
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	e.ID = uuid.New().String()
	if err := h.store.CreateMedicationExecution(c.Request.Context(), &e); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": e})
}

// ListExecutions returns recent medication executions.
func (h *MedicationHandler) ListExecutions(c *gin.Context) {
	personID := c.Param("personId")
	chain := c.Query("chain")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit > 200 {
		limit = 200
	}
	execs, err := h.store.ListMedicationExecutions(c.Request.Context(), personID, model.BusinessChain(chain), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": execs})
}
