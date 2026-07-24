package handler

import (
	"net/http"
	"strconv"

	"eregen.dev/b2b-community-platform/internal/model"
	"eregen.dev/b2b-community-platform/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CarePlanHandler struct {
	store *store.Postgres
	log   *zap.Logger
}

func NewCarePlanHandler(store *store.Postgres, log *zap.Logger) *CarePlanHandler {
	return &CarePlanHandler{store: store, log: log}
}

// POST /api/v2/b2b/care-plans — create a care plan
func (h *CarePlanHandler) Create(c *gin.Context) {
	var req model.CarePlan
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.store.CreateCarePlan(c.Request.Context(), &req); err != nil {
		h.log.Error("create care plan", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create care plan"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": req})
}

// GET /api/v2/b2b/care-plans/:elderly_id — get care plans for elderly person
func (h *CarePlanHandler) GetForElderly(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	plans, err := h.store.GetCarePlansForElderly(c.Request.Context(), elderlyID)
	if err != nil {
		h.log.Error("get care plans", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get care plans"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": plans})
}

// GET /api/v2/b2b/care-plans/:id — get one care plan
func (h *CarePlanHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	plan, err := h.store.GetCarePlanByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "care plan not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": plan})
}

// PUT /api/v2/b2b/care-plans/:id — update a care plan
func (h *CarePlanHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req model.CarePlan
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.store.UpdateCarePlan(c.Request.Context(), id, &req); err != nil {
		h.log.Error("update care plan", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update care plan"})
		return
	}
	req.ID = id
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": req})
}

// DELETE /api/v2/b2b/care-plans/:id — delete a care plan
func (h *CarePlanHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.store.DeleteCarePlan(c.Request.Context(), id); err != nil {
		h.log.Error("delete care plan", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete care plan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Care plan deleted"})
}

func parseIntParam(c *gin.Context, key string, defaultVal int) (int, bool) {
	v := c.Query(key)
	if v == "" {
		return defaultVal, false
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal, false
	}
	return n, true
}
