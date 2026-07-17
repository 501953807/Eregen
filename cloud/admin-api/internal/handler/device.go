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
