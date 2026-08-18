package handler

import (
	"net/http"
	"strconv"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"
	"eregen.dev/shared/validation"

	"github.com/gin-gonic/gin"
)

// PatientHandler serves patient management endpoints.
type PatientHandler struct {
	store store.Store
}

// NewPatientHandler creates a new PatientHandler.
func NewPatientHandler(s store.Store) *PatientHandler {
	return &PatientHandler{store: s}
}

// ListPatients returns paginated patient list.
func (h *PatientHandler) ListPatients(c *gin.Context) {
	var err error
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, pageSize, err = validation.ValidatePagination(page, pageSize, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := c.Query("status")
	patients, err := h.store.ListPatients(c.Request.Context(), page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":      "OK",
		"data":      patients,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetPatient returns a single patient by ID.
func (h *PatientHandler) GetPatient(c *gin.Context) {
	id := c.Param("id")
	patient, err := h.store.GetPatient(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": patient})
}

// CreatePatient registers a new patient.
func (h *PatientHandler) CreatePatient(c *gin.Context) {
	var p model.MedicalPatient
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if p.Status == "" {
		p.Status = "admitted"
	}
	if err := h.store.CreatePatient(c.Request.Context(), &p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": "OK", "data": p})
}

// UpdatePatient modifies an existing patient.
func (h *PatientHandler) UpdatePatient(c *gin.Context) {
	id := c.Param("id")
	var p model.MedicalPatient
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	p.ID = id
	if err := h.store.UpdatePatient(c.Request.Context(), &p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": p})
}

// DeletePatient soft-deletes a patient (marks as discharged).
func (h *PatientHandler) DeletePatient(c *gin.Context) {
	id := c.Param("id")
	if err := h.store.DeletePatient(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "discharged"})
}

// GetByAdmissionNo looks up a patient by admission number.
func (h *PatientHandler) GetByAdmissionNo(c *gin.Context) {
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
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": patient})
}

// BatchImport imports multiple patients from JSON.
func (h *PatientHandler) BatchImport(c *gin.Context) {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "imported", "count": len(patients)})
}

// GetPatientHistory returns treatment history for a patient.
func (h *PatientHandler) GetPatientHistory(c *gin.Context) {
	id := c.Param("id")
	history, err := h.store.GetPatientHistory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "System internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": history})
}
