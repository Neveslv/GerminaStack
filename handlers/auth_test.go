package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"germinaStack/auth"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestLoginReturnsAcceptedChallenge(t *testing.T) {
	t.Parallel()

	service := &authServiceFake{startID: "challenge-id"}
	handler := NewAuthHandler(service, "jwt-test-secret", true, 24*time.Hour)
	recorder := performRequest(handler.Login, http.MethodPost, "/api/login", `{"username":"ana","password":"correct-password"}`, "application/json")

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202; body=%s", recorder.Code, recorder.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("response JSON: %v", err)
	}
	if len(body) != 1 || body["challenge_id"] != "challenge-id" {
		t.Fatalf("body = %#v", body)
	}
	if service.startUsername != "ana" || service.startPassword != "correct-password" {
		t.Fatalf("service inputs = %q/%q", service.startUsername, service.startPassword)
	}
}

func TestLoginReturnsIdenticalUnauthorizedResponseForInvalidCredentials(t *testing.T) {
	t.Parallel()

	makeResponse := func() *httptest.ResponseRecorder {
		service := &authServiceFake{startErr: auth.ErrInvalidCredentials}
		handler := NewAuthHandler(service, "jwt-test-secret", true, 24*time.Hour)
		return performRequest(handler.Login, http.MethodPost, "/api/login", `{"username":"ana","password":"wrong"}`, "application/json")
	}
	first := makeResponse()
	second := makeResponse()
	if first.Code != http.StatusUnauthorized || second.Code != http.StatusUnauthorized {
		t.Fatalf("statuses = %d/%d, want 401/401", first.Code, second.Code)
	}
	if first.Body.String() != second.Body.String() || first.Body.String() != `{"error":"credenciais inválidas"}` {
		t.Fatalf("responses differ or expose detail: %q / %q", first.Body.String(), second.Body.String())
	}
}

func TestLoginRejectsMalformedAndOversizedInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		body        string
		contentType string
	}{
		{name: "malformed JSON", body: `{`, contentType: "application/json"},
		{name: "unknown field", body: `{"username":"ana","password":"secret","extra":true}`, contentType: "application/json"},
		{name: "missing username", body: `{"username":"","password":"secret"}`, contentType: "application/json"},
		{name: "long username", body: `{"username":"` + strings.Repeat("a", 101) + `","password":"secret"}`, contentType: "application/json"},
		{name: "long password", body: `{"username":"ana","password":"` + strings.Repeat("a", 1025) + `"}`, contentType: "application/json"},
		{name: "wrong content type", body: `{"username":"ana","password":"secret"}`, contentType: "text/plain"},
		{name: "oversized body", body: `{"username":"ana","password":"` + strings.Repeat("a", 5000) + `"}`, contentType: "application/json"},
		{name: "trailing JSON", body: `{"username":"ana","password":"secret"} {}`, contentType: "application/json"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			service := &authServiceFake{}
			handler := NewAuthHandler(service, "jwt-test-secret", true, 24*time.Hour)
			recorder := performRequest(handler.Login, http.MethodPost, "/api/login", tt.body, tt.contentType)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body=%s", recorder.Code, recorder.Body.String())
			}
			if service.startCalls != 0 {
				t.Fatalf("service calls = %d, want 0", service.startCalls)
			}
		})
	}
}

func TestLoginMapsDependencyFailureToUnavailable(t *testing.T) {
	t.Parallel()

	handler := NewAuthHandler(&authServiceFake{startErr: auth.ErrUnavailable}, "jwt-test-secret", true, 24*time.Hour)
	recorder := performRequest(handler.Login, http.MethodPost, "/api/login", `{"username":"ana","password":"secret"}`, "application/json")
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", recorder.Code)
	}
	if strings.Contains(recorder.Body.String(), "SMTP") || strings.Contains(recorder.Body.String(), "database") {
		t.Fatalf("response leaked internal detail: %s", recorder.Body.String())
	}
}

func TestCompleteLoginSetsJWTInExplicitSecureCookie(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 30, 14, 30, 0, 0, time.UTC)
	service := &authServiceFake{completeUserID: 42}
	handler := NewAuthHandler(service, "jwt-test-secret", true, 24*time.Hour)
	handler.now = func() time.Time { return now }
	recorder := performRequest(handler.CompleteLogin, http.MethodPost, "/api/login/2fa", `{"challenge_id":"challenge-id","code":"123456"}`, "application/json")

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body=%s", recorder.Code, recorder.Body.String())
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("response body = %q, want empty", recorder.Body.String())
	}
	response := recorder.Result()
	cookies := response.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != auth.CookieName || !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteLaxMode || cookie.Path != "/" {
		t.Fatalf("cookie flags = %#v", cookie)
	}
	if cookie.MaxAge != 86400 || !cookie.Expires.Equal(now.Add(24*time.Hour)) {
		t.Fatalf("cookie lifetime = MaxAge %d Expires %v", cookie.MaxAge, cookie.Expires)
	}
	claims, err := auth.ParseToken(cookie.Value, "jwt-test-secret")
	if err != nil {
		t.Fatalf("ParseToken(cookie) error = %v", err)
	}
	if claims.Subject != "42" || claims.IsAdmin {
		t.Fatalf("claims = %#v", claims)
	}
	if claims.IssuedAt == nil || !claims.IssuedAt.Time.Equal(now) || claims.ExpiresAt == nil || !claims.ExpiresAt.Time.Equal(now.Add(24*time.Hour)) {
		t.Fatalf("temporal claims = iat %v exp %v", claims.IssuedAt, claims.ExpiresAt)
	}
}

func TestCompleteLoginMapsEveryState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "malformed", err: auth.ErrMalformedCode, status: http.StatusBadRequest},
		{name: "incorrect", err: auth.ErrInvalidCode, status: http.StatusUnauthorized},
		{name: "expired", err: auth.ErrChallengeExpired, status: http.StatusGone},
		{name: "used", err: auth.ErrChallengeUsed, status: http.StatusGone},
		{name: "limit", err: auth.ErrTooManyAttempts, status: http.StatusTooManyRequests},
		{name: "unavailable", err: auth.ErrUnavailable, status: http.StatusServiceUnavailable},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handler := NewAuthHandler(&authServiceFake{completeErr: tt.err}, "jwt-test-secret", false, 24*time.Hour)
			recorder := performRequest(handler.CompleteLogin, http.MethodPost, "/api/login/2fa", `{"challenge_id":"challenge-id","code":"123456"}`, "application/json")
			if recorder.Code != tt.status {
				t.Fatalf("status = %d, want %d; body=%s", recorder.Code, tt.status, recorder.Body.String())
			}
		})
	}
}

func TestCompleteLoginRejectsMalformedInputBeforeService(t *testing.T) {
	t.Parallel()

	tests := []string{
		`{"challenge_id":"","code":"123456"}`,
		`{"challenge_id":"challenge-id","code":"12345"}`,
		`{"challenge_id":"challenge-id","code":"12a456"}`,
		`{"challenge_id":"` + strings.Repeat("a", 129) + `","code":"123456"}`,
	}
	for _, body := range tests {
		service := &authServiceFake{}
		handler := NewAuthHandler(service, "jwt-test-secret", true, 24*time.Hour)
		recorder := performRequest(handler.CompleteLogin, http.MethodPost, "/api/login/2fa", body, "application/json")
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400; body=%s", recorder.Code, recorder.Body.String())
		}
		if service.completeCalls != 0 {
			t.Fatalf("service calls = %d, want 0", service.completeCalls)
		}
	}
}

func performRequest(handler gin.HandlerFunc, method, path, body, contentType string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", contentType)
	context.Request = request
	handler(context)
	return recorder
}

type authServiceFake struct {
	startID        string
	startErr       error
	startCalls     int
	startUsername  string
	startPassword  string
	completeUserID int64
	completeErr    error
	completeCalls  int
	completeID     string
	completeCode   string
}

func (f *authServiceFake) StartLogin(_ context.Context, username, password string) (string, error) {
	f.startCalls++
	f.startUsername = username
	f.startPassword = password
	return f.startID, f.startErr
}

func (f *authServiceFake) CompleteLogin(_ context.Context, challengeID, code string) (int64, error) {
	f.completeCalls++
	f.completeID = challengeID
	f.completeCode = code
	return f.completeUserID, f.completeErr
}
