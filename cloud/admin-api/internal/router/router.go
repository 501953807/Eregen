package router

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"eregen.dev/admin-api/internal/handler"
	"eregen.dev/admin-api/internal/middleware"
	"eregen.dev/admin-api/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Setup wires up the Gin engine with all admin routes.
func Setup(db *sql.DB, logger *zap.Logger, dbType string) *gin.Engine {
	s := store.NewStore(db, dbType)
	r := gin.Default()

	// CORS middleware - allow admin-web and other trusted origins
	r.Use(func(c *gin.Context) {
		origins := []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:5173",
		}
		origin := c.Request.Header.Get("Origin")
		allowed := false

		// Check if origin is in allowed list
		if origin != "" {
			for _, o := range origins {
				if o == origin {
					allowed = true
					break
				}
			}
		} else {
			// Empty origin (same-origin or browser-initiated without Origin header) is always allowed
			allowed = true
		}

		if allowed {
			if origin == "" {
				origin = origins[0] // default for same-origin requests
			}
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Device-Token")
			c.Header("Access-Control-Max-Age", "86400")
		}

		// Handle preflight
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required — admin API cannot start without secure authentication")
	}
	adminJWT := middleware.NewAdminJWT(jwtSecret, 24*time.Hour, logger)

	// Auth handler (used for login endpoint which doesn't require auth)
	authHandler := handler.NewAuthHandler(jwtSecret, logger, s)

	// Unprotected health check endpoint — always available, no auth required
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "ok"}})
	})

	// Auth endpoints — no authentication required for login
	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/login", authHandler.Login)
	}

	dashboard := handler.NewDashboardHandler(s)
	device := handler.NewDeviceHandler(s)
	user := handler.NewUserHandler(s)
	alert := handler.NewAlertHandler(s)
	elderly := handler.NewElderlyHandler(s)
	firmware := handler.NewFirmwareHandler(s)
	settings := handler.NewSettingsHandler(s)
	medical := handler.NewMedicalWristbandHandler(s)
	regulatory := handler.NewRegulatoryHandler(s)
	communityWB := handler.NewCommunityWBHandler(s)

	// Rate limiter — fail open if Redis is unavailable
	rateLimiter, rlErr := middleware.NewAdminRateLimiter()
	if rlErr != nil {
		log.Printf("admin rate limiter init failed: %v (will fail open)", rlErr)
	}

	// SSE stream — auth via query param token (EventSource cannot send custom headers)
	r.GET("/api/v1/admin/stream/alerts", func(c *gin.Context) {
		handler.StreamHandler().ServeHTTP(c.Writer, c.Request)
	})

	api := r.Group("/api/v1/admin")
	if rlErr == nil {
		api.Use(rateLimiter.Middleware())
	}
	// JWT Authentication — only applied when JWT_SECRET is set
	if jwtSecret != "" {
		api.Use(adminJWT.AuthMiddleware())
	}
	{
		api.GET("/stats/overview", dashboard.GetOverview)
		api.GET("/stats/subscriptions", dashboard.GetSubscriptionStats)
		api.GET("/devices", device.List)
		api.GET("/users", user.List)
		api.POST("/users", user.Create)
		api.PUT("/users/:id", user.Update)
		api.DELETE("/users/:id", user.Delete)
		// User role management
		api.POST("/users/:id/role", user.SetRole)
		api.GET("/alerts", alert.List)
		// Device config and OTA
		api.POST("/devices/:id/config", device.UpdateConfig)
		api.POST("/devices/:id/ota", device.TriggerOTA)
		// Alert resolution
		api.POST("/alerts/:id/resolve", alert.Resolve)
		api.POST("/alerts/:id/acknowledge", alert.Acknowledge)
		// Elderly person management
		api.GET("/elderly", elderly.List)
		api.GET("/elderly/:id", elderly.Detail)
		api.POST("/elderly", elderly.Create)
		api.PUT("/elderly/:id", elderly.Update)
		api.DELETE("/elderly/:id", elderly.Delete)
		// Elderly detail views
		api.GET("/elderly/:id/health-stats", elderly.HealthStats)
		api.GET("/elderly/:id/health-records", elderly.HealthRecords)
		api.GET("/elderly/:id/medication-rules", elderly.MedicationRules)
		api.POST("/elderly/:id/medication-rules", elderly.CreateMedicationRule)
		api.PUT("/elderly/:id/medication-rules/:rule_id", elderly.UpdateMedicationRule)
		api.DELETE("/elderly/:id/medication-rules/:rule_id", elderly.DeleteMedicationRule)
		api.GET("/elderly/:id/devices", elderly.DeviceList)
		api.GET("/elderly/:id/location-history", elderly.LocationHistory)
		api.GET("/elderly/:id/alert-history", elderly.AlertHistory)

		// Dashboard chart stats
		api.GET("/stats/alert-trend", dashboard.GetAlertTrend)
		api.GET("/stats/alert-distribution", dashboard.GetAlertDistribution)
		api.GET("/stats/user-growth", dashboard.GetUserGrowth)

		// Device detail / unbind / batch OTA
		api.GET("/devices/:id", device.Detail)
		api.DELETE("/devices/:id/unbind", device.Unbind)
		api.POST("/devices/batch-ota", device.BatchOTA)

		// Firmware versions (OTA management)
		fw := api.Group("/firmware-versions")
		{
			fw.GET("", firmware.List)
			fw.POST("", firmware.Create)
			fw.DELETE("/:id", firmware.Delete)
		}
		api.POST("/ota/push", firmware.PushOTA)

		// System settings
		setting := api.Group("/settings")
		{
			setting.GET("/notifications", settings.GetNotificationSettings)
			setting.PUT("/notifications", settings.UpdateNotificationSettings)
			setting.GET("/api-keys", settings.ListAPIKeys)
			setting.POST("/api-keys", settings.CreateAPIKey)
			setting.DELETE("/api-keys/:id", settings.RevokeAPIKey)
			setting.POST("/password", settings.ChangePassword)
		}

		// Medical wristband management
		med := api.Group("/medical")
		{
			// Patient endpoints
			med.GET("/patients", medical.ListPatients)
			med.GET("/patients/:id", medical.GetPatient)
			med.POST("/patients", medical.CreatePatient)
			med.PUT("/patients/:id", medical.UpdatePatient)
			med.DELETE("/patients/:id", medical.DeletePatient)
			med.GET("/patients/by-admission", medical.GetByAdmissionNo)
			med.POST("/patients/batch-import", medical.BatchImport)
			med.GET("/patients/:id/history", medical.GetPatientHistory)

			// Wristband device endpoints
			med.GET("/wristbands", medical.ListWristbands)
			med.POST("/wristbands/bind", medical.BindWristband)
			med.POST("/wristbands/:id/unbind", medical.UnbindWristband)
			med.POST("/wristbands/:id/clear", medical.ClearWristband)
			med.POST("/wristbands/:id/write", medical.WriteToWristband)
			med.GET("/wristbands/:id/firmware", medical.GetFirmware)

			// Expense endpoints
			med.GET("/patients/:id/expenses", medical.ListExpenses)
			med.POST("/expenses", medical.CreateExpense)

			// Medication endpoints
			med.GET("/patients/:id/medications", medical.ListMedications)
			med.POST("/medications", medical.CreateMedication)

			// Test result endpoints
			med.GET("/patients/:id/test-results", medical.ListTestResults)
			med.POST("/test-results", medical.CreateTestResult)

			// Daily entry endpoints
			med.GET("/patients/:id/daily-entries", medical.ListDailyEntries)
			med.POST("/daily-entries", medical.CreateDailyEntry)

			// Verification endpoints
			med.GET("/verifications", medical.ListVerifications)
			med.POST("/verifications", medical.CreateVerification)
			med.PUT("/verifications/:id/status", medical.UpdateVerificationStatus)
			med.GET("/verifications/stats/today", medical.GetTodayVerificationStats)

			// Stats and alert tags
			med.GET("/stats/overview", medical.GetStatsOverview)
			med.GET("/alert-tags", medical.ListAlertTagConfigs)
			med.POST("/alert-tags", medical.CreateAlertTagConfig)

			// Clinical workflow endpoints
			med.POST("/admissions", medical.AdmitPatient)
			med.GET("/admissions", medical.ListAdmissions)
			med.POST("/admissions/:id/discharge", medical.DischargePatient)
			med.GET("/patients/:id/ward-round", medical.GetWardRound)
			med.POST("/patients/:id/ward-round", medical.CompleteWardRound)
		}

		// Regulatory closure
		reg := api.Group("/regulatory")
		{
			reg.GET("/dashboard/patient-overview", regulatory.GetDashboardOverview)
			reg.GET("/dashboard/patient-list", regulatory.ListRegulatoryPatients)
			reg.GET("/alerts", regulatory.ListAlerts)
			reg.GET("/alerts/:id", regulatory.GetAlert)
			reg.POST("/alerts/:id/acknowledge", regulatory.AcknowledgeAlert)
			reg.POST("/alerts/:id/resolve", regulatory.ResolveRegulatoryAlert)
			reg.POST("/alerts", regulatory.CreateRegulatoryAlert)
			reg.GET("/audit/patient/:id", regulatory.GetAuditTrail)
			reg.GET("/rules", regulatory.ListRuleConfigs)
			reg.PUT("/rules/:code/config", regulatory.UpdateRuleConfig)
			reg.POST("/fence/config", regulatory.ConfigureFence)
			reg.GET("/fence/config", regulatory.GetFenceConfig)
			reg.GET("/compliance/report", regulatory.GetComplianceReport)
		}

		// Community elderly wristband
		cwb := api.Group("/community-wb")
		{
			// Elders
			cwb.GET("/elders", communityWB.ListElders)
			cwb.GET("/elders/:id", communityWB.GetElder)
			cwb.POST("/elders", communityWB.CreateElder)
			cwb.PUT("/elders/:id", communityWB.UpdateElder)
			cwb.DELETE("/elders/:id", communityWB.DeleteElder)
			cwb.GET("/elders/:id/welfare", communityWB.GetElderWelfareTags)
			cwb.POST("/elders/:id/welfare/:tag_code", communityWB.AssignWelfareTag)
			cwb.DELETE("/elders/:id/welfare/:tag_code", communityWB.RevokeWelfareTag)
			cwb.GET("/elders/stats", communityWB.GetElderStats)
			// Devices
			cwb.GET("/devices", communityWB.ListDevices)
			cwb.POST("/devices/bind", communityWB.BindElderDevice)
			// Welfare tags config
			cwb.GET("/welfare-tags", communityWB.ListWelfareTags)
			// Sign-in
			cwb.POST("/signin/trigger", communityWB.TriggerSignin)
			cwb.GET("/signin/records", communityWB.ListSigninRecords)
			// Pharmacy
			cwb.POST("/pharmacy/dispense", communityWB.DispenseMedicine)
			// Minzheng
			cwb.POST("/minzheng/import", communityWB.ImportMinzhengData)
			cwb.GET("/minzheng/sync", communityWB.ListMinzhengSync)
			// Batch payments
			cwb.POST("/batch-pay/execute", communityWB.ExecuteBatchPayment)
			cwb.GET("/batch-payments", communityWB.ListBatchPayments)
		}
	}

	return r
}