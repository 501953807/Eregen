package handler

import (
	"net/http"

	"eregen.dev/api-server/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// EmergencyHandler handles emergency response workflow endpoints.
type EmergencyHandler struct {
	workflow *service.EmergencyResponseWorkflow
	log      *zap.Logger
}

func NewEmergencyHandler(workflow *service.EmergencyResponseWorkflow, log *zap.Logger) *EmergencyHandler {
	return &EmergencyHandler{workflow: workflow, log: log}
}

// ResolveAlert resolves an active emergency case.
func (h *EmergencyHandler) ResolveAlert(c *gin.Context) {
	alertID := c.Param("alert_id")

	if err := h.workflow.ResolveAlert(c.Request.Context(), alertID); err != nil {
		h.log.Error("resolve emergency alert", zap.String("alert_id", alertID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "Failed to resolve alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Alert resolved"})
}

// GetActiveCases returns all active emergency cases.
func (h *EmergencyHandler) GetActiveCases(c *gin.Context) {
	cases := h.workflow.GetActiveCases()

	c.JSON(http.StatusOK, gin.H{
		"code":        "OK",
		"active_cases": cases,
		"count":       len(cases),
	})
}
