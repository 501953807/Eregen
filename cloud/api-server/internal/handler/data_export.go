package handler

import (
	"net/http"

	"eregen.dev/api-server/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DataExportHandler handles data export and deletion endpoints.
type DataExportHandler struct {
	service *service.DataExportService
	log     *zap.Logger
}

func NewDataExportHandler(svc *service.DataExportService, log *zap.Logger) *DataExportHandler {
	return &DataExportHandler{service: svc, log: log}
}

// CreateExportRequest initiates a data export request.
func (h *DataExportHandler) CreateExportRequest(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "User ID required"})
		return
	}

	req, err := h.service.CreateExportRequest(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("create data export request", zap.String("user_id", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "Failed to create export request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":          "OK",
		"data":          req,
		"download_url":  req.DownloadURL,
		"status":        req.Status,
	})
}

// GetDataExportStatus returns the status of a data export request.
func (h *DataExportHandler) GetDataExportStatus(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "User ID required"})
		return
	}

	req, err := h.service.GetDataExportStatus(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Export request not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":   "OK",
		"data":   req,
		"status": req.Status,
	})
}

// DownloadExport downloads the completed data export archive.
func (h *DataExportHandler) DownloadExport(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "User ID required"})
		return
	}

	req, err := h.service.GetDataExportStatus(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Export request not found"})
		return
	}

	if req.Status != "completed" || req.DownloadURL == "" {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Export not ready"})
		return
	}

	// In production, serve the actual file from storage
	c.Header("Content-Disposition", "attachment; filename=\"eregen_data_export_"+userID+".json\"")
	c.JSON(http.StatusOK, gin.H{
		"code":       "OK",
		"message":    "Download URL generated",
		"download_url": req.DownloadURL,
	})
}

// RequestDeletion initiates a data deletion request.
func (h *DataExportHandler) RequestDeletion(c *gin.Context) {
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "Reason required"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "User ID required"})
		return
	}

	delReq, err := h.service.DeleteUserData(c.Request.Context(), userID, req.Reason)
	if err != nil {
		h.log.Error("delete user data", zap.String("user_id", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "Failed to process deletion request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":   "OK",
		"data":   delReq,
		"status": delReq.Status,
	})
}

// GetDeletionStatus returns the status of a data deletion request.
func (h *DataExportHandler) GetDeletionStatus(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "User ID required"})
		return
	}

	req, err := h.service.GetDeletionStatus(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get deletion status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":   "OK",
		"data":   req,
		"status": req.Status,
	})
}
