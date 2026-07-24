package router

import (
	"database/sql"
	"log"
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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "change-me-in-production"
	}
	adminJWT := middleware.NewAdminJWT(jwtSecret, 24*time.Hour, logger)

	r.Use(adminJWT.AuthMiddleware())

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

	api := r.Group("/api/v1/admin")
	if rlErr == nil {
		api.Use(rateLimiter.Middleware())
	}
	{
		api.GET("/stats/overview", dashboard.GetOverview)
		api.GET("/stats/subscriptions", dashboard.GetSubscriptionStats)
		api.GET("/devices", device.List)
		api.GET("/users", user.List)
		api.GET("/alerts", alert.List)
		// User role management
		api.POST("/users/:id/role", user.SetRole)
		// Device config and OTA
		api.POST("/devices/:id/config", device.UpdateConfig)
		api.POST("/devices/:id/ota", device.TriggerOTA)
		// Alert resolution
		api.POST("/alerts/:id/resolve", alert.Resolve)
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
