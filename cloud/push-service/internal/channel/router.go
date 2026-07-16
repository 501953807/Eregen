package channel

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// Channel defines the push delivery interface.
type Channel interface {
	// Send delivers a notification. Returns error on failure.
	Send(ctx context.Context, recipient, title, body string) error
}

// Router fans out events to all available channels.
type Router struct {
	fcm    Channel
	wechat Channel
	sms    Channel
}

// NewRouter creates a channel router with all configured channels.
// A channel may be nil if its provider is not configured (dev mode).
func NewRouter(fcm, wechat, sms Channel) *Router {
	return &Router{
		fcm:    fcm,
		wechat: wechat,
		sms:    sms,
	}
}

// Deliver sends a notification to ALL available channels for a recipient.
// Each channel retries up to 2 times with exponential backoff.
func (r *Router) Deliver(ctx context.Context, event, title, body string, familyMembers []FamilyMember) {
	for _, member := range familyMembers {
		r.deliverToMember(ctx, event, title, body, member)
	}
}

func (r *Router) deliverToMember(ctx context.Context, event, title, body string, member FamilyMember) {
	channels := []struct {
		name  string
		ch    Channel
		retry int
	}{
		{"fcm", r.fcm, 2},
		{"wechat", r.wechat, 2},
		{"sms", r.sms, 1},
	}

	for _, ch := range channels {
		if ch.ch == nil {
			continue // provider not configured
		}

		recipient := member.DeviceToken
		if ch.name == "wechat" {
			recipient = member.OpenID
		} else if ch.name == "sms" {
			recipient = member.Phone
		}

		err := r.withRetry(ctx, ch.ch, recipient, title, body, ch.retry)
		if err != nil {
			log.Printf("[router] %s failed for %s: %v", ch.name, recipient, err)
		} else {
			log.Printf("[router] %s delivered to %s", ch.name, recipient)
		}
	}
}

func (r *Router) withRetry(ctx context.Context, ch Channel, recipient, title, body string, maxRetries int) error {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(attempt) * 2 * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
		if err := ch.Send(ctx, recipient, title, body); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	return fmt.Errorf("all %d attempts failed: %w", maxRetries+1, lastErr)
}

// FamilyMember holds delivery targets for one family account.
type FamilyMember struct {
	UserID      string
	DeviceToken string // FCM token
	OpenID      string // WeChat open_id
	Phone       string // Mobile number
}

// ParseRecipient extracts delivery targets from user profile.
// In production this queries PostgreSQL via the store layer.
func ParseMember(userID string) (FamilyMember, error) {
	// Placeholder — real implementation reads from DB
	return FamilyMember{UserID: userID}, nil
}

// DeliverAlert distributes an alert event across all channels.
func (r *Router) DeliverAlert(ctx context.Context, members []FamilyMember, alertType, message string, severity string) {
	title := "颐贞紧急告警"
	if strings.HasPrefix(severity, "P1") {
		title = "颐贞重要提醒"
	} else {
		title = "颐贞通知"
	}

	for _, m := range members {
		r.deliverToMember(ctx, "alert", title, message, m)
	}
}

// DeliverReminder distributes a medication reminder.
func (r *Router) DeliverReminder(ctx context.Context, members []FamilyMember, message string) {
	title := "颐贞用药提醒"
	for _, m := range members {
		r.deliverToMember(ctx, "reminder", title, message, m)
	}
}
