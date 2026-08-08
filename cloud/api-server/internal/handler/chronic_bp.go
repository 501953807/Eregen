package handler

import (
	"net/http"
	"strconv"
	"time"

	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ChronicBPHandler handles blood pressure endpoints under /api/v1/chronic/.
type ChronicBPHandler struct {
	svc *service.ChronicBPService
	log *zap.Logger
}

// NewChronicBPHandler creates a new chronic BP handler.
func NewChronicBPHandler(svc *service.ChronicBPService, log *zap.Logger) *ChronicBPHandler {
	return &ChronicBPHandler{svc: svc, log: log}
}

// POST /api/v1/chronic/:elderly_id/blood-pressure — manual entry
func (h *ChronicBPHandler) CreateRecord(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	var req model.ChronicBPRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "Invalid request body"})
		return
	}
	if req.Systolic <= 0 || req.Diastolic <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_VALUE", "message": "systolic and diastolic must be positive"})
		return
	}
	if req.MeasurementTime.IsZero() {
		req.MeasurementTime = time.Now()
	}

	if err := h.svc.CreateRecord(c.Request.Context(), elderlyID, &req); err != nil {
		h.log.Error("create BP record", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "CREATE_FAILED", "message": "Failed to save blood pressure record"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "message": "Blood pressure record created", "data": gin.H{"id": req.ID}})
}

// GET /api/v1/chronic/:elderly_id/blood-pressure — list records
func (h *ChronicBPHandler) ListRecords(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	records, err := h.svc.ListRecords(c.Request.Context(), elderlyID, days)
	if err != nil {
		h.log.Error("list BP records", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch blood pressure records"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": records})
}

// POST /api/v1/chronic/:elderly_id/bp-device/sync — from BLE/medical device
func (h *ChronicBPHandler) SyncFromDevice(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	var req model.ChronicBPRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "Invalid request body"})
		return
	}
	if req.Systolic <= 0 || req.Diastolic <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_VALUE", "message": "systolic and diastolic must be positive"})
		return
	}
	if req.MeasurementTime.IsZero() {
		req.MeasurementTime = time.Now()
	}

	if err := h.svc.SyncFromDevice(c.Request.Context(), elderlyID, &req); err != nil {
		h.log.Error("sync BP from device", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "CREATE_FAILED", "message": "Failed to save blood pressure record from device"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "message": "Blood pressure record synced from device", "data": gin.H{"id": req.ID}})
}
