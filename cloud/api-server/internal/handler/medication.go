package handler

import (
	"net/http"
	"strconv"
	"time"

	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/service"
	"eregen.dev/api-server/internal/validation"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MedicationHandler handles medication rule and tracking endpoints.
type MedicationHandler struct {
	svc  *service.MedicationService
	log  *zap.Logger
}

// NewMedicationHandler creates a new medication handler.
func NewMedicationHandler(svc *service.MedicationService, log *zap.Logger) *MedicationHandler {
	return &MedicationHandler{svc: svc, log: log}
}

// GET /api/v1/elderly/:elderly_id/medication/rules
func (h *MedicationHandler) ListRules(c *gin.Context) {
	elderlyID := c.Param("elderly_id")

	rules, err := h.svc.ListRules(c.Request.Context(), elderlyID)
	if err != nil {
		h.log.Error("list medication rules", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch rules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": rules})
}

// POST /api/v1/elderly/:elderly_id/medication/rules
func (h *MedicationHandler) CreateRule(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	var req model.CreateMedicationRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "Invalid request body"})
		return
	}

	if err := validateTime(req.ScheduleTime); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_TIME", "message": "schedule_time must be HH:MM format"})
		return
	}

	if err := validation.Medication(req.DoseCount, req.PillType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_MEDICATION", "message": err.Error()})
		return
	}

	if err := validation.DaysOfWeek(req.DaysOfWeek); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_DAYS", "message": "days_of_week: " + err.Error()})
		return
	}

	if err := h.svc.CreateRule(c.Request.Context(), elderlyID, &req); err != nil {
		h.log.Error("create medication rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "CREATE_FAILED", "message": "Failed to create rule"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "message": "Medication rule created and pushed to device"})
}

// PUT /api/v1/elderly/:elderly_id/medication/rules/:rule_id
func (h *MedicationHandler) UpdateRule(c *gin.Context) {
	_ = c.Param("elderly_id") // bound by ResolveElderlyID middleware
	ruleID := c.Param("rule_id")
	var req model.CreateMedicationRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "Invalid request body"})
		return
	}

	if err := validateTime(req.ScheduleTime); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_TIME", "message": "schedule_time must be HH:MM format"})
		return
	}

	if err := validation.Medication(req.DoseCount, req.PillType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_MEDICATION", "message": err.Error()})
		return
	}

	if err := h.svc.UpdateRule(c.Request.Context(), ruleID, &req); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Medication rule not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Medication rule updated"})
}

// DELETE /api/v1/elderly/:elderly_id/medication/rules/:rule_id
func (h *MedicationHandler) DeleteRule(c *gin.Context) {
	_ = c.Param("elderly_id")
	ruleID := c.Param("rule_id")

	if err := h.svc.DeleteRule(c.Request.Context(), ruleID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Medication rule not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Medication rule deleted"})
}

// GET /api/v1/elderly/:elderly_id/medication/today
func (h *MedicationHandler) TodayStatus(c *gin.Context) {
	elderlyID := c.Param("elderly_id")

	records, err := h.svc.GetTodayStatus(c.Request.Context(), elderlyID)
	if err != nil {
		h.log.Error("get today med status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch today's status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": records})
}

// GET /api/v1/elderly/:elderly_id/medication/history
func (h *MedicationHandler) History(c *gin.Context) {
	elderlyID := c.Param("elderly_id")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	records, err := h.svc.GetHistory(c.Request.Context(), elderlyID, days)
	if err != nil {
		h.log.Error("get medication history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "QUERY_FAILED", "message": "Failed to fetch history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": records})
}

func validateTime(t string) error {
	_, err := time.Parse("15:04", t)
	return err
}
