package handler

import (
	"net/http"
	"time"

	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/service"
	"eregen.dev/api-server/internal/validation"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LocationHandler handles location and geofence endpoints.
type LocationHandler struct {
	svc  *service.LocationService
	log  *zap.Logger
}

// NewLocationHandler creates a new location handler.
func NewLocationHandler(svc *service.LocationService, log *zap.Logger) *LocationHandler {
	return &LocationHandler{svc: svc, log: log}
}

// GET /api/v1/elderly/:elderly_id/location/latest
func (h *LocationHandler) Latest(c *gin.Context) {
	elderlyID := c.Param("elderly_id")

	loc, err := h.svc.GetLatest(c.Request.Context(), elderlyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "No location data found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": loc})
}

// GET /api/v1/elderly/:elderly_id/location/history
func (h *LocationHandler) History(c *gin.Context) {
	elderlyID := c.Param("elderly_id")

	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_DATE", "message": "Use YYYY-MM-DD format"})
		return
	}

	from := date.Truncate(24 * time.Hour)
	until := from.Add(24 * time.Hour)

	records, err := h.svc.GetHistory(c.Request.Context(), elderlyID, from, until)
	if err != nil {
		h.log.Error("get location history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch location data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": records})
}

// POST /api/v1/elderly/:elderly_id/geofence
func (h *LocationHandler) SetGeofence(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	var req struct {
		Lat          float64 `json:"lat" binding:"required"`
		Lon          float64 `json:"lon" binding:"required"`
		RadiusMeters int     `json:"radius_meters" binding:"required,min=50,max=10000"`
		Name         string  `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "Invalid request body"})
		return
	}

	if err := validation.Geofence(req.Name, req.Lat, req.Lon, req.RadiusMeters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_GEOFENCE", "message": err.Error()})
		return
	}

	gf := &model.Geofence{
		ElderlyID:    elderlyID,
		Name:         req.Name,
		Latitude:     req.Lat,
		Longitude:    req.Lon,
		RadiusMeters: req.RadiusMeters,
		Active:       true,
	}

	err := h.svc.CreateGeofence(c.Request.Context(), gf)
	if err != nil {
		h.log.Error("create geofence", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "CREATE_FAILED", "message": "Failed to create geofence"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": gf})
}

// GET /api/v1/elderly/:elderly_id/geofence
func (h *LocationHandler) ListGeofences(c *gin.Context) {
	elderlyID := c.Param("elderly_id")

	fences, err := h.svc.ListGeofences(c.Request.Context(), elderlyID)
	if err != nil {
		h.log.Error("list geofences", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch geofences"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": fences})
}
