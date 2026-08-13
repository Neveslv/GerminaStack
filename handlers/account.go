package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"germinaStack/database"
	"germinaStack/model"

	"github.com/gin-gonic/gin"
)

type AccountRepository interface {
	GetProfile(context.Context, int64) (model.User, error)
	GetPublicProfile(context.Context, string) (model.User, error)
	UpdateProfile(context.Context, int64, model.User) (model.User, error)
	GetPreferences(context.Context, int64) (model.Preference, error)
	UpsertPreferences(context.Context, int64, model.Preference) (model.Preference, error)
}

type publicProfileResponse struct {
	ID                      int64      `json:"id"`
	YearID                  int64      `json:"id_year"`
	Name                    string     `json:"name"`
	ProfileImageURL         *string    `json:"profile_image_url"`
	ProfileImageDescription *string    `json:"profile_image_description"`
	Username                string     `json:"username"`
	CreatedAt               *time.Time `json:"created_at"`
}

type AccountHandler struct {
	repository       AccountRepository
	operationTimeout time.Duration
}

func NewAccountHandler(repository AccountRepository, operationTimeout time.Duration) *AccountHandler {
	return &AccountHandler{repository: repository, operationTimeout: operationTimeout}
}

func (h *AccountHandler) GetProfile(c *gin.Context) {
	userID, ok := discussionUserID(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	profile, err := h.repository.GetProfile(ctx, userID)
	if err != nil {
		writeAccountError(c, err)
		return
	}
	c.JSON(http.StatusOK, profile)
}

func (h *AccountHandler) GetPublicProfile(c *gin.Context) {
	username := strings.TrimSpace(c.Param("username"))
	if username == "" || len(username) > 100 {
		writeInvalidAccountRequest(c)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	profile, err := h.repository.GetPublicProfile(ctx, username)
	if err != nil {
		writeAccountError(c, err)
		return
	}
	c.JSON(http.StatusOK, publicProfileResponse{
		ID: profile.ID, YearID: profile.YearID, Name: profile.Name,
		ProfileImageURL: profile.ProfileImageURL, ProfileImageDescription: profile.ProfileImageDescription,
		Username: profile.Username, CreatedAt: profile.CreatedAt,
	})
}

func (h *AccountHandler) UpdateProfile(c *gin.Context) {
	userID, ok := discussionUserID(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	profile, err := h.repository.GetProfile(ctx, userID)
	if err != nil {
		writeAccountError(c, err)
		return
	}
	fields, ok := decodeObject(c)
	if !ok || !applyProfilePatch(fields, &profile) {
		writeInvalidAccountRequest(c)
		return
	}
	updated, err := h.repository.UpdateProfile(ctx, userID, profile)
	if err != nil {
		writeAccountError(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *AccountHandler) GetPreferences(c *gin.Context) {
	userID, ok := discussionUserID(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	preference, err := h.repository.GetPreferences(ctx, userID)
	if errors.Is(err, database.ErrPreferencesNotFound) {
		c.JSON(http.StatusOK, model.Preference{UserID: userID})
		return
	}
	if err != nil {
		writeAccountError(c, err)
		return
	}
	c.JSON(http.StatusOK, preference)
}

func (h *AccountHandler) UpdatePreferences(c *gin.Context) {
	userID, ok := discussionUserID(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	preference, err := h.repository.GetPreferences(ctx, userID)
	if errors.Is(err, database.ErrPreferencesNotFound) {
		preference = model.Preference{UserID: userID}
	} else if err != nil {
		writeAccountError(c, err)
		return
	}
	fields, ok := decodeObject(c)
	if !ok || !applyPreferencePatch(fields, &preference) || preference.Validate() != nil {
		writeInvalidAccountRequest(c)
		return
	}
	updated, err := h.repository.UpsertPreferences(ctx, userID, preference)
	if err != nil {
		writeAccountError(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func decodeObject(c *gin.Context) (map[string]json.RawMessage, bool) {
	var fields map[string]json.RawMessage
	if err := decodeJSON(c, &fields); err != nil || fields == nil {
		return nil, false
	}
	return fields, true
}

func applyProfilePatch(fields map[string]json.RawMessage, profile *model.User) bool {
	if len(fields) == 0 {
		return false
	}
	for key, raw := range fields {
		switch key {
		case "name":
			value, ok := optionalString(raw, false)
			if !ok || value == nil || len([]byte(strings.TrimSpace(*value))) == 0 || len([]byte(*value)) > 150 {
				return false
			}
			profile.Name = strings.TrimSpace(*value)
		case "profile_image_url":
			value, ok := optionalString(raw, true)
			if !ok {
				return false
			}
			profile.ProfileImageURL = value
		case "profile_image_description":
			value, ok := optionalString(raw, true)
			if !ok {
				return false
			}
			profile.ProfileImageDescription = value
		default:
			return false
		}
	}
	return profile.Validate() == nil
}

func applyPreferencePatch(fields map[string]json.RawMessage, preference *model.Preference) bool {
	if len(fields) == 0 {
		return false
	}
	for key, raw := range fields {
		switch key {
		case "contrast_theme":
			value, ok := optionalString(raw, true)
			if !ok {
				return false
			}
			if value == nil {
				preference.ContrastTheme = nil
			} else {
				parsed := model.ContrastTheme(*value)
				preference.ContrastTheme = &parsed
			}
		case "font_family":
			value, ok := optionalString(raw, true)
			if !ok {
				return false
			}
			if value == nil {
				preference.FontFamily = nil
			} else {
				parsed := model.FontFamily(*value)
				preference.FontFamily = &parsed
			}
		case "font_spacing":
			value, ok := optionalString(raw, true)
			if !ok {
				return false
			}
			if value == nil {
				preference.FontSpacing = nil
			} else {
				parsed := model.FontSpacing(*value)
				preference.FontSpacing = &parsed
			}
		case "font_size":
			value, ok := optionalString(raw, true)
			if !ok {
				return false
			}
			if value == nil {
				preference.FontSize = nil
			} else {
				parsed := model.FontSize(*value)
				preference.FontSize = &parsed
			}
		default:
			return false
		}
	}
	return true
}

func optionalString(raw json.RawMessage, allowNull bool) (*string, bool) {
	if string(raw) == "null" {
		return nil, allowNull
	}
	var value string
	if json.Unmarshal(raw, &value) != nil {
		return nil, false
	}
	value = strings.TrimSpace(value)
	return &value, true
}

func writeAccountError(c *gin.Context, err error) {
	if errors.Is(err, database.ErrAccountNotFound) || errors.Is(err, database.ErrPreferencesNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "recurso não encontrado"})
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": "serviço indisponível"})
}

func writeInvalidAccountRequest(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"error": "requisição inválida"})
}
