package handler

import (
	"net/http"
	"strconv"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"
	"eregen.dev/shared/validation"

	"github.com/gin-gonic/gin"
)

type SubscriptionHandler struct {
	store store.SubscriptionStore
}

func NewSubscriptionHandler(s store.SubscriptionStore) *SubscriptionHandler {
	return &SubscriptionHandler{store: s}
}

func (h *SubscriptionHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, pageSize, err := validation.ValidatePagination(page, pageSize, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	status := c.Query("status")
	planTier := c.Query("plan_tier")
	items, err := h.store.ListSubscriptions(c.Request.Context(), page, pageSize, status, planTier)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list subscriptions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": items, "page": page, "page_size": pageSize})
}

func (h *SubscriptionHandler) Get(c *gin.Context) {
	item, err := h.store.GetSubscription(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get subscription"})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": item})
}

func (h *SubscriptionHandler) Create(c *gin.Context) {
	var sub model.SubscriptionItem
	if err := c.ShouldBindJSON(&sub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.CreateSubscription(c.Request.Context(), &sub); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": sub})
}

func (h *SubscriptionHandler) Update(c *gin.Context) {
	var updates map[string]any
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.UpdateSubscription(c.Request.Context(), c.Param("id"), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "updated"})
}

func (h *SubscriptionHandler) Renew(c *gin.Context) {
	var body struct {
		EndDate string `json:"end_date" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_date is required"})
		return
	}
	if err := h.store.RenewSubscription(c.Request.Context(), c.Param("id"), body.EndDate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to renew"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "renewed"})
}
