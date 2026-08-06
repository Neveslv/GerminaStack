package handlers

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"germinaStack/model"
)

func TestCatalogHandlerListsYearsAsJSONIncludingEmptyList(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		years []model.Year
		want  string
	}{
		{name: "values", years: []model.Year{{ID: 2, Year: "2"}}, want: `[{"id":2,"year":"2"}]`},
		{name: "empty", years: nil, want: `[]`},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repository := &catalogRepositoryFake{years: tt.years}
			recorder := performRequest(NewCatalogHandler(repository, time.Second).ListYears, http.MethodGet, "/api/years", "", "")
			if recorder.Code != http.StatusOK || strings.TrimSpace(recorder.Body.String()) != tt.want {
				t.Fatalf("status/body = %d/%s, want 200/%s", recorder.Code, recorder.Body.String(), tt.want)
			}
		})
	}
}

type catalogRepositoryFake struct {
	years              []model.Year
	subjects           []model.Subject
	err                error
	listSubjectsUserID int64
	listSubjectsCalls  int
}

func (f *catalogRepositoryFake) ListYears(context.Context) ([]model.Year, error) {
	return f.years, f.err
}

func (f *catalogRepositoryFake) ListSubjects(_ context.Context, userID int64) ([]model.Subject, error) {
	f.listSubjectsCalls++
	f.listSubjectsUserID = userID
	return f.subjects, f.err
}

func handlerInt64Pointer(value int64) *int64 { return &value }
