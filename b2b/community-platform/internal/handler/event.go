package handler

import (
	"net/http"

	"eregen.dev/b2b-community-platform/internal/model"
	"eregen.dev/b2b-community-platform/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type EventHandler struct {
	store *store.Postgres
	log   *zap.Logger
}

func NewEventHandler(store *store.Postgres, log *zap.Logger) *EventHandler {
	return &EventHandler{store: store, log: log}
}

// POST /api/v2/b2b/events — create a community event
func (h *EventHandler) Create(c *gin.Context) {
	var req model.CommunityEvent
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.store.CreateEvent(c.Request.Context(), &req); err != nil {
		h.log.Error("create event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create event"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": req})
}

// GET /api/v2/b2b/events — list events with optional service type filter
func (h *EventHandler) List(c *gin.Context) {
	serviceType := model.ServiceType(c.Query("service_type"))
	page, _ := parseIntParam(c, "page", 1)
	pageSize, _ := parseIntParam(c, "page_size", 20)

	list, total, err := h.store.ListEvents(c.Request.Context(), serviceType, page, pageSize)
	if err != nil {
		h.log.Error("list events", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": list, "total": total, "page": page})
}

// POST /api/v2/b2b/events/:id/register — register elderly for an event
func (h *EventHandler) Register(c *gin.Context) {
	eventID := c.Param("id")
	var req struct {
		ElderlyID   string  `json:"elderly_id" binding:"required"`
		CaregiverID *string `json:"caregiver_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Check capacity
	count, err := h.store.ActiveRegistrationsCount(c.Request.Context(), eventID)
	if err != nil {
		h.log.Error("check capacity", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check capacity"})
		return
	}

	event, err := h.store.GetEventByID(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	if count >= event.MaxParticipants {
		c.JSON(http.StatusConflict, gin.H{"error": "event is full"})
		return
	}

	reg := &model.EventRegistration{
		EventID:     eventID,
		ElderlyID:   req.ElderlyID,
		CaregiverID: req.CaregiverID,
	}

	if err := h.store.RegisterForEvent(c.Request.Context(), reg); err != nil {
		h.log.Error("register for event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": reg})
}

// POST /api/v2/b2b/events/:id/cancel-register — cancel registration
func (h *EventHandler) CancelRegister(c *gin.Context) {
	eventID := c.Param("id")
	var req struct {
		ElderlyID string `json:"elderly_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.store.CancelEventRegistration(c.Request.Context(), eventID, req.ElderlyID); err != nil {
		h.log.Error("cancel registration", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel registration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Registration cancelled"})
}

// GET /api/v2/b2b/events/:id/registrations — get registrations for an event
func (h *EventHandler) GetRegistrations(c *gin.Context) {
	eventID := c.Param("id")
	regs, err := h.store.GetRegistrationsForEvent(c.Request.Context(), eventID)
	if err != nil {
		h.log.Error("get registrations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get registrations"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": regs})
}

// GET /api/v2/b2b/events/:id — get one event
func (h *EventHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
 evt, err := h.store.GetEventByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": evt})
}

// DELETE /api/v2/b2b/events/:id — delete an event
func (h *EventHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.store.DeleteEvent(c.Request.Context(), id); err != nil {
		h.log.Error("delete event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete event"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Event deleted"})
}
