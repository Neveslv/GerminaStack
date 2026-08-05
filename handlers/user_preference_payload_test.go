package handlers

import (
	"context"
	"germinaStack/database"
	"germinaStack/model"
	"net/http"
	"testing"
	"time"
)

func TestUpdateAccessibilityPreferencesForwardsOnlyValidatedPreferencePayload(t *testing.T) {
	repo := &preferencePayloadRepository{}
	recorder := performAccountRequest(NewUserHandler(repo, time.Second).UpdateAccessibilityPreferences, http.MethodPut, "/api/users/me/preferences", `{"contrast_theme":"dark","font_family":"arial","font_spacing":"grande","font_size":"grande"}`, false, true)
	if recorder.Code != http.StatusOK || repo.userID != 42 {
		t.Fatalf("status/user=%d/%d", recorder.Code, repo.userID)
	}
	if repo.preferences.ContrastTheme == nil || *repo.preferences.ContrastTheme != model.ContrastThemeDark || repo.preferences.FontFamily == nil || *repo.preferences.FontFamily != model.FontFamilyArial || repo.preferences.FontSpacing == nil || *repo.preferences.FontSpacing != model.FontSpacingGrande || repo.preferences.FontSize == nil || *repo.preferences.FontSize != model.FontSizeGrande {
		t.Fatalf("preferences=%#v", repo.preferences)
	}
}

type preferencePayloadRepository struct {
	userID      int64
	preferences database.AccessibilityPreferences
}

func (*preferencePayloadRepository) CreateUser(context.Context, database.UserRegistration) (database.User, error) {
	return database.User{}, nil
}
func (*preferencePayloadRepository) FindUserAccount(context.Context, int64) (database.UserAccount, error) {
	return database.UserAccount{}, nil
}
func (r *preferencePayloadRepository) UpdateAccessibilityPreferences(_ context.Context, userID int64, p database.AccessibilityPreferences) (database.AccessibilityPreferences, error) {
	r.userID = userID
	r.preferences = p
	return p, nil
}
