package router

import (
	"net/http"
	"strings"

	"eregen.dev/api-server/internal/handler"
	"eregen.dev/api-server/internal/middleware"
	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/service"
	"eregen.dev/api-server/internal/store"
	"eregen.dev/api-server/internal/ws"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// New creates the full Gin engine with all route groups.
func New(pg *store.Postgres, redis *store.Redis, nats *service.NatsClient, auth *middleware.JWTAuth, deviceAuth *middleware.DeviceAuth, sms *service.SMSProvider, push *service.PushProvider, log *zap.Logger, wsHub *ws.Hub, corsOrigins []string) *gin.Engine {
	r := gin.Default()

	// Security Headers Middleware - protects against common web vulnerabilities
	r.Use(func(c *gin.Context) {
		// Only apply to API paths that serve the admin UI (or all paths if needed)
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/v1/") || strings.HasPrefix(path, "/admin") {
			// Strict Transport Security - enforce HTTPS
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			// X-Frame-DENY - prevent clickjacking
			c.Header("X-Frame-Options", "DENY")
			// X-Content-Type-Options - prevent MIME sniffing
			c.Header("X-Content-Type-Options", "nosniff")
			// XSS Protection - enable browser XSS filter
			c.Header("X-XSS-Protection", "1; mode=block")
			// Referrer Policy - control referrer information
			c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
			// Content-Security-Policy - configured per environment
			// Development allows localhost; production should be strict
			if c.Request.Host == "localhost:8080" || c.Request.Host == "127.0.0.1:8080" {
				c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';")
			} else {
				// Production - restrict to actual domains
				c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; object-src 'none';")
			}
		}
		c.Next()
	})

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

	// Health check endpoint - returns simple OK without exposing internal details
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "ok", "component": "api-server"}})
	})

	// Full health check with dependency probing - protected internally or via special auth
	// This endpoint checks all downstream dependencies. Used by Kubernetes liveness probes.
	// In production, this should only be accessible within the cluster network.
	r.GET("/api/v1/health/ready", func(c *gin.Context) {
		checks := make(map[string]string)

		// Check PostgreSQL connection
		if err := pg.Pool().Ping(c.Request.Context()); err == nil {
			checks["database"] = "ok"
		} else {
			checks["database"] = "unavailable"
		}

		// Check Redis connection
		checks["redis"] = "unknown"
		if redis != nil {
			if _, err := redis.IsDeviceOnline(c.Request.Context(), "health_check"); err == nil {
				checks["redis"] = "ok"
			} else {
				checks["redis"] = "unavailable"
			}
		}

		// Check NATS connection using the Ping method
		checks["nats"] = "unknown"
		if nats != nil {
			// Simplified: just report connected/unavailable, hide error details
			if err := nats.Ping(c.Request.Context()); err == nil {
				checks["nats"] = "connected"
			} else {
				checks["nats"] = "unavailable"
			}
		}

		// Determine overall status (simplified, no error details exposed)
		overallStatus := "ok"
		for _, checkResult := range checks {
			if checkResult != "ok" && checkResult != "connected" && checkResult != "unknown" {
				overallStatus = "degraded"
				break
			}
		}

		c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": overallStatus, "checks": checks}})
	})

	r.GET("/ws/alerts", func(c *gin.Context) {
		ws.UpgradeHandler(wsHub)(c.Writer, c.Request)
	})

	// Device MQTT gateway endpoint (separate from user-facing API)
	deviceH := handler.NewDeviceHandler(pg, pg, pg, redis, nats, log)
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
	otaSvc := service.NewOTAService(pg, pg, nats, log)
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

	medicalH := handler.NewMedicalHandler(pg, log)

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
		pub.GET("/csrf/get", authH.GetCSRFToken) // New endpoint for CSRF token - no CSRF required since it's idempotent
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
		// Add CSRF protection for state-changing requests (POST/PUT/DELETE)
		protected.Use(auth.CSRFCheck())

		// GET endpoints that are safe/read-only don't require CSRF
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
				elderly.PUT("/medication/rules/:rule_id", auditMW.LogAction(service.ActionMedicationRule, "medication_rule", "", nil), auth.ResolveRuleID(), medUpdateRule(pg, nats))
				elderly.DELETE("/medication/rules/:rule_id", auditMW.LogAction(service.ActionMedicationRule, "medication_rule", "", nil), auth.ResolveRuleID(), medDeleteRule(pg, nats))
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

		// Medical wristband data for family app
		med := protected.Group("/medical")
		{
			med.GET("/patients/:patient_id/history", medicalH.GetPatientHistory)
			med.GET("/patients/:patient_id/expenses", func(c *gin.Context) {
				patientID := c.Param("patient_id")
				data, _ := medicalH.QueryExpenses(c, patientID)
				c.JSON(http.StatusOK, gin.H{"code": "OK", "data": data})
			})
			med.GET("/patients/:patient_id/medications", func(c *gin.Context) {
				patientID := c.Param("patient_id")
				data, _ := medicalH.QueryMedications(c, patientID)
				c.JSON(http.StatusOK, gin.H{"code": "OK", "data": data})
			})
			med.GET("/patients/:patient_id/test-results", func(c *gin.Context) {
				patientID := c.Param("patient_id")
				data, _ := medicalH.QueryTestResults(c, patientID)
				c.JSON(http.StatusOK, gin.H{"code": "OK", "data": data})
			})
			med.GET("/patients/:patient_id/daily-entries", func(c *gin.Context) {
				patientID := c.Param("patient_id")
				data, _ := medicalH.QueryDailyEntries(c, patientID, "")
				c.JSON(http.StatusOK, gin.H{"code": "OK", "data": data})
			})
			med.GET("/verifications", func(c *gin.Context) {
				patientID := c.Query("patient_id")
				data, _ := medicalH.QueryVerifications(c, patientID)
				c.JSON(http.StatusOK, gin.H{"code": "OK", "data": data})
			})
		}

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
		admin.Use(auth.RequireRole(model.RoleInstitution))
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
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Device-Token, X-CSRF-Token")
		c.Header("Access-Control-Max-Age", "86400")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
