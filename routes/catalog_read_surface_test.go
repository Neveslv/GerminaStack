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

func TestCatalogReadSurfaceRegistersOnlyReadRoutes(t *testing.T) {
	router := gin.New()
	RegisterCatalogRoutes(router.Group("/api"), handlers.NewCatalogHandler(&readSurfaceCatalogRepositoryFake{}, time.Second), middleware.APIAuthMiddleware("jwt-secret"))

	got := make(map[string]bool)
	for _, route := range router.Routes() {
		got[route.Method+" "+route.Path] = true
	}
	for _, want := range []string{"GET /api/years", "GET /api/subjects"} {
		if !got[want] {
			t.Fatalf("route %q not registered; got %#v", want, got)
		}
	}
	if len(got) != 2 {
		t.Fatalf("registered routes = %#v, want exactly two read routes", got)
	}
}

func TestCatalogReadSurfaceMakesYearsPublicAndSubjectsAuthenticated(t *testing.T) {
	repository := &readSurfaceCatalogRepositoryFake{subjects: []model.Subject{{ID: 9, Subject: "Geral"}}}
	router := gin.New()
	RegisterCatalogRoutes(router.Group("/api"), handlers.NewCatalogHandler(repository, time.Second), middleware.APIAuthMiddleware("jwt-secret"))

	publicYears := httptest.NewRecorder()
	router.ServeHTTP(publicYears, httptest.NewRequest(http.MethodGet, "/api/years", nil))
	if publicYears.Code != http.StatusOK {
		t.Fatalf("years status = %d; body=%s", publicYears.Code, publicYears.Body.String())
	}

	unauthenticated := httptest.NewRecorder()
	router.ServeHTTP(unauthenticated, httptest.NewRequest(http.MethodGet, "/api/subjects?year_id=999", nil))
	if unauthenticated.Code != http.StatusUnauthorized || repository.listSubjectsCalls != 0 {
		t.Fatalf("unauthenticated subjects status/calls = %d/%d", unauthenticated.Code, repository.listSubjectsCalls)
	}

	now := time.Now().UTC()
	token, err := auth.GenerateTokenWithTimes("42", false, "jwt-secret", now.Add(-time.Minute), now.Add(time.Hour))
	if err != nil {
		t.Fatalf("GenerateTokenWithTimes() error = %v", err)
	}
	authenticatedRequest := httptest.NewRequest(http.MethodGet, "/api/subjects?year_id=999", nil)
	authenticatedRequest.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token})
	authenticated := httptest.NewRecorder()
	router.ServeHTTP(authenticated, authenticatedRequest)
	if authenticated.Code != http.StatusOK || repository.listSubjectsCalls != 1 || repository.listSubjectsUserID != 42 {
		t.Fatalf("authenticated subjects status/calls/user = %d/%d/%d", authenticated.Code, repository.listSubjectsCalls, repository.listSubjectsUserID)
	}
}

func TestCatalogReadSurfaceDoesNotExposeAdministrativeMutations(t *testing.T) {
	router := gin.New()
	RegisterCatalogRoutes(router.Group("/api"), handlers.NewCatalogHandler(&readSurfaceCatalogRepositoryFake{}, time.Second), middleware.APIAuthMiddleware("jwt-secret"))

	for _, tt := range []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/api/admin/years"},
		{method: http.MethodPatch, path: "/api/admin/years/1"},
		{method: http.MethodDelete, path: "/api/admin/years/1"},
		{method: http.MethodPost, path: "/api/admin/subjects"},
		{method: http.MethodPatch, path: "/api/admin/subjects/1"},
		{method: http.MethodDelete, path: "/api/admin/subjects/1"},
	} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(tt.method, tt.path, strings.NewReader(`{}`)))
		if recorder.Code != http.StatusNotFound {
			t.Errorf("%s %s status = %d, want 404", tt.method, tt.path, recorder.Code)
		}
	}
}

type readSurfaceCatalogRepositoryFake struct {
	subjects           []model.Subject
	listSubjectsCalls  int
	listSubjectsUserID int64
}

func (f *readSurfaceCatalogRepositoryFake) ListYears(context.Context) ([]model.Year, error) {
	return []model.Year{{ID: 2, Year: "2"}}, nil
}

func (f *readSurfaceCatalogRepositoryFake) ListSubjects(_ context.Context, userID int64) ([]model.Subject, error) {
	f.listSubjectsCalls++
	f.listSubjectsUserID = userID
	return f.subjects, nil
}
