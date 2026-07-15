// © 2026 Eregen (颐贞). All rights reserved.

package mqtt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"
)

// DeviceAuth handles device authentication and rate limiting.
type DeviceAuth struct {
	secretKey string
	// In production, this would query a database or Redis.
	// For dev, we accept any valid prefix.
	devices     map[string]bool
	deviceMu    sync.RWMutex
	rateLimiters map[string]*RateLimiter
	rlMu        sync.Mutex
}

// RateLimiter enforces per-device message rate limits.
type RateLimiter struct {
	mu         sync.Mutex
	messages   int
	windowStart time.Time
	limit      int
	window     time.Duration
}

// NewDeviceAuth creates a new auth handler.
func NewDeviceAuth(secretKey string) *DeviceAuth {
	return &DeviceAuth{
		secretKey:    secretKey,
		devices:      make(map[string]bool),
		rateLimiters: make(map[string]*RateLimiter),
	}
}

// AuthenticateDevice validates a device's identity using HMAC-SHA256 token.
// For development, it accepts any device ID with valid BR- or PX- prefix.
func (da *DeviceAuth) AuthenticateDevice(deviceID string, token string) error {
	// Validate device ID format.
	if !deviceIDRegex.MatchString(deviceID) {
		return fmt.Errorf("invalid device ID format: %s", deviceID)
	}

	// Verify HMAC token: HMAC-SHA256(deviceID + secret_key)
	expected := hmacSHA256(deviceID+da.secretKey, da.secretKey)
	if token != expected {
		return fmt.Errorf("invalid authentication token for device %s", deviceID)
	}

	// Rate limit check.
	rl := da.getRateLimiter(deviceID)
	if !rl.Allow() {
		log.Printf("WARN: device %s exceeded rate limit", deviceID)
		return fmt.Errorf("rate limit exceeded for device %s", deviceID)
	}

	return nil
}

// BindDeviceToUser records the binding between a device and a user.
// In production this would be stored in Redis with a TTL.
func (da *DeviceAuth) BindDeviceToUser(deviceID, userID string) error {
	if !deviceIDRegex.MatchString(deviceID) {
		return fmt.Errorf("invalid device ID format: %s", deviceID)
	}
	// In production: SET gateway:device:user:{dev_id} {user_id} EX 86400
	da.deviceMu.Lock()
	da.devices[deviceID] = true
	da.deviceMu.Unlock()
	return nil
}

// IsDeviceKnown checks if a device is registered in the system.
func (da *DeviceAuth) IsDeviceKnown(deviceID string) bool {
	da.deviceMu.RLock()
	defer da.deviceMu.RUnlock()
	return da.devices[deviceID]
}

func (da *DeviceAuth) getRateLimiter(deviceID string) *RateLimiter {
	da.rlMu.Lock()
	defer da.rlMu.Unlock()

	rl, ok := da.rateLimiters[deviceID]
	if !ok {
		rl = &RateLimiter{
			limit:       100,
			window:      time.Second,
			windowStart: time.Now(),
		}
		da.rateLimiters[deviceID] = rl
	}
	return rl
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	if now.Sub(rl.windowStart) > rl.window {
		rl.messages = 0
		rl.windowStart = now
	}

	rl.messages++
	return rl.messages <= rl.limit
}

func hmacSHA256(message, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
