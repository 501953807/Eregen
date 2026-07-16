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

	claims := r.Group("/api/v2/b2b/claims")
	{
		claims.POST("", claimH.Create)
		claims.GET("/:id", claimH.Get)
		claims.PUT("/:id/status", claimH.UpdateStatus)
		claims.GET("/elderly/:elderly_id", claimH.GetForElderly)
	}

	policies := r.Group("/api/v2/b2b/policies")
	{
		policies.POST("", policyH.Create)
		policies.GET("/elderly/:elderly_id", policyH.GetForElderly)
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
	}
}
