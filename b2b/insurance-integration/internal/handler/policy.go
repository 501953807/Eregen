package handler

import (
	"net/http"
	"strconv"
	"time"

	"eregen.dev/b2b-insurance-integration/internal/model"
	"eregen.dev/b2b-insurance-integration/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PolicyHandler struct {
	store *store.Postgres
	log   *zap.Logger
}

func NewPolicyHandler(store *store.Postgres, log *zap.Logger) *PolicyHandler {
	return &PolicyHandler{store: store, log: log}
}

// POST /api/v2/b2b/policies — register a new insurance policy
func (h *PolicyHandler) Create(c *gin.Context) {
	var req model.Policy
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.store.CreatePolicy(c.Request.Context(), &req); err != nil {
		h.log.Error("create policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create policy"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": req})
}

// GET /api/v2/b2b/policies/elderly/:elderly_id — get policies for an elderly person
func (h *PolicyHandler) GetForElderly(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	policies, err := h.store.GetPoliciesForElderly(c.Request.Context(), elderlyID)
	if err != nil {
		h.log.Error("get policies", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get policies"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": policies})
}

// POST /api/v2/b2b/reminders — schedule a premium reminder
func (h *PolicyHandler) CreateReminder(c *gin.Context) {
	var req struct {
		PolicyID   string    `json:"policy_id" binding:"required"`
		ElderlyID  string    `json:"elderly_id" binding:"required"`
		FamilyID   string    `json:"family_id"`
		RemindDate time.Time `json:"remind_date" binding:"required"`
		Amount     float64   `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	reminder := &model.PremiumReminder{
		PolicyID:   req.PolicyID,
		ElderlyID:  req.ElderlyID,
		FamilyID:   req.FamilyID,
		RemindDate: req.RemindDate,
		Amount:     req.Amount,
	}

	if err := h.store.CreateReminder(c.Request.Context(), reminder); err != nil {
		h.log.Error("create reminder", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create reminder"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": reminder})
}

// GET /api/v2/b2b/reminders/upcoming — get upcoming premium reminders
func (h *PolicyHandler) GetUpcoming(c *gin.Context) {
	days, _ := parseIntParam(c, "days", 30)
	reminders, err := h.store.GetUpcomingReminders(c.Request.Context(), days)
	if err != nil {
		h.log.Error("get reminders", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get reminders"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": reminders})
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
