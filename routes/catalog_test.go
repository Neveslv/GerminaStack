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

func TestRegisterCatalogRoutesKeepsYearsPublic(t *testing.T) {
	router := gin.New()
	RegisterCatalogRoutes(router.Group("/api"), handlers.NewCatalogHandler(&routeCatalogRepositoryFake{}, time.Second), middleware.APIAuthMiddleware("jwt-secret"))

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/years", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("years status = %d; body=%s", recorder.Code, recorder.Body.String())
	}
}

type routeCatalogRepositoryFake struct{}

func (routeCatalogRepositoryFake) ListYears(context.Context) ([]model.Year, error) {
	return []model.Year{{ID: 2, Year: "2"}}, nil
}

func (routeCatalogRepositoryFake) ListSubjects(context.Context, int64) ([]model.Subject, error) {
	return []model.Subject{{Subject: "Geral"}}, nil
}
