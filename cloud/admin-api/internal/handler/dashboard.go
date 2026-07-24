package handler

import (
	"net/http"
	"strconv"

	"eregen.dev/admin-api/internal/store"

	"github.com/gin-gonic/gin"
)

// DashboardHandler serves dashboard statistics endpoints.
type DashboardHandler struct {
	store store.Store
}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler(s store.Store) *DashboardHandler {
	return &DashboardHandler{store: s}
}

// GetOverview returns the top-level dashboard metrics.
func (h *DashboardHandler) GetOverview(c *gin.Context) {
	stats, err := h.store.GetDashboardStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": stats})
}

// GetSubscriptionStats returns a per-tier subscription breakdown.
func (h *DashboardHandler) GetSubscriptionStats(c *gin.Context) {
	stats, err := h.store.GetSubscriptionStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": stats})
}

// GetAlertTrend returns alert counts grouped by date and device type.
func (h *DashboardHandler) GetAlertTrend(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}
	points, err := h.store.GetAlertTrend(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": points})
}

// GetAlertDistribution returns alert counts by type.
func (h *DashboardHandler) GetAlertDistribution(c *gin.Context) {
	items, err := h.store.GetAlertDistribution(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": items})
}

// GetUserGrowth returns new user counts grouped by month.
func (h *DashboardHandler) GetUserGrowth(c *gin.Context) {
	months, _ := strconv.Atoi(c.DefaultQuery("months", "12"))
	if months < 1 || months > 24 {
		months = 12
	}
	points, err := h.store.GetUserGrowth(c.Request.Context(), months)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": points})
}
