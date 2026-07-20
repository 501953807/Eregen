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

func newTestDeviceHandler(t *testing.T) *DeviceHandler {
	log := zap.NewNop()
	return NewDeviceHandler(&store.Postgres{}, &store.Redis{}, nil, log)
}

// ---------- HandleTelemetry input validation ----------

func TestHandleTelemetry_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestDeviceHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/telemetry", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.HandleTelemetry(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// Note: Tests below require Redis/DB mocks; skipped in unit tests
// func TestHandleTelemetry_HealthMissingElderlyID
// func TestHandleTelemetry_LocationMissingCoords

// ---------- HandleHeartbeat input validation ----------

func TestHandleHeartbeat_MissingDeviceID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestDeviceHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/heartbeat", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.HandleHeartbeat(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ---------- HandleLocation input validation ----------

func TestHandleLocation_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestDeviceHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/location", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.HandleLocation(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandleLocation_ZeroCoords(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestDeviceHandler(t)

	body := map[string]any{
		"device_id":  "BR-TEST01",
		"elderly_id": "elderly-1",
		"lat":        0,
		"lon":        0,
	}
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/location", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.HandleLocation(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ---------- Bind device validation ----------

func TestBind_MissingDeviceID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestDeviceHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/bind", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.Bind(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestBind_InvalidDeviceID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestDeviceHandler(t)

	body := map[string]any{
		"device_id": "INVALID-ID-FORMAT",
	}
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/bind", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.Bind(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestBind_InvalidPrefix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestDeviceHandler(t)

	body := map[string]any{
		"device_id": "XX-12345",
	}
	reqBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/bind", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.Bind(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}
