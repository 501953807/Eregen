package handler

import (
	"net/http"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"

	"strconv"

	"github.com/gin-gonic/gin"
)

// HealthGuidanceHandler manages health guidance rules and deliveries.
type HealthGuidanceHandler struct {
	store store.HealthGuidanceStore
}

func NewHealthGuidanceHandler(s store.HealthGuidanceStore) *HealthGuidanceHandler {
	return &HealthGuidanceHandler{store: s}
}

// CreateRule adds a guidance rule.
func (h *HealthGuidanceHandler) CreateRule(c *gin.Context) {
	var r model.HealthGuidanceRule
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.CreateGuidanceRule(c.Request.Context(), &r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK"})
}

// ListRules returns guidance rules for a chain.
func (h *HealthGuidanceHandler) ListRules(c *gin.Context) {
	chain := c.Query("chain")
	enabledOnly := c.Query("enabled_only") == "true"
	rules, err := h.store.ListGuidanceRules(c.Request.Context(), model.BusinessChain(chain), enabledOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": rules})
}

// Evaluate runs guidance evaluation for a person.
func (h *HealthGuidanceHandler) Evaluate(c *gin.Context) {
	personID := c.Param("personId")
	chain := c.Query("chain")
	var healthData map[string]any
	if err := c.ShouldBindJSON(&healthData); err != nil {
		// Allow empty body
		healthData = map[string]any{}
	}
	rules, err := h.store.EvaluateGuidanceRules(c.Request.Context(), personID, model.BusinessChain(chain), healthData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": rules})
}

// CreateDelivery records a guidance delivery.
func (h *HealthGuidanceHandler) CreateDelivery(c *gin.Context) {
	var d model.HealthGuidanceDelivery
	if err := c.ShouldBindJSON(&d); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.CreateGuidanceDelivery(c.Request.Context(), &d); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK"})
}

// ListDeliveries returns guidance delivery history.
func (h *HealthGuidanceHandler) ListDeliveries(c *gin.Context) {
	personID := c.Param("personId")
	chain := c.Query("chain")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit > 200 {
		limit = 200
	}
	deliveries, err := h.store.ListGuidanceDeliveries(c.Request.Context(), personID, model.BusinessChain(chain), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": deliveries})
}
