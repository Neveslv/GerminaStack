package routes

import (
	"testing"
	"time"

	"germinaStack/handlers"
	"germinaStack/middleware"
)

func TestNewRouterComposesCatalogAndMessagesOnce(t *testing.T) {
	authHandler := handlers.NewAuthHandler(routerAuthServiceFake{}, "jwt-test-secret", true, 24*time.Hour, 2*time.Second)
	userHandler := handlers.NewUserHandler(routerUserRepositoryFake{}, 2*time.Second)
	dependencies := RouterDependencies{
		Catalog:       handlers.NewCatalogHandler(&routeCatalogRepositoryFake{}, time.Second),
		Messages:      handlers.NewMessageHandler(routeMessageRepositoryFake{}, time.Second),
		Authenticated: middleware.APIAuthMiddleware("jwt-test-secret"),
	}
	router := NewRouter(authHandler, userHandler, dependencies)

	got := make(map[string]int)
	for _, route := range router.Routes() {
		got[route.Method+" "+route.Path]++
	}
	for _, route := range []string{
		"GET /api/years", "GET /api/subjects", "GET /api/posts", "POST /api/posts",
		"GET /api/posts/:id", "GET /api/posts/:id/comments", "POST /api/posts/:id/comments",
		"GET /api/comments/:id/replies", "POST /api/comments/:id/replies",
	} {
		if got[route] != 1 {
			t.Fatalf("route %q count = %d; routes=%#v", route, got[route], got)
		}
	}
	for _, route := range []string{"POST /api/admin/years", "PATCH /api/posts/:id", "DELETE /api/posts/:id", "GET /api/notifications/:id"} {
		if got[route] != 0 {
			t.Fatalf("forbidden route %q count = %d", route, got[route])
		}
	}
}
