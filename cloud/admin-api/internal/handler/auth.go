package handler

import (
	"net/http"

	"eregen.dev/admin-api/internal/auth"
	"eregen.dev/admin-api/internal/store"
	"go.uber.org/zap"
	"github.com/gin-gonic/gin"
)

// AuthHandler serves authentication endpoints.
type AuthHandler struct {
	jwtSecret string
	logger    *zap.Logger
	store     store.Store
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(jwtSecret string, logger *zap.Logger, s store.Store) *AuthHandler {
	return &AuthHandler{jwtSecret: jwtSecret, logger: logger, store: s}
}

// Login handles admin login - verifies credentials and returns JWT token.
// Supports dual login: method=email with username+password, or method=phone with phone+OTP.
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Method     string `json:"method" binding:"required,oneof=email phone"`
		Credential string `json:"credential" binding:"required"`
		Secret     string `json:"secret" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid request body"})
		return
	}

	var userInfo *struct{ ID, Name, Role string }

	// Try DB first, fall back to built-in default users
	if h.store != nil {
		dbUser, err := h.store.GetUserByCredential(c.Request.Context(), req.Method, req.Credential, req.Secret)
		if err == nil {
			userInfo = &struct{ ID, Name, Role string }{ID: dbUser.ID, Name: dbUser.Name, Role: dbUser.Role}
		}
	}

	if userInfo == nil {
		var err error
		userInfo, err = auth.VerifyLogin(req.Method, req.Credential, req.Secret)
		if err != nil {
			h.logger.Warn("login failed", zap.String("method", req.Method), zap.String("credential", req.Credential))
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "invalid credentials"})
			return
		}
	}

	token, err := auth.GenerateToken(userInfo.ID, userInfo.Role, h.jwtSecret, h.logger)
	if err != nil {
		h.logger.Error("failed to generate token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "failed to create authentication token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "login successful",
		"data": auth.LoginResult{
			Token: token,
			User: &auth.UserInfo{
				ID:    userInfo.ID,
				Name:  userInfo.Name,
				Phone: req.Credential,
				Role:  userInfo.Role,
			},
		},
	})
}
