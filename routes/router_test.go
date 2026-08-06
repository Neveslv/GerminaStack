package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
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
	catalogHandler := handlers.NewCatalogHandler(&routeCatalogRepositoryFake{}, 2*time.Second)
	discussionHandler := handlers.NewDiscussionHandler(&routeDiscussionRepositoryFake{}, 2*time.Second)
	accountHandler := handlers.NewAccountHandler(&routeAccountRepositoryFake{}, 2*time.Second)
	router := NewRouter(authHandler, userHandler, catalogHandler, discussionHandler, accountHandler, "jwt-test-secret")

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
}

func TestNewRouterServesFrontend(t *testing.T) {
	t.Parallel()

	router := NewRouter(
		handlers.NewAuthHandler(routerAuthServiceFake{}, "jwt-test-secret", true, 24*time.Hour, 2*time.Second),
		handlers.NewUserHandler(routerUserRepositoryFake{}, 2*time.Second),
		handlers.NewCatalogHandler(&routeCatalogRepositoryFake{}, 2*time.Second),
		handlers.NewDiscussionHandler(&routeDiscussionRepositoryFake{}, 2*time.Second),
		handlers.NewAccountHandler(&routeAccountRepositoryFake{}, 2*time.Second),
		"jwt-test-secret",
	)

	got := make(map[string]bool)
	for _, route := range router.Routes() {
		got[route.Method+" "+route.Path] = true
	}
	for _, want := range []string{"GET /", "GET /login", "GET /static/*filepath"} {
		if !got[want] {
			t.Fatalf("route %q is not registered", want)
		}
	}
}

func TestNewRouterAllowsAuthenticatedUsersToListSubjects(t *testing.T) {
	t.Parallel()

	router := NewRouter(
		handlers.NewAuthHandler(routerAuthServiceFake{}, "jwt-test-secret", true, 24*time.Hour, 2*time.Second),
		handlers.NewUserHandler(routerUserRepositoryFake{}, 2*time.Second),
		handlers.NewCatalogHandler(&routeCatalogRepositoryFake{}, 2*time.Second),
		handlers.NewDiscussionHandler(&routeDiscussionRepositoryFake{}, 2*time.Second),
		handlers.NewAccountHandler(&routeAccountRepositoryFake{}, 2*time.Second),
		"jwt-test-secret",
	)

	request := httptest.NewRequest(http.MethodGet, "/api/subjects", nil)
	request.AddCookie(&http.Cookie{Name: auth.CookieName, Value: routerToken(t, false)})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET /api/subjects status = %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
}

func routerToken(t *testing.T, admin bool) string {
	t.Helper()
	token, err := auth.GenerateTokenWithTimes("42", admin, "jwt-test-secret", time.Now().Add(-time.Minute), time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("GenerateTokenWithTimes() error = %v", err)
	}
	return token
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
