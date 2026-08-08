package handler

import (
	"net/http"

	"eregen.dev/api-server/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ChronicTaskHandler handles chronic daily task endpoints.
type ChronicTaskHandler struct {
	svc *service.ChronicTaskService
	log *zap.Logger
}

// NewChronicTaskHandler creates a new chronic task handler.
func NewChronicTaskHandler(svc *service.ChronicTaskService, log *zap.Logger) *ChronicTaskHandler {
	return &ChronicTaskHandler{svc: svc, log: log}
}

// ListToday returns today's (or specified date's) daily tasks for an elderly person.
// GET /api/v1/chronic/:elderly_id/daily-tasks?date=2026-08-08
func (h *ChronicTaskHandler) List(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	taskDate := c.Query("date")

	tasks, err := h.svc.ListTasks(c.Request.Context(), elderlyID, taskDate)
	if err != nil {
		h.log.Error("list chronic daily tasks", zap.String("elderly_id", elderlyID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch daily tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": tasks})
}

// MarkComplete marks a daily task as completed.
// PUT /api/v1/chronic/:elderly_id/daily-tasks/:task_id
func (h *ChronicTaskHandler) MarkComplete(c *gin.Context) {
	taskID := c.Param("task_id")

	if err := h.svc.MarkComplete(c.Request.Context(), taskID); err != nil {
		h.log.Error("mark task complete", zap.String("task_id", taskID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "UPDATE_FAILED", "message": "Failed to mark task as completed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Task marked as completed"})
}
