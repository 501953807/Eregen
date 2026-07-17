package handler

import (
	"net/http"
	"strconv"

	"eregen.dev/admin-api/internal/store"
	"eregen.dev/shared/validation"

	"github.com/gin-gonic/gin"
)

// DeviceHandler serves device management endpoints.
type DeviceHandler struct {
	store *store.PostgresStore
}

// NewDeviceHandler creates a new DeviceHandler.
func NewDeviceHandler(s *store.PostgresStore) *DeviceHandler {
	return &DeviceHandler{store: s}
}

// UpdateConfig updates device settings and triggers NATS push.
func (h *DeviceHandler) UpdateConfig(c *gin.Context) {
	deviceID := c.Param("id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.UpdateDeviceConfig(c.Request.Context(), deviceID, body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "config updated"})
}

// TriggerOTA schedules an OTA firmware update for a device.
func (h *DeviceHandler) TriggerOTA(c *gin.Context) {
	deviceID := c.Param("id")
	var body struct {
		URL      string `json:"url" binding:"required"`
		Hash256  string `json:"hash" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.TriggerOTA(c.Request.Context(), deviceID, body.URL, body.Hash256); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "OTA scheduled"})
}

// List returns a paginated list of devices with optional filters.
func (h *DeviceHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, pageSize, err := validation.ValidatePagination(page, pageSize, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := c.Query("status")
	devType := c.Query("type")
	tier := c.Query("tier")

	devices, err := h.store.ListDevices(c.Request.Context(), page, pageSize, status, devType, tier)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      devices,
		"page":      page,
		"page_size": pageSize,
	})
}
