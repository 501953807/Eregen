package middleware

import (
	"net/http"
	"strings"

	"eregen.dev/b2b-hospital-api/internal/model"
	"eregen.dev/b2b-hospital-api/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ContextKey string

const (
	ContextInstitutionID ContextKey = "institution_id"
	ContextAccessLevel   ContextKey = "access_level"
)

// APIKeyAuth validates the X-API-Key header against stored hashes.
func APIKeyAuth(pgStore *store.Postgres, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawKey := c.GetHeader("X-API-Key")
		if rawKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing API key"})
			c.Abort()
			return
		}

		// Strip "Bearer " prefix if present
		key := strings.TrimPrefix(rawKey, "Bearer ")

		inst, err := pgStore.GetInstitutionByAPIKey(c.Request.Context(), key)
		if err != nil {
			log.Debug("invalid API key", zap.String("key_prefix", key[:min(8, len(key))]))
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid or expired API key"})
			c.Abort()
			return
		}

		c.Set(string(ContextInstitutionID), inst.ID)
		c.Set(string(ContextAccessLevel), string(inst.AccessLevel))
		c.Next()
	}
}

// RequireAccess returns a middleware that checks the institution has at least the required level.
func RequireAccess(required model.AccessLevel) gin.HandlerFunc {
	return func(c *gin.Context) {
		level, exists := c.Get(string(ContextAccessLevel))
		if !exists {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		current := level.(model.AccessLevel)

		order := map[model.AccessLevel]int{
			model.AccessEmergency:  1,
			model.AccessRead:       2,
			model.AccessReadWrite:  3,
		}
		if order[current] < order[required] {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient access level"})
			c.Abort()
			return
		}
		c.Next()
	}
}
