package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"eregen.dev/api-server/internal/config"
	"eregen.dev/api-server/internal/middleware"
	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/router"
	"eregen.dev/api-server/internal/service"
	"eregen.dev/api-server/internal/store"
	"eregen.dev/api-server/internal/ws"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	log, _ := zap.NewProduction()
	defer log.Sync()

	// PostgreSQL via pgxpool
	pgConfig, err := pgxpool.ParseConfig(cfg.DBURL)
	if err != nil {
		log.Fatal("invalid postgres URL", zap.Error(err))
	}
	pgConfig.MaxConns = 10
	pgPool, err := pgxpool.NewWithConfig(context.Background(), pgConfig)
	if err != nil {
		log.Fatal("failed to connect to postgres", zap.Error(err))
	}
	pg := store.NewPostgres(pgPool, log)
	defer pgPool.Close()

	// Redis
	rdbOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatal("invalid redis URL", zap.Error(err))
	}
	rdb := redis.NewClient(rdbOpts)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Warn("redis not available", zap.Error(err))
	}
	redisLayer := store.NewRedis(rdb, log)

	// WebSocket Hub for real-time alerts
	wsHub := ws.NewHub()
	go wsHub.Run(context.Background())

	// NATS
	var natsClient *service.NatsClient
	natsClient, err = service.NewNatsClient(cfg.NATSURL, log)
	if err != nil {
		log.Warn("nats not available", zap.Error(err))
	} else {
		defer natsClient.Close()

		eventHandler := service.NewEventHandler(nil, log)
		eventHandler.SetAlertCallback(func(ctx context.Context, a *model.Alert) error {
			if err := pg.CreateAlert(ctx, a); err != nil {
				return err
			}
			// Broadcast to WebSocket clients
			wsHub.PublishAlert(ws.AlertBroadcast{
				ElderlyID: a.ElderlyID,
				Type:      a.AlertType,
				Payload:   a.Metadata,
				Timestamp: a.CreatedAt,
			})
			return nil
		})
		eventHandler.SetHealthCallback(func(ctx context.Context, r *model.HealthRecord) error {
			return pg.CreateHealthRecord(ctx, r)
		})
		eventHandler.SetLocationCallback(func(ctx context.Context, r *model.LocationRecord) error {
			return pg.CreateLocationRecord(ctx, r)
		})
		eventHandler.SetMedStatusCallback(func(ctx context.Context, r *model.MedStatusRecord) error {
			return pg.CreateMedStatusRecord(ctx, r)
		})

		// Create OTA service and wire progress callback
		otaSvc := service.NewOTAService(pg, natsClient, log)
		eventHandler.SetOTACallback(func(ctx context.Context, jobID, deviceID, status string) error {
			return otaSvc.UpdateProgress(ctx, jobID, deviceID, status)
		})

		go func() {
			ctx := context.Background()
			if err := natsClient.SubscribeDeviceEvents(ctx, eventHandler); err != nil {
				log.Error("nats subscriber failed", zap.Error(err))
			}
		}()
	}

	sms := service.NewSMSProvider(cfg.SMSAccessKey, cfg.SMSAccessSecret, cfg.SMSSignName, cfg.SMPTemplateID, log)
	push := service.NewPushProvider(cfg.FCMKeyPath, cfg.FCMProjectID, log)

	authMW := middleware.NewJWTAuth(cfg.JWTSecret, time.Duration(cfg.TokenExpiry)*time.Second,
		time.Duration(cfg.RefreshExpiry)*time.Second, log, pg)

	deviceAuth := middleware.NewDeviceAuth(pg, log, cfg.DeviceSecret)

	r := router.New(pg, redisLayer, natsClient, authMW, deviceAuth, sms, push, log, wsHub, cfg.CORSOrigins)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Info("starting API server", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server forced shutdown", zap.Error(err))
	}
	log.Info("server exited")
}
