package handler

import (
	"net/http"
	"time"

	"eregen.dev/b2b-hospital-api/internal/model"
	"eregen.dev/b2b-hospital-api/internal/middleware"
	"eregen.dev/b2b-hospital-api/internal/store"

	"github.com/gin-gonic/gin"
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

	h.log.Info("received health data",
		zap.String("patient_id", req.PatientID),
		zap.String("institution", instID.(string)),
		zap.Int("vitals", len(req.Vitals)),
	)

	// In production:
	// 1. Validate patient_id exists in external system
	// 2. If eregen_id provided, link to local elderly profile
	// 3. Store vitals in InfluxDB for time-series analysis
	// 4. Check for abnormal values → trigger alerts
	// 5. Publish to NATS for real-time processing

	c.JSON(http.StatusOK, gin.H{"code": "OK", "message": "Health data received"})
}

// GET /api/v2/b2b/patients/:eregen_id/report — generate health report
func (h *HealthDataHandler) GetReport(c *gin.Context) {
	eregenID := c.Param("eregen_id")
	days, _ := parseIntParam(c, "days", 30)

	// In production: query InfluxDB for time-series data, aggregate
	report := model.HealthReport{
		ElderlyID:   eregenID,
		ReportDate:  time.Now(),
		PeriodStart: time.Now().Add(-time.Duration(days) * 24 * time.Hour),
		PeriodEnd:   time.Now(),
		Summary: model.ReportSummary{
			RiskLevel: "low",
		},
		MedAdherence: 85.0,
	}

	c.JSON(http.StatusOK, gin.H{"code": "OK", "data": report})
}
