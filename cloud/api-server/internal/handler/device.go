package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"eregen.dev/api-server/internal/middleware"
	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/service"
	"eregen.dev/api-server/internal/store"
	"eregen.dev/api-server/internal/validation"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DeviceHandler handles device management endpoints.
type DeviceHandler struct {
	store *store.Postgres
	redis *store.Redis
	nats  *service.NatsClient
	log   *zap.Logger
}

// NewDeviceHandler creates a new device handler.
func NewDeviceHandler(store *store.Postgres, redis *store.Redis, nats *service.NatsClient, log *zap.Logger) *DeviceHandler {
	return &DeviceHandler{store: store, redis: redis, nats: nats, log: log}
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

	// Push settings to device via NATS JetStream
	cmd := map[string]any{
		"type":     "config",
		"settings": req.Settings,
	}
	if err := h.nats.PublishCommand(c.Request.Context(), deviceID, cmd); err != nil {
		h.log.Warn("push device settings via NATS", zap.Error(err))
		// Non-fatal: settings are persisted in DB, push is best-effort
	}

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

// POST /api/v1/devices/bind — bind a device to the authenticated user
func (h *DeviceHandler) Bind(c *gin.Context) {
	userID, _ := c.Get(string(middleware.ContextUserID))

	var req model.BindDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "device_id required"})
		return
	}

	// Input validation: device ID format
	if err := validation.DeviceID(req.DeviceID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_DEVICE_ID", "message": err.Error()})
		return
	}

	// Parse device type and tier from device ID (BR-XXXX = bracelet, PX-XXXX = pillbox)
	deviceType := "bracelet"
	tier := "starter"
	switch {
	case strings.HasPrefix(req.DeviceID, "BR-"):
		deviceType = "bracelet"
	case strings.HasPrefix(req.DeviceID, "PX-"):
		deviceType = "pillbox"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_DEVICE_ID", "message": "Device ID must start with BR- or PX-"})
		return
	}

	device, err := h.store.BindDevice(c.Request.Context(), req.DeviceID, userID.(string), deviceType, tier)
	if err != nil {
		// Device already bound to another user
		existing, getErr := h.store.GetDevice(c.Request.Context(), req.DeviceID)
		if getErr == nil && existing.OwnerUserID != userID.(string) {
			c.JSON(http.StatusConflict, gin.H{"code": "DEVICE_BOUND", "message": "Device already bound to another account"})
			return
		}
		h.log.Error("bind device", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "BIND_FAILED", "message": "Failed to bind device"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code": "OK",
		"data": device,
	})
}

// POST /api/v1/devices/telemetry — device sends health/location data
func (h *DeviceHandler) HandleTelemetry(c *gin.Context) {
	var req struct {
		DeviceID string `json:"device_id" binding:"required"`
		Type     string `json:"type" binding:"required"` // health, location
		Data     map[string]any `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "Invalid request body"})
		return
	}

	// Mark device online
	h.redis.SetDeviceOnline(c.Request.Context(), req.DeviceID)

	switch req.Type {
	case "health":
		elderlyID, _ := req.Data["elderly_id"].(string)
		if elderlyID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": "MISSING_ELDERLY_ID", "message": "elderly_id required"})
			return
		}
		record := &model.HealthRecord{
			ElderlyID: elderlyID,
			Timestamp: time.Now(),
		}
		if hr, ok := req.Data["hr"].(float64); ok {
			v := int(hr)
			record.HR = &v
		}
		if spo2, ok := req.Data["spo2"].(float64); ok {
			v := int(spo2)
			record.SPO2 = &v
		}
		if steps, ok := req.Data["steps"].(float64); ok {
			v := int64(steps)
			record.Steps = &v
		}
		if sleep, ok := req.Data["sleep_hours"].(float64); ok {
			record.SleepHours = &sleep
		}
		if bpSys, ok := req.Data["bp_systolic"].(float64); ok {
			v := int(bpSys)
			record.BPSystolic = &v
		}
		if bpDia, ok := req.Data["bp_diastolic"].(float64); ok {
			v := int(bpDia)
			record.BPDiastolic = &v
		}
		if err := h.store.CreateHealthRecord(c.Request.Context(), record); err != nil {
			h.log.Error("create health record", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"code": "FAILED", "message": "Failed to save health data"})
			return
		}
	case "location":
		elderlyID, _ := req.Data["elderly_id"].(string)
		lat, _ := req.Data["lat"].(float64)
		lon, _ := req.Data["lon"].(float64)
		if elderlyID == "" || lat == 0 || lon == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": "MISSING_LOCATION", "message": "elderly_id, lat, lon required"})
			return
		}
		record := &model.LocationRecord{
			ElderlyID: elderlyID,
			Timestamp: time.Now(),
			Lat:       lat,
			Lon:       lon,
		}
		if acc, ok := req.Data["accuracy"].(float64); ok {
			record.Accuracy = &acc
		}
		if err := h.store.CreateLocationRecord(c.Request.Context(), record); err != nil {
			h.log.Error("create location record", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"code": "FAILED", "message": "Failed to save location"})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_TYPE", "message": "type must be health or location"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Data received"})
}

// POST /api/v1/devices/heartbeat — device heartbeat
func (h *DeviceHandler) HandleHeartbeat(c *gin.Context) {
	var req struct {
		DeviceID string `json:"device_id" binding:"required"`
		Battery  *int   `json:"battery,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "Invalid request body"})
		return
	}

	h.redis.SetDeviceOnline(c.Request.Context(), req.DeviceID)

	// Update device last_seen
	settings := map[string]any{}
	if req.Battery != nil {
		settings["battery"] = *req.Battery
	}
	_ = h.store.UpdateDeviceSettings(c.Request.Context(), req.DeviceID, settings)

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Heartbeat received"})
}

// POST /api/v1/devices/location — device location report
func (h *DeviceHandler) HandleLocation(c *gin.Context) {
	var req struct {
		DeviceID string  `json:"device_id" binding:"required"`
		ElderlyID string `json:"elderly_id" binding:"required"`
		Lat      float64 `json:"lat" binding:"required"`
		Lon      float64 `json:"lon" binding:"required"`
		Accuracy *float64 `json:"accuracy,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "Invalid request body"})
		return
	}

	h.redis.SetDeviceOnline(c.Request.Context(), req.DeviceID)

	record := &model.LocationRecord{
		ElderlyID: req.ElderlyID,
		Timestamp: time.Now(),
		Lat:       req.Lat,
		Lon:       req.Lon,
		Accuracy:  req.Accuracy,
	}
	if err := h.store.CreateLocationRecord(c.Request.Context(), record); err != nil {
		h.log.Error("create location record", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAILED", "message": "Failed to save location"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Location saved"})
}
