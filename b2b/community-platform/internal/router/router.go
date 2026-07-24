package router

import (
	"eregen.dev/b2b-community-platform/internal/handler"
	"eregen.dev/b2b-community-platform/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Register(r *gin.Engine, pg *store.Postgres, log *zap.Logger) {
	eventH := handler.NewEventHandler(pg, log)
	healthH := handler.NewHealthCheckHandler(pg, log)
	careH := handler.NewCarePlanHandler(pg, log)

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
		healthChecks.GET("/:elderly_id", healthH.GetForElderly)
		healthChecks.GET("/:id", healthH.GetByID)
		healthChecks.PUT("/:id", healthH.Update)
		healthChecks.DELETE("/:id", healthH.Delete)
	}

	carePlans := r.Group("/api/v2/b2b/care-plans")
	{
		carePlans.POST("", careH.Create)
		carePlans.GET("/:elderly_id", careH.GetForElderly)
		carePlans.GET("/:id", careH.GetByID)
		carePlans.PUT("/:id", careH.Update)
		carePlans.DELETE("/:id", careH.Delete)
	}
}
