package handler

import (
	"net/http"
	"strconv"

	"eregen.dev/admin-api/internal/store"
	"eregen.dev/shared/validation"

	"github.com/gin-gonic/gin"
)

// WristbandHandler serves medical wristband device management endpoints.
type WristbandHandler struct {
	store store.Store
}

// NewWristbandHandler creates a new WristbandHandler.
func NewWristbandHandler(s store.Store) *WristbandHandler {
	return &WristbandHandler{store: s}
}

// ListWristbands returns paginated wristband devices.
func (h *WristbandHandler) ListWristbands(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, pageSize, err := validation.ValidatePagination(page, pageSize, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "System internal error"})
		return
	}

	status := c.Query("status")
	devices, err := h.store.ListWristbands(c.Request.Context(), page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":      "OK",
		"data":      devices,
		"page":      page,
		"page_size": pageSize,
	})
}

// BindWristband binds a wristband device to a patient.
func (h *WristbandHandler) BindWristband(c *gin.Context) {
	var body struct {
		PatientID string `json:"patient_id" binding:"required"`
		DeviceID  string `json:"device_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.BindWristband(c.Request.Context(), body.PatientID, body.DeviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "bound"})
}

// UnbindWristband unbinds a wristband device from a patient.
func (h *WristbandHandler) UnbindWristband(c *gin.Context) {
	bindingID := c.Param("id")
	if err := h.store.UnbindWristband(c.Request.Context(), bindingID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "unbound"})
}

// ClearWristband clears all data from a wristband device.
func (h *WristbandHandler) ClearWristband(c *gin.Context) {
	deviceID := c.Param("id")
	if err := h.store.ClearWristband(c.Request.Context(), deviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "cleared"})
}

// WriteToWristband pushes data to a wristband device.
func (h *WristbandHandler) WriteToWristband(c *gin.Context) {
	deviceID := c.Param("id")
	var body struct {
		Data string `json:"data" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.WriteToWristband(c.Request.Context(), deviceID, body.Data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "written"})
}

// GetFirmware returns firmware version for a device.
func (h *WristbandHandler) GetFirmware(c *gin.Context) {
	deviceID := c.Param("id")
	fw, err := h.store.GetWristbandFirmware(c.Request.Context(), deviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "wristband not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": fw})
}
