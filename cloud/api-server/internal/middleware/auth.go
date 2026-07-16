package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"eregen.dev/api-server/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// ContextKey is the key used to store user info in gin.Context.
type ContextKey string

const (
	ContextUserID    ContextKey = "user_id"
	ContextUserRole  ContextKey = "user_role"
	ContextElderlyID ContextKey = "elderly_id"
	TokenContextKey  ContextKey = "auth_token"
)

// JWTAuth provides JWT-based authentication middleware.
type JWTAuth struct {
	secret     string
	tokenTTL   time.Duration
	refreshTTL time.Duration
	log        *zap.Logger
}

// NewJWTAuth creates an auth middleware with the given secret.
func NewJWTAuth(secret string, tokenTTL, refreshTTL time.Duration, log *zap.Logger) *JWTAuth {
	return &JWTAuth{
		secret:     secret,
		tokenTTL:   tokenTTL,
		refreshTTL: refreshTTL,
		log:        log,
	}
}

// GenerateAccessToken creates a JWT access token for the given user.
func (a *JWTAuth) GenerateAccessToken(userID string, role model.Role) (string, error) {
	return a.generateToken(userID, string(role), a.tokenTTL)
}

// GenerateRefreshToken creates a JWT refresh token.
func (a *JWTAuth) GenerateRefreshToken(userID string) (string, error) {
	return a.generateToken(userID, "", a.refreshTTL)
}

func (a *JWTAuth) generateToken(userID, role string, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(ttl).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.secret))
}

// AuthMiddleware validates the JWT token from the Authorization header.
func (a *JWTAuth) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			a.unauthorized(c, "MISSING_TOKEN", "Authentication token is required")
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			a.unauthorized(c, "INVALID_FORMAT", "Token must use Bearer scheme")
			return
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(a.secret), nil
		})
		if err != nil || !token.Valid {
			a.unauthorized(c, "INVALID_TOKEN", "Invalid or expired token")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			a.unauthorized(c, "INVALID_TOKEN", "Invalid token claims")
			return
		}

		userID, _ := claims["user_id"].(string)
		roleStr, _ := claims["role"].(string)

		c.Set(string(ContextUserID), userID)
		c.Set(string(ContextUserRole), roleStr)
		c.Set(string(TokenContextKey), tokenStr)

		c.Next()
	}
}

// RequireRole returns middleware that enforces specific roles.
func (a *JWTAuth) RequireRole(roles ...model.Role) gin.HandlerFunc {
	roleSet := make(map[string]bool)
	for _, r := range roles {
		roleSet[string(r)] = true
	}

	return func(c *gin.Context) {
		roleStr, exists := c.Get(string(ContextUserRole))
		if !exists || !roleSet[roleStr.(string)] {
			a.forbidden(c, "INSUFFICIENT_ROLE", "This resource requires a different role")
			return
		}
		c.Next()
	}
}

// ResolveElderlyID extracts the elderly_id from URL params and validates access.
// Family users can only access their own elders; institution users can access all.
func (a *JWTAuth) ResolveElderlyID() gin.HandlerFunc {
	return func(c *gin.Context) {
		elderlyID := c.Param("elderly_id")
		if elderlyID == "" {
			a.badRequest(c, "MISSING_ELDERLY_ID", "elderly_id parameter is required")
			return
		}
		c.Set(string(ContextElderlyID), elderlyID)
		c.Next()
	}
}

// ResolveDeviceID extracts the device_id from URL params.
func (a *JWTAuth) ResolveDeviceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceID := c.Param("device_id")
		if deviceID == "" {
			a.badRequest(c, "MISSING_DEVICE_ID", "device_id parameter is required")
			return
		}
		c.Set("device_id", deviceID)
		c.Next()
	}
}

// ResolveAlertID extracts the alert_id from URL params.
func (a *JWTAuth) ResolveAlertID() gin.HandlerFunc {
	return func(c *gin.Context) {
		alertID := c.Param("alert_id")
		if alertID == "" {
			a.badRequest(c, "MISSING_ALERT_ID", "alert_id parameter is required")
			return
		}
		c.Set("alert_id", alertID)
		c.Next()
	}
}

// ResolveRuleID extracts the rule_id from URL params.
func (a *JWTAuth) ResolveRuleID() gin.HandlerFunc {
	return func(c *gin.Context) {
		ruleID := c.Param("rule_id")
		if ruleID == "" {
			a.badRequest(c, "MISSING_RULE_ID", "rule_id parameter is required")
			return
		}
		c.Set("rule_id", ruleID)
		c.Next()
	}
}

func (a *JWTAuth) unauthorized(c *gin.Context, code, msg string) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"code":    code,
		"message": msg,
	})
	c.Abort()
}

func (a *JWTAuth) forbidden(c *gin.Context, code, msg string) {
	c.JSON(http.StatusForbidden, gin.H{
		"code":    code,
		"message": msg,
	})
	c.Abort()
}

func (a *JWTAuth) badRequest(c *gin.Context, code, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"code":    code,
		"message": msg,
	})
	c.Abort()
}

// RefreshTTL returns the refresh token time-to-live duration.
func (a *JWTAuth) RefreshTTL() time.Duration {
	return a.refreshTTL
}

// TokenExpiry returns the access token expiry in seconds.
func (a *JWTAuth) TokenExpiry() int {
	return int(a.tokenTTL.Seconds())
}

// ParseToken parses a JWT token string and returns its claims.
func (a *JWTAuth) ParseToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(a.secret), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}
	return claims, nil
}
