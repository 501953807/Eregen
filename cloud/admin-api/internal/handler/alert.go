package handler

import (
	"net/http"
	"strconv"

	"eregen.dev/admin-api/internal/store"
	"eregen.dev/shared/validation"

	"github.com/gin-gonic/gin"
)

// AlertHandler serves alert management endpoints.
type AlertHandler struct {
	store *store.PostgresStore
}

// NewAlertHandler creates a new AlertHandler.
func NewAlertHandler(s *store.PostgresStore) *AlertHandler {
	return &AlertHandler{store: s}
}

// Resolve marks an alert as resolved.
func (h *AlertHandler) Resolve(c *gin.Context) {
	alertID := c.Param("id")
	if err := h.store.ResolveAlert(c.Request.Context(), alertID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "alert resolved"})
}

// List returns recent alerts with optional severity and status filters.
func (h *AlertHandler) List(c *gin.Context) {
	var sev, status string

	if sev = c.Query("severity"); sev != "" {
		if err := validation.ValidateEnum(sev, []string{"P0", "P1", "P2"}); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid severity"})
			return
		}
	}
	if status = c.Query("status"); status != "" {
		if err := validation.ValidateEnum(status, []string{"pending", "resolved"}); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if limit < 1 || limit > 200 {
		limit = 50
	}

	alerts, err := h.store.ListAlerts(c.Request.Context(), sev, status, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": alerts})
}
