package handler

import (
	"net/http"
	"strconv"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"
	"eregen.dev/shared/validation"

	"github.com/gin-gonic/gin"
)

// ClinicalHandler serves clinical data endpoints: expenses, medications,
// test results, daily entries, verifications, stats, and alert tags.
type ClinicalHandler struct {
	store store.Store
}

// NewClinicalHandler creates a new ClinicalHandler.
func NewClinicalHandler(s store.Store) *ClinicalHandler {
	return &ClinicalHandler{store: s}
}

// ---------- Expense endpoints ----------

// ListExpenses returns expenses for a patient.
func (h *ClinicalHandler) ListExpenses(c *gin.Context) {
	patientID := c.Param("id")
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, pageSize, err := validation.ValidatePagination(page, pageSize, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "System internal error"})
		return
	}

	expenses, err := h.store.ListExpenses(c.Request.Context(), patientID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":      "OK",
		"data":      expenses,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateExpense adds a new expense record.
func (h *ClinicalHandler) CreateExpense(c *gin.Context) {
	var e model.MedicalExpense
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.CreateExpense(c.Request.Context(), &e); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": e})
}

// ---------- Medication endpoints ----------

// ListMedications returns medications for a patient.
func (h *ClinicalHandler) ListMedications(c *gin.Context) {
	patientID := c.Param("id")
	items, err := h.store.ListMedications(c.Request.Context(), patientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": items})
}

// CreateMedication adds a new medication order.
func (h *ClinicalHandler) CreateMedication(c *gin.Context) {
	var m model.MedicalMedication
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.CreateMedication(c.Request.Context(), &m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": m})
}

// ---------- Test result endpoints ----------

// ListTestResults returns test results for a patient.
func (h *ClinicalHandler) ListTestResults(c *gin.Context) {
	patientID := c.Param("id")
	items, err := h.store.ListTestResults(c.Request.Context(), patientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": items})
}

// CreateTestResult adds a new test result.
func (h *ClinicalHandler) CreateTestResult(c *gin.Context) {
	var r model.MedicalTestResult
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.CreateTestResult(c.Request.Context(), &r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": r})
}

// ---------- Daily entry endpoints ----------

// ListDailyEntries returns daily nursing/doctor entries for a patient.
func (h *ClinicalHandler) ListDailyEntries(c *gin.Context) {
	patientID := c.Param("id")
	date := c.Query("date")
	items, err := h.store.ListDailyEntries(c.Request.Context(), patientID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": items})
}

// CreateDailyEntry adds a new daily entry.
func (h *ClinicalHandler) CreateDailyEntry(c *gin.Context) {
	var e model.MedicalDailyEntry
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.CreateDailyEntry(c.Request.Context(), &e); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": e})
}

// ---------- Verification endpoints ----------

// ListVerifications returns verification records.
func (h *ClinicalHandler) ListVerifications(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, pageSize, err := validation.ValidatePagination(page, pageSize, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "System internal error"})
		return
	}

	items, err := h.store.ListVerifications(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":      "OK",
		"data":      items,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateVerification records a nurse NFC verification scan.
func (h *ClinicalHandler) CreateVerification(c *gin.Context) {
	var v model.MedicalVerification
	if err := c.ShouldBindJSON(&v); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.CreateVerification(c.Request.Context(), &v); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": v})
}

// UpdateVerificationStatus updates verification status.
func (h *ClinicalHandler) UpdateVerificationStatus(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.UpdateVerificationStatus(c.Request.Context(), id, body.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// GetTodayVerificationStats returns today's verification statistics.
func (h *ClinicalHandler) GetTodayVerificationStats(c *gin.Context) {
	stats, err := h.store.GetTodayVerificationStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": stats})
}

// ---------- Stats endpoints ----------

// GetStatsOverview returns overall medical statistics.
func (h *ClinicalHandler) GetStatsOverview(c *gin.Context) {
	stats, err := h.store.GetMedicalStatsOverview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": stats})
}

// ---------- Alert tag config endpoints ----------

// ListAlertTagConfigs returns alert tag configurations.
func (h *ClinicalHandler) ListAlertTagConfigs(c *gin.Context) {
	items, err := h.store.ListAlertTagConfigs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": items})
}

// CreateAlertTagConfig creates an alert tag configuration.
func (h *ClinicalHandler) CreateAlertTagConfig(c *gin.Context) {
	var cfg model.MedicalAlertTagConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.CreateAlertTagConfig(c.Request.Context(), &cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": cfg})
}
