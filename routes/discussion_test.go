package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"germinaStack/auth"
	"germinaStack/database"
	"germinaStack/handlers"
	"germinaStack/middleware"
	"germinaStack/model"

	"github.com/gin-gonic/gin"
)

func TestRegisterDiscussionRoutesRegistersOnlyDocumentedSurface(t *testing.T) {
	t.Parallel()
	router := gin.New()
	RegisterDiscussionRoutes(router.Group("/api"), handlers.NewDiscussionHandler(&routeDiscussionRepositoryFake{}, time.Second), middleware.APIAuthMiddleware("jwt-secret"))
	got := make(map[string]bool)
	for _, route := range router.Routes() {
		got[route.Method+" "+route.Path] = true
	}
	for _, want := range []string{
		"GET /api/posts", "POST /api/posts", "GET /api/posts/:id", "PATCH /api/posts/:id", "DELETE /api/posts/:id",
		"GET /api/posts/:id/comments", "POST /api/posts/:id/comments",
		"GET /api/comments/:id/replies", "POST /api/comments/:id/replies",
		"PATCH /api/comments/:id", "DELETE /api/comments/:id", "PATCH /api/replies/:id", "DELETE /api/replies/:id",
		"PUT /api/posts/:id/reaction", "PUT /api/comments/:id/reaction", "PUT /api/replies/:id/reaction",
		"GET /api/notifications", "PATCH /api/notifications/read-all",
	} {
		if !got[want] {
			t.Fatalf("route %q not registered; got %#v", want, got)
		}
	}
	if got["PATCH /api/notifications/:id/read"] || got["DELETE /api/posts/:id/reaction"] {
		t.Fatalf("obsolete route registered: %#v", got)
	}
}

func TestRegisterDiscussionRoutesRequiresValidJWT(t *testing.T) {
	t.Parallel()
	repository := &routeDiscussionRepositoryFake{}
	router := gin.New()
	RegisterDiscussionRoutes(router.Group("/api"), handlers.NewDiscussionHandler(repository, time.Second), middleware.APIAuthMiddleware("jwt-secret"))
	request := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized || repository.notificationCalls != 0 {
		t.Fatalf("missing jwt status/calls = %d/%d", recorder.Code, repository.notificationCalls)
	}

	now := time.Now().UTC()
	token, err := auth.GenerateTokenWithTimes("42", false, "jwt-secret", now.Add(-time.Minute), now.Add(time.Hour))
	if err != nil {
		t.Fatalf("GenerateTokenWithTimes() error = %v", err)
	}
	request = httptest.NewRequest(http.MethodGet, "/api/notifications", nil)
	request.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token})
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || repository.notificationUserID != 42 {
		t.Fatalf("jwt status/user = %d/%d", recorder.Code, repository.notificationUserID)
	}
}

type routeDiscussionRepositoryFake struct {
	notificationUserID int64
	notificationCalls  int
}

func (f *routeDiscussionRepositoryFake) GetPost(context.Context, int64) (model.Post, error) {
	return model.Post{}, nil
}
func (f *routeDiscussionRepositoryFake) CreatePost(context.Context, int64, database.PostInput) (model.Post, error) {
	return model.Post{}, nil
}
func (f *routeDiscussionRepositoryFake) UpdatePost(context.Context, int64, database.PostInput) (model.Post, error) {
	return model.Post{}, nil
}
func (f *routeDiscussionRepositoryFake) DeletePost(context.Context, int64) error { return nil }
func (f *routeDiscussionRepositoryFake) GetComment(context.Context, int64) (model.Comment, error) {
	return model.Comment{}, nil
}
func (f *routeDiscussionRepositoryFake) CreateComment(context.Context, int64, int64, database.CommentInput) (model.Comment, error) {
	return model.Comment{}, nil
}
func (f *routeDiscussionRepositoryFake) UpdateComment(context.Context, int64, database.CommentInput) (model.Comment, error) {
	return model.Comment{}, nil
}
func (f *routeDiscussionRepositoryFake) DeleteComment(context.Context, int64) error { return nil }
func (f *routeDiscussionRepositoryFake) GetReply(context.Context, int64) (model.CommentOnComment, error) {
	return model.CommentOnComment{}, nil
}
func (f *routeDiscussionRepositoryFake) CreateReply(context.Context, int64, int64, database.CommentInput) (model.CommentOnComment, error) {
	return model.CommentOnComment{}, nil
}
func (f *routeDiscussionRepositoryFake) UpdateReply(context.Context, int64, database.CommentInput) (model.CommentOnComment, error) {
	return model.CommentOnComment{}, nil
}
func (f *routeDiscussionRepositoryFake) DeleteReply(context.Context, int64) error { return nil }

func (f *routeDiscussionRepositoryFake) ListPosts(context.Context, database.PostFilter) ([]model.Post, error) {
	return []model.Post{}, nil
}
func (f *routeDiscussionRepositoryFake) ListComments(context.Context, int64, database.Pagination) ([]model.Comment, error) {
	return []model.Comment{}, nil
}
func (f *routeDiscussionRepositoryFake) ListReplies(context.Context, int64, database.Pagination) ([]model.CommentOnComment, error) {
	return []model.CommentOnComment{}, nil
}
func (f *routeDiscussionRepositoryFake) React(context.Context, int64, int64, string, model.ReactionType) error {
	return nil
}
func (f *routeDiscussionRepositoryFake) ListNotifications(_ context.Context, userID int64, _ database.NotificationFilter) ([]model.Notification, error) {
	f.notificationCalls++
	f.notificationUserID = userID
	return []model.Notification{}, nil
}
func (f *routeDiscussionRepositoryFake) MarkNotificationsRead(context.Context, int64) error {
	return nil
}
