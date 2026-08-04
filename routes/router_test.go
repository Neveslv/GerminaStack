package routes

import (
	"context"
	"net/http"
	"testing"
	"time"

	"germinaStack/auth"
	"germinaStack/database"
	"germinaStack/handlers"
)

func TestNewRouterRegistersIssueTenAPIRoutes(t *testing.T) {
	t.Parallel()

	authHandler := handlers.NewAuthHandler(routerAuthServiceFake{}, "jwt-test-secret", true, 24*time.Hour, 2*time.Second)
	userHandler := handlers.NewUserHandler(routerUserRepositoryFake{}, 2*time.Second)
	router := NewRouter(authHandler, userHandler)

	got := make(map[string]bool)
	for _, route := range router.Routes() {
		got[route.Method+" "+route.Path] = true
	}
	for _, want := range []string{
		http.MethodPost + " /api/users",
		http.MethodPost + " /api/login",
		http.MethodPost + " /api/login/2fa",
		http.MethodPost + " /api/logout",
	} {
		if !got[want] {
			t.Fatalf("route %q is not registered; got %#v", want, got)
		}
	}
	if len(got) != 4 {
		t.Fatalf("registered routes = %#v, want exactly four issue #10 routes", got)
	}
}

type routerAuthServiceFake struct{}

func (routerAuthServiceFake) StartLogin(context.Context, string, string) (string, error) {
	return "challenge-id", nil
}

func (routerAuthServiceFake) CompleteLogin(context.Context, string, string) (auth.Principal, error) {
	return auth.Principal{ID: 42}, nil
}

type routerUserRepositoryFake struct{}

func (routerUserRepositoryFake) CreateUser(context.Context, database.UserRegistration) (database.User, error) {
	return database.User{}, nil
}
