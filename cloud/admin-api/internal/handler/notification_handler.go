package handler

import (
	"net/http"
	"strconv"
	"time"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"

	"github.com/gin-gonic/gin"
)

// NotificationHandler manages notification templates and logs.
type NotificationHandler struct {
	store store.NotificationStore
}

func NewNotificationHandler(s store.NotificationStore) *NotificationHandler {
	return &NotificationHandler{store: s}
}

// CreateTemplate adds a notification template.
func (h *NotificationHandler) CreateTemplate(c *gin.Context) {
	var t model.NotificationTemplate
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.CreateNotificationTemplate(c.Request.Context(), &t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK"})
}

// ListTemplates returns notification templates for a chain.
func (h *NotificationHandler) ListTemplates(c *gin.Context) {
	chain := c.Query("chain")
	templates, err := h.store.ListNotificationTemplates(c.Request.Context(), model.BusinessChain(chain))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": templates})
}

// CreateLog records a notification send.
func (h *NotificationHandler) CreateLog(c *gin.Context) {
	var l model.NotificationLog
	if err := c.ShouldBindJSON(&l); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.CreateNotificationLog(c.Request.Context(), &l); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK"})
}

// UpdateStatus updates notification delivery status.
func (h *NotificationHandler) UpdateStatus(c *gin.Context) {
	logID := c.Param("logId")
	var body struct {
		Status string  `json:"status" binding:"required"`
		SentAt *time.Time `json:"sent_at,omitempty"`
		ReadAt *time.Time `json:"read_at,omitempty"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.UpdateNotificationStatus(c.Request.Context(), logID, body.Status, body.SentAt, body.ReadAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK"})
}

// ListLogs returns notification logs for a person.
func (h *NotificationHandler) ListLogs(c *gin.Context) {
	personID := c.Param("personId")
	chain := c.Query("chain")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit > 200 {
		limit = 200
	}
	logs, err := h.store.ListNotificationLogs(c.Request.Context(), personID, model.BusinessChain(chain), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": logs})
}
