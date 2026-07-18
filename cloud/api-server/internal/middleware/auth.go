package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// DeviceAuth provides middleware for device-to-cloud mutual authentication.
type DeviceAuth struct {
	store *store.Postgres
	log   *zap.Logger
}

// NewDeviceAuth creates a device auth handler.
func NewDeviceAuth(pg *store.Postgres, log *zap.Logger) *DeviceAuth {
	return &DeviceAuth{store: pg, log: log}
}

// DeviceAuthMiddleware validates device tokens signed by the cloud server.
// Devices present a short-lived JWT with device_id + owner_user_id claims.
func (d *DeviceAuth) DeviceAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("X-Device-Token")
		if tokenStr == "" {
			d.unauthorized(c, "MISSING_DEVICE_TOKEN", "Device token required")
			return
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte("device-secret"), nil // separate signing key for device tokens
		})
		if err != nil || !token.Valid {
			d.unauthorized(c, "INVALID_DEVICE_TOKEN", "Invalid or expired device token")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			d.unauthorized(c, "INVALID_DEVICE_TOKEN", "Invalid device token claims")
			return
		}

		deviceID, _ := claims["device_id"].(string)
		ownerID, _ := claims["owner_id"].(string)

		// Verify device is registered and active
		dev, err := d.store.GetDeviceByDeviceID(c.Request.Context(), deviceID)
		if err != nil || dev == nil {
			d.unauthorized(c, "DEVICE_NOT_FOUND", "Device not registered")
			return
		}
		if dev.Status != model.DeviceOnline {
			d.unauthorized(c, "DEVICE_OFFLINE", "Device is offline")
			return
		}

		c.Set("device_id", deviceID)
		c.Set("device_owner", ownerID)
		c.Next()
	}
}

func (d *DeviceAuth) unauthorized(c *gin.Context, code, msg string) {
	d.log.Warn("device auth failed",
		zap.String("ip", c.ClientIP()),
		zap.String("code", code),
	)
	c.JSON(http.StatusUnauthorized, gin.H{"code": code, "message": msg})
	c.Abort()
}

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
	store      *store.Postgres
}

// NewJWTAuth creates an auth middleware with the given secret.
func NewJWTAuth(secret string, tokenTTL, refreshTTL time.Duration, log *zap.Logger, pg *store.Postgres) *JWTAuth {
	return &JWTAuth{
		secret:     secret,
		tokenTTL:   tokenTTL,
		refreshTTL: refreshTTL,
		log:        log,
		store:      pg,
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

		userID, _ := c.Get(string(ContextUserID))
		roleStr, _ := c.Get(string(ContextUserRole))

		// Institution users can access any elder
		if roleStr == string(model.RoleInstitution) {
			c.Set(string(ContextElderlyID), elderlyID)
			c.Next()
			return
		}

		// Family/elderly users must own the profile
		var count int
		err := a.store.Pool().QueryRow(c.Request.Context(),
			"SELECT COUNT(*) FROM elderly_profiles WHERE id = $1 AND user_id = $2",
			elderlyID, userID.(string),
		).Scan(&count)
		if err != nil || count == 0 {
			a.forbidden(c, "ACCESS_DENIED", "You don't have access to this elder")
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
	a.log.Warn("authentication failed",
		zap.String("ip", c.ClientIP()),
		zap.String("path", c.Request.URL.Path),
		zap.String("code", code),
	)
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
