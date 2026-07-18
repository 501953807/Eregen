package store

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	deviceOnlineKey   = "device:online:%s"
	latestHealthKey   = "health:latest:%s"
	latestLocationKey = "location:latest:%s"
	deviceTTL         = 5 * time.Minute
	healthTTL         = 5 * time.Minute
	locationTTL       = 5 * time.Minute
)

// Redis wraps cache operations using go-redis.
type Redis struct {
	client *redis.Client
	log    *zap.Logger
}

// NewRedis creates a new cache layer backed by the given client.
func NewRedis(client *redis.Client, log *zap.Logger) *Redis {
	return &Redis{client: client, log: log}
}

// SetDeviceOnline marks a device as online with a TTL.
func (r *Redis) SetDeviceOnline(ctx context.Context, deviceID string) error {
	key := formatDeviceKey(deviceID)
	return r.client.Set(ctx, key, "online", deviceTTL).Err()
}

// IsDeviceOnline checks if a device is currently online.
func (r *Redis) IsDeviceOnline(ctx context.Context, deviceID string) (bool, error) {
	key := formatDeviceKey(deviceID)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == "online", nil
}

// InvalidateDevice removes cached device status.
func (r *Redis) InvalidateDevice(ctx context.Context, deviceID string) error {
	key := formatDeviceKey(deviceID)
	return r.client.Del(ctx, key).Err()
}

// SetLatestHealth caches today's health summary for an elderly person.
func (r *Redis) SetLatestHealth(ctx context.Context, elderlyID string, data map[string]any) error {
	key := formatHealthKey(elderlyID)
	return r.client.Set(ctx, key, data, healthTTL).Err()
}

// GetLatestHealth retrieves cached latest health readings.
func (r *Redis) GetLatestHealth(ctx context.Context, elderlyID string) (map[string]any, error) {
	key := formatHealthKey(elderlyID)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// Simple string-to-map; callers parse as needed
	data := make(map[string]any)
	data["raw"] = val
	return data, nil
}

// SetLatestLocation caches the most recent GPS fix.
func (r *Redis) SetLatestLocation(ctx context.Context, elderlyID string, data map[string]any) error {
	key := formatLocationKey(elderlyID)
	return r.client.Set(ctx, key, data, locationTTL).Err()
}

// GetLatestLocation retrieves cached latest location.
func (r *Redis) GetLatestLocation(ctx context.Context, elderlyID string) (map[string]any, error) {
	key := formatLocationKey(elderlyID)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	data := make(map[string]any)
	data["raw"] = val
	return data, nil
}

// SetRefreshToken stores a refresh token in Redis with expiry.
func (r *Redis) SetRefreshToken(ctx context.Context, token string, userID string, ttl time.Duration) error {
	key := "token:refresh:" + token
	return r.client.Set(ctx, key, userID, ttl).Err()
}

// ValidateRefreshToken checks if a refresh token exists and returns its user ID.
func (r *Redis) ValidateRefreshToken(ctx context.Context, token string) (string, error) {
	key := "token:refresh:" + token
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // token not found, will be treated as invalid
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

// InvalidateRefreshToken blacklists a refresh token.
func (r *Redis) InvalidateRefreshToken(ctx context.Context, token string) error {
	key := "token:refresh:" + token
	return r.client.Del(ctx, key).Err()
}

// SetOTP stores an OTP code with short TTL.
func (r *Redis) SetOTP(ctx context.Context, phoneOrEmail, code string, ttl time.Duration) error {
	key := "otp:" + normalizeKey(phoneOrEmail)
	return r.client.Set(ctx, key, code, ttl).Err()
}

// VerifyOTP checks an OTP code and deletes it (one-time use).
func (r *Redis) VerifyOTP(ctx context.Context, phoneOrEmail, code string) error {
	key := "otp:" + normalizeKey(phoneOrEmail)
	stored, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	// Delete immediately — one-time use
	r.client.Del(ctx, key)
	if stored != code {
		return redis.Nil // treat as wrong code
	}
	return nil
}

// DelByPattern deletes all keys matching a Redis glob pattern.
// Used for bulk token revocation on logout/session management.
func (r *Redis) DelByPattern(ctx context.Context, pattern string) error {
	var cursor uint64
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			r.client.Del(ctx, keys...)
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}

// SetResetToken stores a password reset token.
func (r *Redis) SetResetToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	key := "reset:" + token
	return r.client.Set(ctx, key, userID, ttl).Err()
}

// GetResetToken retrieves and validates a reset token.
func (r *Redis) GetResetToken(ctx context.Context, token string) (string, error) {
	key := "reset:" + token
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// Store implements OTPStore.Store — stores an OTP code with TTL.
func (r *Redis) Store(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.SetOTP(ctx, key, value, ttl)
}

// Verify implements OTPStore.Verify — checks and consumes a one-time OTP.
func (r *Redis) Verify(ctx context.Context, key, value string) error {
	return r.VerifyOTP(ctx, key, value)
}

func formatDeviceKey(deviceID string) string {
	return "device:online:" + deviceID
}

func formatHealthKey(elderlyID string) string {
	return "health:latest:" + elderlyID
}

func formatLocationKey(elderlyID string) string {
	return "location:latest:" + elderlyID
}

func normalizeKey(s string) string {
	return s
}
