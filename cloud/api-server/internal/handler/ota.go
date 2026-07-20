package handler

import (
	"net/http"

	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// OTAHandler handles firmware OTA endpoints.
type OTAHandler struct {
	svc *service.OTAService
	log *zap.Logger
}

// NewOTAHandler creates a new OTA handler.
func NewOTAHandler(svc *service.OTAService, log *zap.Logger) *OTAHandler {
	return &OTAHandler{svc: svc, log: log}
}

// POST /api/v1/admin/firmware — create a new firmware release
func (h *OTAHandler) CreateFirmware(c *gin.Context) {
	var req model.CreateFirmwareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "device_type, tier, version, url, sha256_hash required"})
		return
	}

	release, err := h.svc.CreateFirmwareRelease(c.Request.Context(), &req)
	if err != nil {
		h.log.Error("create firmware release", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "Failed to create firmware release"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": release})
}

// GET /api/v1/admin/firmware — list firmware releases
func (h *OTAHandler) ListFirmware(c *gin.Context) {
	deviceType := c.Query("device_type")
	tier := c.Query("tier")

	releases, err := h.svc.ListFirmwareReleases(c.Request.Context(), deviceType, tier)
	if err != nil {
		h.log.Error("list firmware", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "Failed to list firmware"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": releases})
}

// GET /api/v1/admin/firmware/:id — get single firmware release
func (h *OTAHandler) GetFirmware(c *gin.Context) {
	id := c.Param("id")

	release, err := h.svc.GetFirmwareRelease(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Firmware release not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": release})
}

// POST /api/v1/admin/ota/push — trigger an OTA push
func (h *OTAHandler) PushOTA(c *gin.Context) {
	var req model.PushOTARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "firmware_id required"})
		return
	}

	firmwareID := req.FirmwareID
	if firmwareID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "firmware_id required"})
		return
	}

	release, err := h.svc.GetFirmwareRelease(c.Request.Context(), firmwareID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Firmware release not found"})
		return
	}

	// Resolve target devices
	var deviceIDs []string
	if len(req.DeviceIDs) > 0 {
		deviceIDs = req.DeviceIDs
	} else {
		// Match all devices of the same type and tier
		devices, _ := h.svc.GetMatchingDevices(c.Request.Context(), release.DeviceType, release.Tier)
		for _, d := range devices {
			deviceIDs = append(deviceIDs, d.DeviceID)
		}
	}

	if len(deviceIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "NO_TARGET_DEVICES", "message": "No matching devices found"})
		return
	}

	job, err := h.svc.CreateOTAJob(c.Request.Context(), firmwareID, deviceIDs)
	if err != nil {
		h.log.Error("create OTA job", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "Failed to create OTA job"})
		return
	}

	// Push commands to all target devices
	go func() {
		ctx := c.Request.Context()
		if err := h.svc.PushToDevices(ctx, job, release); err != nil {
			h.log.Error("push OTA devices", zap.Error(err))
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"code":    "OK",
		"message": "OTA push initiated",
		"data": gin.H{
			"job_id":           job.ID,
			"target_count":     len(deviceIDs),
			"firmware_version": release.Version,
		},
	})
}

// GET /api/v1/admin/ota/jobs/:id — get OTA job status
func (h *OTAHandler) GetOTAJob(c *gin.Context) {
	id := c.Param("id")

	job, err := h.svc.GetOTAJob(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "OTA job not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": job})
}

// POST /api/v1/admin/firmware/:id/verify — verify firmware signature
func (h *OTAHandler) VerifyFirmware(c *gin.Context) {
	id := c.Param("id")

	valid, status, err := h.svc.VerifyFirmwareSignature(c.Request.Context(), id)
	if err != nil {
		h.log.Error("verify firmware failed", zap.String("firmware_id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "Failed to verify firmware signature"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "OK",
		"data": gin.H{
			"firmware_id": id,
			"valid":       valid,
			"status":      status,
		},
	})
}
