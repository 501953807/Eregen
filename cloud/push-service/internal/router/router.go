package router

import (
	"context"
	"log"
	"strings"
	"time"

	smsChannel "eregen.dev/push/internal/channel/sms"
	wechatChannel "eregen.dev/push/internal/channel/wechat"
	"eregen.dev/push/internal/fcm"
	"eregen.dev/push/internal/model"
)

// Router coordinates event distribution to push channels.
type Router struct {
	fcm    *fcm.Client
	wechat *wechatChannel.WeChatClient
	sms    *smsChannel.SMSClient
}

// NewRouter creates a router with all configured channels.
func NewRouter(fcmClient *fcm.Client, wechatClient *wechatChannel.WeChatClient, smsClient *smsChannel.SMSClient) *Router {
	return &Router{fcm: fcmClient, wechat: wechatClient, sms: smsClient}
}

// DeliverAlert distributes an alert across all available channels.
func (r *Router) DeliverAlert(ctx context.Context, ev model.AlertPushEvent, members []Member) {
	title := "颐贞紧急告警"
	if strings.Contains(string(ev.Severity), "P1") {
		title = "颐贞重要提醒"
	}

	for _, m := range members {
		r.deliverToMember(ctx, m, title, ev.Message)
	}
}

// DeliverReminder sends medication or health reminders.
func (r *Router) DeliverReminder(ctx context.Context, message string, members []Member) {
	for _, m := range members {
		r.deliverToMember(ctx, m, "颐贞用药提醒", message)
	}
}

func (r *Router) deliverToMember(ctx context.Context, m Member, title, body string) {
	// FCM — primary channel
	if m.DeviceToken != "" {
		if err := r.fcm.SendToDevice(ctx, m.DeviceToken, title, body); err != nil {
			log.Printf("[router] fcm failed %s: %v", m.UserID, err)
		}
	}

	// WeChat — secondary channel
	if m.OpenID != "" {
		if err := r.wechat.SendTemplateMessage(m.OpenID, "", map[string]wechatChannel.WeChatData{
			"thing1": {Value: title},
			"thing2": {Value: body},
			"time3":  {Value: time.Now().Format("2006-01-02 15:04")},
		}); err != nil {
			log.Printf("[router] wechat failed %s: %v", m.UserID, err)
		}
	}

	// SMS — fallback for critical alerts
	if m.Phone != "" && (strings.Contains(title, "紧急") || strings.Contains(title, "重要")) {
		if err := r.sms.SendAlert(m.Phone, body); err != nil {
			log.Printf("[router] sms failed %s: %v", m.UserID, err)
		}
	}
}

// Member holds delivery targets for one family account.
type Member struct {
	UserID      string
	DeviceToken string // FCM token
	OpenID      string // WeChat open_id
	Phone       string // Mobile number
}
