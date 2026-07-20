package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func newTestJWTAuth() *JWTAuth {
	log, _ := zap.NewDevelopment()
	return NewJWTAuth("test-secret-key-min-32-bytes-long", 15*time.Minute, 7*24*time.Hour, log, &store.Postgres{})
}

func TestGenerateAccessToken(t *testing.T) {
	auth := newTestJWTAuth()
	tokenStr, err := auth.GenerateAccessToken("user-123", model.RoleFamily)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := auth.ParseToken(tokenStr)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}
	if claims["user_id"] != "user-123" {
		t.Errorf("user_id = %v, want user-123", claims["user_id"])
	}
	if claims["role"] != string(model.RoleFamily) {
		t.Errorf("role = %v, want family", claims["role"])
	}
	if _, ok := claims["jti"].(string); !ok || claims["jti"] == "" {
		t.Error("missing jti claim")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	auth := newTestJWTAuth()
	tokenStr, err := auth.GenerateRefreshToken("user-456")
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	claims, err := auth.ParseToken(tokenStr)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}
	if claims["user_id"] != "user-456" {
		t.Errorf("user_id = %v, want user-456", claims["user_id"])
	}
	if claims["role"] != "" {
		t.Errorf("refresh token should have empty role, got %v", claims["role"])
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth := newTestJWTAuth()
	tokenStr, _ := auth.GenerateAccessToken("user-789", model.RoleElderly)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	auth.AuthMiddleware()(c)

	uid, exists := c.Get(string(ContextUserID))
	if !exists || uid.(string) != "user-789" {
		t.Errorf("user_id not set correctly: %v", uid)
	}
	role, exists := c.Get(string(ContextUserRole))
	if !exists || role.(string) != string(model.RoleElderly) {
		t.Errorf("role not set correctly: %v", role)
	}
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth := newTestJWTAuth()

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	auth.AuthMiddleware()(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth := newTestJWTAuth()

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Token invalid-token-here")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	auth.AuthMiddleware()(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth := NewJWTAuth("test-secret-key-min-32-bytes-long", -1*time.Hour, 7*24*time.Hour, zap.NewNop(), &store.Postgres{})
	tokenStr, _ := auth.GenerateAccessToken("user-expired", model.RoleFamily)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	auth.AuthMiddleware()(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func mustToken(auth *JWTAuth, userID string, role model.Role) string {
	token, _ := auth.GenerateAccessToken(userID, role)
	return token
}

func TestRequireRole_AccessGranted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth := newTestJWTAuth()

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+mustToken(auth, "user-1", model.RoleElderly))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	auth.AuthMiddleware()(c)
	auth.RequireRole(model.RoleElderly)(c)

	uid, exists := c.Get(string(ContextUserID))
	if !exists || uid.(string) != "user-1" {
		t.Errorf("user_id not preserved through RequireRole: %v", uid)
	}
}

func TestRequireRole_AccessDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth := newTestJWTAuth()

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+mustToken(auth, "user-2", model.RoleFamily))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	auth.AuthMiddleware()(c)
	auth.RequireRole(model.RoleElderly)(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}
