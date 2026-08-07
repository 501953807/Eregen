package handler

import (
	"net/http"

	"eregen.dev/admin-api/internal/auth"
	"eregen.dev/admin-api/internal/store"
	"go.uber.org/zap"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	jwtSecret string
	logger    *zap.Logger
	store     store.Store
}

func NewAuthHandler(jwtSecret string, logger *zap.Logger, s store.Store) *AuthHandler {
	return &AuthHandler{jwtSecret: jwtSecret, logger: logger, store: s}
}

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

	if h.store != nil {
		dbUser, err := h.store.GetUserByCredential(c.Request.Context(), req.Method, req.Credential, req.Secret)
		if err == nil {
			userInfo = &struct{ ID, Name, Role string }{ID: dbUser.ID, Name: dbUser.Name, Role: dbUser.Role}
		} else {
			h.logger.Warn("db auth failed", zap.Error(err))
		}
	}

	if userInfo == nil {
		h.logger.Warn("no auth path succeeded", zap.String("method", req.Method), zap.String("credential", req.Credential))
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "invalid credentials"})
		return
	}

	token, err := auth.GenerateToken(userInfo.ID, userInfo.Role, h.jwtSecret)
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
