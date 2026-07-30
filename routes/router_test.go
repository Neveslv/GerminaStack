package routes

import (
	"context"
	"net/http"
	"testing"
	"time"

	"germinaStack/handlers"
)

func TestNewRouterRegistersTwoStepLoginRoutes(t *testing.T) {
	t.Parallel()

	handler := handlers.NewAuthHandler(routerAuthServiceFake{}, "jwt-test-secret", true, 24*time.Hour)
	router := NewRouter(handler)

	got := make(map[string]bool)
	for _, route := range router.Routes() {
		got[route.Method+" "+route.Path] = true
	}
	for _, want := range []string{
		http.MethodPost + " /api/login",
		http.MethodPost + " /api/login/2fa",
	} {
		if !got[want] {
			t.Fatalf("route %q is not registered; got %#v", want, got)
		}
	}
}

type routerAuthServiceFake struct{}

func (routerAuthServiceFake) StartLogin(context.Context, string, string) (string, error) {
	return "challenge-id", nil
}

func (routerAuthServiceFake) CompleteLogin(context.Context, string, string) (int64, error) {
	return 42, nil
}
