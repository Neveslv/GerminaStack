package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"germinaStack/auth"
	"germinaStack/handlers"
	"germinaStack/middleware"
	"germinaStack/model"

	"github.com/gin-gonic/gin"
)

func TestRegisterCatalogRoutesRegistersPublicAndAdminSurface(t *testing.T) {
	t.Parallel()
	router := gin.New()
	api := router.Group("/api")
	RegisterCatalogRoutes(api, handlers.NewCatalogHandler(&routeCatalogRepositoryFake{}, time.Second), middleware.APIAdminAuthMiddleware("jwt-secret"))

	got := make(map[string]bool)
	for _, route := range router.Routes() {
		got[route.Method+" "+route.Path] = true
	}
	for _, want := range []string{
		"GET /api/years", "GET /api/subjects",
		"POST /api/admin/years", "PATCH /api/admin/years/:id", "DELETE /api/admin/years/:id",
		"POST /api/admin/subjects", "PATCH /api/admin/subjects/:id", "DELETE /api/admin/subjects/:id",
	} {
		if !got[want] {
			t.Fatalf("route %q not registered; got %#v", want, got)
		}
	}
	if len(got) != 8 {
		t.Fatalf("registered routes = %#v, want exactly eight", got)
	}
}

func TestRegisterCatalogRoutesLeavesReadsPublicAndProtectsAllWrites(t *testing.T) {
	t.Parallel()
	repository := &routeCatalogRepositoryFake{year: model.Year{ID: 3, Year: "2026"}}
	router := gin.New()
	RegisterCatalogRoutes(router.Group("/api"), handlers.NewCatalogHandler(repository, time.Second), middleware.APIAdminAuthMiddleware("jwt-secret"))

	publicRequest := httptest.NewRequest(http.MethodGet, "/api/years", nil)
	publicRecorder := httptest.NewRecorder()
	router.ServeHTTP(publicRecorder, publicRequest)
	if publicRecorder.Code != http.StatusOK {
		t.Fatalf("public status = %d; body=%s", publicRecorder.Code, publicRecorder.Body.String())
	}

	for _, tt := range []struct {
		name    string
		isAdmin *bool
		status  int
		calls   int
	}{
		{name: "missing", status: http.StatusUnauthorized},
		{name: "non-admin", isAdmin: boolPointer(false), status: http.StatusForbidden},
		{name: "admin", isAdmin: boolPointer(true), status: http.StatusCreated, calls: 1},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			before := repository.createYearCalls
			request := httptest.NewRequest(http.MethodPost, "/api/admin/years", strings.NewReader(`{"year":"2026"}`))
			request.Header.Set("Content-Type", "application/json")
			if tt.isAdmin != nil {
				now := time.Now().UTC()
				token, err := auth.GenerateTokenWithTimes("42", *tt.isAdmin, "jwt-secret", now.Add(-time.Minute), now.Add(time.Hour))
				if err != nil {
					t.Fatalf("GenerateTokenWithTimes() error = %v", err)
				}
				request.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token})
			}
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != tt.status || repository.createYearCalls-before != tt.calls {
				t.Fatalf("status/new calls = %d/%d, want %d/%d; body=%s", recorder.Code, repository.createYearCalls-before, tt.status, tt.calls, recorder.Body.String())
			}
		})
	}
}

type routeCatalogRepositoryFake struct {
	year            model.Year
	createYearCalls int
}

func (f routeCatalogRepositoryFake) ListYears(context.Context) ([]model.Year, error) {
	return []model.Year{}, nil
}
func (f *routeCatalogRepositoryFake) CreateYear(context.Context, string) (model.Year, error) {
	f.createYearCalls++
	return f.year, nil
}
func (f routeCatalogRepositoryFake) UpdateYear(context.Context, int64, string) (model.Year, error) {
	return f.year, nil
}
func (f routeCatalogRepositoryFake) DeleteYear(context.Context, int64) error { return nil }
func (f routeCatalogRepositoryFake) ListSubjects(context.Context, *int64) ([]model.Subject, error) {
	return []model.Subject{}, nil
}
func (f routeCatalogRepositoryFake) CreateSubject(context.Context, string, *int64) (model.Subject, error) {
	return model.Subject{}, nil
}
func (f routeCatalogRepositoryFake) UpdateSubject(context.Context, int64, string, *int64) (model.Subject, error) {
	return model.Subject{}, nil
}
func (f routeCatalogRepositoryFake) DeleteSubject(context.Context, int64) error { return nil }

func boolPointer(value bool) *bool { return &value }
