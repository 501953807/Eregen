package handler

import (
	"net/http"

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
	var r model.HealthReport
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.CreateReport(c.Request.Context(), &r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK"})
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
