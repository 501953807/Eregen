package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"eregen.dev/api-server/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func newTestAuthHandler(t *testing.T) *AuthHandler {
	log := zap.NewNop()
	return NewAuthHandler(&store.Postgres{}, &store.Redis{}, nil, nil, log)
}

// ---------- Register validation tests (no DB needed) ----------

func TestRegister_PasswordMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestAuthHandler(t)

	body := map[string]any{
		"name":             "Test User",
		"password":         "Password123",
		"confirm_password": "Different456",
	}
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.Register(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestRegister_WeakPasswordShort(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestAuthHandler(t)

	body := map[string]any{
		"name":             "Test User",
		"password":         "weak",
		"confirm_password": "weak",
	}
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.Register(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestRegister_WeakPasswordNoUpper(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestAuthHandler(t)

	body := map[string]any{
		"name":             "Test User",
		"password":         "password123",
		"confirm_password": "password123",
	}
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.Register(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestRegister_MissingIdentifier(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestAuthHandler(t)

	body := map[string]any{
		"name":             "Test User",
		"password":         "Password123",
		"confirm_password": "Password123",
	}
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.Register(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ---------- SendOTP / SendCode validation tests ----------

func TestSendOTP_NoPhoneOrEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestAuthHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/send-otp", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.SendOTP(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestSendCode_Alias(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestAuthHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/send-code", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.SendCode(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ---------- PhoneLogin / WechatLogin validation tests ----------

func TestPhoneLogin_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestAuthHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/phone-login", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.PhoneLogin(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestWechatLogin_MissingCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestAuthHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/wechat/login", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.WechatLogin(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ---------- ForgotPassword validation tests ----------

func TestForgotPassword_NoPhoneOrEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestAuthHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.ForgotPassword(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ---------- Refresh / Logout validation tests ----------

func TestRefresh_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestAuthHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.Refresh(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestLogout_WithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestAuthHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.Logout(c)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (logout is lenient)", w.Code)
	}
}

func TestRevokeAllSessions_NotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestAuthHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/revoke-all-sessions", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.RevokeAllSessions(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestRegisterDevice_NotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestAuthHandler(t)

	body := map[string]any{
		"device_id":    "BR-TEST01",
		"device_type":  "bracelet",
		"tier":         "plus",
		"fingerprint":  "test-fp",
		"serial_number": "SN001",
	}
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/device/register", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.RegisterDevice(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// ---------- generateOTP correctness tests ----------

func TestGenerateOTP_Length(t *testing.T) {
	code := generateOTP()
	if len(code) != 6 {
		t.Errorf("generateOTP() length = %d, want 6", len(code))
	}
	for _, r := range code {
		if r < '0' || r > '9' {
			t.Errorf("generateOTP() contains non-digit: %c", r)
		}
	}
}

func TestGenerateOTP_Diversity(t *testing.T) {
	set := make(map[string]bool)
	for i := 0; i < 100; i++ {
		set[generateOTP()] = true
	}
	if len(set) < 50 {
		t.Errorf("generateOTP() produced only %d unique values in 100 calls, expected more", len(set))
	}
}

func TestGenerateOTP_Range(t *testing.T) {
	for i := 0; i < 100; i++ {
		code := generateOTP()
		n := int64(0)
		for _, r := range code {
			n = n*10 + int64(r-'0')
		}
		if n < 100000 || n >= 1000000 {
			t.Errorf("generateOTP() = %s (%d), want [100000, 999999]", code, n)
			return
		}
	}
}
