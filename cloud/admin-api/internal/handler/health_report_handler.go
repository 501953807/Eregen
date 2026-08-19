package handler

import (
	"fmt"
	"net/http"
	"time"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"

	"strconv"

	"github.com/gin-gonic/gin"
)

// HealthReportHandler manages health report templates and reports.
type HealthReportHandler struct {
	store store.HealthReportStore
}

func NewHealthReportHandler(s store.HealthReportStore) *HealthReportHandler {
	return &HealthReportHandler{store: s}
}

// CreateTemplate adds a report template.
func (h *HealthReportHandler) CreateTemplate(c *gin.Context) {
	var t model.HealthReportTemplate
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.CreateReportTemplate(c.Request.Context(), &t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK"})
}

// ListTemplates returns report templates for a chain.
func (h *HealthReportHandler) ListTemplates(c *gin.Context) {
	chain := c.Query("chain")
	templates, err := h.store.ListReportTemplates(c.Request.Context(), model.BusinessChain(chain))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": templates})
}

// CreateReport generates a health report.
func (h *HealthReportHandler) CreateReport(c *gin.Context) {
	var raw map[string]interface{}
	if err := c.ShouldBindJSON(&raw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	r := model.HealthReport{
		PersonID:      asString(raw, "person_id"),
		BusinessChain: asString(raw, "business_chain"),
		Status:        asString(raw, "status"),
		TemplateID:    asString(raw, "template_id"),
	}
	if r.ReportPeriodStart, _ = parseDate(asString(raw, "report_period_start")); r.ReportPeriodStart.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "report_period_start is required"})
		return
	}
	if r.ReportPeriodEnd, _ = parseDate(asString(raw, "report_period_end")); r.ReportPeriodEnd.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "report_period_end is required"})
		return
	}

	if err := h.store.CreateReport(c.Request.Context(), &r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK"})
}

func asString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func parseDate(s string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date: %s", s)
}

// ListReports returns reports for a person.
func (h *HealthReportHandler) ListReports(c *gin.Context) {
	personID := c.Param("personId")
	chain := c.Query("chain")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 100 {
		limit = 100
	}
	reports, err := h.store.ListReports(c.Request.Context(), personID, model.BusinessChain(chain), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": reports})
}
