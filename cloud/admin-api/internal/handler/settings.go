package handler

import (
	"net/http"
	"time"

	"eregen.dev/admin-api/internal/auth"
	"eregen.dev/admin-api/internal/middleware"
	"eregen.dev/admin-api/internal/store"

	"github.com/gin-gonic/gin"
)

// SettingsHandler serves system settings endpoints.
type SettingsHandler struct {
	store store.Store
}

// NewSettingsHandler creates a new SettingsHandler.
func NewSettingsHandler(s store.Store) *SettingsHandler {
	return &SettingsHandler{store: s}
}

// GetNotificationSettings retrieves notification config.
func (h *SettingsHandler) GetNotificationSettings(c *gin.Context) {
	settings, err := h.store.GetNotificationSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": settings})
}

// UpdateNotificationSettings persists notification config.
func (h *SettingsHandler) UpdateNotificationSettings(c *gin.Context) {
	var body map[string]any
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.UpdateNotificationSettings(c.Request.Context(), body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "settings updated"})
}

// ListAPIKeys returns registered B2B API keys.
func (h *SettingsHandler) ListAPIKeys(c *gin.Context) {
	keys, err := h.store.ListAPIKeys(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": keys})
}

// CreateAPIKey registers a new B2B API key.
func (h *SettingsHandler) CreateAPIKey(c *gin.Context) {
	var body struct {
		Name      string  `json:"name"`
		ExpiresAt *string `json:"expires_at,omitempty"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	// Generate a random key hash (in production use a proper crypto rand + hash)
	keyHash := "placeholder-hash-" + body.Name
	var expiresAt *time.Time
	if body.ExpiresAt != nil && *body.ExpiresAt != "" {
		if t, err := time.Parse("2006-01-02", *body.ExpiresAt); err == nil {
			expiresAt = &t
		}
	}
	id, err := h.store.CreateAPIKey(c.Request.Context(), body.Name, keyHash, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": gin.H{"id": id}})
}

// RevokeAPIKey deactivates a B2B API key.
func (h *SettingsHandler) RevokeAPIKey(c *gin.Context) {
	id := c.Param("id")
	if err := h.store.RevokeAPIKey(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "API key revoked"})
}

// ChangePassword updates the authenticated admin user's password.
func (h *SettingsHandler) ChangePassword(c *gin.Context) {
	userID := c.GetString(string(middleware.ContextUserID))
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var body struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	hash, err := auth.HashPassword(body.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	if err := h.store.ChangeAdminPassword(c.Request.Context(), userID, hash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "password updated"})
}
