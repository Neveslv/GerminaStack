package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"germinaStack/database"
	"germinaStack/model"

	"github.com/gin-gonic/gin"
	"net/http/httptest"
)

func TestCatalogHandlerListsYearsAsJSONIncludingEmptyList(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		years []model.Year
		want  string
	}{
		{name: "values", years: []model.Year{{ID: 2, Year: "1st year"}}, want: `[{"id":2,"year":"1st year"}]`},
		{name: "empty", years: nil, want: `[]`},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := &catalogRepositoryFake{years: tt.years}
			recorder := performRequest(NewCatalogHandler(repo, time.Second).ListYears, http.MethodGet, "/api/years", "", "")
			if recorder.Code != http.StatusOK || strings.TrimSpace(recorder.Body.String()) != tt.want {
				t.Fatalf("status/body = %d/%s, want 200/%s", recorder.Code, recorder.Body.String(), tt.want)
			}
		})
	}
}

func TestCatalogHandlerListsSubjectsAndValidatesOptionalFilter(t *testing.T) {
	t.Parallel()
	repo := &catalogRepositoryFake{subjects: []model.Subject{{ID: 3, YearID: handlerInt64Pointer(7), Subject: "Biology"}}}
	recorder := performRequest(NewCatalogHandler(repo, time.Second).ListSubjects, http.MethodGet, "/api/subjects?year_id=7", "", "")
	if recorder.Code != http.StatusOK || repo.listSubjectsYearID == nil || *repo.listSubjectsYearID != 7 {
		t.Fatalf("status/filter = %d/%v; body=%s", recorder.Code, repo.listSubjectsYearID, recorder.Body.String())
	}
	var got []map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil || len(got) != 1 || got[0]["year_id"] != float64(7) {
		t.Fatalf("response = %#v, error=%v", got, err)
	}
	if _, exists := got[0]["id_year"]; exists {
		t.Fatalf("response exposes persistence name: %#v", got[0])
	}

	for _, query := range []string{"?year_id=", "?year_id=abc", "?year_id=0", "?year_id=-1", "?year_id=1&year_id=2"} {
		invalidRepo := &catalogRepositoryFake{}
		response := performRequest(NewCatalogHandler(invalidRepo, time.Second).ListSubjects, http.MethodGet, "/api/subjects"+query, "", "")
		if response.Code != http.StatusBadRequest || invalidRepo.listSubjectsCalls != 0 {
			t.Fatalf("query %q status/calls = %d/%d", query, response.Code, invalidRepo.listSubjectsCalls)
		}
	}
}

func TestCatalogHandlerYearMutationsValidateAndMapErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		handle func(*CatalogHandler) gin.HandlerFunc
		method string
		path   string
		body   string
		err    error
		status int
	}{
		{name: "create", handle: wrapGin((*CatalogHandler).CreateYear), method: http.MethodPost, path: "/api/admin/years", body: `{"year":" 2026 "}`, status: http.StatusCreated},
		{name: "update", handle: wrapGin((*CatalogHandler).UpdateYear), method: http.MethodPatch, path: "/api/admin/years/3", body: `{"year":"2027"}`, status: http.StatusOK},
		{name: "delete", handle: wrapGin((*CatalogHandler).DeleteYear), method: http.MethodDelete, path: "/api/admin/years/3", status: http.StatusNoContent},
		{name: "duplicate", handle: wrapGin((*CatalogHandler).CreateYear), method: http.MethodPost, path: "/api/admin/years", body: `{"year":"2026"}`, err: database.ErrCatalogConflict, status: http.StatusConflict},
		{name: "missing", handle: wrapGin((*CatalogHandler).UpdateYear), method: http.MethodPatch, path: "/api/admin/years/404", body: `{"year":"2026"}`, err: database.ErrYearNotFound, status: http.StatusNotFound},
		{name: "referenced", handle: wrapGin((*CatalogHandler).DeleteYear), method: http.MethodDelete, path: "/api/admin/years/3", err: database.ErrCatalogReferenced, status: http.StatusConflict},
		{name: "outage", handle: wrapGin((*CatalogHandler).CreateYear), method: http.MethodPost, path: "/api/admin/years", body: `{"year":"2026"}`, err: errors.New("private SQL detail"), status: http.StatusServiceUnavailable},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := &catalogRepositoryFake{year: model.Year{ID: 3, Year: "2026"}, err: tt.err}
			handler := NewCatalogHandler(repo, time.Second)
			recorder := performCatalogRequest(tt.handle(handler), tt.method, tt.path, tt.body, contentTypeFor(tt.body))
			if recorder.Code != tt.status || strings.Contains(recorder.Body.String(), "private") {
				t.Fatalf("status = %d, want %d; body=%s", recorder.Code, tt.status, recorder.Body.String())
			}
			if tt.name == "create" && repo.yearLabel != "2026" {
				t.Fatalf("trimmed label = %q", repo.yearLabel)
			}
		})
	}

	for _, tt := range []struct{ name, path, body string }{
		{name: "missing content type", path: "/api/admin/years", body: `{"year":"2026"}`},
		{name: "empty label", path: "/api/admin/years", body: `{"year":"  "}`},
		{name: "long label", path: "/api/admin/years", body: `{"year":"` + strings.Repeat("x", 101) + `"}`},
		{name: "unknown field", path: "/api/admin/years", body: `{"year":"2026","extra":true}`},
		{name: "trailing JSON", path: "/api/admin/years", body: `{"year":"2026"}{}`},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := &catalogRepositoryFake{}
			contentType := "application/json"
			if tt.name == "missing content type" {
				contentType = ""
			}
			recorder := performRequest(NewCatalogHandler(repo, time.Second).CreateYear, http.MethodPost, tt.path, tt.body, contentType)
			if recorder.Code != http.StatusBadRequest || repo.createYearCalls != 0 {
				t.Fatalf("status/calls = %d/%d; body=%s", recorder.Code, repo.createYearCalls, recorder.Body.String())
			}
		})
	}

	repo := &catalogRepositoryFake{}
	badID := performCatalogRequest(NewCatalogHandler(repo, time.Second).UpdateYear, http.MethodPatch, "/api/admin/years/0", `{"year":"2026"}`, "application/json")
	if badID.Code != http.StatusBadRequest || repo.updateYearCalls != 0 {
		t.Fatalf("invalid id status/calls = %d/%d", badID.Code, repo.updateYearCalls)
	}
}

func TestCatalogHandlerSubjectMutationsValidateNullableYearAndMapErrors(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		body   string
		err    error
		status int
	}{
		{name: "nullable year", body: `{"subject":" General ","year_id":null}`, status: http.StatusCreated},
		{name: "positive year", body: `{"subject":"Biology","year_id":7}`, status: http.StatusCreated},
		{name: "duplicate", body: `{"subject":"Biology","year_id":7}`, err: database.ErrCatalogConflict, status: http.StatusConflict},
		{name: "unknown year", body: `{"subject":"Biology","year_id":7}`, err: database.ErrYearNotFound, status: http.StatusNotFound},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := &catalogRepositoryFake{subject: model.Subject{ID: 3, Subject: "Biology"}, err: tt.err}
			recorder := performRequest(NewCatalogHandler(repo, time.Second).CreateSubject, http.MethodPost, "/api/admin/subjects", tt.body, "application/json")
			if recorder.Code != tt.status {
				t.Fatalf("status = %d, want %d; body=%s", recorder.Code, tt.status, recorder.Body.String())
			}
			if tt.name == "nullable year" && (repo.subjectLabel != "General" || repo.subjectYearID != nil) {
				t.Fatalf("normalized request = %q/%v", repo.subjectLabel, repo.subjectYearID)
			}
		})
	}

	for _, body := range []string{
		`{"subject":"Biology","year_id":0}`,
		`{"subject":"Biology","year_id":-1}`,
		`{"subject":"Biology","year_id":"7"}`,
		`{"subject":"","year_id":null}`,
		`{"subject":"Biology"}`,
	} {
		repo := &catalogRepositoryFake{}
		recorder := performRequest(NewCatalogHandler(repo, time.Second).CreateSubject, http.MethodPost, "/api/admin/subjects", body, "application/json")
		if recorder.Code != http.StatusBadRequest || repo.createSubjectCalls != 0 {
			t.Fatalf("body %s status/calls = %d/%d", body, recorder.Code, repo.createSubjectCalls)
		}
	}

	repo := &catalogRepositoryFake{subject: model.Subject{ID: 3, YearID: handlerInt64Pointer(7), Subject: "Biology"}}
	updated := performCatalogRequest(NewCatalogHandler(repo, time.Second).UpdateSubject, http.MethodPatch, "/api/admin/subjects/3", `{"subject":"Biology","year_id":7}`, "application/json")
	if updated.Code != http.StatusOK || repo.updateSubjectCalls != 1 {
		t.Fatalf("update status/calls = %d/%d; body=%s", updated.Code, repo.updateSubjectCalls, updated.Body.String())
	}

	repo = &catalogRepositoryFake{err: database.ErrCatalogReferenced}
	deleted := performCatalogRequest(NewCatalogHandler(repo, time.Second).DeleteSubject, http.MethodDelete, "/api/admin/subjects/3", "", "")
	if deleted.Code != http.StatusConflict {
		t.Fatalf("delete referenced status = %d; body=%s", deleted.Code, deleted.Body.String())
	}
}

type catalogRepositoryFake struct {
	years              []model.Year
	subjects           []model.Subject
	year               model.Year
	subject            model.Subject
	err                error
	listSubjectsYearID *int64
	subjectYearID      *int64
	yearLabel          string
	subjectLabel       string
	listSubjectsCalls  int
	createYearCalls    int
	updateYearCalls    int
	createSubjectCalls int
	updateSubjectCalls int
}

func (f *catalogRepositoryFake) ListYears(context.Context) ([]model.Year, error) {
	return f.years, f.err
}
func (f *catalogRepositoryFake) CreateYear(_ context.Context, label string) (model.Year, error) {
	f.createYearCalls++
	f.yearLabel = label
	return f.year, f.err
}
func (f *catalogRepositoryFake) UpdateYear(_ context.Context, _ int64, label string) (model.Year, error) {
	f.updateYearCalls++
	f.yearLabel = label
	return f.year, f.err
}
func (f *catalogRepositoryFake) DeleteYear(context.Context, int64) error { return f.err }
func (f *catalogRepositoryFake) ListSubjects(_ context.Context, yearID *int64) ([]model.Subject, error) {
	f.listSubjectsCalls++
	f.listSubjectsYearID = yearID
	return f.subjects, f.err
}
func (f *catalogRepositoryFake) CreateSubject(_ context.Context, label string, yearID *int64) (model.Subject, error) {
	f.createSubjectCalls++
	f.subjectLabel, f.subjectYearID = label, yearID
	return f.subject, f.err
}
func (f *catalogRepositoryFake) UpdateSubject(_ context.Context, _ int64, label string, yearID *int64) (model.Subject, error) {
	f.updateSubjectCalls++
	f.subjectLabel, f.subjectYearID = label, yearID
	return f.subject, f.err
}
func (f *catalogRepositoryFake) DeleteSubject(context.Context, int64) error { return f.err }

func handlerInt64Pointer(value int64) *int64 { return &value }

func contentTypeFor(body string) string {
	if body == "" {
		return ""
	}
	return "application/json"
}

func wrapGin(method func(*CatalogHandler, *gin.Context)) func(*CatalogHandler) gin.HandlerFunc {
	return func(handler *CatalogHandler) gin.HandlerFunc {
		return func(c *gin.Context) { method(handler, c) }
	}
}

func performCatalogRequest(handler gin.HandlerFunc, method, path, body, contentType string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	router := gin.New()
	routePath := path
	if method == http.MethodPatch || method == http.MethodDelete {
		lastSlash := strings.LastIndex(path, "/")
		routePath = path[:lastSlash] + "/:id"
	}
	router.Handle(method, routePath, handler)
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", contentType)
	router.ServeHTTP(recorder, request)
	return recorder
}
