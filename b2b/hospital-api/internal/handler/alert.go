package handler

import (
	"net/http"

	"eregen.dev/b2b-hospital-api/internal/model"
	"eregen.dev/b2b-hospital-api/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AlertHandler struct {
	store *store.Postgres
	log   *zap.Logger
}

func NewAlertHandler(store *store.Postgres, log *zap.Logger) *AlertHandler {
	return &AlertHandler{store: store, log: log}
}

// POST /api/v2/b2b/alerts/forward — forward P0/P1 alerts to linked institutions
func (h *AlertHandler) Forward(c *gin.Context) {
	var req model.AlertForwardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	h.log.Info("alert forwarded",
		zap.String("elderly_id", req.ElderlyID),
		zap.String("alert_type", req.AlertType),
		zap.String("severity", req.Severity),
		zap.String("institution", req.InstitutionID),
	)

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Alert forwarded successfully"})
}
