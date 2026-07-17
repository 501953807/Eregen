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
func Setup(db *sql.DB, logger *zap.Logger) *gin.Engine {
	s := store.NewStore(db)
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
	}

	return r
}
