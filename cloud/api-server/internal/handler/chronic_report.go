package handler

import (
	"net/http"

	"eregen.dev/api-server/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ChronicReportHandler handles chronic health report endpoints.
type ChronicReportHandler struct {
	svc *service.ChronicReportService
	log *zap.Logger
}

// NewChronicReportHandler creates a new chronic report handler.
func NewChronicReportHandler(svc *service.ChronicReportService, log *zap.Logger) *ChronicReportHandler {
	return &ChronicReportHandler{svc: svc, log: log}
}

// GET /api/v1/chronic/:elderly_id/report/:type — get existing report
func (h *ChronicReportHandler) GetReport(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	reportType := c.Param("type")

	report, err := h.svc.GetReport(c.Request.Context(), elderlyID, reportType)
	if err != nil {
		h.log.Error("get report", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"code": "REPORT_NOT_FOUND", "message": "No report found for the specified type"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": report})
}

// POST /api/v1/chronic/:elderly_id/report/generate — generate a new report
func (h *ChronicReportHandler) GenerateReport(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	var req struct {
		ReportType string `json:"report_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "report_type is required"})
		return
	}

	validTypes := map[string]bool{"weekly": true, "monthly": true, "annual": true}
	if !validTypes[req.ReportType] {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_TYPE", "message": "report_type must be one of: weekly, monthly, annual"})
		return
	}

	report, err := h.svc.GenerateReport(c.Request.Context(), elderlyID, req.ReportType)
	if err != nil {
		h.log.Error("generate report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "GENERATE_FAILED", "message": "Failed to generate report"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Report generated successfully", "data": report})
}
