package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"germinaStack/auth"
	"germinaStack/model"

	"github.com/gin-gonic/gin"
)

func TestCatalogHandlerListsSubjectsForAuthenticatedUserIgnoringYearQuery(t *testing.T) {
	repository := &readOnlyCatalogRepositoryFake{
		subjects: []model.Subject{{ID: 1, YearID: handlerInt64Pointer(7), Subject: "Biologia ESG"}, {ID: 2, Subject: "Geral"}},
	}
	handler := NewCatalogHandler(repository, time.Second)
	recorder := performRequest(func(c *gin.Context) {
		c.Set(auth.ContextUserID, int64(42))
		handler.ListSubjects(c)
	}, http.MethodGet, "/api/subjects?year_id=999", "", "")

	if recorder.Code != http.StatusOK || repository.listSubjectsUserID != 42 {
		t.Fatalf("status/user = %d/%d, want 200/42; body=%s", recorder.Code, repository.listSubjectsUserID, recorder.Body.String())
	}
	var response []map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil || len(response) != 2 || response[1]["subject"] != "Geral" {
		t.Fatalf("response = %#v, error = %v", response, err)
	}
}

func TestCatalogHandlerRejectsSubjectsWithoutAuthenticatedUser(t *testing.T) {
	repository := &readOnlyCatalogRepositoryFake{}
	recorder := performRequest(NewCatalogHandler(repository, time.Second).ListSubjects, http.MethodGet, "/api/subjects", "", "")

	if recorder.Code != http.StatusUnauthorized || repository.listSubjectsCalls != 0 {
		t.Fatalf("status/calls = %d/%d, want 401/0", recorder.Code, repository.listSubjectsCalls)
	}
}

func TestCatalogHandlerMapsReadErrorsWithoutLeakingDetails(t *testing.T) {
	repository := &readOnlyCatalogRepositoryFake{err: errors.New("private SQL detail")}
	recorder := performRequest(NewCatalogHandler(repository, time.Second).ListYears, http.MethodGet, "/api/years", "", "")

	if recorder.Code != http.StatusServiceUnavailable || strings.Contains(recorder.Body.String(), "private") {
		t.Fatalf("status/body = %d/%s, want sanitized 503", recorder.Code, recorder.Body.String())
	}
}

type readOnlyCatalogRepositoryFake struct {
	years              []model.Year
	subjects           []model.Subject
	err                error
	listSubjectsUserID int64
	listSubjectsCalls  int
}

func (f *readOnlyCatalogRepositoryFake) ListYears(context.Context) ([]model.Year, error) {
	return f.years, f.err
}

func (f *readOnlyCatalogRepositoryFake) ListSubjects(_ context.Context, userID int64) ([]model.Subject, error) {
	f.listSubjectsCalls++
	f.listSubjectsUserID = userID
	return f.subjects, f.err
}
