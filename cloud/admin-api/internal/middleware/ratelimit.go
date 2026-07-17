// © 2026 Eregen (颐贞). All rights reserved.

package middleware

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// AdminRateLimiter enforces a strict 30 req/min limit for admin users.
type AdminRateLimiter struct {
	rdb *redis.Client
}

// NewAdminRateLimiter creates a new admin rate limiter using Redis DB 1.
func NewAdminRateLimiter() (*AdminRateLimiter, error) {
	addr := getEnv("REDIS_ADDR", "localhost:6379")
	password := getEnv("REDIS_PASSWORD", "")

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       1, // separate DB for admin traffic
	})
	// Fail open on startup — if Redis is down, admin routes proceed without limiting.
	return &AdminRateLimiter{rdb: rdb}, nil
}

func (l *AdminRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		adminRole, _ := c.Get("admin_role")
		key := "admin:"
		if role, ok := adminRole.(string); ok {
			key += role + ":"
		}
		key += c.GetString("user_id")

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
			log.Printf("admin rate limiter error: %v", err)
			c.Next()
			return
		}
		if count.Val() > 30 {
			c.JSON(429, gin.H{"error": "admin rate limit exceeded"})
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
