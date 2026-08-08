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

// ChronicGlucoseHandler handles blood glucose endpoints under /api/v1/chronic/.
type ChronicGlucoseHandler struct {
	svc *service.ChronicGlucoseService
	log *zap.Logger
}

// NewChronicGlucoseHandler creates a new chronic glucose handler.
func NewChronicGlucoseHandler(svc *service.ChronicGlucoseService, log *zap.Logger) *ChronicGlucoseHandler {
	return &ChronicGlucoseHandler{svc: svc, log: log}
}

// POST /api/v1/chronic/:elderly_id/glucose — manual entry
func (h *ChronicGlucoseHandler) CreateRecord(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	var req model.ChronicGlucoseRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "Invalid request body"})
		return
	}
	if req.Value <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_VALUE", "message": "value must be a positive number"})
		return
	}
	if req.MeasurementTime.IsZero() {
		req.MeasurementTime = time.Now()
	}

	if err := h.svc.CreateRecord(c.Request.Context(), elderlyID, &req); err != nil {
		h.log.Error("create glucose record", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "CREATE_FAILED", "message": "Failed to save glucose record"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "message": "Glucose record created", "data": gin.H{"id": req.ID}})
}

// GET /api/v1/chronic/:elderly_id/glucose — list records
func (h *ChronicGlucoseHandler) ListRecords(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	records, err := h.svc.ListRecords(c.Request.Context(), elderlyID, days)
	if err != nil {
		h.log.Error("list glucose records", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch glucose records"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": records})
}

// GET /api/v1/chronic/:elderly_id/glucose/trend — aggregated trend data
func (h *ChronicGlucoseHandler) GetTrend(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	trend, err := h.svc.GetTrend(c.Request.Context(), elderlyID, days)
	if err != nil {
		h.log.Error("get glucose trend", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch glucose trend"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": trend})
}

// POST /api/v1/chronic/:elderly_id/test-strip/read — from bracelet device
func (h *ChronicGlucoseHandler) TestStripRead(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	var req model.ChronicGlucoseRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "Invalid request body"})
		return
	}
	if req.Value <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_VALUE", "message": "value must be a positive number"})
		return
	}
	if req.MeasurementTime.IsZero() {
		req.MeasurementTime = time.Now()
	}

	if err := h.svc.CreateTestStripRead(c.Request.Context(), elderlyID, &req); err != nil {
		h.log.Error("create test-strip read", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "CREATE_FAILED", "message": "Failed to save test-strip reading"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "message": "Test-strip reading recorded", "data": gin.H{"id": req.ID}})
}
