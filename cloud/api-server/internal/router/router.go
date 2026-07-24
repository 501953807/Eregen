package router

import (
	"net/http"
	"eregen.dev/api-server/internal/handler"
	"eregen.dev/api-server/internal/middleware"
	"eregen.dev/api-server/internal/service"
	"eregen.dev/api-server/internal/store"
	"eregen.dev/api-server/internal/ws"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// New creates the full Gin engine with all route groups.
func New(pg *store.Postgres, redis *store.Redis, nats *service.NatsClient, auth *middleware.JWTAuth, deviceAuth *middleware.DeviceAuth, sms *service.SMSProvider, push *service.PushProvider, log *zap.Logger, wsHub *ws.Hub, corsOrigins []string) *gin.Engine {
	r := gin.Default()

	r.Use(corsMiddleware(corsOrigins))

	// Request body size limit: 1MB for normal API, 10MB for OTA uploads
	r.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/api/v1/admin/firmware" {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20)
		} else if c.Request.ContentLength > 1<<20 {
			c.AbortWithStatusJSON(413, gin.H{"code": "PAYLOAD_TOO_LARGE", "message": "Request body exceeds 1MB limit"})
			return
		}
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"code": "OK", "message": "Eregen API server is running"})
	})

	r.GET("/ws/alerts", func(c *gin.Context) {
		ws.UpgradeHandler(wsHub)(c.Writer, c.Request)
	})

	// Device MQTT gateway endpoint (separate from user-facing API)
	deviceH := handler.NewDeviceHandler(pg, redis, nats, log)
	deviceMW := deviceAuth.DeviceAuthMiddleware()
	devicesPub := r.Group("/api/v1/devices")
	devicesPub.Use(deviceMW)
	{
		devicesPub.POST("/telemetry", deviceH.HandleTelemetry)
		devicesPub.POST("/heartbeat", deviceH.HandleHeartbeat)
		devicesPub.POST("/location", deviceH.HandleLocation)
	}

	authH := handler.NewAuthHandler(pg, redis, auth, sms, log)
	userH := handler.NewUserHandler(pg, redis, log)
	alertSvc := service.NewAlertService(pg, push, redis, nats, log)
	alertH := handler.NewAlertHandler(pg, alertSvc, log)
	insightEngine := service.NewInsightEngine(pg, log)
	insightsH := handler.NewInsightsHandler(insightEngine, log)
	otaSvc := service.NewOTAService(pg, nats, log)
	otaH := handler.NewOTAHandler(otaSvc, log)

	// Medication interaction checker
	interactionChecker := service.NewMedicationInteractionChecker(log)
	interactionH := handler.NewMedicationInteractionHandler(interactionChecker, log)

	// Emergency response workflow
	emergencyStore := pg
	emergencyWF := service.NewEmergencyResponseWorkflow(emergencyStore, push, nil, log)
	alertSvc.SetEmergencyWorkflow(emergencyWF)
	emergencyH := handler.NewEmergencyHandler(emergencyWF, log)

	// New aggregate handlers
	healthAgg := handler.NewHealthAggregateHandler(pg, log)
	subH := handler.NewSubscriptionHandler(pg, log)
	userListH := handler.NewUserListHandler(pg, log)
	medTakeH := handler.NewMedicationTakeHandler(pg, log)
	alertHandleH := handler.NewAlertHandleHandler(pg, log)
	dataExportSvc := service.NewDataExportService(pg, log)
	dataExportH := handler.NewDataExportHandler(dataExportSvc, log)
	statsH := handler.NewAdminStatsHandler(pg, log)

	// Audit logger and handler
	auditLogger := service.NewAuditLogger(10000, log)
	auditH := handler.NewAuditHandler(auditLogger, log)
	auditMW := middleware.NewAuditMiddleware(auditLogger)

	rateLimiter, rlErr := middleware.NewSlidingWindowLimiter(log)
	if rlErr != nil {
		log.Warn("rate limiter init failed (will fail open)", zap.Error(rlErr))
	}

	pub := r.Group("/api/v1/auth")
	if rlErr == nil {
		pub.Use(rateLimiter.Anonymous())
	}
	{
		pub.POST("/register", auditMW.LogAction(service.ActionUserRegister, "user", "", nil), authH.Register)
		pub.POST("/login", auditMW.LogAction(service.ActionUserLogin, "user", "", nil), authH.Login)
		pub.POST("/logout", auditMW.LogAction(service.ActionUserLogout, "user", "", nil), authH.Logout)
		pub.POST("/revoke-all-sessions", auditMW.LogAction(service.ActionUserLogout, "user", "all-sessions", nil), authH.RevokeAllSessions)
		pub.POST("/device/register", authH.RegisterDevice)
		pub.POST("/send-otp", authH.SendOTP)
		pub.POST("/send-code", authH.SendCode)
		pub.POST("/phone-login", auditMW.LogAction(service.ActionUserLogin, "user", "", nil), authH.PhoneLogin)
		pub.POST("/wechat/login", auditMW.LogAction(service.ActionUserLogin, "user", "", nil), authH.WechatLogin)
		pub.POST("/forgot-password", authH.ForgotPassword)
	}

	protected := r.Group("/api/v1")
	protected.Use(auth.AuthMiddleware())
	if rlErr == nil {
		protected.Use(rateLimiter.Authenticated())
	}
	{
		protected.GET("/users/me", userH.GetMe)
		protected.PUT("/users/me", auditMW.LogAction(service.ActionUserUpdate, "user", "", nil), userH.UpdateMe)

		devices := protected.Group("/devices")
		{
			devices.GET("", deviceH.List)
			devices.POST("", auditMW.LogAction(service.ActionDeviceBind, "device", "", nil), deviceH.Bind)
			devices.GET("/:device_id", auth.ResolveDeviceID(), deviceH.Get)
			devices.PUT("/:device_id/settings", auth.ResolveDeviceID(), deviceH.UpdateSettings)
			devices.DELETE("/:device_id", auditMW.LogAction(service.ActionDeviceUnbind, "device", "", nil), auth.ResolveDeviceID(), deviceH.Delete)
		}

		elderlyGroup := protected.Group("/elderly")
		{
			elderlyGroup.GET("", userH.ListElderly)
			elderlyGroup.POST("", auditMW.LogAction(service.ActionElderlyCreate, "elderly", "", nil), userH.CreateElderly)

			elderly := elderlyGroup.Group("/:elderly_id")
			elderly.Use(auth.ResolveElderlyID())
			{
				elderly.GET("/profile", userH.GetElderlyProfile)
				elderly.PUT("/profile", auditMW.LogAction(service.ActionElderlyUpdate, "elderly", "", nil), userH.UpdateElderlyProfile)
				elderly.POST("/link-device", auditMW.LogAction(service.ActionDeviceBind, "device", "", nil), userH.LinkDeviceToElderly)

				elderly.GET("/health/summary", healthSummary(pg))
				elderly.GET("/health/history", healthHistory(pg))
				elderly.GET("/health/trend", healthTrend(pg))

				elderly.GET("/location/latest", locationLatest(pg))
				elderly.GET("/location/history", locationHistory(pg))
				elderly.POST("/geofence", auditMW.LogAction(service.ActionAdminAction, "geofence", "", nil), geofenceSet(pg))
				elderly.GET("/geofence", geofenceList(pg))
				elderly.PUT("/geofence/:geofence_id", auditMW.LogAction(service.ActionAdminAction, "geofence", "", nil), geofenceUpdate(pg))
				elderly.DELETE("/geofence/:geofence_id", auditMW.LogAction(service.ActionAdminAction, "geofence", "", nil), geofenceDelete(pg))

				elderly.GET("/medication/rules", medRules(pg))
				elderly.POST("/medication/rules", auditMW.LogAction(service.ActionMedicationRule, "medication_rule", "", nil), medCreateRule(pg, nats))
				elderly.PUT("/medication/rules/:rule_id", auditMW.LogAction(service.ActionMedicationRule, "medication_rule", "", nil), auth.ResolveRuleID(), medUpdateRule(pg))
				elderly.DELETE("/medication/rules/:rule_id", auditMW.LogAction(service.ActionMedicationRule, "medication_rule", "", nil), auth.ResolveRuleID(), medDeleteRule(pg))
				elderly.GET("/medication/today", medToday(pg))
				elderly.GET("/medication/history", medHistory(pg))
				elderly.POST("/medication/check-interactions", interactionH.CheckInteractions)
				elderly.POST("/medication/check-conditions", interactionH.CheckConditions)

				insights := elderly.Group("/insights")
				{
					insights.GET("/daily", insightsH.DailyInsight)
					insights.GET("/weekly", insightsH.WeeklyInsight)
				}
			}
		}

		protected.GET("/health/latest", healthAgg.Latest)
		protected.GET("/health/records", healthAgg.Records)
		protected.GET("/health/risk-score", healthAgg.RiskScore)

		protected.GET("/subscriptions", subH.List)
		protected.GET("/subscriptions/stats", subH.Stats)

		protected.GET("/users", userListH.List)
		protected.GET("/users/:id", userListH.Get)

		protected.POST("/medication/:rule_id/take", medTakeH.Take)

		alerts := protected.Group("/alerts")
		{
			alerts.GET("", alertH.List)
			alerts.GET("/:alert_id", auth.ResolveAlertID(), alertH.Get)
			alerts.PUT("/:alert_id", auth.ResolveAlertID(), alertH.Update)
			alerts.PUT("/:id/handle", auditMW.LogAction(service.ActionAlertResolve, "alert", "", nil), alertHandleH.Handle)
			alerts.POST("/share-location", auditMW.LogAction(service.ActionAdminAction, "alert", "", nil), alertHandleH.ShareLocation)
			alerts.POST("/sos/call", alertH.SOSCall)
			alerts.PUT("/:alert_id/resolve", auditMW.LogAction(service.ActionAlertResolve, "alert", "", nil), emergencyH.ResolveAlert)
			alerts.GET("/active-cases", emergencyH.GetActiveCases)
		}

		admin := protected.Group("/admin")
		{
			// Dashboard statistics
			admin.GET("/stats/overview", statsH.Overview)
			admin.GET("/stats/alert-trend", statsH.AlertTrend)
			admin.GET("/stats/alert-distribution", statsH.AlertDistribution)
			admin.GET("/stats/user-growth", statsH.UserGrowth)

			// User management
			admin.PUT("/users/:id/role", userListH.UpdateRole)

			// Device management (admin)
			admin.GET("/devices", deviceH.AdminList)
			admin.GET("/devices/:id", deviceH.AdminGetDevice)
			admin.PUT("/devices/:id/settings", deviceH.AdminUpdateSettings)
			admin.DELETE("/devices/:id", deviceH.AdminDeleteDevice)
			admin.POST("/devices/:id/ota", deviceH.AdminOTAPush)
			admin.POST("/devices/batch-ota", deviceH.AdminBatchOTAPush)

			firmware := admin.Group("/firmware")
			{
				firmware.POST("", auditMW.LogAction(service.ActionAdminAction, "firmware", "", nil), otaH.CreateFirmware)
				firmware.GET("", otaH.ListFirmware)
				firmware.GET("/:id", otaH.GetFirmware)
				firmware.POST("/:id/verify", otaH.VerifyFirmware)
			}
			admin.POST("/ota/push", auditMW.LogAction(service.ActionOTAUpdate, "ota_job", "", nil), otaH.PushOTA)
			admin.GET("/ota/jobs/:id", otaH.GetOTAJob)

			// Audit log endpoints
			admin.GET("/audit-logs", auditH.List)
		}

		// User's own audit logs
		protected.GET("/users/me/audit-logs", auditH.MyLogs)

		data := protected.Group("/data")
		{
			data.POST("/export", auditMW.LogAction(service.ActionAdminAction, "data_export", "", nil), dataExportH.CreateExportRequest)
			data.GET("/export/status", dataExportH.GetDataExportStatus)
			data.GET("/export/:user_id/download", dataExportH.DownloadExport)
			data.POST("/delete", auditMW.LogAction(service.ActionAdminAction, "data_deletion", "", nil), dataExportH.RequestDeletion)
			data.GET("/delete/status", dataExportH.GetDeletionStatus)
		}
	}

	return r
}

func corsMiddleware(origins []string) gin.HandlerFunc {
	allowed := make(map[string]bool)
	if len(origins) == 0 {
		origins = []string{"http://localhost:3000", "http://localhost:5173", "http://127.0.0.1:3000"}
	}
	for _, o := range origins {
		allowed[o] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if allowed[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		} else {
			c.AbortWithStatusJSON(403, gin.H{"code": "CORS_DENIED", "message": "Origin not allowed"})
			return
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Device-Token")
		c.Header("Access-Control-Max-Age", "86400")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
