package router

import (
	"eregen.dev/api-server/internal/handler"
	"eregen.dev/api-server/internal/middleware"
	"eregen.dev/api-server/internal/service"
	"eregen.dev/api-server/internal/store"
	"eregen.dev/api-server/internal/ws"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// New creates the full Gin engine with all route groups.
func New(pg *store.Postgres, redis *store.Redis, nats *service.NatsClient, auth *middleware.JWTAuth, sms *service.SMSProvider, push *service.PushProvider, log *zap.Logger, wsHub *ws.Hub) *gin.Engine {
	r := gin.Default()

	r.Use(corsMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"code": "OK", "message": "Eregen API server is running"})
	})

	r.GET("/ws/alerts", func(c *gin.Context) {
		ws.UpgradeHandler(wsHub)(c.Writer, c.Request)
	})

	authH := handler.NewAuthHandler(pg, redis, auth, service.NewSMSProvider("", "", log), log)
	userH := handler.NewUserHandler(pg, redis, log)
	deviceH := handler.NewDeviceHandler(pg, redis, log)
	alertSvc := service.NewAlertService(pg, push, nats, log)
	alertH := handler.NewAlertHandler(pg, alertSvc, log)
	insightEngine := service.NewInsightEngine(pg, log)
	insightsH := handler.NewInsightsHandler(insightEngine, log)
	otaSvc := service.NewOTAService(pg, nats, log)
	otaH := handler.NewOTAHandler(otaSvc, log)

	rateLimiter, rlErr := middleware.NewSlidingWindowLimiter(log)
	if rlErr != nil {
		log.Warn("rate limiter init failed (will fail open)", zap.Error(rlErr))
	}

	pub := r.Group("/api/v1/auth")
	if rlErr == nil {
		pub.Use(rateLimiter.Anonymous())
	}
	{
		pub.POST("/register", authH.Register)
		pub.POST("/login", authH.Login)
		pub.POST("/refresh", authH.Refresh)
		pub.POST("/logout", authH.Logout)
		pub.POST("/revoke-all-sessions", authH.RevokeAllSessions)
		pub.POST("/device/register", authH.RegisterDevice)
		pub.POST("/send-otp", authH.SendOTP)
		pub.POST("/forgot-password", authH.ForgotPassword)
	}

	protected := r.Group("/api/v1")
	protected.Use(auth.AuthMiddleware())
	if rlErr == nil {
		protected.Use(rateLimiter.Authenticated())
	}
	{
		protected.GET("/users/me", userH.GetMe)
		protected.PUT("/users/me", userH.UpdateMe)

		devices := protected.Group("/devices")
		{
			devices.GET("", deviceH.List)
			devices.POST("", deviceH.Bind)
			devices.GET("/:device_id", auth.ResolveDeviceID(), deviceH.Get)
			devices.PUT("/:device_id/settings", auth.ResolveDeviceID(), deviceH.UpdateSettings)
			devices.DELETE("/:device_id", auth.ResolveDeviceID(), deviceH.Delete)
		}

		elderlyGroup := protected.Group("/elderly")
		{
			elderlyGroup.GET("", userH.ListElderly)
			elderlyGroup.POST("", userH.CreateElderly)

			elderly := elderlyGroup.Group("/:elderly_id")
			elderly.Use(auth.ResolveElderlyID())
			{
				elderly.GET("/profile", userH.GetElderlyProfile)
				elderly.PUT("/profile", userH.UpdateElderlyProfile)
				elderly.POST("/link-device", userH.LinkDeviceToElderly)

				elderly.GET("/health/summary", healthSummary(pg))
				elderly.GET("/health/history", healthHistory(pg))
				elderly.GET("/health/trend", healthTrend(pg))

				elderly.GET("/location/latest", locationLatest(pg))
				elderly.GET("/location/history", locationHistory(pg))
				elderly.POST("/geofence", geofenceSet(pg))
				elderly.GET("/geofence", geofenceList(pg))
				elderly.PUT("/geofence/:geofence_id", geofenceUpdate(pg))
				elderly.DELETE("/geofence/:geofence_id", geofenceDelete(pg))

				elderly.GET("/medication/rules", medRules(pg))
				elderly.POST("/medication/rules", medCreateRule(pg, nats))
				elderly.PUT("/medication/rules/:rule_id", auth.ResolveRuleID(), medUpdateRule(pg))
				elderly.DELETE("/medication/rules/:rule_id", auth.ResolveRuleID(), medDeleteRule(pg))
				elderly.GET("/medication/today", medToday(pg))
				elderly.GET("/medication/history", medHistory(pg))

					insights := elderly.Group("/insights")
					{
						insights.GET("/daily", insightsH.DailyInsight)
						insights.GET("/weekly", insightsH.WeeklyInsight)
					}
				}
			}

		alerts := protected.Group("/alerts")
		{
			alerts.GET("", alertH.List)
			alerts.GET("/:alert_id", auth.ResolveAlertID(), alertH.Get)
			alerts.PUT("/:alert_id", auth.ResolveAlertID(), alertH.Update)
			alerts.POST("/sos/call", alertH.SOSCall)
		}

		admin := protected.Group("/admin")
		{
			firmware := admin.Group("/firmware")
			{
				firmware.POST("", otaH.CreateFirmware)
				firmware.GET("", otaH.ListFirmware)
				firmware.GET("/:id", otaH.GetFirmware)
			}
			admin.POST("/ota/push", otaH.PushOTA)
			admin.GET("/ota/jobs/:id", otaH.GetOTAJob)
		}
	}

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
