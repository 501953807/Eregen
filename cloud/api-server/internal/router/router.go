package router

import (
	"net/http"
	"strings"

	"eregen.dev/api-server/internal/handler"
	"eregen.dev/api-server/internal/middleware"
	"eregen.dev/api-server/internal/service"
	"eregen.dev/api-server/internal/store"
	"eregen.dev/api-server/internal/ws"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// New creates the Gin engine with IoT device routes only.
func New(pg store.Store, redis *store.Redis, nats *service.NatsClient, auth *middleware.JWTAuth, deviceAuth *middleware.DeviceAuth, log *zap.Logger, wsHub *ws.Hub, corsOrigins []string) *gin.Engine {
	r := gin.Default()

	// Security Headers Middleware
	r.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/v1/") || strings.HasPrefix(path, "/admin") {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			c.Header("X-Frame-Options", "DENY")
			c.Header("X-Content-Type-Options", "nosniff")
			c.Header("X-XSS-Protection", "1; mode=block")
			c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
			if c.Request.Host == "localhost:8080" || c.Request.Host == "127.0.0.1:8080" {
				c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';")
			} else {
				c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; object-src 'none';")
			}
		}
		c.Next()
	})

	r.Use(corsMiddleware(corsOrigins))

	// Request body size limit
	r.Use(func(c *gin.Context) {
		if c.Request.ContentLength > 10<<20 {
			c.AbortWithStatusJSON(413, gin.H{"code": "PAYLOAD_TOO_LARGE", "message": "Request body exceeds 10MB limit"})
			return
		}
		c.Next()
	})

	// Health check
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "ok", "component": "api-server"}})
	})

	// Readiness probe
	r.GET("/api/v1/health/ready", func(c *gin.Context) {
		checks := make(map[string]string)
		if err := pg.Pool().Ping(c.Request.Context()); err == nil {
			checks["database"] = "ok"
		} else {
			checks["database"] = "unavailable"
		}
		checks["redis"] = "unknown"
		if redis != nil {
			if _, err := redis.IsDeviceOnline(c.Request.Context(), "health_check"); err == nil {
				checks["redis"] = "ok"
			} else {
				checks["redis"] = "unavailable"
			}
		}
		checks["nats"] = "unknown"
		if nats != nil {
			if err := nats.Ping(c.Request.Context()); err == nil {
				checks["nats"] = "connected"
			} else {
				checks["nats"] = "unavailable"
			}
		}
		overallStatus := "ok"
		for _, v := range checks {
			if v != "ok" && v != "connected" && v != "unknown" {
				overallStatus = "degraded"
				break
			}
		}
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": overallStatus, "checks": checks}})
	})

	// WebSocket alerts
	r.GET("/ws/alerts", func(c *gin.Context) {
		ws.UpgradeHandler(wsHub)(c.Writer, c.Request)
	})

	// Device handler
	deviceH := handler.NewDeviceHandler(pg, pg, pg, redis, nats, log)
	deviceMW := deviceAuth.DeviceAuthMiddleware()
	devicesPub := r.Group("/api/v1/devices")
	devicesPub.Use(deviceMW)
	{
		devicesPub.POST("/telemetry", deviceH.HandleTelemetry)
		devicesPub.POST("/heartbeat", deviceH.HandleHeartbeat)
		devicesPub.POST("/location", deviceH.HandleLocation)
	}

	// Auth for device registration
	authH := handler.NewAuthHandler(pg, redis, auth, nil, log)

	// User-facing device management (JWT auth)
	protected := r.Group("/api/v1")
	protected.Use(auth.AuthMiddleware())
	{
		devices := protected.Group("/devices")
		{
			devices.GET("", deviceH.List)
			devices.POST("", deviceH.Bind)
			devices.GET("/:device_id", auth.ResolveDeviceID(), deviceH.Get)
			devices.PUT("/:device_id/settings", auth.ResolveDeviceID(), deviceH.UpdateSettings)
			devices.DELETE("/:device_id", auth.ResolveDeviceID(), deviceH.Delete)
		}

		protected.POST("/auth/device/register", authH.RegisterDevice)
	}

	// Admin device & OTA management (role-gated)
	admin := protected.Group("/admin")
	admin.Use(auth.RequireRole("institution"))
	{
		admin.GET("/devices", deviceH.AdminList)
		admin.GET("/devices/:id", deviceH.AdminGetDevice)
		admin.PUT("/devices/:id/settings", deviceH.AdminUpdateSettings)
		admin.DELETE("/devices/:id", deviceH.AdminDeleteDevice)
		admin.POST("/devices/:id/ota", deviceH.AdminOTAPush)
		admin.POST("/devices/batch-ota", deviceH.AdminBatchOTAPush)

		otaSvc := service.NewOTAService(pg, pg, nats, log)
		otaH := handler.NewOTAHandler(otaSvc, log)

		firmware := admin.Group("/firmware")
		{
			firmware.POST("", otaH.CreateFirmware)
			firmware.GET("", otaH.ListFirmware)
			firmware.GET("/:id", otaH.GetFirmware)
			firmware.POST("/:id/verify", otaH.VerifyFirmware)
		}
		admin.POST("/ota/push", otaH.PushOTA)
		admin.GET("/ota/jobs/:id", otaH.GetOTAJob)
	}

	return r
}

func corsMiddleware(origins []string) gin.HandlerFunc {
	allowed := make(map[string]bool)
	if len(origins) == 0 {
		origins = []string{"http://localhost:3000", "http://localhost:3100", "http://localhost:5173", "http://127.0.0.1:3000", "http://127.0.0.1:3100"}
	}
	for _, o := range origins {
		allowed[o] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			c.Next()
			return
		}
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
