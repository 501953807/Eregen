package router

import (
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
func Setup(s store.Store, logger *zap.Logger) *gin.Engine {
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
patient := handler.NewPatientHandler(s)
	_wristband := handler.NewWristbandHandler(s)
	clinical := handler.NewClinicalHandler(s)
	admission := handler.NewAdmissionHandler(s)
	regulatory := handler.NewRegulatoryHandler(s)
	communityWB := handler.NewCommunityWBHandler(s)
	subscription := handler.NewSubscriptionHandler(s)
	institution := handler.NewInstitutionHandler(s)
	person := handler.NewPersonHandler(s)
	lifecycle := handler.NewLifecycleHandler(s)
	alertRule := handler.NewAlertRuleHandler(s)
	medication := handler.NewMedicationHandler(s)
	healthRecord := handler.NewHealthRecordHandler(s)
	healthGuidance := handler.NewHealthGuidanceHandler(s)
	healthReport := handler.NewHealthReportHandler(s)
	compliance := handler.NewComplianceHandler(s)
	deviceBinding := handler.NewDeviceBindingHandler(s)
	notification := handler.NewNotificationHandler(s)

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
		// Subscription management
		api.GET("/subscriptions", subscription.List)
		api.GET("/subscriptions/:id", subscription.Get)
		api.POST("/subscriptions", subscription.Create)
		api.PUT("/subscriptions/:id", subscription.Update)
		api.POST("/subscriptions/:id/renew", subscription.Renew)
		// Institution management
		api.GET("/institutions", institution.List)
		api.GET("/institutions/:id", institution.Get)
		api.POST("/institutions", institution.Create)
		api.PUT("/institutions/:id", institution.Update)
		api.DELETE("/institutions/:id", institution.Delete)
		api.POST("/institutions/:id/api-keys", institution.CreateAPIKey)
		api.DELETE("/institutions/:id/api-keys/:key_id", institution.RevokeAPIKey)
		api.GET("/devices", device.List)
		api.GET("/users", user.List)
		api.POST("/users", user.Create)
		api.PUT("/users/:id", user.Update)
		api.DELETE("/users/:id", user.Delete)
		// User role management
		api.POST("/users/:id/role", user.SetRole)
		api.GET("/alerts", alert.List)
		api.POST("/alerts", alert.Create)
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
		api.POST("/elderly/:id/health-records", elderly.CreateHealthRecord)
		api.POST("/elderly/:id/locations", elderly.CreateLocation)

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

		// Person unified management (self/hospital/community chains)
		persons := api.Group("/persons")
		{
			persons.GET("", person.List)
			persons.GET("/:id", person.Get)
			persons.POST("", person.Create)
			persons.PUT("/:id", person.Update)
			persons.DELETE("/:id", person.Delete)
			persons.POST("/profile", person.CreateProfile)
			persons.GET("/:id/profile", person.GetProfile)
			persons.POST("/welfare-tags", person.AssignWelfareTag)
			persons.DELETE("/:id/welfare-tags/:tag_code", person.RevokeWelfareTag)
			persons.GET("/:id/welfare-tags", person.ListWelfareTags)
		}
		// Person lifecycle / cross-chain transitions
		api.PUT("/persons/:id/status", lifecycle.TransitionStatus)
		api.POST("/persons/link", lifecycle.LinkPerson)
		// Medical wristband management
		med := api.Group("/medical")
		{
			// Patient endpoints
			med.GET("/patients", patient.ListPatients)
			med.GET("/patients/:id", patient.GetPatient)
			med.POST("/patients", patient.CreatePatient)
			med.PUT("/patients/:id", patient.UpdatePatient)
			med.DELETE("/patients/:id", patient.DeletePatient)
			med.GET("/patients/by-admission", patient.GetByAdmissionNo)
			med.POST("/patients/batch-import", patient.BatchImport)
			med.GET("/patients/:id/history", patient.GetPatientHistory)

			// Wristband device endpoints
			med.GET("/wristbands", _wristband.ListWristbands)
			med.POST("/wristbands/bind", _wristband.BindWristband)
			med.POST("/wristbands/:id/unbind", _wristband.UnbindWristband)
			med.POST("/wristbands/:id/clear", _wristband.ClearWristband)
			med.POST("/wristbands/:id/write", _wristband.WriteToWristband)
			med.GET("/wristbands/:id/firmware", _wristband.GetFirmware)

			// Expense endpoints
			med.GET("/patients/:id/expenses", clinical.ListExpenses)
			med.POST("/expenses", clinical.CreateExpense)

			// Medication endpoints
			med.GET("/patients/:id/medications", clinical.ListMedications)
			med.POST("/medications", clinical.CreateMedication)

			// Test result endpoints
			med.GET("/patients/:id/test-results", clinical.ListTestResults)
			med.POST("/test-results", clinical.CreateTestResult)

			// Daily entry endpoints
			med.GET("/patients/:id/daily-entries", clinical.ListDailyEntries)
			med.POST("/daily-entries", clinical.CreateDailyEntry)

			// Verification endpoints
			med.GET("/verifications", clinical.ListVerifications)
			med.POST("/verifications", clinical.CreateVerification)
			med.PUT("/verifications/:id/status", clinical.UpdateVerificationStatus)
			med.GET("/verifications/stats/today", clinical.GetTodayVerificationStats)

			// Stats and alert tags
			med.GET("/stats/overview", clinical.GetStatsOverview)
			med.GET("/alert-tags", clinical.ListAlertTagConfigs)
			med.POST("/alert-tags", clinical.CreateAlertTagConfig)

			// Clinical workflow endpoints
			med.POST("/admissions", admission.AdmitPatient)
			med.GET("/admissions", admission.ListAdmissions)
			med.POST("/admissions/:id/discharge", admission.DischargePatient)
			med.GET("/patients/:id/ward-round", admission.GetWardRound)
			med.POST("/patients/:id/ward-round", admission.CompleteWardRound)
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
			cwb.GET("/pharmacy/logs", communityWB.ListPharmacyLogs)
			// Minzheng
			cwb.POST("/minzheng/import", communityWB.ImportMinzhengData)
			cwb.GET("/minzheng/sync", communityWB.ListMinzhengSync)
			// Batch payments
			cwb.POST("/batch-pay/execute", communityWB.ExecuteBatchPayment)
			cwb.GET("/batch-payments", communityWB.ListBatchPayments)
		}

		// ========== Business chain route groups with RequireChain middleware ==========

		// Self chain: elderly profiles, health reports, guidance
		selfGroup := api.Group("/self")
		selfGroup.Use(adminJWT.RequireChain("self"))
		{
			selfGroup.GET("/elderly", elderly.List)
			selfGroup.GET("/elderly/:id", elderly.Detail)
			selfGroup.POST("/elderly", elderly.Create)
			selfGroup.PUT("/elderly/:id", elderly.Update)
			selfGroup.DELETE("/elderly/:id", elderly.Delete)
			selfGroup.GET("/elderly/:id/health-report", elderly.HealthStats)
			selfGroup.POST("/elderly/:id/guidance", func(c *gin.Context) {
				// Guidance evaluation uses the same store; inline for now
				c.JSON(http.StatusOK, gin.H{"code": "OK", "data": []interface{}{}})
			})
		}

		// Hospital chain: patients, admissions, daily entries, verify
		hospitalGroup := api.Group("/hospital")
		hospitalGroup.Use(adminJWT.RequireChain("hospital"))
		{
			hospitalGroup.GET("/patients", patient.ListPatients)
			hospitalGroup.POST("/patients", patient.CreatePatient)
			hospitalGroup.POST("/admissions", admission.AdmitPatient)
			hospitalGroup.POST("/admissions/:id/discharge", admission.DischargePatient)
			hospitalGroup.GET("/patients/:id/daily", clinical.ListDailyEntries)
			hospitalGroup.POST("/patients/:id/verify", func(c *gin.Context) {
				// Verification is handled via clinical handler
				c.JSON(http.StatusOK, gin.H{"code": "OK"})
			})
		}

		// Community chain: elders, signin, welfare, stats
		communityGroup := api.Group("/community")
		communityGroup.Use(adminJWT.RequireChain("community"))
		{
			communityGroup.GET("/elders", communityWB.ListElders)
			communityGroup.POST("/elders", communityWB.CreateElder)
			communityGroup.PUT("/elders/:id", communityWB.UpdateElder)
			communityGroup.POST("/elders/:id/signin", communityWB.TriggerSignin)
			communityGroup.POST("/elders/:id/welfare", communityWB.AssignWelfareTag)
			communityGroup.GET("/elders/:id/stats", communityWB.GetElderStats)
		}

		// Regulatory chain: compliance, audit, reports
		regulatoryGroup := api.Group("/regulatory")
		regulatoryGroup.Use(adminJWT.RequireChain("regulatory"))
		{
			regulatoryGroup.GET("/compliance", regulatory.GetComplianceReport)
			regulatoryGroup.POST("/compliance/run", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"code": "OK", "data": "compliance check queued"})
			})
			regulatoryGroup.GET("/audit/:patientId", regulatory.GetAuditTrail)
			regulatoryGroup.GET("/reports", regulatory.GetComplianceReport)
		}
	}

	// Public medical endpoints (for family app / mini program — no JWT auth required)
	publicMed := r.Group("/api/v1/medical")
	{
		publicMed.GET("/patients/:id/history", patient.GetPatientHistory)
		publicMed.GET("/patients/:id/expenses", clinical.ListExpenses)
		publicMed.GET("/patients/:id/medications", clinical.ListMedications)
		publicMed.GET("/patients/:id/test-results", clinical.ListTestResults)
		publicMed.GET("/patients/:id/daily-entries", clinical.ListDailyEntries)
		publicMed.GET("/verifications", clinical.ListVerifications)
	}

	// Unified person-centric API routes (issues 05-14 handlers)
	// Alert rules per chain
	alertRules := api.Group("/alert-rules")
	{
		alertRules.GET("", alertRule.List)
		alertRules.POST("", alertRule.Create)
		alertRules.PUT("/:id", alertRule.Update)
		alertRules.DELETE("/:id", alertRule.Delete)
	}
	// Medication rules & executions per person
	medR := api.Group("/medications")
	{
		medR.GET("", medication.ListRules)
		medR.POST("", medication.CreateRule)
		medR.PUT("/:ruleId", medication.UpdateRule)
		medR.DELETE("/:ruleId", medication.DeleteRule)
		medR.POST("/executions", medication.CreateExecution)
		medR.GET("/executions", medication.ListExecutions)
	}
	// Health records per person
	healthR := api.Group("/health-records")
	{
		healthR.POST("", healthRecord.Create)
		healthR.GET("", healthRecord.List)
		healthR.GET("/summary", healthRecord.GetSummary)
		healthR.PUT("/summary", healthRecord.UpdateSummary)
	}
	// Health guidance per person
	guidanceR := api.Group("/guidance")
	{
		guidanceR.POST("", healthGuidance.CreateDelivery)
		guidanceR.GET("", healthGuidance.ListDeliveries)
		guidanceR.POST("/evaluate", healthGuidance.Evaluate)
	}
	// Health report templates & reports
	templatesR := api.Group("/health-report-templates")
	{
		templatesR.POST("", healthReport.CreateTemplate)
		templatesR.GET("", healthReport.ListTemplates)
	}
	reportsR := api.Group("/health-reports")
	{
		reportsR.POST("", healthReport.CreateReport)
		reportsR.GET("", healthReport.ListReports)
	}
	// Compliance checks per person
	complianceR := api.Group("/compliance-checks")
	{
		complianceR.POST("/run", compliance.RunCheck)
		complianceR.GET("", compliance.ListChecks)
		complianceR.PUT("/checks/:checkId", compliance.ReviewCheck)
	}
	complianceRulesR := api.Group("/compliance-rules")
	{
		complianceRulesR.POST("", compliance.CreateRule)
		complianceRulesR.GET("", compliance.ListRules)
	}
	// Device bindings per person
	bindingsR := api.Group("/device-bindings")
	{
		bindingsR.POST("", deviceBinding.Bind)
		bindingsR.GET("", deviceBinding.ListBindings)
		bindingsR.DELETE("/:bindingId", deviceBinding.Unbind)
	}
	bindingsDevicesR := api.Group("/person-devices")
	{
		bindingsDevicesR.GET("", deviceBinding.ListDevices)
	}
	// Notifications
	notifTemplatesR := api.Group("/notification-templates")
	{
		notifTemplatesR.POST("", notification.CreateTemplate)
		notifTemplatesR.GET("", notification.ListTemplates)
	}
	notifLogsR := api.Group("/notifications")
	{
		notifLogsR.POST("", notification.CreateLog)
		notifLogsR.GET("", notification.ListLogs)
		notifLogsR.PUT("/:logId/status", notification.UpdateStatus)
	}

	return r
}