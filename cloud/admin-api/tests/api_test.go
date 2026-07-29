// tests/api_test verifies the /api/v1/* endpoints return correct responses.
// Run: go test -v ./tests/...
package tests

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"eregen.dev/admin-api/internal/router"
	"eregen.dev/admin-api/internal/store"

	"go.uber.org/zap"
)

func TestAPIV1Endpoints(t *testing.T) {
	// Set JWT_SECRET for router setup
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing")

	// Create in-memory SQLite DB
	db, err := store.NewSqlite("/tmp/regen-test.db")
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}
	defer db.Close()

	// Seed test data
	seedTestData(db)

	// Setup router with test db
	logger, _ := zap.NewProduction()
	r := router.Setup(db, logger, "sqlite")

	// Create test request recorder
	testCases := []struct {
		name     string
		method   string
		path     string
		wantStatus int
	}{
		{
			name:     "GET /api/v1/health",
			method:   http.MethodGet,
			path:     "/api/v1/health",
			wantStatus: http.StatusOK,
		},
		{
			name:     "GET /api/v1/elderly",
			method:   http.MethodGet,
			path:     "/api/v1/admin/elderly",
			wantStatus: http.StatusOK,
		},
		{
			name:     "GET /api/v1/users?role=family",
			method:   http.MethodGet,
			path:     "/api/v1/admin/users?role=family",
			wantStatus: http.StatusOK,
		},
		{
			name:     "GET /api/v1/devices",
			method:   http.MethodGet,
			path:     "/api/v1/admin/devices",
			wantStatus: http.StatusOK,
		},
		{
			name:     "GET /api/v1/alerts",
			method:   http.MethodGet,
			path:     "/api/v1/admin/alerts",
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			// Add Authorization header for protected endpoints (except health)
			if tc.path != "/api/v1/health" {
				req.Header.Set("Authorization", "Bearer test-jwt-token-for-testing")
			}
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("expected status %d, got %d", tc.wantStatus, rec.Code)
				t.Log(rec.Body.String())
			}

			// Additional validation for specific endpoints
			switch tc.path {
			case "/api/v1/health":
				var resp map[string]interface{}
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatal("failed to parse health response")
				}
				data, ok := resp["data"].(map[string]interface{})
				if !ok || data["status"] != "ok" {
					t.Fatalf("health response missing {data: {status: 'ok'}}, got=%v", resp)
				}
			case "/api/v1/elderly":
				var resp map[string]interface{}
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatal("failed to parse elderly response")
				}
				if data, ok := resp["data"].([]interface{}); !ok || len(data) == 0 {
					t.Fatal("elderly endpoint should return array of profiles")
				}
			case "/api/v1/users":
				var resp map[string]interface{}
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatal("failed to parse users response")
				}
				if _, ok := resp["data"]; !ok || resp["meta"] == nil {
					t.Fatal("users response missing data or meta")
				}
			}
		})
	}
}

func seedTestData(db *sql.DB) {
	// Insert test users
	users := []struct {
		id, name, email, role, phone, password_hash string
	}{
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

	// Insert test elderly profiles
	elders := []struct {
		id, userID, name, birthDate string
		healthTiers string // JSON array
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

	// Insert test devices WITH id column
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

	// Insert test alert
	if err := dbExec(`INSERT OR REPLACE INTO alerts (id, elderly_id, alert_type, severity, status, message, device_id) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"alert-001", "eld-1", "sos", "high", "pending", "老人按下SOS按钮", "dev-br-001"); err != nil {
		log.Printf("failed to insert alert alert-001: %v", err)
	}
}
