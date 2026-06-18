package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"taskflow-pro/backend/internal/config"
	"taskflow-pro/backend/internal/service"

	"github.com/gin-gonic/gin"
)

func TestAuthMiddlewareAcceptsValidJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth := service.NewAuthService(nil, nil, config.JWTConfig{
		Secret: "jwt-test-secret",
		TTL:    time.Hour,
	})
	token, err := auth.IssueToken(42)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	r := gin.New()
	r.GET("/protected", Auth(auth), func(c *gin.Context) {
		if CurrentUserID(c) != 42 {
			t.Fatalf("expected user id 42, got %d", CurrentUserID(c))
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuthMiddlewareRejectsMissingJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth := service.NewAuthService(nil, nil, config.JWTConfig{
		Secret: "jwt-test-secret",
		TTL:    time.Hour,
	})

	r := gin.New()
	r.GET("/protected", Auth(auth), func(c *gin.Context) {
		t.Fatal("protected handler should not run")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
