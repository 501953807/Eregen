package handler

import (
	"net/http"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"

	"github.com/gin-gonic/gin"
)

// AlertRuleHandler manages alert_rules per business chain.
type AlertRuleHandler struct {
	store store.Store
}

func NewAlertRuleHandler(s store.Store) *AlertRuleHandler {
	return &AlertRuleHandler{store: s}
}

// List returns alert rules filtered by business_chain.
func (h *AlertRuleHandler) List(c *gin.Context) {
	chain := c.Query("chain")
	rules, err := h.store.ListAlertRules(c.Request.Context(), model.BusinessChain(chain))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": rules})
}

// Create adds a new alert rule.
func (h *AlertRuleHandler) Create(c *gin.Context) {
	var r struct {
		Name          string `json:"name" binding:"required"`
		BusinessChain string `json:"business_chain" binding:"required"`
		AlertType     string `json:"alert_type" binding:"required"`
		Severity      string `json:"severity" binding:"required"`
	}
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	alertRule := &model.AlertRule{
		Name:          r.Name,
		BusinessChain: r.BusinessChain,
		AlertType:     r.AlertType,
		Severity:      r.Severity,
	}
	if err := h.store.CreateAlertRule(c.Request.Context(), alertRule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK"})
}

// Update modifies an existing alert rule.
func (h *AlertRuleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]any
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.UpdateAlertRule(c.Request.Context(), id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK"})
}

// Delete removes an alert rule.
func (h *AlertRuleHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.store.DeleteAlertRule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK"})
}
