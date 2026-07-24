package handler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store"
	"eregen.dev/shared/validation"

	"github.com/gin-gonic/gin"
)

type CommunityWBHandler struct {
	store store.Store
}

func NewCommunityWBHandler(s store.Store) *CommunityWBHandler {
	return &CommunityWBHandler{store: s}
}

// Elder profiles
func (h *CommunityWBHandler) ListElders(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, pageSize, err := validation.ValidatePagination(page, pageSize, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	elders, err := h.store.ListCommunityElders(c.Request.Context(), page, pageSize, c.Query("status"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": elders, "page": page, "page_size": pageSize})
}

func (h *CommunityWBHandler) GetElder(c *gin.Context) {
	elder, err := h.store.GetCommunityElder(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "elder not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": elder})
}

func (h *CommunityWBHandler) CreateElder(c *gin.Context) {
	var elder model.CommunityElder
	if err := c.ShouldBindJSON(&elder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.store.CreateCommunityElder(c.Request.Context(), &elder); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": elder})
}

func (h *CommunityWBHandler) UpdateElder(c *gin.Context) {
	var elder model.CommunityElder
	if err := c.ShouldBindJSON(&elder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	elder.ID = c.Param("id")
	if err := h.store.UpdateCommunityElder(c.Request.Context(), &elder); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *CommunityWBHandler) DeleteElder(c *gin.Context) {
	if err := h.store.DeleteCommunityElder(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (h *CommunityWBHandler) GetElderStats(c *gin.Context) {
	stats, err := h.store.GetCommunityElderStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// Device management
func (h *CommunityWBHandler) ListDevices(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, pageSize, err := validation.ValidatePagination(page, pageSize, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	devices, err := h.store.ListCommunityDevices(c.Request.Context(), page, pageSize, c.Query("status"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": devices, "page": page, "page_size": pageSize})
}

func (h *CommunityWBHandler) BindElderDevice(c *gin.Context) {
	var body struct {
		ElderID  string `json:"elder_id" binding:"required"`
		DeviceID string `json:"device_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.store.BindCommunityElderDevice(c.Request.Context(), body.ElderID, body.DeviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "bound"})
}

// Welfare tags
func (h *CommunityWBHandler) ListWelfareTags(c *gin.Context) {
	configs, err := h.store.ListWelfareTagConfigs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": configs})
}

func (h *CommunityWBHandler) GetElderWelfareTags(c *gin.Context) {
	tags, err := h.store.ListElderWelfareTags(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tags})
}

func (h *CommunityWBHandler) AssignWelfareTag(c *gin.Context) {
	var welfare model.CommunityElderWelfare
	if err := c.ShouldBindJSON(&welfare); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	welfare.ElderID = c.Param("id")
	if err := h.store.AssignWelfareTag(c.Request.Context(), &welfare); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "assigned"})
}

func (h *CommunityWBHandler) RevokeWelfareTag(c *gin.Context) {
	tagCode := c.Param("tag_code")
	if err := h.store.RevokeWelfareTag(c.Request.Context(), c.Param("id"), tagCode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "revoked"})
}

// Sign-in
func (h *CommunityWBHandler) TriggerSignin(c *gin.Context) {
	var body struct {
		ElderID       string   `json:"elder_id" binding:"required"`
		DeviceID      string   `json:"device_id" binding:"required"`
		HospitalID    string   `json:"hospital_id" binding:"required"`
		Period        string   `json:"period" binding:"required"`
		ActivatedTags []string `json:"activated_tags"`
		IsMedical     bool     `json:"is_medical_signin"`
		IsWelfare     bool     `json:"is_welfare_signin"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Fetch elder to get id_card for cross-hospital dedup
	elder, err := h.store.GetCommunityElder(c.Request.Context(), body.ElderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "elder not found"})
		return
	}

	// 2. Check same elder_id + device_id + period already signed in (unique constraint guard)
	existing, _ := h.store.GetSigninSummary(c.Request.Context(), body.ElderID, body.Period)
	if existing != nil && existing.DeviceID == body.DeviceID {
		c.JSON(http.StatusConflict, gin.H{"error": "already signed in this period"})
		return
	}

	// 3. Cross-hospital duplicate check via id_card (R_C01)
	if elder.IDCard != "" {
		recs, err := h.store.ListSigninRecords(c.Request.Context(), "", body.Period, "", 1, 100)
		if err == nil {
			for _, r := range recs {
				if r.ElderID == body.ElderID && r.HospitalID != body.HospitalID {
					// Same elder at different hospital in same period — R_C01 alert
					alert := model.RegulatoryAlert{
						RuleCode: "R_C01", Severity: "high", AlertType: "duplicate_signin",
						Detail: fmt.Sprintf("elder %s (%s) signed at %s and %s in period %s",
							elder.Name, elder.IDCard, body.HospitalID, r.HospitalID, body.Period),
					}
					h.store.CreateRegulatoryAlert(c.Request.Context(), &alert)
					break
				}
			}
		}
	}

	rec := model.CommunitySigninRecord{
		ElderID:         body.ElderID,
		DeviceID:        body.DeviceID,
		HospitalID:      body.HospitalID,
		Period:          body.Period,
		IDCard:          elder.IDCard,
		IsMedicalSignin: body.IsMedical,
		IsWelfareSignin: body.IsWelfare,
	}
	if len(body.ActivatedTags) > 0 {
		tagsJSON, _ := json.Marshal(body.ActivatedTags)
		rec.ActivatedTags = string(tagsJSON)
	}
	if err := h.store.CreateSigninRecord(c.Request.Context(), &rec); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "signed_in", "record_id": rec.ID})
}

func (h *CommunityWBHandler) ListSigninRecords(c *gin.Context) {
	records, err := h.store.ListSigninRecords(c.Request.Context(),
		c.Query("elder_id"), c.Query("period"), c.Query("hospital_id"), 1, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": records})
}

// Pharmacy
func (h *CommunityWBHandler) DispenseMedicine(c *gin.Context) {
	var body struct {
		ElderID        string   `json:"elder_id" binding:"required"`
		HospitalID     string   `json:"hospital_id" binding:"required"`
		PharmacistID   string   `json:"pharmacist_id"`
		Period         string   `json:"period" binding:"required"`
		Items          []string `json:"items" binding:"required"`
		TotalCost      float64  `json:"total_cost"`
		InsuranceCovered float64 `json:"insurance_covered"`
		SelfPay        float64  `json:"self_pay"`
		Notes          string   `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log := model.CommunityPharmacyLog{
		ElderID:        body.ElderID,
		HospitalID:     body.HospitalID,
		PharmacistID:   body.PharmacistID,
		Period:         body.Period,
		TotalCost:      body.TotalCost,
		InsuranceCovered: body.InsuranceCovered,
		SelfPay:        body.SelfPay,
		Notes:          body.Notes,
	}
	log.Items = "[]" // TODO: marshal items to JSON
	if err := h.store.CreatePharmacyLog(c.Request.Context(), &log); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "dispensed", "log_id": log.ID})
}

// Minzheng sync
func (h *CommunityWBHandler) ImportMinzhengData(c *gin.Context) {
	var body struct {
		Source string `json:"source" binding:"required"`
		Filename string `json:"filename"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sync := model.CommunityMinzhengSync{
		Source:     body.Source,
		Filename:   body.Filename,
		Status:     "processing",
		ImportedCount: 0,
		MatchedCount:  0,
	}
	if err := h.store.CreateMinzhengSync(c.Request.Context(), &sync); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": sync})
}

func (h *CommunityWBHandler) ListMinzhengSync(c *gin.Context) {
	syncs, err := h.store.ListMinzhengSync(c.Request.Context(), 1, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": syncs})
}

// Batch payments
func (h *CommunityWBHandler) ExecuteBatchPayment(c *gin.Context) {
	var body struct {
		BatchID  string   `json:"batch_id" binding:"required"`
		Period   string   `json:"period" binding:"required"`
		PayType  string   `json:"pay_type" binding:"required"`
		ElderIDs []string `json:"elder_ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payments := make([]model.CommunityBatchPayment, 0, len(body.ElderIDs))
	for _, elderID := range body.ElderIDs {
		// MVP stub: simulate 90% success rate
		status := "success"
		if randFloat() < 0.1 {
			status = "failed"
		}
		payments = append(payments, model.CommunityBatchPayment{
			BatchID: body.BatchID, Period: body.Period, PayType: body.PayType,
			ElderID: elderID, Status: status, Amount: 0,
		})
	}
	if len(payments) > 0 {
		h.store.BulkCreateBatchPayments(c.Request.Context(), payments)
	}
	c.JSON(http.StatusOK, gin.H{"status": "completed", "batch_id": body.BatchID, "total": len(payments)})
}

func (h *CommunityWBHandler) ListBatchPayments(c *gin.Context) {
	payments, err := h.store.ListBatchPayments(c.Request.Context(), c.Query("batch_id"), 1, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": payments})
}

// randFloat returns a pseudo-random float64 in [0, 1).
var randFloat = func() float64 { return rand.Float64() }
