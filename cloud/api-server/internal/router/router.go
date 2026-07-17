package router

import (
	"eregen.dev/api-server/internal/handler"
	"eregen.dev/api-server/internal/middleware"
	"eregen.dev/api-server/internal/service"
	"eregen.dev/api-server/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// New creates the full Gin engine with all route groups.
func New(pg *store.Postgres, redis *store.Redis, nats *service.NatsClient, auth *middleware.JWTAuth, sms *service.SMSProvider, push *service.PushProvider, log *zap.Logger) *gin.Engine {
	r := gin.Default()

	r.Use(corsMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"code": "OK", "message": "Eregen API server is running"})
	})

	authH := handler.NewAuthHandler(pg, redis, auth, service.NewSMSProvider("", "", log), log)
	userH := handler.NewUserHandler(pg, redis, log)
	deviceH := handler.NewDeviceHandler(pg, redis, log)
	alertSvc := service.NewAlertService(pg, push, nats, log)
	alertH := handler.NewAlertHandler(pg, alertSvc, log)

	// Rate limiter — fail open if Redis is unavailable at startup
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
			devices.GET("/:device_id", auth.ResolveDeviceID(), deviceH.Get)
			devices.PUT("/:device_id/settings", auth.ResolveDeviceID(), deviceH.UpdateSettings)
			devices.DELETE("/:device_id", auth.ResolveDeviceID(), deviceH.Delete)
		}

		elderly := protected.Group("/elderly/:elderly_id")
		elderly.Use(auth.ResolveElderlyID())
		{
			elderly.GET("/profile", userH.GetElderlyProfile)
			elderly.PUT("/profile", userH.UpdateElderlyProfile)

			// Health endpoints delegate to store
			elderly.GET("/health/summary", healthSummary(pg))
			elderly.GET("/health/history", healthHistory(pg))
			elderly.GET("/health/trend", healthTrend(pg))

			// Location endpoints
			elderly.GET("/location/latest", locationLatest(pg))
			elderly.GET("/location/history", locationHistory(pg))
			elderly.POST("/geofence", geofenceSet())
			elderly.GET("/geofence", geofenceList())

			// Medication endpoints
			elderly.GET("/medication/rules", medRules(pg))
			elderly.POST("/medication/rules", medCreateRule(pg, nats))
			elderly.PUT("/medication/rules/:rule_id", auth.ResolveRuleID(), medUpdateRule(pg))
			elderly.DELETE("/medication/rules/:rule_id", auth.ResolveRuleID(), medDeleteRule(pg))
			elderly.GET("/medication/today", medToday(pg))
			elderly.GET("/medication/history", medHistory(pg))
		}

		alerts := protected.Group("/alerts")
		{
			alerts.GET("", alertH.List)
			alerts.GET("/:alert_id", auth.ResolveAlertID(), alertH.Get)
			alerts.PUT("/:alert_id", auth.ResolveAlertID(), alertH.Update)
			alerts.POST("/sos/call", alertH.SOSCall)
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
