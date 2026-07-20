package handler

import (
	"net/http"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"

	"github.com/gin-gonic/gin"
)

// FirmwareHandler serves firmware version management endpoints.
type FirmwareHandler struct {
	store *store.PostgresStore
}

// NewFirmwareHandler creates a new FirmwareHandler.
func NewFirmwareHandler(s *store.PostgresStore) *FirmwareHandler {
	return &FirmwareHandler{store: s}
}

// List returns all firmware versions.
func (h *FirmwareHandler) List(c *gin.Context) {
	versions, err := h.store.ListFirmwareVersions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": versions})
}

// Create adds a new firmware release record.
func (h *FirmwareHandler) Create(c *gin.Context) {
	var body struct {
		DeviceType    string `json:"device_type" binding:"required"`
		Tier          string `json:"tier" binding:"required"`
		Version       string `json:"version" binding:"required"`
		URL           string `json:"url" binding:"required"`
		Sha256Hash    string `json:"sha256_hash" binding:"required"`
		Changelog     string `json:"changelog"`
		MinAppVersion string `json:"min_app_version"`
		ForceUpdate   bool   `json:"force_update"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	v := &model.FirmwareVersion{
		DeviceType: body.DeviceType, Tier: body.Tier, Version: body.Version,
		DownloadURL: body.URL, Sha256Hash: body.Sha256Hash, Changelog: body.Changelog,
		MinAppVersion: body.MinAppVersion, ForceUpdate: body.ForceUpdate,
	}
	if err := h.store.CreateFirmwareVersion(c.Request.Context(), v); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "message": "firmware version created"})
}

// Delete soft-deletes a firmware release.
func (h *FirmwareHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.store.DeleteFirmwareVersion(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "firmware version deleted"})
}

// PushOTA triggers an OTA push for a firmware version.
func (h *FirmwareHandler) PushOTA(c *gin.Context) {
	var body struct {
		FirmwareID string   `json:"firmware_id" binding:"required"`
		DeviceIDs  []string `json:"device_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.PushOTAJob(c.Request.Context(), body.FirmwareID, body.DeviceIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "OTA push scheduled"})
}
