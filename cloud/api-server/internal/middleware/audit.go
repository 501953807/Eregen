package middleware

import (
	"eregen.dev/api-server/internal/service"

	"github.com/gin-gonic/gin"
)

// AuditMiddleware injects audit logging into request context.
type AuditMiddleware struct {
	logger *service.AuditLogger
}

func NewAuditMiddleware(logger *service.AuditLogger) *AuditMiddleware {
	return &AuditMiddleware{logger: logger}
}

// LogAction logs an audit entry after the request completes.
func (m *AuditMiddleware) LogAction(action service.AuditAction, resource string, resourceID string, details map[string]any) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		userID, _ := c.Get("user_id")
		userIDStr, ok := userID.(string)
		if !ok {
			userIDStr = "anonymous"
		}

		ip := c.ClientIP()
		userAgent := c.Request.UserAgent()

		m.logger.Log(c.Request.Context(), userIDStr, action, resource, resourceID, details, ip, userAgent)
	}
}
