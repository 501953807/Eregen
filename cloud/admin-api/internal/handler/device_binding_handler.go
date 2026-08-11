package handler

import (
	"net/http"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"

	"github.com/gin-gonic/gin"
)

// DeviceBindingHandler manages device bindings between persons and devices.
type DeviceBindingHandler struct {
	store store.DeviceBindingStore
}

func NewDeviceBindingHandler(s store.DeviceBindingStore) *DeviceBindingHandler {
	return &DeviceBindingHandler{store: s}
}

// Bind binds a device to a person.
func (h *DeviceBindingHandler) Bind(c *gin.Context) {
	var b model.DeviceBinding
	if err := c.ShouldBindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.BindDevice(c.Request.Context(), &b); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK"})
}

// Unbind removes a device binding.
func (h *DeviceBindingHandler) Unbind(c *gin.Context) {
	bindingID := c.Param("bindingId")
	if err := h.store.UnbindDevice(c.Request.Context(), bindingID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK"})
}

// ListBindings returns device bindings for a person.
func (h *DeviceBindingHandler) ListBindings(c *gin.Context) {
	personID := c.Param("personId")
	chain := c.Query("chain")
	bindings, err := h.store.ListDeviceBindings(c.Request.Context(), personID, model.BusinessChain(chain))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": bindings})
}

// ListDevices returns devices bound to a person.
func (h *DeviceBindingHandler) ListDevices(c *gin.Context) {
	personID := c.Param("personId")
	devices, err := h.store.ListDevicesByPerson(c.Request.Context(), personID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": devices})
}
