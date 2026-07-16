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
		events.POST("/:id/register", eventH.Register)
		events.GET("/:id/registrations", eventH.GetRegistrations)
	}

	healthChecks := r.Group("/api/v2/b2b/health-checks")
	{
		healthChecks.POST("", healthH.Create)
		healthChecks.GET("/:elderly_id", healthH.GetForElderly)
	}

	carePlans := r.Group("/api/v2/b2b/care-plans")
	{
		carePlans.POST("", careH.Create)
		carePlans.GET("/:elderly_id", careH.GetForElderly)
	}
}
