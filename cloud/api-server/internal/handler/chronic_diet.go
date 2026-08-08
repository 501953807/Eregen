package handler

import (
	"net/http"

	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ChronicDietHandler handles diet endpoints under /api/v1/chronic/.
type ChronicDietHandler struct {
	svc *service.ChronicDietService
	log *zap.Logger
}

// NewChronicDietHandler creates a new chronic diet handler.
func NewChronicDietHandler(svc *service.ChronicDietService, log *zap.Logger) *ChronicDietHandler {
	return &ChronicDietHandler{svc: svc, log: log}
}

// POST /api/v1/chronic/:elderly_id/diet — record a meal
func (h *ChronicDietHandler) CreateRecord(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	var req model.ChronicDietRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "Invalid request body"})
		return
	}

	if err := h.svc.CreateRecord(c.Request.Context(), elderlyID, &req); err != nil {
		h.log.Error("create diet record", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "CREATE_FAILED", "message": "Failed to save diet record"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "message": "Diet record created", "data": gin.H{"id": req.ID}})
}

// GET /api/v1/chronic/:elderly_id/diet — list diet records
func (h *ChronicDietHandler) ListRecords(c *gin.Context) {
	elderlyID := c.Param("elderly_id")

	records, err := h.svc.ListRecords(c.Request.Context(), elderlyID, 30)
	if err != nil {
		h.log.Error("list diet records", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch diet records"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": records})
}
