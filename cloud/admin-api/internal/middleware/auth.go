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
	ContextUserID    ContextKey = "user_id"
	ContextRole      ContextKey = "role"
	ContextAdminRole ContextKey = "admin_role"
)

type AdminJWT struct {
	secret   string
	tokenTTL time.Duration
	log      *zap.Logger
}

func NewAdminJWT(secret string, tokenTTL time.Duration, log *zap.Logger) *AdminJWT {
	return &AdminJWT{secret: secret, tokenTTL: tokenTTL, log: log}
}

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

		if claims.Role == "admin" || claims.Role == "operator" || claims.Role == "nurse" || claims.Role == "regulator" {
			c.Set(string(ContextAdminRole), claims.Role)
		}

		c.Next()
	}
}

func (j *AdminJWT) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get(string(ContextUserID))
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized access"})
			return
		}
		c.Next()
	}
}

func (j *AdminJWT) RequireRole(minRole string) gin.HandlerFunc {
	roleOrder := map[string]int{
		"viewer":      1,
		"operator":    2,
		"super_admin": 3,
		"nurse":       2,
		"regulator":   3,
	}

	minLevel, ok := roleOrder[minRole]
	if !ok {
		return func(c *gin.Context) { c.Next() }
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

func (j *AdminJWT) RequireAdminRole(minRole string) gin.HandlerFunc {
	return j.RequireRole(minRole)
}

// ChainPermissions maps role string → allowed business chains.
var ChainPermissions = map[string][]string{
	"super_admin":     {"self", "hospital", "community", "regulatory"},
	"operator":        {"self", "regulatory"},
	"hospital_doc":    {"hospital"},
	"nurse":           {"hospital"},
	"community_staff": {"community"},
	"regulator":       {"hospital", "community", "regulatory"},
}

// RequireChain returns middleware that checks the user has access to the given business chain.
func (j *AdminJWT) RequireChain(chain string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(string(ContextRole))
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
			return
		}
		roleStr, ok := role.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid role"})
			return
		}
		allowed, ok := ChainPermissions[roleStr]
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "unknown role"})
			return
		}
		hasAccess := false
		for _, c := range allowed {
			if c == chain {
				hasAccess = true
				break
			}
		}
		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied for this business chain"})
			c.Abort()
			return
		}
		c.Next()
	}
}
