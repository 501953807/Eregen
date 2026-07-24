package handler

import (
	"net/http"
	"strconv"

	"eregen.dev/api-server/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AdminStatsHandler serves admin dashboard statistics.
type AdminStatsHandler struct {
	pg *store.Postgres
	log *zap.Logger
}

func NewAdminStatsHandler(pg *store.Postgres, log *zap.Logger) *AdminStatsHandler {
	return &AdminStatsHandler{pg: pg, log: log}
}

// GET /api/v1/admin/stats/overview
func (h *AdminStatsHandler) Overview(c *gin.Context) {
	stats, err := h.pg.AdminStatsOverview(c.Request.Context())
	if err != nil {
		h.log.Error("admin stats overview", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch stats"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": stats})
}

// GET /api/v1/admin/stats/alert-trend
func (h *AdminStatsHandler) AlertTrend(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	points, err := h.pg.AdminStatsAlertTrend(c.Request.Context(), days)
	if err != nil {
		h.log.Error("admin alert trend", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": points})
}

// GET /api/v1/admin/stats/alert-distribution
func (h *AdminStatsHandler) AlertDistribution(c *gin.Context) {
	items, err := h.pg.AdminStatsAlertDistribution(c.Request.Context())
	if err != nil {
		h.log.Error("admin alert distribution", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": items})
}

// GET /api/v1/admin/stats/user-growth
func (h *AdminStatsHandler) UserGrowth(c *gin.Context) {
	months, _ := strconv.Atoi(c.DefaultQuery("months", "6"))
	points, err := h.pg.AdminStatsUserGrowth(c.Request.Context(), months)
	if err != nil {
		h.log.Error("admin user growth", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": points})
}
