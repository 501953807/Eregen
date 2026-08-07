package handler

import (
	"net/http"
	"strconv"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"
	"eregen.dev/shared/validation"

	"github.com/gin-gonic/gin"
)

// AlertHandler serves alert management endpoints.
type AlertHandler struct {
	store store.Store
}

// NewAlertHandler creates a new AlertHandler.
func NewAlertHandler(s store.Store) *AlertHandler {
	return &AlertHandler{store: s}
}

// Resolve marks an alert as resolved and broadcasts to SSE clients.
func (h *AlertHandler) Resolve(c *gin.Context) {
	alertID := c.Param("id")
	if err := h.store.ResolveAlert(c.Request.Context(), alertID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "OK", "msg": "failed to resolve alert"})
		return
	}
	// Broadcast stats update after resolution.
	if stats, err := h.store.GetDashboardStats(c.Request.Context()); err == nil {
		AlertsHubInstance.PushStats(*stats)
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "alert resolved"})
}

// Acknowledge marks an alert as acknowledged and broadcasts to SSE clients.
func (h *AlertHandler) Acknowledge(c *gin.Context) {
	alertID := c.Param("id")
	if err := h.store.UpdateAlertStatus(c.Request.Context(), alertID, "acknowledged"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "OK", "msg": "failed to acknowledge alert"})
		return
	}
	if stats, err := h.store.GetDashboardStats(c.Request.Context()); err == nil {
		AlertsHubInstance.PushStats(*stats)
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "alert acknowledged"})
}

// Create adds a new alert.
func (h *AlertHandler) Create(c *gin.Context) {
	var body struct {
		ElderlyID string `json:"elderly_id"`
		AlertType string `json:"alert_type" binding:"required"`
		Severity  string `json:"severity" binding:"required"`
		DeviceID  string `json:"device_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	alert := &model.AlertSummary{
		ElderlyID: body.ElderlyID,
		AlertType: body.AlertType,
		Severity:  body.Severity,
		Status:    "pending",
		DeviceID:  body.DeviceID,
	}
	if err := h.store.CreateAlert(c.Request.Context(), alert); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": alert})
}


// List returns recent alerts with optional severity and status filters.
func (h *AlertHandler) List(c *gin.Context) {
	var sev, status string

	if sev = c.Query("severity"); sev != "" {
		if err := validation.ValidateEnum(sev, []string{"P0", "P1", "P2"}); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "OK", "msg": "invalid severity"})
			return
		}
	}
	if status = c.Query("status"); status != "" {
		if err := validation.ValidateEnum(status, []string{"pending", "resolved"}); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "OK", "msg": "invalid status"})
			return
		}
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if limit < 1 || limit > 200 {
		limit = 50
	}

	alerts, err := h.store.ListAlerts(c.Request.Context(), sev, status, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "OK", "msg": "failed to list alerts"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": alerts})
}

// StreamHandler returns the SSE handler for /api/v1/admin/stream/alerts.
func StreamHandler() http.HandlerFunc {
	return AlertsHubInstance.Handler()
}
