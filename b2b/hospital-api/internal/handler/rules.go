package handler

import (
	"net/http"
	"time"

	"eregen.dev/b2b-hospital-api/internal/model"
	"eregen.dev/b2b-hospital-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type RulesHandler struct {
	store store.Database
	log   *zap.Logger
}

func NewRulesHandler(store store.Database, log *zap.Logger) *RulesHandler {
	return &RulesHandler{store: store, log: log}
}

// POST /api/v2/b2b/hospitals/:id/rules — create medication rule
func (h *RulesHandler) Create(c *gin.Context) {
	instID := c.Param("id")
	var req model.MedicationRuleV2
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ID = uuid.New().String()
	req.BusinessChain = "hospital"
	req.SourceType = "doctor_order"
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	req.Active = true

	if err := h.store.CreateMedicationRule(c.Request.Context(), &req); err != nil {
		h.log.Error("create medication rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create rule"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": req})
}

// GET /api/v2/b2b/hospitals/:id/rules — list medication rules
func (h *RulesHandler) List(c *gin.Context) {
	instID := c.Param("id")
	rules, err := h.store.GetMedicationRulesByInstitution(c.Request.Context(), instID)
	if err != nil {
		h.log.Error("list medication rules", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list rules"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": rules})
}
