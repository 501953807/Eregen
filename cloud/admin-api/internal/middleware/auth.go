package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type ContextKey string

const (
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
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader || tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or malformed token"})
			return
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(j.secret), nil
		})
		if err != nil || !token.Valid {
			j.log.Warn("admin auth failed", zap.String("ip", c.ClientIP()))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			return
		}

		if role, ok := claims["role"].(string); ok {
			c.Set(string(ContextAdminRole), role)
		}
		c.Next()
	}
}

func (j *AdminJWT) RequireAdminRole(minRole string) gin.HandlerFunc {
	roleOrder := map[string]int{"viewer": 1, "operator": 2, "super_admin": 3}
	minLevel := roleOrder[minRole]

	return func(c *gin.Context) {
		role, exists := c.Get(string(ContextAdminRole))
		if !exists {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		level := roleOrder[role.(string)]
		if level < minLevel {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient admin privileges"})
			c.Abort()
			return
		}
		c.Next()
	}
}
