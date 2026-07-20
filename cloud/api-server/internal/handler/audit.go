package handler

import (
	"net/http"
	"strconv"

	"eregen.dev/api-server/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuditHandler handles audit log query endpoints.
type AuditHandler struct {
	logger *service.AuditLogger
	log    *zap.Logger
}

// NewAuditHandler creates a new audit handler.
func NewAuditHandler(logger *service.AuditLogger, log *zap.Logger) *AuditHandler {
	return &AuditHandler{logger: logger, log: log}
}

// GET /api/v1/admin/audit-logs — list recent audit entries (admin)
func (h *AuditHandler) List(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 500 {
		limit = 50
	}

	entries := h.logger.GetEntries(limit)
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": entries})
}

// GET /api/v1/users/me/audit-logs — list user's own audit entries
func (h *AuditHandler) MyLogs(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDStr, ok := userID.(string)
	if !ok {
		userIDStr = "anonymous"
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 500 {
		limit = 50
	}

	entries := h.logger.GetEntriesByUser(userIDStr, limit)
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": entries})
}
