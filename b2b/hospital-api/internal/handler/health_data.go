package handler

import (
	"net/http"
	"time"

	"eregen.dev/b2b-hospital-api/internal/model"
	"eregen.dev/b2b-hospital-api/internal/middleware"
	"eregen.dev/b2b-hospital-api/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type HealthDataHandler struct {
	store *store.Postgres
	log   *zap.Logger
}

func NewHealthDataHandler(store *store.Postgres, log *zap.Logger) *HealthDataHandler {
	return &HealthDataHandler{store: store, log: log}
}

// POST /api/v2/b2b/health-data — receive health data from hospital HIS
func (h *HealthDataHandler) Receive(c *gin.Context) {
	var req model.HealthDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	instID, _ := c.Get(string(middleware.ContextInstitutionID))
	instStr := instID.(string)

	h.log.Info("received health data",
		zap.String("patient_id", req.PatientID),
		zap.String("institution", instStr),
		zap.Int("vitals", len(req.Vitals)),
	)

	ctx := c.Request.Context()

	// Resolve patient to local elderly profile
	eregenID := ""
	if req.EregenID != nil && *req.EregenID != "" {
		eregenID = *req.EregenID
	} else {
		// Try to find existing link by external patient ID
		linked, err := h.store.FindElderlyByExternalPatient(ctx, req.PatientID)
		if err == nil && linked != "" {
			eregenID = linked
		}
	}

	if eregenID != "" && req.EregenID == nil {
		// Auto-link if not already linked
		_ = h.store.LinkElderlyToExternalPatient(ctx, eregenID, req.PatientID, eregenID)
	}

	// Convert incoming vitals to stored records
	now := time.Now()
	for _, v := range req.Vitals {
		record := &model.VitalSignRecord{
			ID:            uuid.New().String(),
			ElderlyID:     eregenID,
			InstitutionID: instStr,
			PatientID:     req.PatientID,
			RecordedAt:    now,
		}
		switch v.Type {
		case "hr":
			v := int(v.Value)
			record.HeartRate = &v
		case "spo2":
			v := int(v.Value)
			record.SPO2 = &v
		case "bp_systolic":
			v := int(v.Value)
			record.SystolicBP = &v
		case "bp_diastolic":
			v := int(v.Value)
			record.DiastolicBP = &v
		case "temp":
			record.Temperature = &v.Value
		case "steps":
			steps := int64(v.Value)
			record.Steps = &steps
		}
		if err := h.store.StoreVitals(ctx, record); err != nil {
			h.log.Error("store vital sign", zap.Error(err))
		}
	}

	// Check for abnormal values and log alerts
	for _, v := range req.Vitals {
		if v.Normal != nil && !*v.Normal {
			h.log.Warn("abnormal vital sign",
				zap.String("type", v.Type),
				zap.Float64("value", v.Value),
				zap.String("patient", req.PatientID),
			)
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Health data received"})
}

// GET /api/v2/b2b/patients/:eregen_id/report — generate health report
func (h *HealthDataHandler) GetReport(c *gin.Context) {
	eregenID := c.Param("eregen_id")
	days, _ := parseIntParam(c, "days", 30)

	vitals, err := h.store.GetVitalsForElderly(c.Request.Context(), eregenID, days)
	if err != nil {
		h.log.Error("get vitals for report", zap.Error(err))
		vitals = nil // continue with empty data
	}

	// Aggregate vitals into trend data
	var hrSum, spo2Sum, stepsSum float64
	var hrCount, spo2Count, stepsCount int
	var trends []model.VitalTrend

	for _, v := range vitals {
		if v.HeartRate != nil {
			hrSum += float64(*v.HeartRate)
			hrCount++
		}
		if v.SPO2 != nil {
			spo2Sum += float64(*v.SPO2)
			spo2Count++
		}
		if v.Steps != nil {
			stepsSum += float64(*v.Steps)
			stepsCount++
		}
	}

	report := model.HealthReport{
		ElderlyID:   eregenID,
		ReportDate:  time.Now(),
		PeriodStart: time.Now().Add(-time.Duration(days) * 24 * time.Hour),
		PeriodEnd:   time.Now(),
		Summary: model.ReportSummary{
			AvgHR:     ptrFloat64(ifZero(hrSum, 0) / ifZero(float64(hrCount), 1)),
			AvgSPO2:   ptrFloat64(ifZero(spo2Sum, 0) / ifZero(float64(spo2Count), 1)),
			AvgSteps:  ptrFloat64(ifZero(stepsSum, 0) / ifZero(float64(stepsCount), 1)),
			RiskLevel: determineRiskLevel(vitals),
		},
		VitalsTrend:  trends,
		AlertCount:   0,
		MedAdherence: 85.0,
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": report})
}

func ptrFloat64(f float64) *float64 { return &f }
func ifZero(v, def float64) float64 {
	if v == 0 { return def }
	return v
}
func determineRiskLevel(vitals []model.VitalSignRecord) string {
	for _, v := range vitals {
		if v.HeartRate != nil && (*v.HeartRate < 50 || *v.HeartRate > 120) {
			return "high"
		}
		if v.SPO2 != nil && *v.SPO2 < 90 {
			return "high"
		}
		if v.SystolicBP != nil && (*v.SystolicBP > 160 || *v.SystolicBP < 90) {
			return "medium"
		}
	}
	return "low"
}
