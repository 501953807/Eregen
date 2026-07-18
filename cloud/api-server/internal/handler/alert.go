package handler

import (
	"net/http"
	"strconv"

	"eregen.dev/api-server/internal/middleware"
	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/service"
	"eregen.dev/api-server/internal/store"
	"eregen.dev/api-server/internal/validation"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AlertHandler handles alert management endpoints.
type AlertHandler struct {
	store *store.Postgres
	svc   *service.AlertService
	log   *zap.Logger
}

// NewAlertHandler creates a new alert handler.
func NewAlertHandler(store *store.Postgres, svc *service.AlertService, log *zap.Logger) *AlertHandler {
	return &AlertHandler{store: store, svc: svc, log: log}
}

// GET /api/v1/alerts
func (h *AlertHandler) List(c *gin.Context) {
	userID, _ := c.Get(string(middleware.ContextUserID))
	elderIDs := []string{userID.(string)}

	severity := c.Query("severity")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := &model.AlertFilter{}
	if severity != "" {
		s := model.AlertSeverity(severity)
		filter.Severity = &s
	}
	if status != "" {
		st := model.AlertStatus(status)
		filter.Status = &st
	}

	alerts, total, err := h.store.ListAlerts(c.Request.Context(), elderIDs, filter, page, pageSize)
	if err != nil {
		h.log.Error("list alerts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": gin.H{
		"alerts":    alerts,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}})
}

// GET /api/v1/alerts/:alert_id
func (h *AlertHandler) Get(c *gin.Context) {
	alertID := c.Param("alert_id")

	alert, err := h.store.GetAlert(c.Request.Context(), alertID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Alert not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": alert})
}

// PUT /api/v1/alerts/:alert_id
func (h *AlertHandler) Update(c *gin.Context) {
	alertID := c.Param("alert_id")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "status required"})
		return
	}

	if req.Status != "resolved" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_STATUS", "message": "status must be 'resolved'"})
		return
	}

	if err := h.svc.ResolveAlert(c.Request.Context(), alertID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Alert not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Alert resolved"})
}

// POST /api/v1/alerts/sos/call
func (h *AlertHandler) SOSCall(c *gin.Context) {
	var req struct {
		ElderlyID string  `json:"elderly_id" binding:"required"`
		DeviceID  string  `json:"device_id"`
		Lat       float64 `json:"lat"`
		Lon       float64 `json:"lon"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "elderly_id required"})
		return
	}

	if err := validation.Location(req.Lat, req.Lon); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_LOCATION", "message": err.Error()})
		return
	}

	if err := h.svc.CreateSOSAlert(c.Request.Context(), req.ElderlyID, req.DeviceID, req.Lat, req.Lon); err != nil {
		h.log.Error("create SOS alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "ALERT_FAILED", "message": "Failed to create SOS alert"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "message": "SOS alert created and notification sent"})
}
