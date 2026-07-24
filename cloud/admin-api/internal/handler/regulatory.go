package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"
	"eregen.dev/shared/validation"

	"github.com/gin-gonic/gin"
)

type RegulatoryHandler struct {
	store store.Store
}

func NewRegulatoryHandler(s store.Store) *RegulatoryHandler {
	return &RegulatoryHandler{store: s}
}

// GetDashboardOverview returns summary stats for the regulatory dashboard.
func (h *RegulatoryHandler) GetDashboardOverview(c *gin.Context) {
	dept := c.Query("department")
	ov, err := h.store.GetRegulatoryOverview(c.Request.Context(), dept)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": ov})
}

// ListRegulatoryPatients returns the patient list with fence/alert status.
func (h *RegulatoryHandler) ListRegulatoryPatients(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, pageSize, err := validation.ValidatePagination(page, pageSize, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dept := c.Query("department")
	patients, err := h.store.ListRegulatoryPatients(c.Request.Context(), dept, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": patients, "page": page, "page_size": pageSize})
}

// ListAlerts returns regulatory alerts with filtering.
func (h *RegulatoryHandler) ListAlerts(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, pageSize, err := validation.ValidatePagination(page, pageSize, 200)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	alerts, err := h.store.ListRegulatoryAlerts(c.Request.Context(),
		c.Query("rule_code"), c.Query("level"), c.Query("status"), c.Query("department"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": alerts, "page": page, "page_size": pageSize})
}

// GetAlert returns a single regulatory alert.
func (h *RegulatoryHandler) GetAlert(c *gin.Context) {
	alert, err := h.store.GetRegulatoryAlert(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": alert})
}

// AcknowledgeAlert confirms a regulatory alert.
func (h *RegulatoryHandler) AcknowledgeAlert(c *gin.Context) {
	var body struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.store.AcknowledgeAlert(c.Request.Context(), c.Param("id"), body.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "acknowledged"})
}

// ResolveRegulatoryAlert marks a regulatory alert as resolved.
func (h *RegulatoryHandler) ResolveRegulatoryAlert(c *gin.Context) {
	var body struct {
		UserID string `json:"user_id" binding:"required"`
		Notes  string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.store.ResolveRegulatoryAlert(c.Request.Context(), c.Param("id"), body.UserID, body.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "resolved"})
}

// CreateRegulatoryAlert creates an alert (used by rule engine).
func (h *RegulatoryHandler) CreateRegulatoryAlert(c *gin.Context) {
	var alert model.RegulatoryAlert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.store.CreateRegulatoryAlert(c.Request.Context(), &alert); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": alert})
}

// GetAuditTrail returns full audit trail for a patient.
func (h *RegulatoryHandler) GetAuditTrail(c *gin.Context) {
	trail, err := h.store.GetRegulatoryAuditTrail(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": trail})
}

// ListRuleConfigs returns all rule configurations.
func (h *RegulatoryHandler) ListRuleConfigs(c *gin.Context) {
	configs, err := h.store.ListRuleConfigs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": configs})
}

// UpdateRuleConfig updates a rule configuration.
func (h *RegulatoryHandler) UpdateRuleConfig(c *gin.Context) {
	ruleCode := c.Param("code")
	var body struct {
		Config map[string]interface{} `json:"config" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	configJSON, _ := json.Marshal(body.Config)
	if err := h.store.UpdateRuleConfig(c.Request.Context(), ruleCode, string(configJSON)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// ConfigureFence sets or updates hospital geofence config.
func (h *RegulatoryHandler) ConfigureFence(c *gin.Context) {
	var fc model.RegulatoryFenceConfig
	if err := c.ShouldBindJSON(&fc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.store.CreateFenceConfig(c.Request.Context(), &fc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": fc, "status": "configured"})
}

// GetFenceConfig returns geofence configuration for a hospital.
func (h *RegulatoryHandler) GetFenceConfig(c *gin.Context) {
	hospitalID := c.Query("hospital_id")
	fc, err := h.store.GetFenceConfig(c.Request.Context(), hospitalID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "fence config not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": fc})
}

// GetComplianceReport returns periodic compliance report.
func (h *RegulatoryHandler) GetComplianceReport(c *gin.Context) {
	report, err := h.store.GetComplianceReport(c.Request.Context(),
		c.Query("hospital_id"), c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": report})
}
