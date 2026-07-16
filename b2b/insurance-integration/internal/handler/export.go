package handler

import (
	"net/http"
	"time"

	"eregen.dev/b2b-insurance-integration/internal/model"
	"eregen.dev/b2b-insurance-integration/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ExportHandler struct {
	store *store.Postgres
	log   *zap.Logger
}

func NewExportHandler(store *store.Postgres, log *zap.Logger) *ExportHandler {
	return &ExportHandler{store: store, log: log}
}

// POST /api/v2/b2b/exports — generate a health data export for insurance
func (h *ExportHandler) Create(c *gin.Context) {
	var req struct {
		ElderlyID  string    `json:"elderly_id" binding:"required"`
		ClaimID    *string   `json:"claim_id"`
		ExportType string    `json:"export_type" binding:"required"`
		Days       int       `json:"days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	days := req.Days
	if days <= 0 {
		days = 90
	}

	export := &model.HealthDataExport{
		ElderlyID:   req.ElderlyID,
		ClaimID:     req.ClaimID,
		ExportType:  req.ExportType,
		PeriodStart: time.Now().Add(-time.Duration(days) * 24 * time.Hour),
		PeriodEnd:   time.Now(),
	}

	if err := h.store.CreateExport(c.Request.Context(), export); err != nil {
		h.log.Error("create export", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create export"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": export})
}
