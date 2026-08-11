package router

import (
	"eregen.dev/b2b-community-platform/internal/handler"
	"eregen.dev/b2b-community-platform/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Register(r *gin.Engine, st store.Store, log *zap.Logger) {
	eventH := handler.NewEventHandler(st, log)
	healthH := handler.NewHealthCheckHandler(st, log)
	careH := handler.NewCarePlanHandler(st, log)

	events := r.Group("/api/v2/b2b/events")
	{
		events.POST("", eventH.Create)
		events.GET("", eventH.List)
		events.GET("/:id", eventH.GetByID)
		events.DELETE("/:id", eventH.Delete)
		events.POST("/:id/register", eventH.Register)
		events.POST("/:id/cancel-register", eventH.CancelRegister)
		events.GET("/:id/registrations", eventH.GetRegistrations)
	}

	healthChecks := r.Group("/api/v2/b2b/health-checks")
	{
		healthChecks.POST("", healthH.Create)
		// Use query param for elderly lookup to avoid Gin wildcard conflict
		healthChecks.GET("", healthH.GetForElderly)
		healthChecks.GET("/:id", healthH.GetByID)
		healthChecks.PUT("/:id", healthH.Update)
		healthChecks.DELETE("/:id", healthH.Delete)
	}

	carePlans := r.Group("/api/v2/b2b/care-plans")
	{
		carePlans.POST("", careH.Create)
		// Use query param for elderly lookup to avoid Gin wildcard conflict
		carePlans.GET("", careH.GetForElderly)
		carePlans.GET("/:id", careH.GetByID)
		carePlans.PUT("/:id", careH.Update)
		carePlans.DELETE("/:id", careH.Delete)
	}
}
