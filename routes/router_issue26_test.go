package routes

import (
	"context"
	"testing"
	"time"

	"germinaStack/database"
	"germinaStack/handlers"
	"germinaStack/middleware"
	"germinaStack/model"
)

func TestNewRouterRegistersEveryIssue26APIRouteExactlyOnce(t *testing.T) {
	authHandler := handlers.NewAuthHandler(routerAuthServiceFake{}, "jwt-test-secret", true, 24*time.Hour, 2*time.Second)
	userHandler := handlers.NewUserHandler(routerUserRepositoryFake{}, 2*time.Second)
	router := NewRouter(authHandler, userHandler, RouterDependencies{
		Catalog:       handlers.NewCatalogHandler(routerIssue26CatalogRepositoryFake{}, 2*time.Second),
		Messages:      handlers.NewMessageHandler(routerIssue26MessageRepositoryFake{}, 2*time.Second),
		Reactions:     handlers.NewReactionHandler(routerIssue26ReactionRepositoryFake{}, 2*time.Second),
		Notifications: handlers.NewNotificationHandler(routerIssue26NotificationRepositoryFake{}, 2*time.Second),
		Authenticated: middleware.APIAuthMiddleware("jwt-test-secret"),
	})

	got := make(map[string]int)
	for _, route := range router.Routes() {
		got[route.Method+" "+route.Path]++
	}
	want := []string{
		"POST /api/users", "POST /api/login", "POST /api/login/2fa", "POST /api/logout",
		"GET /api/years", "GET /api/subjects",
		"GET /api/posts", "POST /api/posts", "GET /api/posts/:id",
		"GET /api/posts/:id/comments", "POST /api/posts/:id/comments",
		"GET /api/comments/:id/replies", "POST /api/comments/:id/replies",
		"GET /api/users/me", "PUT /api/users/me/preferences",
		"GET /api/preferences", "PUT /api/preferences",
		"POST /api/reactions", "GET /api/reactions",
		"GET /api/notifications", "POST /api/notifications/read",
	}
	for _, route := range want {
		if got[route] != 1 {
			t.Fatalf("route %q count = %d; all routes = %#v", route, got[route], got)
		}
	}
	if len(got) != len(want) {
		t.Fatalf("registered routes = %#v, want exactly %d routes", got, len(want))
	}
	for _, forbidden := range []string{
		"POST /api/admin/years", "PATCH /api/admin/years/:id", "DELETE /api/admin/years/:id",
		"POST /api/admin/subjects", "PATCH /api/admin/subjects/:id", "DELETE /api/admin/subjects/:id",
		"PATCH /api/posts/:id", "PUT /api/posts/:id", "DELETE /api/posts/:id",
		"GET /api/notifications/:id",
	} {
		if got[forbidden] != 0 {
			t.Fatalf("forbidden route registered: %s", forbidden)
		}
	}
}

type routerIssue26CatalogRepositoryFake struct{}

func (routerIssue26CatalogRepositoryFake) ListYears(context.Context) ([]model.Year, error) {
	return nil, nil
}
func (routerIssue26CatalogRepositoryFake) ListSubjects(context.Context, int64) ([]model.Subject, error) {
	return nil, nil
}

type routerIssue26MessageRepositoryFake struct{}

func (routerIssue26MessageRepositoryFake) CreatePost(context.Context, int64, int64, string, *string, *string, string) (model.Post, error) {
	return model.Post{}, nil
}
func (routerIssue26MessageRepositoryFake) CreateComment(context.Context, int64, int64, string) (model.Comment, error) {
	return model.Comment{}, nil
}
func (routerIssue26MessageRepositoryFake) CreateReply(context.Context, int64, int64, string) (model.CommentOnComment, error) {
	return model.CommentOnComment{}, nil
}
func (routerIssue26MessageRepositoryFake) GetPost(context.Context, int64) (model.Post, error) {
	return model.Post{}, nil
}
func (routerIssue26MessageRepositoryFake) ListPosts(context.Context, *int64) ([]model.Post, error) {
	return nil, nil
}
func (routerIssue26MessageRepositoryFake) GetComment(context.Context, int64) (model.Comment, error) {
	return model.Comment{}, nil
}
func (routerIssue26MessageRepositoryFake) ListComments(context.Context, int64) ([]model.Comment, error) {
	return nil, nil
}
func (routerIssue26MessageRepositoryFake) GetReply(context.Context, int64) (model.CommentOnComment, error) {
	return model.CommentOnComment{}, nil
}
func (routerIssue26MessageRepositoryFake) ListReplies(context.Context, int64) ([]model.CommentOnComment, error) {
	return nil, nil
}

type routerIssue26ReactionRepositoryFake struct{}

func (routerIssue26ReactionRepositoryFake) ToggleReaction(context.Context, int64, string, int64, model.ReactionType) (handlers.ReactionResult, error) {
	return handlers.ReactionResult{}, nil
}
func (routerIssue26ReactionRepositoryFake) GetReaction(context.Context, int64, string, int64) (*model.ReactionType, error) {
	return nil, nil
}

type routerIssue26NotificationRepositoryFake struct{}

func (routerIssue26NotificationRepositoryFake) ListNotifications(context.Context, int64) ([]model.Notification, error) {
	return nil, nil
}
func (routerIssue26NotificationRepositoryFake) MarkNotificationsAsRead(context.Context, int64) error {
	return nil
}

var _ database.MessageRepository = routerIssue26MessageRepositoryFake{}
