package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"germinaStack/auth"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAPIAuthMiddlewareSetsTypedPrincipalContext(t *testing.T) {
	t.Parallel()
	token, err := auth.GenerateTokenWithTimes("42", true, "jwt-secret", time.Now().Add(-time.Minute), time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("GenerateTokenWithTimes() error = %v", err)
	}

	called := false
	recorder := performAPIMiddlewareRequest(APIAuthMiddleware("jwt-secret"), token, func(c *gin.Context) {
		called = true
		userID, userOK := c.Get(auth.ContextUserID)
		isAdmin, adminOK := c.Get(auth.ContextIsAdmin)
		if !userOK || !adminOK || userID != int64(42) || isAdmin != true {
			t.Fatalf("context principal = %#v/%#v", userID, isAdmin)
		}
		c.Status(http.StatusNoContent)
	})
	if !called || recorder.Code != http.StatusNoContent {
		t.Fatalf("called/status = %v/%d", called, recorder.Code)
	}
}

func TestAPIAuthMiddlewareRejectsInvalidTokensAsJSON(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	expired, _ := auth.GenerateTokenWithTimes("42", false, "jwt-secret", now.Add(-2*time.Hour), now.Add(-time.Hour))
	nonnumeric, _ := auth.GenerateTokenWithTimes("abc", false, "jwt-secret", now, now.Add(time.Hour))
	zero, _ := auth.GenerateTokenWithTimes("0", false, "jwt-secret", now, now.Add(time.Hour))
	wrongSecret, _ := auth.GenerateTokenWithTimes("42", false, "other-secret", now, now.Add(time.Hour))
	tests := []struct {
		name  string
		token string
	}{
		{name: "missing"},
		{name: "malformed", token: "not-a-token"},
		{name: "expired", token: expired},
		{name: "nonnumeric subject", token: nonnumeric},
		{name: "nonpositive subject", token: zero},
		{name: "wrong injected secret", token: wrongSecret},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			called := false
			recorder := performAPIMiddlewareRequest(APIAuthMiddleware("jwt-secret"), tt.token, func(c *gin.Context) { called = true })
			if recorder.Code != http.StatusUnauthorized || called {
				t.Fatalf("status/called = %d/%v, want 401/false; body=%s", recorder.Code, called, recorder.Body.String())
			}
			if recorder.Header().Get("Content-Type") != "application/json; charset=utf-8" {
				t.Fatalf("content type = %q", recorder.Header().Get("Content-Type"))
			}
		})
	}
}

func TestAPIAdminAuthMiddlewareForbidsNonAdminAndAllowsAdmin(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	for _, isAdmin := range []bool{false, true} {
		token, _ := auth.GenerateTokenWithTimes("42", isAdmin, "jwt-secret", now, now.Add(time.Hour))
		called := false
		recorder := performAPIMiddlewareRequest(APIAdminAuthMiddleware("jwt-secret"), token, func(c *gin.Context) {
			called = true
			c.Status(http.StatusNoContent)
		})
		if isAdmin {
			if recorder.Code != http.StatusNoContent || !called {
				t.Fatalf("admin status/called = %d/%v", recorder.Code, called)
			}
		} else if recorder.Code != http.StatusForbidden || called {
			t.Fatalf("non-admin status/called = %d/%v", recorder.Code, called)
		}
	}
}

func performAPIMiddlewareRequest(middleware gin.HandlerFunc, token string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	router := gin.New()
	router.GET("/api/protected", middleware, handler)
	request := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	if token != "" {
		request.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token})
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}
