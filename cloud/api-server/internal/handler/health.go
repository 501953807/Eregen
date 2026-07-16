package handler

import (
	"net/http"
	"strconv"
	"time"

	"eregen.dev/api-server/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HealthHandler handles health data endpoints.
type HealthHandler struct {
	svc *service.HealthService
	log *zap.Logger
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(svc *service.HealthService, log *zap.Logger) *HealthHandler {
	return &HealthHandler{svc: svc, log: log}
}

// GET /api/v1/elderly/:elderly_id/health/summary
func (h *HealthHandler) Summary(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	dayStr := c.DefaultQuery("day", time.Now().Format("2006-01-02"))
	day, err := time.Parse("2006-01-02", dayStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_DATE", "message": "Invalid date format, use YYYY-MM-DD"})
		return
	}

	record, err := h.svc.GetSummary(c.Request.Context(), elderlyID, day)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "No health data found for this date"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": record})
}

// GET /api/v1/elderly/:elderly_id/health/history
func (h *HealthHandler) History(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	records, err := h.svc.GetHistory(c.Request.Context(), elderlyID, days)
	if err != nil {
		h.log.Error("get health history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch health data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": records})
}

// GET /api/v1/elderly/:elderly_id/health/trend
func (h *HealthHandler) Trend(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	metric := c.DefaultQuery("metric", "hr")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	validMetrics := map[string]bool{"hr": true, "spo2": true, "steps": true, "sleep_hours": true, "bp_systolic": true, "bp_diastolic": true}
	if !validMetrics[metric] {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_METRIC", "message": "Valid metrics: hr, spo2, steps, sleep_hours, bp_systolic, bp_diastolic"})
		return
	}

	records, err := h.svc.GetTrend(c.Request.Context(), elderlyID, metric, days)
	if err != nil {
		h.log.Error("get health trend", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch trend data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": records})
}
