package handlers

import (
	"context"
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

func TestAccountHandlerScopesProfileAndPreferencesToJWT(t *testing.T) {
	t.Parallel()
	repository := &accountRepositoryFake{profile: model.User{ID: 42, Name: "Ana"}}
	handler := NewAccountHandler(repository, time.Second)
	profile := performAccountRequest(handler.GetProfile, http.MethodGet, "/api/me", "", 42)
	updated := performAccountRequest(handler.UpdateProfile, http.MethodPatch, "/api/me", `{"name":" Ana Silva ","email":"changed@example.com"}`, 42)
	preference := performAccountRequest(handler.UpdatePreferences, http.MethodPatch, "/api/me/preferences", `{"font_size":"grande","font_family":"lexend"}`, 42)
	if profile.Code != http.StatusOK || updated.Code != http.StatusBadRequest || preference.Code != http.StatusOK || repository.profileUserID != 42 || repository.preferenceUserID != 42 || repository.preference.FontSize == nil || *repository.preference.FontSize != model.FontSizeGrande {
		t.Fatalf("status/profile/preferences = %d/%d/%d; repository=%#v", profile.Code, updated.Code, preference.Code, repository)
	}

	for _, body := range []string{`{"font_size":"invalid"}`, `{"font_family":"arial","font_spacing":"wrong"}`} {
		response := performAccountRequest(handler.UpdatePreferences, http.MethodPatch, "/api/me/preferences", body, 42)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("body %s status = %d", body, response.Code)
		}
	}
	unauthorized := performAccountRequest(handler.GetProfile, http.MethodGet, "/api/me", "", 0)
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d", unauthorized.Code)
	}
}

func TestAccountHandlerValidatesProfileImagePairAndDefaultsPreferences(t *testing.T) {
	t.Parallel()
	repository := &accountRepositoryFake{profile: model.User{ID: 42, Name: "Ana"}, preferenceNotFound: true}
	handler := NewAccountHandler(repository, time.Second)
	unpaired := performAccountRequest(handler.UpdateProfile, http.MethodPatch, "/api/me", `{"profile_image_url":"https://image"}`, 42)
	paired := performAccountRequest(handler.UpdateProfile, http.MethodPatch, "/api/me", `{"profile_image_url":"https://image","profile_image_description":"Ana em frente à tela"}`, 42)
	preferences := performAccountRequest(handler.GetPreferences, http.MethodGet, "/api/me/preferences", "", 42)
	if unpaired.Code != http.StatusBadRequest || paired.Code != http.StatusOK || preferences.Code != http.StatusOK || repository.profile.ProfileImageURL == nil || repository.profile.ProfileImageDescription == nil {
		t.Fatalf("status/profile/preferences = %d/%d/%d; repository=%#v", unpaired.Code, paired.Code, preferences.Code, repository)
	}
}

func TestAccountHandlerReturnsPublicProfileWithoutEmail(t *testing.T) {
	t.Parallel()
	repository := &accountRepositoryFake{publicProfile: model.User{ID: 7, YearID: 2, Name: "Bruno", Username: "bruno.salles", Email: "private@example.com"}}
	response := performPublicProfileRequest(NewAccountHandler(repository, time.Second).GetPublicProfile, "bruno.salles", 42)
	if response.Code != http.StatusOK || strings.Contains(response.Body.String(), "private@example.com") || repository.publicUsername != "bruno.salles" {
		t.Fatalf("status/body/repository = %d/%s/%#v", response.Code, response.Body.String(), repository)
	}
}

func performPublicProfileRequest(handler gin.HandlerFunc, username string, userID int64) *httptest.ResponseRecorder {
	router := gin.New()
	router.GET("/api/users/:username", func(c *gin.Context) {
		c.Set(auth.ContextUserID, userID)
		handler(c)
	})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/users/"+username, nil))
	return recorder
}

type accountRepositoryFake struct {
	profile            model.User
	profileUserID      int64
	publicProfile      model.User
	publicUsername     string
	preference         model.Preference
	preferenceUserID   int64
	preferenceNotFound bool
}

func (f *accountRepositoryFake) GetProfile(_ context.Context, userID int64) (model.User, error) {
	f.profileUserID = userID
	return f.profile, nil
}
func (f *accountRepositoryFake) GetPublicProfile(_ context.Context, username string) (model.User, error) {
	f.publicUsername = username
	return f.publicProfile, nil
}
func (f *accountRepositoryFake) UpdateProfile(_ context.Context, userID int64, profile model.User) (model.User, error) {
	f.profileUserID = userID
	f.profile = profile
	return profile, nil
}
func (f *accountRepositoryFake) GetPreferences(_ context.Context, userID int64) (model.Preference, error) {
	f.preferenceUserID = userID
	if f.preferenceNotFound {
		return model.Preference{}, database.ErrPreferencesNotFound
	}
	return f.preference, nil
}
func (f *accountRepositoryFake) UpsertPreferences(_ context.Context, userID int64, preference model.Preference) (model.Preference, error) {
	f.preferenceUserID = userID
	f.preference = preference
	return preference, nil
}

func performAccountRequest(handler gin.HandlerFunc, method, path, body string, values ...any) *httptest.ResponseRecorder {
	userID := int64(0)
	isAdmin := false
	authenticated := false
	if len(values) == 1 {
		switch value := values[0].(type) {
		case int:
			userID = int64(value)
		case int64:
			userID = value
		}
		authenticated = userID > 0
	} else if len(values) >= 2 {
		isAdmin, _ = values[0].(bool)
		authenticated, _ = values[1].(bool)
		if authenticated {
			userID = 42
		}
	}
	router := gin.New()
	router.Handle(method, path, func(c *gin.Context) {
		if authenticated {
			c.Set(auth.ContextUserID, userID)
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
