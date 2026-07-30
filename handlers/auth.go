package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"germinaStack/auth"

	"github.com/gin-gonic/gin"
)

const maxRequestBodyBytes int64 = 4 * 1024

var codePattern = regexp.MustCompile(`^[0-9]{6}$`)

type AuthService interface {
	StartLogin(context.Context, string, string) (string, error)
	CompleteLogin(context.Context, string, string) (int64, error)
}

type AuthHandler struct {
	service      AuthService
	jwtSecret    string
	cookieSecure bool
	tokenTTL     time.Duration
	now          func() time.Time
}

func NewAuthHandler(service AuthService, jwtSecret string, cookieSecure bool, tokenTTL time.Duration) *AuthHandler {
	return &AuthHandler{
		service:      service,
		jwtSecret:    jwtSecret,
		cookieSecure: cookieSecure,
		tokenTTL:     tokenTTL,
		now:          time.Now,
	}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var request loginRequest
	if err := decodeJSON(c, &request); err != nil ||
		len(request.Username) < 1 ||
		len(request.Username) > 100 ||
		len(request.Password) < 1 ||
		len(request.Password) > 1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requisição inválida"})
		return
	}

	challengeID, err := h.service.StartLogin(c.Request.Context(), request.Username, request.Password)
	switch {
	case err == nil:
		c.JSON(http.StatusAccepted, gin.H{"challenge_id": challengeID})
	case errors.Is(err, auth.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": auth.ErrInvalidCredentials.Error()})
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": auth.ErrUnavailable.Error()})
	}
}

type completeLoginRequest struct {
	ChallengeID string `json:"challenge_id"`
	Code        string `json:"code"`
}

func (h *AuthHandler) CompleteLogin(c *gin.Context) {
	var request completeLoginRequest
	if err := decodeJSON(c, &request); err != nil ||
		len(request.ChallengeID) < 1 ||
		len(request.ChallengeID) > 128 ||
		!codePattern.MatchString(request.Code) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requisição inválida"})
		return
	}

	userID, err := h.service.CompleteLogin(c.Request.Context(), request.ChallengeID, request.Code)
	if err != nil {
		h.writeCompleteLoginError(c, err)
		return
	}

	issuedAt := h.now().UTC()
	expiresAt := issuedAt.Add(h.tokenTTL)
	token, err := auth.GenerateTokenWithTimes(strconv.FormatInt(userID, 10), false, h.jwtSecret, issuedAt, expiresAt)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": auth.ErrUnavailable.Error()})
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     auth.CookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   int(h.tokenTTL / time.Second),
		Secure:   h.cookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	c.Status(http.StatusNoContent)
	c.Writer.WriteHeaderNow()
}

func (h *AuthHandler) writeCompleteLoginError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, auth.ErrMalformedCode):
		c.JSON(http.StatusBadRequest, gin.H{"error": "requisição inválida"})
	case errors.Is(err, auth.ErrInvalidCode):
		c.JSON(http.StatusUnauthorized, gin.H{"error": auth.ErrInvalidCode.Error()})
	case errors.Is(err, auth.ErrChallengeExpired):
		c.JSON(http.StatusGone, gin.H{"error": auth.ErrChallengeExpired.Error()})
	case errors.Is(err, auth.ErrChallengeUsed):
		c.JSON(http.StatusGone, gin.H{"error": auth.ErrChallengeUsed.Error()})
	case errors.Is(err, auth.ErrTooManyAttempts):
		c.JSON(http.StatusTooManyRequests, gin.H{"error": auth.ErrTooManyAttempts.Error()})
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": auth.ErrUnavailable.Error()})
	}
}

func decodeJSON(c *gin.Context, destination any) error {
	mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return errors.New("content type must be application/json")
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodyBytes)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request must contain one JSON object")
	}
	return nil
}
