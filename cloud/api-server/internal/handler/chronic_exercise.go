package handler

import (
	"net/http"

	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ChronicExerciseHandler handles exercise endpoints under /api/v1/chronic/.
type ChronicExerciseHandler struct {
	svc *service.ChronicExerciseService
	log *zap.Logger
}

// NewChronicExerciseHandler creates a new chronic exercise handler.
func NewChronicExerciseHandler(svc *service.ChronicExerciseService, log *zap.Logger) *ChronicExerciseHandler {
	return &ChronicExerciseHandler{svc: svc, log: log}
}

// POST /api/v1/chronic/:elderly_id/exercise — record an exercise session
func (h *ChronicExerciseHandler) CreateRecord(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	var req model.ChronicExerciseRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "Invalid request body"})
		return
	}

	if err := h.svc.CreateRecord(c.Request.Context(), elderlyID, &req); err != nil {
		h.log.Error("create exercise record", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "CREATE_FAILED", "message": "Failed to save exercise record"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "message": "Exercise record created", "data": gin.H{"id": req.ID}})
}

// GET /api/v1/chronic/:elderly_id/exercise — list exercise records
func (h *ChronicExerciseHandler) ListRecords(c *gin.Context) {
	elderlyID := c.Param("elderly_id")

	records, err := h.svc.ListRecords(c.Request.Context(), elderlyID, 30)
	if err != nil {
		h.log.Error("list exercise records", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch exercise records"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": records})
}
