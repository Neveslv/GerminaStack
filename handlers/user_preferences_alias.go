package handlers

import (
	"context"
	"net/http"

	"germinaStack/auth"

	"github.com/gin-gonic/gin"
)

// GetAccessibilityPreferences returns the preference object expected by the
// browser API. The account endpoint keeps the same data nested under
// `preferences` for the profile surface.
func (h *UserHandler) GetAccessibilityPreferences(c *gin.Context) {
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
	c.JSON(http.StatusOK, accessibilityPreferencesResponse{
		ContrastTheme: account.Preferences.ContrastTheme,
		FontFamily:    account.Preferences.FontFamily,
		FontSpacing:   account.Preferences.FontSpacing,
		FontSize:      account.Preferences.FontSize,
	})
}
