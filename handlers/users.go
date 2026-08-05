package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"germinaStack/auth"
	"germinaStack/database"

	"github.com/gin-gonic/gin"
)

type UserRepository interface {
	CreateUser(context.Context, database.UserRegistration) (database.User, error)
}

type UserHandler struct {
	repository       UserRepository
	operationTimeout time.Duration
}

func NewUserHandler(repository UserRepository, operationTimeout time.Duration) *UserHandler {
	return &UserHandler{repository: repository, operationTimeout: operationTimeout}
}

type registerUserRequest struct {
	Email                string `json:"email"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
	YearID               int64  `json:"year_id"`
}

type userResponse struct {
	ID       int64  `json:"id"`
	YearID   int64  `json:"year_id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"is_admin"`
}

func (h *UserHandler) Register(c *gin.Context) {
	var request registerUserRequest
	if err := decodeJSON(c, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requisi\u00e7\u00e3o inv\u00e1lida"})
		return
	}

	identity, err := auth.ParseInstitutionalEmail(request.Email)
	passwordBytes := len([]byte(request.Password))
	if err != nil || passwordBytes < 8 || passwordBytes > 72 || request.Password != request.PasswordConfirmation || request.YearID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requisi\u00e7\u00e3o inv\u00e1lida"})
		return
	}

	passwordHash, err := auth.HashPassword(request.Password)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": auth.ErrUnavailable.Error()})
		return
	}

	operationContext, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	user, err := h.repository.CreateUser(operationContext, database.UserRegistration{
		YearID:       request.YearID,
		Name:         identity.Name,
		Username:     identity.Username,
		Email:        request.Email,
		PasswordHash: passwordHash,
	})
	switch {
	case err == nil:
		c.JSON(http.StatusCreated, userResponse{
			ID: user.ID, YearID: user.YearID, Name: user.Name, Username: user.Username, Email: user.Email, IsAdmin: user.IsAdmin,
		})
	case errors.Is(err, database.ErrCredentialConflict):
		c.JSON(http.StatusConflict, gin.H{"error": "e-mail ou usu\u00e1rio j\u00e1 cadastrado"})
	case errors.Is(err, database.ErrYearNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "ano n\u00e3o encontrado"})
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": auth.ErrUnavailable.Error()})
	}
}
