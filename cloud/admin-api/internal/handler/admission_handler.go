package handler

import (
	"net/http"
	"strconv"
	"time"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"
	"eregen.dev/shared/validation"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdmissionHandler serves hospital admission and ward round endpoints.
type AdmissionHandler struct {
	store store.Store
}

// NewAdmissionHandler creates a new AdmissionHandler.
func NewAdmissionHandler(s store.Store) *AdmissionHandler {
	return &AdmissionHandler{store: s}
}

// ListAdmissions returns paginated admission list.
func (h *AdmissionHandler) ListAdmissions(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, pageSize, err := validation.ValidatePagination(page, pageSize, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "System internal error"})
		return
	}

	department := c.Query("department")
	status := c.Query("status")
	admissions, err := h.store.ListAdmissions(c.Request.Context(), page, pageSize, department, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":      "OK",
		"data":      admissions,
		"page":      page,
		"page_size": pageSize,
	})
}

// AdmitPatient registers a new hospital admission with wristband binding.
func (h *AdmissionHandler) AdmitPatient(c *gin.Context) {
	var req struct {
		PatientID        string `json:"patient_id" binding:"required"`
		BedNo            string `json:"bed_no" binding:"required"`
		Department       string `json:"department" binding:"required"`
		Diagnosis        string `json:"diagnosis"`
		EmergencyContact string `json:"emergency_contact"`
		Allergies        string `json:"allergies"`
		ExpectedStayDays int    `json:"expected_stay_days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	admission := &model.HospitalAdmission{
		ID:               uuid.New().String(),
		PatientID:        req.PatientID,
		BedNo:            req.BedNo,
		Department:       req.Department,
		Diagnosis:        req.Diagnosis,
		EmergencyContact: req.EmergencyContact,
		Allergies:        req.Allergies,
	}
	if req.ExpectedStayDays > 0 {
		t := time.Now().AddDate(0, 0, req.ExpectedStayDays)
		admission.ExpectedDischargeAt = &t
	}

	if err := h.store.CreateAdmission(c.Request.Context(), admission); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}

	// Evaluate regulatory rules
	h.store.EvaluateRegulatoryRules(c.Request.Context(), "patient_admitted", map[string]string{
		"patient_id": req.PatientID,
	})

	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": admission})
}

// DischargePatient completes an admission.
func (h *AdmissionHandler) DischargePatient(c *gin.Context) {
	admissionID := c.Param("id")
	var body struct {
		DischargeType string `json:"discharge_type" binding:"required"`
		Notes         string `json:"notes"`
		TransferredTo string `json:"transferred_to"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.store.CompleteAdmission(c.Request.Context(), admissionID, body.DischargeType, body.Notes, body.TransferredTo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "discharged"})
}

// GetWardRound returns scheduled/completed ward rounds for a patient.
func (h *AdmissionHandler) GetWardRound(c *gin.Context) {
	patientID := c.Param("id")
	rounds, err := h.store.ListWardRounds(c.Request.Context(), patientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": rounds})
}

// CompleteWardRound records a nursing round entry with vitals.
func (h *AdmissionHandler) CompleteWardRound(c *gin.Context) {
	patientID := c.Param("id")
	var entry model.WardRoundEntry
	entry.PatientID = patientID
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	entry.ID = uuid.New().String()
	if err := h.store.CreateWardRound(c.Request.Context(), &entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": entry})
}
