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
	tokenStr := c.GetHeader("Authorization")
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

	claims, err := h.auth.ParseToken(tokenStr)
	if err == nil {
		if uid, ok := claims["user_id"].(string); ok {
			h.redis.SetRefreshToken(c.Request.Context(), tokenStr, uid, -1)
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Logged out"})
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
