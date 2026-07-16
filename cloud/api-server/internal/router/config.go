package router

import (
	"time"

	"eregen.dev/api-server/internal/store"
	"eregen.dev/api-server/internal/service"

	"go.uber.org/zap"
)

// Config holds the dependencies needed to wire the router.
type Config struct {
	PG           *store.Postgres
	Redis        *store.Redis
	Nats         *service.NatsClient
	JWTSecret    string
	TokenExpiry  time.Duration
	RefreshExpiry time.Duration
	Log          *zap.Logger

	SMSSignName     string
	SMSTemplateID   string
	FCMProjectID    string
	FCMServerKey    string
}
