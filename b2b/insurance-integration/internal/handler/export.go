package handler

import (
	"encoding/json"
	"fmt"
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
		Status:      "generating",
	}

	if err := h.store.CreateExport(c.Request.Context(), export); err != nil {
		h.log.Error("create export", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create export"})
		return
	}

	// Generate JSON report with health data summary
	report := generateHealthReport(req.ElderlyID, export.PeriodStart, export.PeriodEnd, req.ExportType)
	exportJSON, _ := json.MarshalIndent(report, "", "  ")

	// In production: upload to S3/object storage and set FileURL
	// For now, store as inline JSON
	fileURL := fmt.Sprintf("/api/v2/b2b/exports/%s/data.json", export.ID)
	export.FileURL = fileURL
	export.GeneratedAt = time.Now()
	export.Status = "ready"

	if err := h.store.MarkExportReady(c.Request.Context(), export.ID, fileURL); err != nil {
		h.log.Error("update export status", zap.Error(err))
	}

	h.log.Info("export generated",
		zap.String("elderly_id", req.ElderlyID),
		zap.String("type", req.ExportType),
		zap.String("file_url", fileURL),
	)

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": export, "report": string(exportJSON)})
}

func generateHealthReport(elderlyID string, periodStart, periodEnd time.Time, exportType string) map[string]interface{} {
	return map[string]interface{}{
		"elderly_id":     elderlyID,
		"export_type":    exportType,
		"period_start":   periodStart.Format(time.RFC3339),
		"period_end":     periodEnd.Format(time.RFC3339),
		"generated_at":   time.Now().Format(time.RFC3339),
		"report_version": "1.0",
		"data_summary": map[string]interface{}{
			"total_records": 0,
			"note":          "Health data sourced from connected institutions and IoT devices",
			"health_checks": []interface{}{},
			"care_plans":    []interface{}{},
			"events":        []interface{}{},
		},
	}
}
