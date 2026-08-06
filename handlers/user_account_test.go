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

func TestGetCurrentUserReturnsAccountAndNeverPassword(t *testing.T) {
	yearID := int64(7)
	dark := model.ContrastThemeDark
	repository := &userAccountRepositoryFake{account: database.UserAccount{
		ID: 42, YearID: &yearID, Name: "Ana", Username: "ana", Email: "ana@example.test", Preferences: database.AccessibilityPreferences{ContrastTheme: &dark},
	}}
	recorder := performUserAccountRequest(NewUserHandler(repository, time.Second).GetCurrentUser, http.MethodGet, "/api/users/me", "", true)
	if recorder.Code != http.StatusOK || repository.findUserID != 42 {
		t.Fatalf("status/user = %d/%d", recorder.Code, repository.findUserID)
	}
	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, exposed := payload["password"]; exposed || payload["year_id"] != float64(7) {
		t.Fatalf("account response = %#v", payload)
	}
}

func TestAccountHandlersRequireAuthenticationAndValidatePreferences(t *testing.T) {
	repository := &userAccountRepositoryFake{}
	handler := NewUserHandler(repository, time.Second)
	if got := performUserAccountRequest(handler.GetCurrentUser, http.MethodGet, "/api/users/me", "", false).Code; got != http.StatusUnauthorized {
		t.Fatalf("GET unauthenticated status = %d", got)
	}
	if got := performUserAccountRequest(handler.UpdateAccessibilityPreferences, http.MethodPut, "/api/preferences", `{"contrast_theme":"neon"}`, true).Code; got != http.StatusBadRequest {
		t.Fatalf("PUT invalid status = %d", got)
	}
	if repository.updateCalls != 0 {
		t.Fatalf("invalid preference update calls = %d", repository.updateCalls)
	}
}

func TestUpdateAccessibilityPreferencesForwardsOnlyPreferenceFields(t *testing.T) {
	repository := &userAccountRepositoryFake{}
	recorder := performUserAccountRequest(NewUserHandler(repository, time.Second).UpdateAccessibilityPreferences, http.MethodPut, "/api/users/me/preferences", `{"contrast_theme":"dark"}`, true)
	if recorder.Code != http.StatusOK || repository.updateUserID != 42 || repository.updated.ContrastTheme == nil || *repository.updated.ContrastTheme != model.ContrastThemeDark {
		t.Fatalf("status/user/preferences = %d/%d/%#v", recorder.Code, repository.updateUserID, repository.updated)
	}
}

type userAccountRepositoryFake struct {
	account      database.UserAccount
	findUserID   int64
	updated      database.AccessibilityPreferences
	updateUserID int64
	updateCalls  int
}

func (f *userAccountRepositoryFake) CreateUser(context.Context, database.UserRegistration) (database.User, error) {
	return database.User{}, nil
}
func (f *userAccountRepositoryFake) FindUserAccount(_ context.Context, userID int64) (database.UserAccount, error) {
	f.findUserID = userID
	return f.account, nil
}
func (f *userAccountRepositoryFake) UpdateAccessibilityPreferences(_ context.Context, userID int64, preferences database.AccessibilityPreferences) (database.AccessibilityPreferences, error) {
	f.updateCalls++
	f.updateUserID = userID
	f.updated = preferences
	return preferences, nil
}

func performUserAccountRequest(handler gin.HandlerFunc, method, path, body string, authenticated bool) *httptest.ResponseRecorder {
	router := gin.New()
	router.Handle(method, path, func(c *gin.Context) {
		if authenticated {
			c.Set(auth.ContextUserID, int64(42))
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
