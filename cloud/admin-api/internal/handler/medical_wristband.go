package handler

import (
	"net/http"
	"strconv"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"
	"eregen.dev/shared/validation"

	"github.com/gin-gonic/gin"
)

// MedicalWristbandHandler serves medical wristband management endpoints.
type MedicalWristbandHandler struct {
	store store.Store
}

// NewMedicalWristbandHandler creates a new MedicalWristbandHandler.
func NewMedicalWristbandHandler(s store.Store) *MedicalWristbandHandler {
	return &MedicalWristbandHandler{store: s}
}

// ---------- Patient endpoints ----------

// ListPatients returns paginated patient list.
func (h *MedicalWristbandHandler) ListPatients(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, pageSize, err := validation.ValidatePagination(page, pageSize, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := c.Query("status")
	patients, err := h.store.ListPatients(c.Request.Context(), page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      patients,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetPatient returns a single patient by ID.
func (h *MedicalWristbandHandler) GetPatient(c *gin.Context) {
	id := c.Param("id")
	patient, err := h.store.GetPatient(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": patient})
}

// CreatePatient registers a new patient.
func (h *MedicalWristbandHandler) CreatePatient(c *gin.Context) {
	var p model.MedicalPatient
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if p.Status == "" {
		p.Status = "admitted"
	}
	if err := h.store.CreatePatient(c.Request.Context(), &p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": p})
}

// UpdatePatient modifies an existing patient.
func (h *MedicalWristbandHandler) UpdatePatient(c *gin.Context) {
	id := c.Param("id")
	var p model.MedicalPatient
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	p.ID = id
	if err := h.store.UpdatePatient(c.Request.Context(), &p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": p})
}

// DeletePatient soft-deletes a patient (marks as discharged).
func (h *MedicalWristbandHandler) DeletePatient(c *gin.Context) {
	id := c.Param("id")
	if err := h.store.DeletePatient(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "discharged"})
}

// GetPatientByAdmissionNo looks up a patient by admission number.
func (h *MedicalWristbandHandler) GetByAdmissionNo(c *gin.Context) {
	admissionNo := c.Query("admission_no")
	if admissionNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "admission_no required"})
		return
	}
	patient, err := h.store.GetPatientByAdmissionNo(c.Request.Context(), admissionNo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": patient})
}

// BatchImportPatients imports multiple patients from CSV/JSON.
func (h *MedicalWristbandHandler) BatchImport(c *gin.Context) {
	var patients []model.MedicalPatient
	if err := c.ShouldBindJSON(&patients); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body, expected array"})
		return
	}
	if len(patients) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty patient list"})
		return
	}
	if err := h.store.BatchImportPatients(c.Request.Context(), patients); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "imported", "count": len(patients)})
}

// GetPatientHistory returns treatment history for a patient.
func (h *MedicalWristbandHandler) GetPatientHistory(c *gin.Context) {
	id := c.Param("id")
	history, err := h.store.GetPatientHistory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": history})
}

// ---------- Wristband device endpoints ----------

// ListWristbands returns paginated wristband devices.
func (h *MedicalWristbandHandler) ListWristbands(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, pageSize, err := validation.ValidatePagination(page, pageSize, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := c.Query("status")
	devices, err := h.store.ListWristbands(c.Request.Context(), page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      devices,
		"page":      page,
		"page_size": pageSize,
	})
}

// BindWristband binds a wristband device to a patient.
func (h *MedicalWristbandHandler) BindWristband(c *gin.Context) {
	var body struct {
		PatientID string `json:"patient_id" binding:"required"`
		DeviceID  string `json:"device_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.BindWristband(c.Request.Context(), body.PatientID, body.DeviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "bound"})
}

// UnbindWristband unbinds a wristband device from a patient.
func (h *MedicalWristbandHandler) UnbindWristband(c *gin.Context) {
	bindingID := c.Param("id")
	if err := h.store.UnbindWristband(c.Request.Context(), bindingID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "unbound"})
}

// ClearWristband clears all data from a wristband device.
func (h *MedicalWristbandHandler) ClearWristband(c *gin.Context) {
	deviceID := c.Param("id")
	if err := h.store.ClearWristband(c.Request.Context(), deviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "cleared"})
}

// WriteToWristband pushes data to a wristband device.
func (h *MedicalWristbandHandler) WriteToWristband(c *gin.Context) {
	deviceID := c.Param("id")
	var body struct {
		Data string `json:"data" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.WriteToWristband(c.Request.Context(), deviceID, body.Data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "written"})
}

// GetWristbandFirmware returns firmware version for a device.
func (h *MedicalWristbandHandler) GetFirmware(c *gin.Context) {
	deviceID := c.Param("id")
	fw, err := h.store.GetWristbandFirmware(c.Request.Context(), deviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "wristband not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": fw})
}

// ---------- Expense endpoints ----------

// ListExpenses returns expenses for a patient.
func (h *MedicalWristbandHandler) ListExpenses(c *gin.Context) {
	patientID := c.Param("id")
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, pageSize, err := validation.ValidatePagination(page, pageSize, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	expenses, err := h.store.ListExpenses(c.Request.Context(), patientID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      expenses,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateExpense adds a new expense record.
func (h *MedicalWristbandHandler) CreateExpense(c *gin.Context) {
	var e model.MedicalExpense
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.CreateExpense(c.Request.Context(), &e); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": e})
}

// ---------- Medication endpoints ----------

// ListMedications returns medications for a patient.
func (h *MedicalWristbandHandler) ListMedications(c *gin.Context) {
	patientID := c.Param("id")
	items, err := h.store.ListMedications(c.Request.Context(), patientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// CreateMedication adds a new medication order.
func (h *MedicalWristbandHandler) CreateMedication(c *gin.Context) {
	var m model.MedicalMedication
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.CreateMedication(c.Request.Context(), &m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": m})
}

// ---------- Test result endpoints ----------

// ListTestResults returns test results for a patient.
func (h *MedicalWristbandHandler) ListTestResults(c *gin.Context) {
	patientID := c.Param("id")
	items, err := h.store.ListTestResults(c.Request.Context(), patientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// CreateTestResult adds a new test result.
func (h *MedicalWristbandHandler) CreateTestResult(c *gin.Context) {
	var r model.MedicalTestResult
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.CreateTestResult(c.Request.Context(), &r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": r})
}

// ---------- Daily entry endpoints ----------

// ListDailyEntries returns daily nursing/doctor entries for a patient.
func (h *MedicalWristbandHandler) ListDailyEntries(c *gin.Context) {
	patientID := c.Param("id")
	date := c.Query("date")
	items, err := h.store.ListDailyEntries(c.Request.Context(), patientID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// CreateDailyEntry adds a new daily entry.
func (h *MedicalWristbandHandler) CreateDailyEntry(c *gin.Context) {
	var e model.MedicalDailyEntry
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.CreateDailyEntry(c.Request.Context(), &e); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": e})
}

// ---------- Verification endpoints ----------

// ListVerifications returns verification records.
func (h *MedicalWristbandHandler) ListVerifications(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, pageSize, err := validation.ValidatePagination(page, pageSize, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items, err := h.store.ListVerifications(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      items,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateVerification records a nurse BLE verification scan.
func (h *MedicalWristbandHandler) CreateVerification(c *gin.Context) {
	var v model.MedicalVerification
	if err := c.ShouldBindJSON(&v); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.CreateVerification(c.Request.Context(), &v); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": v})
}

// UpdateVerificationStatus updates verification status.
func (h *MedicalWristbandHandler) UpdateVerificationStatus(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.UpdateVerificationStatus(c.Request.Context(), id, body.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// GetTodayVerificationStats returns today's verification statistics.
func (h *MedicalWristbandHandler) GetTodayVerificationStats(c *gin.Context) {
	stats, err := h.store.GetTodayVerificationStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// ---------- Stats endpoints ----------

// GetMedicalStatsOverview returns overall medical statistics.
func (h *MedicalWristbandHandler) GetStatsOverview(c *gin.Context) {
	stats, err := h.store.GetMedicalStatsOverview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// ---------- Alert tag config endpoints ----------

// ListAlertTagConfigs returns alert tag configurations.
func (h *MedicalWristbandHandler) ListAlertTagConfigs(c *gin.Context) {
	items, err := h.store.ListAlertTagConfigs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// CreateAlertTagConfig creates an alert tag configuration.
func (h *MedicalWristbandHandler) CreateAlertTagConfig(c *gin.Context) {
	var cfg model.MedicalAlertTagConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.store.CreateAlertTagConfig(c.Request.Context(), &cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": cfg})
}
