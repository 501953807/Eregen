package router

import (
	"eregen.dev/b2b-insurance-integration/internal/handler"
	"eregen.dev/b2b-insurance-integration/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Register(r *gin.Engine, pg *store.Postgres, log *zap.Logger) {
	claimH := handler.NewClaimHandler(pg, log)
	policyH := handler.NewPolicyHandler(pg, log)
	evidenceH := handler.NewEvidenceHandler(pg, log)
	exportH := handler.NewExportHandler(pg, log)
	providerH := handler.NewProviderHandler(pg, log)

	providers := r.Group("/api/v2/b2b/providers")
	{
		providers.POST("", providerH.Create)
		providers.GET("", providerH.List)
		providers.GET("/:id", providerH.GetByID)
		providers.PUT("/:id", providerH.Update)
	}

	claims := r.Group("/api/v2/b2b/claims")
	{
		claims.POST("", claimH.Create)
		claims.GET("", claimH.List)
		claims.GET("/:id", claimH.Get)
		claims.PUT("/:id/status", claimH.UpdateStatus)
		claims.GET("/elderly/:elderly_id", claimH.GetForElderly)
	}

	policies := r.Group("/api/v2/b2b/policies")
	{
		policies.POST("", policyH.Create)
		policies.GET("/elderly/:elderly_id", policyH.GetForElderly)
		policies.GET("/:id", policyH.GetByID)
		policies.PUT("/:id", policyH.Update)
	}

	reminders := r.Group("/api/v2/b2b/reminders")
	{
		reminders.POST("", policyH.CreateReminder)
		reminders.GET("/upcoming", policyH.GetUpcoming)
	}

	evidence := r.Group("/api/v2/b2b/evidence")
	{
		evidence.POST("", evidenceH.Upload)
		evidence.GET("/claim/:claim_id", evidenceH.ListByClaim)
	}

	exports := r.Group("/api/v2/b2b/exports")
	{
		exports.POST("", exportH.Create)
		exports.GET("/:id", exportH.GetByID)
	}
}
