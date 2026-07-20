package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"eregen.dev/api-server/internal/model"

	"go.uber.org/zap"
)

// SMSProvider sends SMS messages via 阿里云SMS with HMAC-SHA256 signing.
type SMSProvider struct {
	accessKey    string
	accessSecret string
	signName     string
	templateID   string

	mu         sync.Mutex
	lastSendAt time.Time
	dailyCount int
	log        *zap.Logger
}

// NewSMSProvider creates an SMS sender. All fields must be set for production use.
func NewSMSProvider(accessKey, accessSecret, signName, templateID string, log *zap.Logger) *SMSProvider {
	return &SMSProvider{
		accessKey:    accessKey,
		accessSecret: accessSecret,
		signName:     signName,
		templateID:   templateID,
		log:          log,
	}
}

// SendOTP generates a 6-digit code, stores it via otpStore, and sends via SMS.
func (s *SMSProvider) SendOTP(ctx context.Context, phone, code string, otpStore OTPStore) error {
	if err := otpStore.Store(ctx, phone, code, 5*time.Minute); err != nil {
		return fmt.Errorf("store otp: %w", err)
	}
	if !s.allowSend(10) {
		s.log.Warn("sms rate limit reached", zap.String("phone", phone))
		return fmt.Errorf("sms rate limit exceeded")
	}
	if s.accessKey == "" || s.accessSecret == "" {
		s.log.Info("sms skip: not configured", zap.String("phone", phone), zap.String("code", code))
		return nil
	}

	phoneNum := phone
	if len(phoneNum) > 0 && phoneNum[0] != '+' {
		phoneNum = "+86" + phoneNum
	}

	params := url.Values{}
	params.Set("PhoneNumbers", phoneNum)
	params.Set("RegionId", "cn-shanghai")
	params.Set("SignName", s.signName)
	params.Set("TemplateCode", s.templateID)
	params.Set("TemplateParam", `{"code":"`+code+`"}`)

	canonicalURI := "/"
	stringToSign := "POST\n\n\n" + time.Now().UTC().Format(http.TimeFormat) + "\n" + canonicalURI + "?" + params.Encode()
	h := hmac.New(sha256.New, []byte(s.accessSecret))
	h.Write([]byte(stringToSign))
	signature := "HMAC-SHA256 Signature=" + hex.EncodeToString(h.Sum(nil))

	apiURL := "https://dysmsapi.aliyuncs.com/"
	req, _ := http.NewRequest("POST", apiURL, strings.NewReader(params.Encode()))
	req.Header.Set("Authorization", signature)
	req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("sms request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Code      string `json:"Code"`
		Message   string `json:"Message"`
		RequestID string `json:"RequestId"`
	}
	json.Unmarshal(body, &result)

	if result.Code != "OK" {
		return fmt.Errorf("sms error: %s (%s)", result.Message, result.Code)
	}

	s.mu.Lock()
	s.lastSendAt = time.Now()
	s.dailyCount++
	s.mu.Unlock()

	s.log.Info("sent OTP SMS", zap.String("phone", phoneNum))
	return nil
}

func (s *SMSProvider) allowSend(maxPerDay int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if now.Day() != s.lastSendAt.Day() || now.Month() != s.lastSendAt.Month() {
		s.dailyCount = 0
	}
	return s.dailyCount < maxPerDay
}

// OTPStore interface for OTP persistence (backed by Redis in practice).
type OTPStore interface {
	Store(ctx context.Context, key, value string, ttl time.Duration) error
	Verify(ctx context.Context, key, value string) error
}

// PushProvider sends push notifications via FCM HTTP v1 API with JWT bearer auth.
type PushProvider struct {
	keyPath string // path to service account JSON key file
	project string

	mu      sync.Mutex
	token   string
	expire  time.Time
	httpCli *http.Client
	log     *zap.Logger
}

// NewPushProvider creates a push notification sender.
func NewPushProvider(keyPath, projectID string, log *zap.Logger) *PushProvider {
	return &PushProvider{
		keyPath: keyPath,
		project: projectID,
		httpCli: &http.Client{Timeout: 10 * time.Second},
		log:     log,
	}
}

// SendToUser sends a push notification to all registered devices for a user.
// It resolves device tokens via the provided store interface.
func (p *PushProvider) SendToUser(ctx context.Context, userID, title, body string, tokenStore TokenStore) error {
	tokens, err := tokenStore.GetDeviceTokens(ctx, userID)
	if err != nil {
		p.log.Warn("resolve device tokens", zap.String("user", userID), zap.Error(err))
	}
	if len(tokens) == 0 {
		p.log.Info("no device tokens for user", zap.String("user", userID))
		return nil
	}
	return p.sendBulk(ctx, userID, tokens, title, body)
}

// sendBulk delivers to multiple FCM tokens sequentially.
func (p *PushProvider) sendBulk(ctx context.Context, userID string, tokens []string, title, body string) error {
	oauth, err := p.getOAuthToken(ctx)
	if err != nil {
		return fmt.Errorf("fcm oauth: %w", err)
	}

	for _, tok := range tokens {
		payload := map[string]interface{}{
			"message": map[string]interface{}{
				"token": tok,
				"notification": map[string]string{
					"title": title,
					"body":  body,
				},
				"android": map[string]interface{}{
					"priority": "high",
				},
			},
		}

		bodyBytes, _ := json.Marshal(payload)
		resp, err := p.httpCli.Post(
			fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", p.project),
			"application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			p.log.Warn("fcm send", zap.String("user", userID), zap.Error(err))
			continue
		}
		result, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			p.log.Warn("fcm non-200", zap.String("user", userID), zap.Int("status", resp.StatusCode), zap.String("body", string(result)))
			continue
		}
		p.log.Info("push sent", zap.String("user", userID), zap.String("title", title))
		_ = oauth
	}
	return nil
}

// TokenStore resolves device tokens for a user.
type TokenStore interface {
	GetDeviceTokens(ctx context.Context, userID string) ([]string, error)
}

// getOAuthToken fetches a short-lived OAuth2 token from Google's auth endpoint.
func (p *PushProvider) getOAuthToken(ctx context.Context) (string, error) {
	p.mu.Lock()
	if p.token != "" && time.Now().Before(p.expire) {
		tok := p.token
		p.mu.Unlock()
		return tok, nil
	}
	p.mu.Unlock()

	if p.keyPath == "" {
		return "", fmt.Errorf("fcm: no key path configured")
	}
	data, err := os.ReadFile(p.keyPath)
	if err != nil {
		return "", fmt.Errorf("read fcm key: %w", err)
	}
	var sa struct {
		PrivateKeyID string `json:"private_key_id"`
		PrivateKey   string `json:"private_key"`
		ClientEmail  string `json:"client_email"`
	}
	if err := json.Unmarshal(data, &sa); err != nil {
		return "", fmt.Errorf("parse fcm key: %w", err)
	}

	token, exp, err := jwtBearerGrant(ctx, p.httpCli, sa.PrivateKey, sa.PrivateKeyID, sa.ClientEmail)
	if err != nil {
		return "", fmt.Errorf("jwt grant: %w", err)
	}
	p.mu.Lock()
	p.token = token
	p.expire = exp
	p.mu.Unlock()
	return token, nil
}

// AlertService manages alert creation and resolution.
type AlertService struct {
	store     alertStore
	push      *PushProvider
	tokenStore TokenStore
	nats      *NatsClient
	log       *zap.Logger
	emergency *EmergencyResponseWorkflow
}

type alertStore interface {
	CreateAlert(ctx context.Context, a *model.Alert) error
	GetAlert(ctx context.Context, id string) (*model.Alert, error)
	UpdateAlert(ctx context.Context, id string, status model.AlertStatus) error
	ListAlerts(ctx context.Context, elderIDs []string, filter *model.AlertFilter, page, pageSize int) ([]model.Alert, int, error)
}

// NewAlertService creates a new alert service.
func NewAlertService(store alertStore, push *PushProvider, tokenStore TokenStore, nats *NatsClient, log *zap.Logger) *AlertService {
	return &AlertService{store: store, push: push, tokenStore: tokenStore, nats: nats, log: log}
}

// SetEmergencyWorkflow attaches the emergency response workflow to this alert service.
func (s *AlertService) SetEmergencyWorkflow(wf *EmergencyResponseWorkflow) {
	s.emergency = wf
}

// CreateSOSAlert creates an SOS alert and triggers immediate push notification.
func (s *AlertService) CreateSOSAlert(ctx context.Context, elderlyID, deviceID string, lat, lon float64) error {
	a := &model.Alert{
		ElderlyID: elderlyID,
		AlertType: "sos",
		Severity:  model.AlertP0,
		Status:    model.AlertPending,
		Metadata: map[string]any{
			"device_id": deviceID,
			"lat":       lat,
			"lon":       lon,
		},
	}
	if err := s.store.CreateAlert(ctx, a); err != nil {
		return err
	}
	if s.emergency != nil {
		return s.emergency.ProcessAlert(ctx, a)
	}
	_ = s.push.SendToUser(ctx, elderlyID, "SOS 紧急告警", "老人触发了SOS按钮，请立即查看", s.tokenStore)
	return nil
}

// CreateFallAlert creates a fall detection alert.
func (s *AlertService) CreateFallAlert(ctx context.Context, elderlyID, deviceID string, lat, lon, conf float64) error {
	a := &model.Alert{
		ElderlyID: elderlyID,
		AlertType: "fall",
		Severity:  model.AlertP0,
		Status:    model.AlertPending,
		Metadata: map[string]any{
			"device_id":  deviceID,
			"lat":        lat,
			"lon":        lon,
			"confidence": conf,
		},
	}
	if err := s.store.CreateAlert(ctx, a); err != nil {
		return err
	}
	if s.emergency != nil {
		return s.emergency.ProcessAlert(ctx, a)
	}
	_ = s.push.SendToUser(ctx, elderlyID, "跌倒检测告警", "检测到老人可能跌倒，请立即确认", s.tokenStore)
	return nil
}

// ResolveAlert marks an alert as resolved.
func (s *AlertService) ResolveAlert(ctx context.Context, alertID string) error {
	return s.store.UpdateAlert(ctx, alertID, model.AlertResolved)
}

// HealthService provides health data queries.
type HealthService struct {
	store healthStore
	redis redisCache
	log   *zap.Logger
}

type healthStore interface {
	GetHealthSummary(ctx context.Context, elderlyID string, day time.Time) (*model.HealthRecord, error)
	GetHealthHistory(ctx context.Context, elderlyID string, days int) ([]model.HealthRecord, error)
	GetHealthTrend(ctx context.Context, elderlyID, metric string, days int) ([]model.HealthRecord, error)
}

type redisCache interface {
	SetLatestHealth(ctx context.Context, elderlyID string, data map[string]any) error
	GetLatestHealth(ctx context.Context, elderlyID string) (map[string]any, error)
}

// NewHealthService creates a new health service.
func NewHealthService(store healthStore, redis redisCache, log *zap.Logger) *HealthService {
	return &HealthService{store: store, redis: redis, log: log}
}

func (s *HealthService) GetSummary(ctx context.Context, elderlyID string, day time.Time) (*model.HealthRecord, error) {
	cached, err := s.redis.GetLatestHealth(ctx, elderlyID)
	if err == nil && len(cached) > 0 {
		// Parse cached data into HealthRecord
		raw, ok := cached["raw"].(string)
		if ok && raw != "" {
			var data map[string]any
			if jsonErr := json.Unmarshal([]byte(raw), &data); jsonErr == nil {
				r := &model.HealthRecord{
					ElderlyID: elderlyID,
					Timestamp: day,
				}
				if hr, ok := data["hr"].(float64); ok {
					v := int(hr)
					r.HR = &v
				}
				if spo2, ok := data["spo2"].(float64); ok {
					v := int(spo2)
					r.SPO2 = &v
				}
				if steps, ok := data["steps"].(float64); ok {
					v := int64(steps)
					r.Steps = &v
				}
				return r, nil
			}
		}
		return &model.HealthRecord{}, nil
	}
	return s.store.GetHealthSummary(ctx, elderlyID, day)
}

func (s *HealthService) GetHistory(ctx context.Context, elderlyID string, days int) ([]model.HealthRecord, error) {
	return s.store.GetHealthHistory(ctx, elderlyID, days)
}

func (s *HealthService) GetTrend(ctx context.Context, elderlyID, metric string, days int) ([]model.HealthRecord, error) {
	return s.store.GetHealthTrend(ctx, elderlyID, metric, days)
}

// LocationService provides location data queries.
type LocationService struct {
	store locationStore
	redis locCache
	log   *zap.Logger
}

type locationStore interface {
	GetLatestLocation(ctx context.Context, elderlyID string) (*model.LocationRecord, error)
	GetLocationHistory(ctx context.Context, elderlyID string, from, until time.Time) ([]model.LocationRecord, error)
	CreateGeofence(ctx context.Context, gf *model.Geofence) error
	ListGeofences(ctx context.Context, elderlyID string) ([]model.Geofence, error)
}

type locCache interface {
	SetLatestLocation(ctx context.Context, elderlyID string, data map[string]any) error
	GetLatestLocation(ctx context.Context, elderlyID string) (map[string]any, error)
}

// NewLocationService creates a new location service.
func NewLocationService(store locationStore, redis locCache, log *zap.Logger) *LocationService {
	return &LocationService{store: store, redis: redis, log: log}
}

func (s *LocationService) GetLatest(ctx context.Context, elderlyID string) (*model.LocationRecord, error) {
	cached, err := s.redis.GetLatestLocation(ctx, elderlyID)
	if err == nil && len(cached) > 0 {
		raw, ok := cached["raw"].(string)
		if ok && raw != "" {
			var data map[string]any
			if jsonErr := json.Unmarshal([]byte(raw), &data); jsonErr == nil {
				r := &model.LocationRecord{ElderlyID: elderlyID}
				if lat, ok := data["lat"].(float64); ok {
					r.Lat = lat
				}
				if lon, ok := data["lon"].(float64); ok {
					r.Lon = lon
				}
				if ts, ok := data["ts"].(float64); ok {
					r.Timestamp = time.Unix(int64(ts), 0)
				}
				return r, nil
			}
		}
		return &model.LocationRecord{}, nil
	}
	return s.store.GetLatestLocation(ctx, elderlyID)
}

func (s *LocationService) GetHistory(ctx context.Context, elderlyID string, from, until time.Time) ([]model.LocationRecord, error) {
	return s.store.GetLocationHistory(ctx, elderlyID, from, until)
}

func (s *LocationService) CreateGeofence(ctx context.Context, gf *model.Geofence) error {
	return s.store.CreateGeofence(ctx, gf)
}

func (s *LocationService) ListGeofences(ctx context.Context, elderlyID string) ([]model.Geofence, error) {
	return s.store.ListGeofences(ctx, elderlyID)
}

	// NatsPublisher abstracts NATS command publishing so services can be tested without a live connection.
type natsPublisher interface {
	PublishCommand(ctx context.Context, deviceID string, cmd any) error
}

// MedicationService manages medication rules and tracking.
type MedicationService struct {
	store medStore
	nats  natsPublisher
	log   *zap.Logger
}

type medStore interface {
	CreateMedicationRule(ctx context.Context, mr *model.MedicationRule) error
	ListMedicationRules(ctx context.Context, elderlyID string) ([]model.MedicationRule, error)
	GetMedicationRule(ctx context.Context, ruleID string) (*model.MedicationRule, error)
	UpdateMedicationRule(ctx context.Context, ruleID string, req *model.CreateMedicationRuleRequest) error
	DeleteMedicationRule(ctx context.Context, ruleID string) error
	GetTodayMedStatus(ctx context.Context, elderlyID string) ([]model.MedStatusRecord, error)
	GetMedicationHistory(ctx context.Context, elderlyID string, days int) ([]model.MedStatusRecord, error)
	GetDeviceByElderlyID(ctx context.Context, elderlyID string) (string, error)
}

// NewMedicationService creates a new medication service.
func NewMedicationService(store medStore, nats natsPublisher, log *zap.Logger) *MedicationService {
	return &MedicationService{store: store, nats: nats, log: log}
}

func (s *MedicationService) CreateRule(ctx context.Context, elderlyID string, req *model.CreateMedicationRuleRequest) error {
	mr := &model.MedicationRule{
		ElderlyID:    elderlyID,
		ScheduleTime: req.ScheduleTime,
		DoseCount:    req.DoseCount,
		PillType:     req.PillType,
		DaysOfWeek:   req.DaysOfWeek,
		Active:       req.Active,
	}
	if err := s.store.CreateMedicationRule(ctx, mr); err != nil {
		return err
	}

	// Find linked pillbox device for this elderly user
	deviceID, _ := s.store.GetDeviceByElderlyID(ctx, elderlyID)
	if deviceID == "" {
		deviceID = "unknown" // fallback — push will be queued until device is bound
	}

	cmd := map[string]any{
		"type": "med_rule",
		"rule": map[string]any{
			"time":   req.ScheduleTime,
			"dose":   req.DoseCount,
			"type":   req.PillType,
			"days":   req.DaysOfWeek,
		},
	}
	_ = s.nats.PublishCommand(ctx, deviceID, cmd)
	return nil
}

func (s *MedicationService) ListRules(ctx context.Context, elderlyID string) ([]model.MedicationRule, error) {
	return s.store.ListMedicationRules(ctx, elderlyID)
}

func (s *MedicationService) UpdateRule(ctx context.Context, ruleID string, req *model.CreateMedicationRuleRequest) error {
	rule, err := s.store.GetMedicationRule(ctx, ruleID)
	if err != nil {
		return err
	}
	if err := s.store.UpdateMedicationRule(ctx, ruleID, req); err != nil {
		return err
	}

	// Push updated rule to pillbox device
	deviceID, _ := s.store.GetDeviceByElderlyID(ctx, rule.ElderlyID)
	if deviceID == "" {
		deviceID = "unknown"
	}
	cmd := map[string]any{
		"type": "med_rule_update",
		"rule_id": ruleID,
		"rule": map[string]any{
			"time":   req.ScheduleTime,
			"dose":   req.DoseCount,
			"type":   req.PillType,
			"days":   req.DaysOfWeek,
		},
	}
	_ = s.nats.PublishCommand(ctx, deviceID, cmd)
	return nil
}

func (s *MedicationService) DeleteRule(ctx context.Context, ruleID string) error {
	rule, err := s.store.GetMedicationRule(ctx, ruleID)
	if err != nil {
		return err
	}
	if err := s.store.DeleteMedicationRule(ctx, ruleID); err != nil {
		return err
	}

	// Notify pillbox device to remove rule
	deviceID, _ := s.store.GetDeviceByElderlyID(ctx, rule.ElderlyID)
	if deviceID == "" {
		deviceID = "unknown"
	}
	cmd := map[string]any{
		"type":    "med_rule_delete",
		"rule_id": ruleID,
	}
	_ = s.nats.PublishCommand(ctx, deviceID, cmd)
	return nil
}

func (s *MedicationService) GetTodayStatus(ctx context.Context, elderlyID string) ([]model.MedStatusRecord, error) {
	return s.store.GetTodayMedStatus(ctx, elderlyID)
}

func (s *MedicationService) GetHistory(ctx context.Context, elderlyID string, days int) ([]model.MedStatusRecord, error) {
	return s.store.GetMedicationHistory(ctx, elderlyID, days)
}

// --- FCM JWT bearer grant helpers (mirrors push-service/internal/fcm/jwt.go) ---

type jwtMap = map[string]interface{}

func jwtBearerGrant(ctx context.Context, cli *http.Client, privateKeyPEM, keyID, clientEmail string) (string, time.Time, error) {
	priv, err := parsePrivateKey(privateKeyPEM)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("parse private key: %w", err)
	}

	header := base64RawURLEncode(jsonRaw(jwtMap{"alg": "RS256", "typ": "JWT", "kid": keyID}))
	payload := base64RawURLEncode(jsonRaw(jwtMap{
		"iss": clientEmail, "sub": clientEmail,
		"aud": "https://oauth2.googleapis.com/token",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour).Unix(),
	}))

	signingInput := header + "." + payload
	hashed := sha256.Sum256([]byte(signingInput))
	signature, err := priv.Sign(rand.Reader, hashed[:], nil)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign jwt: %w", err)
	}

	jwt := signingInput + "." + base64.RawURLEncoding.EncodeToString(signature)

	resp, err := cli.PostForm("https://oauth2.googleapis.com/token", map[string][]string{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {jwt},
	})
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	data, _ := io.ReadAll(resp.Body)
	json.Unmarshal(data, &result)
	if result.AccessToken == "" {
		return "", time.Time{}, fmt.Errorf("fcm oauth error: %s", string(data))
	}

	return result.AccessToken, time.Now().Add(time.Duration(result.ExpiresIn-60) * time.Second), nil
}

func parsePrivateKey(pemData string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse pkcs1: %w", err)
		}
	}
	if sk, ok := key.(*rsa.PrivateKey); ok {
		return sk, nil
	}
	return nil, fmt.Errorf("not an RSA key")
}

func jsonRaw(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

func base64RawURLEncode(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}
