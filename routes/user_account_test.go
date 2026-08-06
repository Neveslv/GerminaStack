package routes

import (
	"context"
	"germinaStack/database"
	"germinaStack/handlers"
	"germinaStack/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
	"testing"
	"time"
)

func TestRegisterUserAccountRoutesRegistersOnlyAuthenticatedAccountEndpoints(t *testing.T) {
	router := gin.New()
	api := router.Group("/api")
	RegisterUserAccountRoutes(api, handlers.NewUserHandler(accountRouteRepository{}, time.Second), middleware.APIAuthMiddleware("secret"))
	routes := map[string]bool{}
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, want := range []string{"GET /api/users/me", "PUT /api/users/me/preferences", "GET /api/preferences", "PUT /api/preferences"} {
		if !routes[want] {
			t.Fatalf("missing %s: %#v", want, routes)
		}
	}
	if len(routes) != 4 {
		t.Fatalf("routes=%#v", routes)
	}
	for _, tt := range []struct{ method, path string }{{http.MethodGet, "/api/users/me"}, {http.MethodPut, "/api/users/me/preferences"}, {http.MethodGet, "/api/preferences"}, {http.MethodPut, "/api/preferences"}} {
		request, _ := http.NewRequest(tt.method, "http://example"+tt.path, nil)
		response := newTestResponseWriter()
		router.ServeHTTP(response, request)
		if response.status != http.StatusUnauthorized {
			t.Fatalf("%s %s status=%d", tt.method, tt.path, response.status)
		}
	}
}

type accountRouteRepository struct{}

func (accountRouteRepository) CreateUser(context.Context, database.UserRegistration) (database.User, error) {
	return database.User{}, nil
}
func (accountRouteRepository) FindUserAccount(context.Context, int64) (database.UserAccount, error) {
	return database.UserAccount{}, nil
}
func (accountRouteRepository) UpdateAccessibilityPreferences(context.Context, int64, database.AccessibilityPreferences) (database.AccessibilityPreferences, error) {
	return database.AccessibilityPreferences{}, nil
}

type testResponseWriter struct {
	header http.Header
	status int
}

func newTestResponseWriter() *testResponseWriter {
	return &testResponseWriter{header: make(http.Header)}
}
func (w *testResponseWriter) Header() http.Header { return w.header }
func (w *testResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	return len(b), nil
}
func (w *testResponseWriter) WriteHeader(status int) { w.status = status }
