package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"eregen.dev/admin-api/internal/auth"
)

const testSecret = "test-jwt-secret"

func newTestRouter(jwt *AdminJWT, extraMws ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(jwt.AuthMiddleware())
	for _, m := range extraMws {
		r.Use(m)
	}
	return r
}

func TestChainPermissionsMapping(t *testing.T) {
	expected := map[string][]string{
		"super_admin":     {"self", "hospital", "community", "regulatory"},
		"operator":        {"self", "regulatory"},
		"hospital_doc":    {"hospital"},
		"nurse":           {"hospital"},
		"community_staff": {"community"},
		"regulator":       {"hospital", "community", "regulatory"},
	}
	for role, chains := range expected {
		got, ok := ChainPermissions[role]
		if !ok {
			t.Fatalf("missing role %q in ChainPermissions", role)
		}
		if len(got) != len(chains) {
			t.Errorf("role %q: expected %d chains, got %d", role, len(chains), len(got))
		}
	}
}

func TestRequireChain_AuthenticatedAllowed(t *testing.T) {
	log := zap.NewNop()
	jwt := NewAdminJWT(testSecret, 24, log)

	// super_admin can access any chain
	token, _ := auth.GenerateToken("user-1", "super_admin", testSecret)
	r := newTestRouter(jwt, jwt.RequireChain("hospital"))
	r.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("super_admin accessing hospital: expected 200, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestRequireChain_AuthenticatedDenied(t *testing.T) {
	log := zap.NewNop()
	jwt := NewAdminJWT(testSecret, 24, log)

	// community_staff CANNOT access hospital chain
	token, _ := auth.GenerateToken("user-2", "community_staff", testSecret)
	r := newTestRouter(jwt, jwt.RequireChain("hospital"))
	r.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("community_staff accessing hospital: expected 403, got %d (%s)", w.Code, w.Body.String())
	}

	// nurse CANNOT access self chain
	token2, _ := auth.GenerateToken("user-3", "nurse", testSecret)
	r2 := newTestRouter(jwt, jwt.RequireChain("self"))
	r2.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("Authorization", "Bearer "+token2)
	r2.ServeHTTP(w2, req2)
	if w2.Code != http.StatusForbidden {
		t.Errorf("nurse accessing self: expected 403, got %d (%s)", w2.Code, w2.Body.String())
	}
}

func TestRequireChain_Unauthenticated(t *testing.T) {
	log := zap.NewNop()
	jwt := NewAdminJWT(testSecret, 24, log)
	r := newTestRouter(jwt, jwt.RequireChain("self"))
	r.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for invalid token, got %d", w.Code)
	}

	// No auth header at all
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for no auth header, got %d", w2.Code)
	}
}

func TestRequireChain_UnknownRole(t *testing.T) {
	log := zap.NewNop()
	jwt := NewAdminJWT(testSecret, 24, log)
	token, _ := auth.GenerateToken("user-4", "unknown_role", testSecret)
	r := newTestRouter(jwt, jwt.RequireChain("self"))
	r.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for unknown role, got %d", w.Code)
	}
}
