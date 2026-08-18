package handler

import (
	"net/http"
	"time"

	"eregen.dev/b2b-hospital-api/internal/model"
	"eregen.dev/b2b-hospital-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ExportHandler struct {
	store store.Database
	log   *zap.Logger
}

func NewExportHandler(store store.Database, log *zap.Logger) *ExportHandler {
	return &ExportHandler{store: store, log: log}
}

// POST /api/v2/b2b/hospitals/:id/export — request health data export
func (h *ExportHandler) Create(c *gin.Context) {
	instID := c.Param("id")
	var req struct {
		ElderlyID   string `json:"elderly_id" binding:"required"`
		PeriodStart string `json:"period_start" binding:"required"`
		PeriodEnd   string `json:"period_end" binding:"required"`
		ExportType  string `json:"export_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	export := &model.ExportRequest{
		ID:            uuid.New().String(),
		ElderlyID:     req.ElderlyID,
		InstitutionID: instID,
		ExportType:    req.ExportType,
		PeriodStart:   req.PeriodStart,
		PeriodEnd:     req.PeriodEnd,
		Status:        "generating",
		CreatedAt:     time.Now(),
	}

	if err := h.store.CreateExport(c.Request.Context(), export); err != nil {
		h.log.Error("create export", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create export"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": export})
}

// GET /api/v2/b2b/hospitals/:id/export/:export_id — get export status
func (h *ExportHandler) Get(c *gin.Context) {
	exportID := c.Param("export_id")
	export, err := h.store.GetExportByID(c.Request.Context(), exportID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "export not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": export})
}
