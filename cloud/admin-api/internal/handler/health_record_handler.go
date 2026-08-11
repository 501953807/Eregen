package handler

import (
	"net/http"
	"strconv"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"

	"github.com/gin-gonic/gin"
)

// HealthRecordHandler manages health records per person + chain.
type HealthRecordHandler struct {
	store store.HealthRecordStore
}

func NewHealthRecordHandler(s store.HealthRecordStore) *HealthRecordHandler {
	return &HealthRecordHandler{store: s}
}

// Create adds a health record.
func (h *HealthRecordHandler) Create(c *gin.Context) {
	var r model.HealthRecordV2
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.CreateHealthRecordV2(c.Request.Context(), &r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK"})
}

// List returns health records for a person.
func (h *HealthRecordHandler) List(c *gin.Context) {
	personID := c.Param("personId")
	chain := c.Query("chain")
	recordType := c.Query("record_type")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit > 200 {
		limit = 200
	}
	records, err := h.store.ListHealthRecordsV2(c.Request.Context(), personID, model.BusinessChain(chain), recordType, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": records})
}

// GetSummary returns the health summary for a person.
func (h *HealthRecordHandler) GetSummary(c *gin.Context) {
	personID := c.Param("personId")
	chain := c.Query("chain")
	summary, err := h.store.GetHealthSummaryV2(c.Request.Context(), personID, model.BusinessChain(chain))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": summary})
}

// UpdateSummary updates the health summary.
func (h *HealthRecordHandler) UpdateSummary(c *gin.Context) {
	var summary model.PersonHealthSummary
	if err := c.ShouldBindJSON(&summary); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.UpdateHealthSummaryV2(c.Request.Context(), &summary); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK"})
}
