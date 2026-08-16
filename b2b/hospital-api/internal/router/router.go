package router

import (
	"eregen.dev/b2b-hospital-api/internal/handler"
	"eregen.dev/b2b-hospital-api/internal/middleware"
	"eregen.dev/b2b-hospital-api/internal/model"
	"eregen.dev/b2b-hospital-api/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Register(r *gin.Engine, st store.Database, log *zap.Logger) {
	instH := handler.NewInstitutionHandler(st, log)
	healthH := handler.NewHealthDataHandler(st, log)
	linkH := handler.NewLinkHandler(st, log)
	alertH := handler.NewAlertHandler(st, log)

	// Public: institution registration (no auth required)
	r.POST("/api/v2/b2b/institutions", instH.Create)

	// Protected by API key auth
	b2b := r.Group("/api/v2/b2b")
	b2b.Use(middleware.APIKeyAuth(st, log))
	{
		// Institution management
		institutions := b2b.Group("/institutions")
		{
			institutions.GET("", instH.List)
			institutions.GET("/:id", instH.Get)
			institutions.PUT("/:id", instH.Update)
			institutions.POST("/:id/api-keys", instH.CreateAPIKey)
		}

		// Health data ingestion from hospitals
		healthGroup := b2b.Group("")
		healthGroup.Use(middleware.RequireAccess(model.AccessReadWrite))
		{
			healthGroup.POST("/health-data", healthH.Receive)
			healthGroup.GET("/patients/:eregen_id/report", healthH.GetReport)
		}

		// Elderly-institution linking
		links := b2b.Group("/links")
		{
			links.POST("", linkH.Create)
			links.GET("/institutions/:id", linkH.ListByInstitution)
			links.GET("/elderly/:id", linkH.ListByElderly)
		}

		// Alert forwarding
		alerts := b2b.Group("/alerts")
		{
			alerts.POST("/forward", alertH.Forward)
		}
	}
}
