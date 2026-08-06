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
	"germinaStack/model"

	"github.com/gin-gonic/gin"
)

func TestReactionHandlerRequiresAuthValidatesTargetAndReturnsClientShape(t *testing.T) {
	repository := &reactionHandlerRepositoryFake{result: ReactionResult{Reaction: reactionTypePointer(model.ReactionTypeLike), Likes: 3, Dislikes: 1}}
	handler := NewReactionHandler(repository, time.Second)
	request := performReactionNotificationRequest(handler.ToggleReaction, http.MethodPost, "/api/reactions", `{"message_type":"post","id_message":7,"reaction_type":"like"}`, true)
	if request.Code != http.StatusOK || repository.userID != 42 || repository.messageID != 7 {
		t.Fatalf("status/user/message = %d/%d/%d", request.Code, repository.userID, repository.messageID)
	}
	var body map[string]any
	if err := json.Unmarshal(request.Body.Bytes(), &body); err != nil || body["reacao"] != "like" || body["likes"] != float64(3) {
		t.Fatalf("reaction response = %s", request.Body.String())
	}
	if got := performReactionNotificationRequest(handler.ToggleReaction, http.MethodPost, "/api/reactions", `{"message_type":"invalid","id_message":7,"reaction_type":"like"}`, true).Code; got != http.StatusBadRequest {
		t.Fatalf("invalid reaction status = %d", got)
	}
	if got := performReactionNotificationRequest(handler.ToggleReaction, http.MethodPost, "/api/reactions", `{"message_type":"post","id_message":7,"reaction_type":"like"}`, false).Code; got != http.StatusUnauthorized {
		t.Fatalf("unauthenticated reaction status = %d", got)
	}
}

func TestNotificationHandlerListsAndMarksOnlyAuthenticatedUser(t *testing.T) {
	repository := &notificationHandlerRepositoryFake{items: []model.Notification{{ID: 1, UserID: 42}}}
	handler := NewNotificationHandler(repository, time.Second)
	if got := performReactionNotificationRequest(handler.ListNotifications, http.MethodGet, "/api/notifications", "", false).Code; got != http.StatusUnauthorized {
		t.Fatalf("unauthenticated notification status = %d", got)
	}
	if got := performReactionNotificationRequest(handler.ListNotifications, http.MethodGet, "/api/notifications", "", true).Code; got != http.StatusOK || repository.listUserID != 42 {
		t.Fatalf("list status/user = %d/%d", got, repository.listUserID)
	}
	if got := performReactionNotificationRequest(handler.MarkNotificationsAsRead, http.MethodPost, "/api/notifications/read", "", true).Code; got != http.StatusNoContent || repository.markUserID != 42 {
		t.Fatalf("mark status/user = %d/%d", got, repository.markUserID)
	}
}

type reactionHandlerRepositoryFake struct {
	result    ReactionResult
	userID    int64
	messageID int64
}

func (f *reactionHandlerRepositoryFake) ToggleReaction(_ context.Context, userID int64, _ string, messageID int64, _ model.ReactionType) (ReactionResult, error) {
	f.userID = userID
	f.messageID = messageID
	return f.result, nil
}
func (*reactionHandlerRepositoryFake) GetReaction(context.Context, int64, string, int64) (*model.ReactionType, error) {
	return nil, nil
}

type notificationHandlerRepositoryFake struct {
	items      []model.Notification
	listUserID int64
	markUserID int64
}

func (f *notificationHandlerRepositoryFake) ListNotifications(_ context.Context, userID int64) ([]model.Notification, error) {
	f.listUserID = userID
	return f.items, nil
}
func (f *notificationHandlerRepositoryFake) MarkNotificationsAsRead(_ context.Context, userID int64) error {
	f.markUserID = userID
	return nil
}

func reactionTypePointer(value model.ReactionType) *model.ReactionType {
	return &value
}

func performReactionNotificationRequest(handler gin.HandlerFunc, method, path, body string, authenticated bool) *httptest.ResponseRecorder {
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
