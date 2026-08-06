package handlers

import (
	"context"
	"errors"
	"net/http"

	"germinaStack/auth"
	"germinaStack/database"
	"germinaStack/model"

	"github.com/gin-gonic/gin"
)

type UserAccountRepository interface {
	FindUserAccount(context.Context, int64) (database.UserAccount, error)
	UpdateAccessibilityPreferences(context.Context, int64, database.AccessibilityPreferences) (database.AccessibilityPreferences, error)
}

type accessibilityPreferencesRequest struct {
	ContrastTheme *model.ContrastTheme `json:"contrast_theme"`
	FontFamily    *model.FontFamily    `json:"font_family"`
	FontSpacing   *model.FontSpacing   `json:"font_spacing"`
	FontSize      *model.FontSize      `json:"font_size"`
}
type accessibilityPreferencesResponse struct {
	ContrastTheme *model.ContrastTheme `json:"contrast_theme"`
	FontFamily    *model.FontFamily    `json:"font_family"`
	FontSpacing   *model.FontSpacing   `json:"font_spacing"`
	FontSize      *model.FontSize      `json:"font_size"`
}
type currentUserResponse struct {
	ID          int64                            `json:"id"`
	YearID      *int64                           `json:"year_id"`
	Name        string                           `json:"name"`
	Username    string                           `json:"username"`
	Email       string                           `json:"email"`
	IsAdmin     bool                             `json:"is_admin"`
	Preferences accessibilityPreferencesResponse `json:"preferences"`
}

func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autorizado"})
		return
	}
	repository, ok := h.repository.(UserAccountRepository)
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": auth.ErrUnavailable.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	account, err := repository.FindUserAccount(ctx, userID)
	if err != nil {
		writeAccountRepositoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, currentUserResponse{ID: account.ID, YearID: account.YearID, Name: account.Name, Username: account.Username, Email: account.Email, IsAdmin: account.IsAdmin, Preferences: accessibilityPreferencesResponse{ContrastTheme: account.Preferences.ContrastTheme, FontFamily: account.Preferences.FontFamily, FontSpacing: account.Preferences.FontSpacing, FontSize: account.Preferences.FontSize}})
}

func (h *UserHandler) UpdateAccessibilityPreferences(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autorizado"})
		return
	}
	var request accessibilityPreferencesRequest
	if err := decodeJSON(c, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requisição inválida"})
		return
	}
	preference := model.Preference{UserID: userID, ContrastTheme: request.ContrastTheme, FontFamily: request.FontFamily, FontSpacing: request.FontSpacing, FontSize: request.FontSize}
	if err := preference.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requisição inválida"})
		return
	}
	repository, ok := h.repository.(UserAccountRepository)
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": auth.ErrUnavailable.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	updated, err := repository.UpdateAccessibilityPreferences(ctx, userID, database.AccessibilityPreferences{ContrastTheme: request.ContrastTheme, FontFamily: request.FontFamily, FontSpacing: request.FontSpacing, FontSize: request.FontSize})
	if err != nil {
		writeAccountRepositoryError(c, err)
		return
	}
	c.JSON(http.StatusOK, accessibilityPreferencesResponse{ContrastTheme: updated.ContrastTheme, FontFamily: updated.FontFamily, FontSpacing: updated.FontSpacing, FontSize: updated.FontSize})
}

func authenticatedUserID(c *gin.Context) (int64, bool) {
	value, exists := c.Get(auth.ContextUserID)
	if !exists {
		return 0, false
	}
	userID, ok := value.(int64)
	return userID, ok && userID > 0
}
func writeAccountRepositoryError(c *gin.Context, err error) {
	if errors.Is(err, database.ErrCredentialNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "conta não encontrada"})
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": auth.ErrUnavailable.Error()})
}
