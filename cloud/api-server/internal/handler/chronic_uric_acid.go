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

// ChronicUricAcidHandler handles uric acid endpoints under /api/v1/chronic/.
type ChronicUricAcidHandler struct {
	svc *service.ChronicUricAcidService
	log *zap.Logger
}

// NewChronicUricAcidHandler creates a new chronic uric acid handler.
func NewChronicUricAcidHandler(svc *service.ChronicUricAcidService, log *zap.Logger) *ChronicUricAcidHandler {
	return &ChronicUricAcidHandler{svc: svc, log: log}
}

// POST /api/v1/chronic/:elderly_id/uric-acid — manual entry
func (h *ChronicUricAcidHandler) CreateRecord(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	var req model.ChronicUricAcidRecord
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
		h.log.Error("create uric acid record", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "CREATE_FAILED", "message": "Failed to save uric acid record"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "message": "Uric acid record created", "data": gin.H{"id": req.ID}})
}

// GET /api/v1/chronic/:elderly_id/uric-acid — list records
func (h *ChronicUricAcidHandler) ListRecords(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	records, err := h.svc.ListRecords(c.Request.Context(), elderlyID, days)
	if err != nil {
		h.log.Error("list uric acid records", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch uric acid records"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": records})
}
