// © 2026 Eregen (颐贞). All rights reserved.

package middleware

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// SlidingWindowLimiter is a minimal sliding-window counter using Redis sorted sets.
type SlidingWindowLimiter struct {
	rdb *redis.Client
	log *zap.Logger
}

// NewSlidingWindowLimiter creates a new limiter from config and logger.
// It does NOT fail if Redis is unavailable; requests will fail open instead.
func NewSlidingWindowLimiter(log *zap.Logger) (*SlidingWindowLimiter, error) {
	addr := getEnv("REDIS_ADDR", "localhost:6379")
	password := getEnv("REDIS_PASSWORD", "")
	db := getEnvAsInt("REDIS_DB", 0)

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	// We do not block startup on Redis availability; Allow() will fail open.
	return &SlidingWindowLimiter{rdb: rdb, log: log}, nil
}

// Allow checks if a key is within the rate limit using a sorted-set sliding window.
func (l *SlidingWindowLimiter) Allow(key string, limit int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	now := time.Now().Unix()
	windowKey := fmt.Sprintf("rl:%s", key)

	pipe := l.rdb.Pipeline()
	pipe.ZRemRangeByScore(ctx, windowKey, "0", fmt.Sprintf("%d", now-60))
	pipe.ZAdd(ctx, windowKey, redis.Z{
		Score:  float64(now),
		Member: float64(now),
	})
	pipe.Expire(ctx, windowKey, 61*time.Second)
	count := pipe.ZCard(ctx, windowKey)
	_, err := pipe.Exec(ctx)
	if err != nil {
		l.log.Warn("rate limiter pipeline error", zap.Error(err))
		return true // fail open
	}
	return count.Val() <= int64(limit)
}

// Authenticated returns a Gin middleware for authenticated users (500 req/min).
// Falls back to IP-based key when user_id is not present in context.
func (l *SlidingWindowLimiter) Authenticated() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		key := "u:"
		if uid, ok := userID.(string); ok && uid != "" {
			key += uid
		} else {
			key += "ip:" + c.ClientIP()
		}
		if !l.Allow(key, 500) {
			c.JSON(429, gin.H{"code": "RATE_LIMITED", "message": "Too many requests"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Anonymous returns a Gin middleware for unauthenticated users (100 req/min).
func (l *SlidingWindowLimiter) Anonymous() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "ip:" + c.ClientIP()
		if !l.Allow(key, 100) {
			c.JSON(429, gin.H{"code": "RATE_LIMITED", "message": "Too many requests"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
