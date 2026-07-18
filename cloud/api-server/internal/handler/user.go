package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

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

// NewUserHandler creates a new UserHandler.
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

// GET /api/v1/elderly — list all elderly profiles for the authenticated user
func (h *UserHandler) ListElderly(c *gin.Context) {
	userID, _ := c.Get(string(middleware.ContextUserID))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	profiles, total, err := h.store.ListElderlyProfiles(c.Request.Context(), userID.(string), page, pageSize)
	if err != nil {
		h.log.Error("list elderly profiles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED", "message": "Failed to list profiles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": gin.H{
		"profiles":  profiles,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}})
}

// POST /api/v1/elderly — create a new elderly profile for the authenticated user
func (h *UserHandler) CreateElderly(c *gin.Context) {
	userID, _ := c.Get(string(middleware.ContextUserID))

	var req model.CreateElderlyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "name is required"})
		return
	}

	ep := &model.ElderlyProfile{
		UserID:     userID.(string),
		Name:       req.Name,
		AvatarURL:  req.AvatarURL,
		HealthTiers: req.HealthTiers,
	}
	if req.BirthDate != nil {
		if t, err := time.Parse("2006-01-02", *req.BirthDate); err == nil {
			ep.BirthDate = &t
		}
	}

	if err := h.store.CreateElderlyProfile(c.Request.Context(), ep); err != nil {
		h.log.Error("create elderly profile", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "CREATE_FAILED", "message": "Failed to create profile"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": ep})
}

// GET /api/v1/elderly/:elderly_id/profile
func (h *UserHandler) GetElderlyProfile(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	userID, _ := c.Get(string(middleware.ContextUserID))

	// Verify ownership (ResolveElderlyID already runs as middleware, but double-check here too)
	if !h.checkElderlyAccess(c.Request.Context(), elderlyID, userID.(string)) {
		c.JSON(http.StatusForbidden, gin.H{"code": "ACCESS_DENIED", "message": "You don't have access to this elder"})
		return
	}

	ep, err := h.store.GetElderlyProfile(c.Request.Context(), elderlyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Elderly profile not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": ep})
}

// PUT /api/v1/elderly/:elderly_id/profile
func (h *UserHandler) UpdateElderlyProfile(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	userID, _ := c.Get(string(middleware.ContextUserID))

	if !h.checkElderlyAccess(c.Request.Context(), elderlyID, userID.(string)) {
		c.JSON(http.StatusForbidden, gin.H{"code": "ACCESS_DENIED", "message": "You don't have access to this elder"})
		return
	}

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

// POST /api/v1/elderly/:elderly_id/link-device — bind a device to an elderly profile
func (h *UserHandler) LinkDeviceToElderly(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	userID, _ := c.Get(string(middleware.ContextUserID))

	if !h.checkElderlyAccess(c.Request.Context(), elderlyID, userID.(string)) {
		c.JSON(http.StatusForbidden, gin.H{"code": "ACCESS_DENIED", "message": "You don't have access to this elder"})
		return
	}

	var req model.LinkDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "device_id required"})
		return
	}

	// Ensure device is bound to user first
	device, err := h.store.GetDevice(c.Request.Context(), req.DeviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "DEVICE_NOT_FOUND", "message": "Device not found or not bound to you"})
		return
	}
	if device.OwnerUserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"code": "ACCESS_DENIED", "message": "Device not bound to your account"})
		return
	}

	// Link device to elderly profile via device settings
	settings := device.Settings
	if settings == nil {
		settings = make(map[string]any)
	}
	settings["elderly_id"] = elderlyID
	if err := h.store.UpdateDeviceSettings(c.Request.Context(), req.DeviceID, settings); err != nil {
		h.log.Error("link device to elderly", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "LINK_FAILED", "message": "Failed to link device"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Device linked to elder"})
}

// checkElderlyAccess returns true if the user owns the elderly profile.
func (h *UserHandler) checkElderlyAccess(ctx context.Context, elderlyID, userID string) bool {
	var count int
	err := h.store.Pool().QueryRow(ctx,
		"SELECT COUNT(*) FROM elderly_profiles WHERE id = $1 AND user_id = $2",
		elderlyID, userID,
	).Scan(&count)
	return err == nil && count > 0
}

func sanitizeUser(u *model.User) map[string]any {
	return map[string]any{
		"id":         u.ID,
		"email":      u.Email,
		"phone":      u.Phone,
		"name":       u.Name,
		"role":       u.Role,
		"created_at": u.CreatedAt,
	}
}
