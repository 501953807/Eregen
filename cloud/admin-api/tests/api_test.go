// tests/api_test verifies the /api/v1/* endpoints return correct responses.
// Run: go test -v ./tests/...
package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"eregen.dev/admin-api/internal/auth"
	"eregen.dev/admin-api/internal/router"
	"eregen.dev/admin-api/internal/store"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// generateTestToken creates a valid JWT for the test environment.
func generateTestToken(t *testing.T) string {
	t.Helper()
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "test-secret-key-for-testing"
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "usr-admin",
		"role":    "admin",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tok, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}
	return tok
}

func TestAPIV1Endpoints(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing")

	db, err := store.NewSqlite(":memory:")
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}
	defer db.Close()

	seedTestData(db)

	logger, _ := zap.NewProduction()
	r := router.Setup(store.NewSqliteStore(db), logger)
	token := generateTestToken(t)

	type testCase struct {
		name       string
		method     string
		path       string
		wantStatus int
	}
	testCases := []testCase{
		{
			name:       "GET /api/v1/health",
			method:     http.MethodGet,
			path:       "/api/v1/health",
			wantStatus: http.StatusOK,
		},
		{
			name:       "GET /api/v1/admin/elderly",
			method:     http.MethodGet,
			path:       "/api/v1/admin/elderly",
			wantStatus: http.StatusOK,
		},
		{
			name:       "GET /api/v1/admin/users?role=family",
			method:     http.MethodGet,
			path:       "/api/v1/admin/users?role=family",
			wantStatus: http.StatusOK,
		},
		{
			name:       "GET /api/v1/admin/devices",
			method:     http.MethodGet,
			path:       "/api/v1/admin/devices",
			wantStatus: http.StatusOK,
		},
		{
			name:       "GET /api/v1/admin/alerts",
			method:     http.MethodGet,
			path:       "/api/v1/admin/alerts",
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.path != "/api/v1/health" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("expected status %d, got %d", tc.wantStatus, rec.Code)
				t.Log(rec.Body.String())
			}

			var resp map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &resp)

			switch tc.path {
			case "/api/v1/health":
				data, ok := resp["data"].(map[string]interface{})
				if !ok || data["status"] != "ok" {
					t.Fatalf("health response missing {data: {status: 'ok'}}, got=%v", resp)
				}
			case "/api/v1/admin/elderly":
				if data, ok := resp["data"].([]interface{}); !ok || len(data) == 0 {
					t.Fatal("elderly endpoint should return array of profiles")
				}
			case "/api/v1/admin/users":
				if _, ok := resp["data"]; !ok || resp["meta"] == nil {
					t.Fatal("users response missing data or meta")
				}
			}
		})
	}
}

func TestAPIV1Endpoints_Unauthorized(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing")

	db, err := store.NewSqlite(":memory:")
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}
	defer db.Close()

	seedTestData(db)

	logger, _ := zap.NewProduction()
	r := router.Setup(store.NewSqliteStore(db), logger)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/elderly", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLoginEndpoint(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing")

	db, err := store.NewSqlite(":memory:")
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}
	defer db.Close()

	seedTestData(db)

	logger, _ := zap.NewProduction()
	r := router.Setup(store.NewSqliteStore(db), logger)

	loginBody := map[string]string{"method": "email", "credential": "admin@eregen.com", "secret": "Admin@123"}
	bodyBytes, _ := json.Marshal(loginBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["data"] == nil {
		t.Fatal("login response missing data")
	}
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data is not an object, got %T", resp["data"])
	}
	if _, hasToken := data["token"]; !hasToken {
		t.Fatal("login response missing token")
	}
}

func TestCreateUser(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing")

	db, err := store.NewSqlite(":memory:")
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}
	defer db.Close()

	seedTestData(db)

	logger, _ := zap.NewProduction()
	r := router.Setup(store.NewSqliteStore(db), logger)
	token := generateTestToken(t)

	createBody := map[string]string{
		"name":     "测试用户",
		"phone":    "13900000001",
		"role":     "family",
		"password": "Test@12345",
	}
	bodyBytes, _ := json.Marshal(createBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing")

	db, err := store.NewSqlite(":memory:")
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}
	defer db.Close()

	seedTestData(db)

	logger, _ := zap.NewProduction()
	r := router.Setup(store.NewSqliteStore(db), logger)

	loginBody := map[string]string{"method": "email", "credential": "admin@eregen.com", "secret": "wrong-password"}
	bodyBytes, _ := json.Marshal(loginBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func seedTestData(db *sql.DB) {
	adminHash, err := auth.HashPassword("Admin@123")
	if err != nil {
		log.Fatalf("failed to hash admin password: %v", err)
	}

	users := []struct {
		id, name, email, role, phone, password_hash string
	}{
		{"usr-admin", "系统管理员", "admin@eregen.com", "admin", "", adminHash},
		{"usr-family-1", "张伟", "zhangwei@example.com", "family", "12345678900", "$2a$10$92Ub3fyY.sN1LZ2s8QyLmOZ4j3Kp5q7r8t9u0i1o2p3s4t5u6v7w8x9y0z1"},
		{"usr-family-2", "李娜", "lina@example.com", "family", "13800138000", "$2a$10$92Ub3fyY.sN1LZ2s8QyLmOZ4j3Kp5q7r8t9u0i1o2p3s4t5u6v7w8x9y0z1"},
	}
	for _, u := range users {
		_, err := db.Exec(`INSERT OR REPLACE INTO users (id, name, email, role, phone, password_hash) VALUES (?, ?, ?, ?, ?, ?)`,
			u.id, u.name, u.email, u.role, u.phone, u.password_hash)
		if err != nil {
			log.Printf("failed to insert user %s: %v", u.id, err)
		}
	}

	elders := []struct {
		id, userID, name, birthDate string
		healthTiers string
	}{
		{"eld-1", "usr-family-1", "张建国", "1950-01-01", `["基础版"]`},
		{"eld-2", "usr-family-2", "李秀英", "1948-05-05", `["防跌倒"]`},
	}
	for _, e := range elders {
		_, err := db.Exec(`INSERT OR REPLACE INTO elderly_profiles (id, user_id, name, birth_date, health_tiers) VALUES (?, ?, ?, ?, ?)`,
			e.id, e.userID, e.name, e.birthDate, e.healthTiers)
		if err != nil {
			log.Printf("failed to insert elderly %s: %v", e.id, err)
		}
	}

	dbExec := func(query string, args ...interface{}) error {
		_, err := db.Exec(query, args...)
		return err
	}
	if err := dbExec(`INSERT OR REPLACE INTO devices (id, device_id, device_type, tier, status, owner_user_id, last_seen, settings) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"dev-br-001", "dev-br-001", "bracelet", "pro", "online", "usr-family-1", "2026-07-27T10:00:00Z", `{"fw_version": "2.1.0"}`); err != nil {
		log.Printf("failed to insert device dev-br-001: %v", err)
	}
	if err := dbExec(`INSERT OR REPLACE INTO devices (id, device_id, device_type, tier, status, owner_user_id, last_seen, settings) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"dev-px-001", "dev-px-001", "pillbox", "standard", "online", "usr-family-2", "2026-07-27T09:30:00Z", `{"fw_version": "1.5.2"}`); err != nil {
		log.Printf("failed to insert device dev-px-001: %v", err)
	}
	if err := dbExec(`INSERT OR REPLACE INTO alerts (id, elderly_id, alert_type, severity, status, message, device_id) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"alert-001", "eld-1", "sos", "high", "pending", "老人按下SOS按钮", "dev-br-001"); err != nil {
		log.Printf("failed to insert alert alert-001: %v", err)
	}
}
