// © 2026 Eregen (颐贞). All rights reserved.

// Package ratelimit provides a Redis-backed sliding-window rate limiter.
package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Limiter implements a Redis-based sliding window rate limiter using sorted sets.
type Limiter struct {
	rdb *redis.Client
}

// NewLimiter creates a new rate limiter backed by the given Redis client.
func NewLimiter(rdb *redis.Client) *Limiter {
	return &Limiter{rdb: rdb}
}

// Allow checks if the request identified by key is allowed under the given per-minute limit.
// It uses a sorted-set sliding window: each request is stored with its Unix timestamp as score,
// old entries are pruned, and the cardinality is compared against the limit.
// Returns true when the request is within the limit, false when it would exceed it.
// On Redis errors the method fails open (allows the request).
func (l *Limiter) Allow(ctx context.Context, key string, limitPerMinute int) bool {
	now := time.Now().Unix()
	windowKey := fmt.Sprintf("ratelimit:%s", key)
	windowStart := now - 60

	pipe := l.rdb.Pipeline()
	pipe.ZRemRangeByScore(ctx, windowKey, "0", fmt.Sprintf("%d", windowStart))
	pipe.ZAdd(ctx, windowKey, redis.Z{
		Score:  float64(now),
		Member: float64(now),
	})
	pipe.Expire(ctx, windowKey, 61*time.Second)
	countCmd := pipe.ZCard(ctx, windowKey)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return true // fail open on Redis errors
	}
	return countCmd.Val() <= int64(limitPerMinute)
}

// Middleware returns a Gin middleware for per-user rate limiting.
// It extracts user_id from gin.Context (set by JWT auth middleware);
// if unavailable it falls back to ClientIP.
func (l *Limiter) Middleware(limitPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		key := "user:"
		if uid, ok := userID.(string); ok && uid != "" {
			key += uid
		} else {
			key += c.ClientIP()
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 200*time.Millisecond)
		defer cancel()

		if !l.Allow(ctx, key, limitPerMinute) {
			c.JSON(429, gin.H{
				"code":    "RATE_LIMITED",
				"message": "Too many requests, please try again later",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// MiddlewareIP returns a Gin middleware for per-IP rate limiting.
func (l *Limiter) MiddlewareIP(limitPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		ctx, cancel := context.WithTimeout(c.Request.Context(), 200*time.Millisecond)
		defer cancel()

		if !l.Allow(ctx, "ip:"+ip, limitPerMinute) {
			c.JSON(429, gin.H{
				"code":    "RATE_LIMITED",
				"message": "Too many requests from your IP",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
