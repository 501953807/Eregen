package handler

import (
	"net/http"

	"eregen.dev/api-server/internal/middleware"
	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UserHandler handles user profile endpoints.
type UserHandler struct {
	store *store.Postgres
	redis *store.Redis
	log   *zap.Logger
}

// NewUserHandler creates a new user handler.
func NewUserHandler(store *store.Postgres, redis *store.Redis, log *zap.Logger) *UserHandler {
	return &UserHandler{store: store, redis: redis, log: log}
}

// GET /api/v1/users/me
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, _ := c.Get(string(middleware.ContextUserID))
	user, err := h.store.GetUserByID(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "USER_NOT_FOUND", "message": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": sanitizeUser(user)})
}

// PUT /api/v1/users/me
func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID, _ := c.Get(string(middleware.ContextUserID))
	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "Invalid request body"})
		return
	}

	if err := h.store.UpdateUser(c.Request.Context(), userID.(string), req.Name, req.Phone, req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "UPDATE_FAILED", "message": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Profile updated"})
}

// GET /api/v1/users/:elderly_id/profile
func (h *UserHandler) GetElderlyProfile(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	ep, err := h.store.GetElderlyProfile(c.Request.Context(), elderlyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Elderly profile not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": ep})
}

// PUT /api/v1/users/:elderly_id/profile
func (h *UserHandler) UpdateElderlyProfile(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	var req model.UpdateElderlyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "Invalid request body"})
		return
	}

	if err := h.store.UpdateElderlyProfile(c.Request.Context(), elderlyID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "UPDATE_FAILED", "message": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Profile updated"})
}

func sanitizeUser(u *model.User) map[string]any {
	return map[string]any{
		"id":        u.ID,
		"email":     u.Email,
		"phone":     u.Phone,
		"name":      u.Name,
		"role":      u.Role,
		"created_at": u.CreatedAt,
	}
}
