package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"eregen.dev/admin-api/internal/auth"
)

type ContextKey string

const (
	ContextUserID      ContextKey = "user_id"
	ContextRole        ContextKey = "role"
	ContextAdminRole   ContextKey = "admin_role"
)

type AdminJWT struct {
	secret   string
	tokenTTL time.Duration
	log      *zap.Logger
}

func NewAdminJWT(secret string, tokenTTL time.Duration, log *zap.Logger) *AdminJWT {
	return &AdminJWT{secret: secret, tokenTTL: tokenTTL, log: log}
}

// AuthMiddleware validates JWT tokens and extracts user claims
func (j *AdminJWT) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "malformed authorization header, use 'Bearer <token>'"})
			return
		}

		tokenStr := parts[1]
		claims, err := auth.ValidateToken(tokenStr, j.secret)
		if err != nil {
			j.log.Warn("admin auth failed", zap.String("ip", c.ClientIP()), zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(string(ContextUserID), claims.UserID)
		c.Set(string(ContextRole), claims.Role)

		// For admin-specific role context (backward compatibility)
		if claims.Role == "admin" || claims.Role == "operator" || claims.Role == "nurse" || claims.Role == "regulator" {
			c.Set(string(ContextAdminRole), claims.Role)
		}

		c.Next()
	}
}

// RequireAuth ensures the request has a valid authenticated user
func (j *AdminJWT) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get(string(ContextUserID))
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized access"})
			return
		}
		_ = userID // suppress unused variable warning if needed
		c.Next()
	}
}

// RequireRole checks if the user's role meets or exceeds the minimum required role
func (j *AdminJWT) RequireRole(minRole string) gin.HandlerFunc {
	roleOrder := map[string]int{
		"viewer":    1,
		"operator":  2,
		"super_admin": 3,
		"nurse":     2,
		"regulator": 3,
	}

	minLevel, ok := roleOrder[minRole]
	if !ok {
		return func(c *gin.Context) { c.Next() } // unknown role, allow through
	}

	return func(c *gin.Context) {
		role, exists := c.Get(string(ContextRole))
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
			return
		}
		roleStr, ok := role.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid role type"})
			c.Abort()
			return
		}

		level, roleOK := roleOrder[roleStr]
		if !roleOK || level < minLevel {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient privileges"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdminRole is an alias for RequireRole with higher privilege threshold
// DEPRECATED: Use RequireRole instead
func (j *AdminJWT) RequireAdminRole(minRole string) gin.HandlerFunc {
	return j.RequireRole(minRole)
}
