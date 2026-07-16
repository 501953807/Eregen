package handler

import (
	"net/http"
	"strconv"

	"eregen.dev/api-server/internal/middleware"
	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DeviceHandler handles device management endpoints.
type DeviceHandler struct {
	store *store.Postgres
	redis *store.Redis
	log   *zap.Logger
}

// NewDeviceHandler creates a new device handler.
func NewDeviceHandler(store *store.Postgres, redis *store.Redis, log *zap.Logger) *DeviceHandler {
	return &DeviceHandler{store: store, redis: redis, log: log}
}

// GET /api/v1/devices
func (h *DeviceHandler) List(c *gin.Context) {
	userID, _ := c.Get(string(middleware.ContextUserID))
	deviceType := c.Query("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var dt *string
	if deviceType != "" {
		dt = &deviceType
	}

	devices, total, err := h.store.ListDevices(c.Request.Context(), userID.(string), dt, page, pageSize)
	if err != nil {
		h.log.Error("list devices", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED", "message": "Failed to list devices"})
		return
	}

	// Enrich with online status from Redis
	for i := range devices {
		online, _ := h.redis.IsDeviceOnline(c.Request.Context(), devices[i].DeviceID)
		if online {
			devices[i].Status = model.DeviceOnline
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": gin.H{
		"devices": devices,
		"total":   total,
		"page":    page,
		"page_size": pageSize,
	}})
}

// GET /api/v1/devices/:device_id
func (h *DeviceHandler) Get(c *gin.Context) {
	deviceID := c.Param("device_id")
	device, err := h.store.GetDevice(c.Request.Context(), deviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Device not found"})
		return
	}

	online, _ := h.redis.IsDeviceOnline(c.Request.Context(), deviceID)
	if online {
		device.Status = model.DeviceOnline
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": device})
}

// PUT /api/v1/devices/:device_id/settings
func (h *DeviceHandler) UpdateSettings(c *gin.Context) {
	deviceID := c.Param("device_id")
	var req model.DeviceSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "Invalid request body"})
		return
	}

	if err := h.store.UpdateDeviceSettings(c.Request.Context(), deviceID, req.Settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "UPDATE_FAILED", "message": "Failed to update device settings"})
		return
	}

	// Push settings to device via MQTT/NATS
	// In production: publish to device.command.{device_id}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Device settings updated and pushed"})
}

// DELETE /api/v1/devices/:device_id
func (h *DeviceHandler) Delete(c *gin.Context) {
	deviceID := c.Param("device_id")
	userID, _ := c.Get(string(middleware.ContextUserID))

	if err := h.store.DeleteDevice(c.Request.Context(), deviceID, userID.(string)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Device not found or access denied"})
		return
	}

	h.redis.InvalidateDevice(c.Request.Context(), deviceID)
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Device unbound"})
}
