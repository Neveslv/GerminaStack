package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"germinaStack/handlers"
	"germinaStack/middleware"
	"germinaStack/model"

	"github.com/gin-gonic/gin"
)

func TestRegisterAccountRoutesProtectsAllAccountEndpoints(t *testing.T) {
	t.Parallel()
	router := gin.New()
	RegisterAccountRoutes(router.Group("/api"), handlers.NewAccountHandler(&routeAccountRepositoryFake{}, time.Second), middleware.APIAuthMiddleware("jwt-secret"))
	got := make(map[string]bool)
	for _, route := range router.Routes() {
		got[route.Method+" "+route.Path] = true
	}
	for _, want := range []string{"GET /api/users/:username", "GET /api/me", "PATCH /api/me", "GET /api/me/preferences", "PATCH /api/me/preferences"} {
		if !got[want] {
			t.Fatalf("route %q not registered; got %#v", want, got)
		}
	}
	request := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("missing jwt status = %d", recorder.Code)
	}
}

type routeAccountRepositoryFake struct{}

func (routeAccountRepositoryFake) GetProfile(context.Context, int64) (model.User, error) {
	return model.User{}, nil
}
func (routeAccountRepositoryFake) GetPublicProfile(context.Context, string) (model.User, error) {
	return model.User{}, nil
}
func (routeAccountRepositoryFake) UpdateProfile(context.Context, int64, model.User) (model.User, error) {
	return model.User{}, nil
}
func (routeAccountRepositoryFake) GetPreferences(context.Context, int64) (model.Preference, error) {
	return model.Preference{}, nil
}
func (routeAccountRepositoryFake) UpsertPreferences(context.Context, int64, model.Preference) (model.Preference, error) {
	return model.Preference{}, nil
}
