package handler

import (
	"net/http"

	"eregen.dev/admin-api/internal/auth"
	"go.uber.org/zap"
	"github.com/gin-gonic/gin"
)

// AuthHandler serves authentication endpoints.
type AuthHandler struct {
	jwtSecret string
	logger    *zap.Logger
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(jwtSecret string, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{jwtSecret: jwtSecret, logger: logger}
}

// Login handles admin login - verifies credentials and returns JWT token.
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Verify credentials against default admin user
	if !auth.VerifyLogin(req.Username, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(auth.DefaultAdminUser.Username, auth.DefaultAdminUser.Role, h.jwtSecret, h.logger)
	if err != nil {
		h.logger.Error("failed to generate token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create authentication token"})
		return
	}

	// Return token and user info
	c.JSON(http.StatusOK, gin.H{
		"data": auth.LoginResult{
			Token: token,
			User: &auth.UserInfo{
				Username: auth.DefaultAdminUser.Username,
				Role:     auth.DefaultAdminUser.Role,
			},
		},
	})
}
