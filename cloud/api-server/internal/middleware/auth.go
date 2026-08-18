package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"github.com/redis/go-redis/v9"
)

// CSRFToken provides cross-site request forgery protection.
type CSRFToken struct {
	signKey []byte
	client  *redis.Client // Direct Redis access for storage
	ttl     time.Duration
}

// NewCSRFToken creates a CSRF token manager.
func NewCSRFToken(signKey []byte, redisClient *redis.Client, ttl time.Duration) *CSRFToken {
	if len(signKey) == 0 {
		signKey = []byte("csrf-secret-change-in-production")
	}
	return &CSRFToken{
		signKey: signKey,
		client:  redisClient,
		ttl:     ttl,
	}
}

// GenerateCSRFToken generates a random CSRF token for the user session.
func (c *CSRFToken) GenerateCSRFToken(userID string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(b)
	// Store token signed with secret in Redis linked to user ID
	signature := HMACSign(c.signKey, token)
	key := "csrf:" + userID
	_, err := c.client.Set(context.Background(), key, token+"|"+signature, c.ttl).Result()
	if err != nil {
		return "", fmt.Errorf("redis set failed: %w", err)
	}
	return token, nil
}

// ValidateCSRFToken validates a CSRF token against the stored signature for a user.
func (c *CSRFToken) ValidateCSRFToken(ctx context.Context, userID, token string) bool {
	if token == "" {
		return false
	}
	key := "csrf:" + userID
	val, err := c.client.Get(ctx, key).Result()
	if err != nil || val == "" {
		return false
	}
	parts := strings.Split(val, "|")
	if len(parts) != 2 {
		return false
	}
	storedToken := parts[0]
	if storedToken != token {
		return false
	}
	expectedSignature := HMACSign(c.signKey, token)
	return parts[1] == expectedSignature
}

// HMACSign computes HMAC-SHA256 signature for a token.
func HMACSign(signKey []byte, token string) string {
	h := hmac.New(sha256.New, signKey)
	h.Write([]byte(token))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

// DeviceAuth provides middleware for device-to-cloud mutual authentication.
type DeviceAuth struct {
	store      store.Store
	log        *zap.Logger
	deviceKey  []byte // HMAC key for device JWT signing
}

// NewDeviceAuth creates a device auth handler.
func NewDeviceAuth(pg store.Store, log *zap.Logger, deviceSecret string) *DeviceAuth {
	if len(deviceSecret) == 0 {
		log.Fatal("DEVICE_SECRET environment variable is required for device authentication")
	}
	return &DeviceAuth{store: pg, log: log, deviceKey: []byte(deviceSecret)}
}

// DeviceAuthMiddleware validates device tokens signed by the cloud server.
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
			return d.deviceKey, nil
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

// JWTAuth provides JWT-based authentication supporting both header and cookie modes.
type JWTAuth struct {
	secret     string
	tokenTTL   time.Duration
	refreshTTL time.Duration
	log        *zap.Logger
	store      store.Store
	csrf       *CSRFToken
}

// NewJWTAuth creates an auth middleware with the given secret.
func NewJWTAuth(secret string, tokenTTL, refreshTTL time.Duration, log *zap.Logger, pg store.Store, redisClient *redis.Client, csrfSecret string, csrfTTL time.Duration) *JWTAuth {
csrf := NewCSRFToken([]byte(csrfSecret), redisClient, csrfTTL)
	return &JWTAuth{
		secret:     secret,
		tokenTTL:   tokenTTL,
		refreshTTL: refreshTTL,
		log:        log,
		store:      pg,
		csrf:       csrf,
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
	nbf := time.Now().Add(-5 * time.Second)
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(ttl).Unix(),
		"iat":     time.Now().Unix(),
		"nbf":     nbf.Unix(),
		"jti":     uuid.New().String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.secret))
}

// AuthMiddleware validates the JWT token either from Authorization header OR from httpOnly cookie.
func (a *JWTAuth) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try httpOnly cookie first (for browser clients)
		cookie, err := c.Cookie("access_token")
		if err == nil {
			userID, role, errMsg := a.validateTokenFromCookie(cookie)
			if errMsg == "" {
				c.Set(string(ContextUserID), userID)
				c.Set(string(ContextUserRole), role)
				c.Next()
				return
			}
		}

		// Fallback to Authorization header (for API clients, mobile apps, etc.)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenStr != authHeader {
				userID, role, errMsg := a.validateToken(tokenStr)
				if errMsg == "" {
					c.Set(string(ContextUserID), userID)
					c.Set(string(ContextUserRole), role)
					c.Next()
					return
				} else {
					a.unauthorized(c, "INVALID_TOKEN", errMsg)
					return
				}
			}
		}

		// Neither method succeeded
		a.unauthorized(c, "MISSING_TOKEN", "Authentication token is required")
		c.Abort()
	}
}

// validateTokenFromCookie validates JWT from httpOnly access_token cookie.
// Returns (userID, role, errorMessage) - empty means invalid/fail.
func (a *JWTAuth) validateTokenFromCookie(token string) (string, string, string) {
	// Parse the token
	parsed, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(a.secret), nil
	})
	if err != nil || !parsed.Valid {
		return "", "", "invalid token"
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", "invalid token claims"
	}

	// Validate nbf (not before)
	if nbf, ok := claims["nbf"].(float64); ok && time.Now().Unix() < int64(nbf) {
		return "", "", "token not yet valid"
	}

	userID, ok2 := claims["user_id"].(string)
	if !ok2 {
		return "", "", "missing user_id claim"
	}

	role, ok2 := claims["role"].(string)
	if !ok2 {
		return "", "", "missing role claim in token"
	}

	return userID, role, ""
}

// validateToken validates JWT token string (from header or other source).
// Returns (userID, role, errorMessage).
func (a *JWTAuth) validateToken(tokenStr string) (string, string, string) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(a.secret), nil
	})
	if err != nil || !token.Valid {
		return "", "", "invalid or expired token"
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", "invalid token claims"
	}

	// Validate nbf (not before)
	if nbf, ok := claims["nbf"].(float64); ok && time.Now().Unix() < int64(nbf) {
		return "", "", "token not yet valid"
	}

	userID, ok2 := claims["user_id"].(string)
	if !ok2 {
		return "", "", "missing user_id claim"
	}

	role, ok2 := claims["role"].(string)
	if !ok2 {
		return "", "", "missing role claim in token"
	}

	return userID, role, ""
}

// RequireRole returns middleware that enforces specific roles.
func (a *JWTAuth) RequireRole(roles ...model.Role) gin.HandlerFunc {
	roleSet := make(map[string]bool)
	for _, r := range roles {
		roleSet[string(r)] = true
	}

	return func(c *gin.Context) {
		// Get the role from context, checking for existence
		if rawVal, exists := c.Get(string(ContextUserRole)); exists {
			roleStr, ok := rawVal.(string)
			if ok && roleSet[roleStr] {
				c.Next()
				return
			}
		}
		a.forbid(c, "INSUFFICIENT_ROLE", "This resource requires a different role")
	}
}

// ResolveElderlyID extracts the elderly_id from URL params and validates access.
func (a *JWTAuth) ResolveElderlyID() gin.HandlerFunc {
	return func(c *gin.Context) {
		elderlyID := c.Param("elderly_id")
		if elderlyID == "" {
			a.badRequest(c, "MISSING_ELDERLY_ID", "elderly_id parameter is required")
			return
		}

		// Check if authenticated - get user ID from context
		var userID string
		if rawVal, exists := c.Get(string(ContextUserID)); exists {
			userID, _ = rawVal.(string)
		}
		if userID == "" {
			// No auth, should have been caught by AuthMiddleware but safe to fail
			a.forbid(c, "UNAUTHORIZED", "User must be authenticated")
			return

		}

		// Check role - get user role from context
		var roleStr string
		if rawVal, exists := c.Get(string(ContextUserRole)); exists {
			roleStr, _ = rawVal.(string)
		}

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
			elderlyID, userID,
		).Scan(&count)
		if err != nil || count == 0 {
			a.forbid(c, "ACCESS_DENIED", "You don't have access to this elder")
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

// CSRFCheck middleware validates CSRF token for state-changing requests.
func (a *JWTAuth) CSRFCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to non-GET requests that modify state
		if c.Request.Method != "POST" && c.Request.Method != "PUT" && c.Request.Method != "DELETE" {
			c.Next()
			return
		}

		// Check if we have an authenticated user - get user ID from context correctly
		var userID string
		if rawVal, exists := c.Get(string(ContextUserID)); exists {
			if v, ok := rawVal.(string); ok && v != "" {
				userID = v
			} else {
				userID = ""
			}
		}

		if userID == "" {
			c.Next() // Not authenticated, skip CSRF check
			return
		}

		// Try to get CSRF token from header
		csrfToken := c.GetHeader("X-CSRF-Token")
		if csrfToken == "" {
			csrfToken = c.GetHeader("X-XSRF-Token")
		}

		if csrfToken == "" {
			a.forbid(c, "MISSING_CSRF_TOKEN", "CSRF token required")
			c.Abort()
			return
		}

		// Validate the token
		if !a.csrf.ValidateCSRFToken(c.Request.Context(), userID, csrfToken) {
			a.forbid(c, "INVALID_CSRF_TOKEN", "Invalid or expired CSRF token")
			c.Abort()
			return
		}

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

func (a *JWTAuth) forbid(c *gin.Context, code, msg string) {
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

// GetCSRFToken returns the CSRF token instance for external access.
func (a *JWTAuth) GetCSRFToken() *CSRFToken {
	return a.csrf
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
