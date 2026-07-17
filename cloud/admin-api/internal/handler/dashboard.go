package handler

import (
	"net/http"
	"eregen.dev/admin-api/internal/store"

	"github.com/gin-gonic/gin"
)

// DashboardHandler serves dashboard statistics endpoints.
type DashboardHandler struct {
	store *store.PostgresStore
}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler(s *store.PostgresStore) *DashboardHandler {
	return &DashboardHandler{store: s}
}

// GetOverview returns the top-level dashboard metrics.
func (h *DashboardHandler) GetOverview(c *gin.Context) {
	stats, err := h.store.GetDashboardStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetSubscriptionStats returns a per-tier subscription breakdown.
func (h *DashboardHandler) GetSubscriptionStats(c *gin.Context) {
	stats, err := h.store.GetSubscriptionStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}
