package handler

import (
	"crypto/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"eregen.dev/api-server/internal/middleware"
	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/service"
	"eregen.dev/api-server/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	store *store.Postgres
	redis *store.Redis
	auth  *middleware.JWTAuth
	sms   *service.SMSProvider
	log   *zap.Logger
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(store *store.Postgres, redis *store.Redis, auth *middleware.JWTAuth, sms *service.SMSProvider, log *zap.Logger) *AuthHandler {
	return &AuthHandler{store: store, redis: redis, auth: auth, sms: sms, log: log}
}

type RegisterRequest struct {
	Phone         *string `json:"phone,omitempty"`
	Email         *string `json:"email,omitempty"`
	Password      string  `json:"password"`
	ConfirmPasswd string  `json:"confirm_password"`
	OTPCode       string  `json:"otp_code"`
	Name          string  `json:"name"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "Invalid request body"})
		return
	}

	if req.Password != req.ConfirmPasswd {
		c.JSON(http.StatusBadRequest, gin.H{"code": "PASSWORD_MISMATCH", "message": "Passwords do not match"})
		return
	}

	if len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "WEAK_PASSWORD", "message": "Password must be at least 8 characters"})
		return
	}

	// Check password complexity
	hasUpper, hasLower, hasDigit := false, false, false
	for _, r := range req.Password {
		switch {
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= '0' && r <= '9':
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		c.JSON(http.StatusBadRequest, gin.H{"code": "WEAK_PASSWORD", "message": "Password must be at least 8 chars with uppercase, lowercase, and digit"})
		return
	}

	if req.Phone == nil && req.Email == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "IDENTIFIER_REQUIRED", "message": "Phone or email required"})
		return
	}

	target := getIdentifier(req)
	if err := h.redis.VerifyOTP(c.Request.Context(), target, req.OTPCode); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "INVALID_OTP", "message": "Invalid or expired OTP code"})
		return
	}

	hashedPW, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.log.Error("hash password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "Registration failed"})
		return
	}

	u := &model.User{
		Name:         req.Name,
		Phone:        req.Phone,
		Email:        req.Email,
		PasswordHash: string(hashedPW),
		Role:         model.RoleFamily,
	}
	if err := h.store.CreateUser(c.Request.Context(), u); err != nil {
		c.JSON(http.StatusConflict, gin.H{"code": "DUPLICATE", "message": "Phone or email already registered"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": map[string]any{"user_id": u.ID}})
}

func getIdentifier(req RegisterRequest) string {
	if req.Phone != nil {
		return *req.Phone
	}
	if req.Email != nil {
		return *req.Email
	}
	return ""
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "Invalid request body"})
		return
	}

	var user *model.User
	identifier := strings.ToLower(req.Identifier)
	if strings.Contains(identifier, "@") {
		var err error
		user, err = h.store.GetUserByEmail(c.Request.Context(), identifier)
		if err != nil {
			user = nil
		}
	}
	if user == nil {
		var err error
		user, err = h.store.GetUserByPhone(c.Request.Context(), identifier)
		if err != nil {
			user = nil
		}
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "INVALID_CREDENTIALS", "message": "Invalid phone/email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "INVALID_CREDENTIALS", "message": "Invalid phone/email or password"})
		return
	}

	accessToken, err := h.auth.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "Failed to generate token"})
		return
	}

	refreshToken, err := h.auth.GenerateRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "Failed to generate refresh token"})
		return
	}

	h.redis.SetRefreshToken(c.Request.Context(), refreshToken, user.ID, h.auth.RefreshTTL())

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": LoginResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    h.auth.TokenExpiry(),
		RefreshToken: refreshToken,
	}})
}

// POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "refresh_token required"})
		return
	}

	userID, err := h.redis.ValidateRefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "INVALID_REFRESH_TOKEN", "message": "Invalid or expired refresh token"})
		return
	}

	user, err := h.store.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "USER_NOT_FOUND", "message": "User not found"})
		return
	}

	accessToken, err := h.auth.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "Failed to generate token"})
		return
	}

	newRefresh, err := h.auth.GenerateRefreshToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "Failed to generate refresh token"})
		return
	}
	h.redis.InvalidateRefreshToken(c.Request.Context(), req.RefreshToken)
	h.redis.SetRefreshToken(c.Request.Context(), newRefresh, userID, h.auth.RefreshTTL())

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": LoginResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    h.auth.TokenExpiry(),
		RefreshToken: newRefresh,
	}})
}

// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Blacklist the current access token so it can't be used again before expiry
	authHeader := c.GetHeader("Authorization")
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenStr != authHeader && tokenStr != "" {
		if claims, err := h.auth.ParseToken(tokenStr); err == nil {
			if uid, ok := claims["user_id"].(string); ok {
				h.redis.SetRefreshToken(c.Request.Context(), tokenStr, uid+"|access", -1)
			}
		}
	}

	// Also blacklist the refresh token if present
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err == nil && req.RefreshToken != "" {
		h.redis.InvalidateRefreshToken(c.Request.Context(), req.RefreshToken)
	}

	// Invalidate all other sessions for this user (force logout everywhere)
	if userID, exists := c.Get(string(middleware.ContextUserID)); exists {
		key := "token:user:" + userID.(string) + ":*"
		h.redis.DelByPattern(c.Request.Context(), key)
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Logged out"})
}

// POST /api/v1/auth/revoke-all-sessions
// Revokes all refresh tokens for the current user except the one used in this request.
func (h *AuthHandler) RevokeAllSessions(c *gin.Context) {
	userID, exists := c.Get(string(middleware.ContextUserID))
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "Not authenticated"})
		return
	}

	// Blacklist all tokens for this user
	key := "token:user:" + userID.(string) + ":*"
	h.redis.DelByPattern(c.Request.Context(), key)

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "All sessions revoked"})
}

// POST /api/v1/auth/device/register
// Registers a device with its fingerprint and returns a device credential.
func (h *AuthHandler) RegisterDevice(c *gin.Context) {
	var req struct {
		DeviceID     string `json:"device_id" binding:"required"`
		DeviceType   string `json:"device_type" binding:"required"` // bracelet | pillbox
		Tier         string `json:"tier"`                           // starter/plus/pro
		Fingerprint  string `json:"fingerprint"`                    // TLS client cert SHA256
		SerialNumber string `json:"serial_number"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "device_id and device_type required"})
		return
	}

	userID, exists := c.Get(string(middleware.ContextUserID))
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "Not authenticated"})
		return
	}

	dev := &model.Device{
		ID:          uuid.New().String(),
		DeviceID:    req.DeviceID,
		DeviceType:  req.DeviceType,
		Tier:        req.Tier,
		OwnerUserID: userID.(string),
		Status:      model.DeviceOffline,
		Settings: map[string]any{
			"fingerprint": req.Fingerprint,
			"serial":      req.SerialNumber,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := h.store.CreateDevice(c.Request.Context(), dev); err != nil {
		c.JSON(http.StatusConflict, gin.H{"code": "DEVICE_EXISTS", "message": "Device already registered"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": map[string]string{
		"device_id":     dev.DeviceID,
		"device_uuid":   dev.ID,
		"registered_at": time.Now().UTC().Format(time.RFC3339),
	}})
}

// POST /api/v1/auth/send-otp
func (h *AuthHandler) SendOTP(c *gin.Context) {
	var req struct {
		Phone *string `json:"phone,omitempty"`
		Email *string `json:"email,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || (req.Phone == nil && req.Email == nil) {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "Phone or email required"})
		return
	}

	code := generateOTP()
	target := ""
	if req.Phone != nil {
		target = *req.Phone
	} else {
		target = *req.Email
	}

	if err := h.sms.SendOTP(c.Request.Context(), target, code, h.redis); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "SEND_FAILED", "message": "Failed to send verification code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Verification code sent"})
}

// POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Phone *string `json:"phone,omitempty"`
		Email *string `json:"email,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "Phone or email required"})
		return
	}

	identifier := ""
	if req.Phone != nil {
		identifier = *req.Phone
	} else if req.Email != nil {
		identifier = *req.Email
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "Phone or email required"})
		return
	}

	var user *model.User
	var err error
	if strings.Contains(identifier, "@") {
		user, err = h.store.GetUserByEmail(c.Request.Context(), identifier)
	} else {
		user, err = h.store.GetUserByPhone(c.Request.Context(), identifier)
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "If the account exists, a reset link has been sent"})
		return
	}

	resetToken := uuid.New().String()
	h.redis.SetResetToken(c.Request.Context(), resetToken, user.ID, 1*time.Hour)

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "If the account exists, a reset link has been sent"})
}

func generateOTP() string {
	buf := make([]byte, 6)
	rand.Read(buf)
	code := uint(0)
	for _, b := range buf {
		code = (code << 8) | uint(b)
	}
	code %= 900000
	code += 100000
	return strconv.Itoa(int(code))
}
