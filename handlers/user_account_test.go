package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"germinaStack/auth"
	"germinaStack/database"
	"germinaStack/model"

	"github.com/gin-gonic/gin"
)

func TestGetCurrentUserReadsOnlyAuthenticatedAccount(t *testing.T) {
	dark := model.ContrastThemeDark
	studentYear := int64(7)
	for _, tt := range []struct {
		name    string
		isAdmin bool
		yearID  *int64
	}{{name: "student", yearID: &studentYear}, {name: "admin without year", isAdmin: true}} {
		t.Run(tt.name, func(t *testing.T) {
			repository := &accountRepositoryFake{account: database.UserAccount{ID: 42, YearID: tt.yearID, Name: "Ana Silva", Username: "ana.silva", Email: "ana.silva@institutojef.org.br", IsAdmin: tt.isAdmin, Preferences: database.AccessibilityPreferences{ContrastTheme: &dark}}}
			recorder := performAccountRequest(NewUserHandler(repository, time.Second).GetCurrentUser, http.MethodGet, "/api/users/me", "", tt.isAdmin, true)
			if recorder.Code != http.StatusOK || repository.findUserID != 42 {
				t.Fatalf("status/user ID = %d/%d", recorder.Code, repository.findUserID)
			}
			var response map[string]any
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			preferences, ok := response["preferences"].(map[string]any)
			if response["id"] != float64(42) || response["is_admin"] != tt.isAdmin || !ok || preferences["contrast_theme"] != "dark" {
				t.Fatalf("account response = %#v", response)
			}
			if tt.yearID == nil && response["year_id"] != nil {
				t.Fatalf("admin year_id = %#v, want null", response["year_id"])
			}
			if _, exposed := response["password"]; exposed {
				t.Fatal("response exposes password")
			}
		})
	}
}

func TestCurrentAccountHandlersRequireAuthentication(t *testing.T) {
	repository := &accountRepositoryFake{}
	handler := NewUserHandler(repository, time.Second)
	getResponse := performAccountRequest(handler.GetCurrentUser, http.MethodGet, "/api/users/me", "", false, false)
	updateResponse := performAccountRequest(handler.UpdateAccessibilityPreferences, http.MethodPut, "/api/users/me/preferences", `{}`, false, false)
	if getResponse.Code != http.StatusUnauthorized || updateResponse.Code != http.StatusUnauthorized || repository.findCalls != 0 || repository.updateCalls != 0 {
		t.Fatalf("statuses/calls = %d/%d/%d/%d", getResponse.Code, updateResponse.Code, repository.findCalls, repository.updateCalls)
	}
}

func TestUpdateAccessibilityPreferencesAllowsStudentAndAdmin(t *testing.T) {
	for _, isAdmin := range []bool{false, true} {
		t.Run(map[bool]string{false: "student", true: "admin"}[isAdmin], func(t *testing.T) {
			dark := model.ContrastThemeDark
			arial := model.FontFamilyArial
			spacing := model.FontSpacingGrande
			size := model.FontSizeGrande
			repository := &accountRepositoryFake{updated: database.AccessibilityPreferences{ContrastTheme: &dark, FontFamily: &arial, FontSpacing: &spacing, FontSize: &size}}
			recorder := performAccountRequest(NewUserHandler(repository, time.Second).UpdateAccessibilityPreferences, http.MethodPut, "/api/users/me/preferences", `{"contrast_theme":"dark","font_family":"arial","font_spacing":"grande","font_size":"grande"}`, isAdmin, true)
			if recorder.Code != http.StatusOK || repository.updateUserID != 42 || repository.updateCalls != 1 {
				t.Fatalf("status/user/calls = %d/%d/%d", recorder.Code, repository.updateUserID, repository.updateCalls)
			}
		})
	}
}

func TestUpdateAccessibilityPreferencesRejectsInvalidOrAccountFields(t *testing.T) {
	for _, body := range []string{`{`, `{"contrast_theme":"neon"}`, `{"name":"Outra Pessoa"}`, `{"email":"outra.pessoa@institutojef.org.br"}`, `{"is_admin":true}`} {
		repository := &accountRepositoryFake{}
		recorder := performAccountRequest(NewUserHandler(repository, time.Second).UpdateAccessibilityPreferences, http.MethodPut, "/api/users/me/preferences", body, false, true)
		if recorder.Code != http.StatusBadRequest || repository.updateCalls != 0 {
			t.Fatalf("body %s: status/calls = %d/%d", body, recorder.Code, repository.updateCalls)
		}
	}
}

type accountRepositoryFake struct {
	account      database.UserAccount
	findCalls    int
	findUserID   int64
	updated      database.AccessibilityPreferences
	updateCalls  int
	updateUserID int64
}

func (f *accountRepositoryFake) CreateUser(context.Context, database.UserRegistration) (database.User, error) {
	return database.User{}, nil
}
func (f *accountRepositoryFake) FindUserAccount(_ context.Context, userID int64) (database.UserAccount, error) {
	f.findCalls++
	f.findUserID = userID
	return f.account, nil
}
func (f *accountRepositoryFake) UpdateAccessibilityPreferences(_ context.Context, userID int64, _ database.AccessibilityPreferences) (database.AccessibilityPreferences, error) {
	f.updateCalls++
	f.updateUserID = userID
	return f.updated, nil
}
func performAccountRequest(handler gin.HandlerFunc, method, path, body string, isAdmin, authenticated bool) *httptest.ResponseRecorder {
	router := gin.New()
	router.Handle(method, path, func(c *gin.Context) {
		if authenticated {
			c.Set(auth.ContextUserID, int64(42))
			c.Set(auth.ContextIsAdmin, isAdmin)
		}
		handler(c)
	})
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}
