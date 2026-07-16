package handler

import (
	"net/http"

	"eregen.dev/b2b-community-platform/internal/model"
	"eregen.dev/b2b-community-platform/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HealthCheckHandler struct {
	store *store.Postgres
	log   *zap.Logger
}

func NewHealthCheckHandler(store *store.Postgres, log *zap.Logger) *HealthCheckHandler {
	return &HealthCheckHandler{store: store, log: log}
}

// POST /api/v2/b2b/health-checks — record a community health check
func (h *HealthCheckHandler) Create(c *gin.Context) {
	var req model.HealthCheckRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.store.CreateHealthCheck(c.Request.Context(), &req); err != nil {
		h.log.Error("create health check", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save health check"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": req})
}

// GET /api/v2/b2b/health-checks/:elderly_id — get health checks for an elderly person
func (h *HealthCheckHandler) GetForElderly(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	limit, _ := parseIntParam(c, "limit", 50)

	records, err := h.store.GetHealthChecksForElderly(c.Request.Context(), elderlyID, limit)
	if err != nil {
		h.log.Error("get health checks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get health checks"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": records})
}
