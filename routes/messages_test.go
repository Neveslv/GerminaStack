package routes

import (
	"context"
	"testing"
	"time"

	"germinaStack/database"
	"germinaStack/handlers"
	"germinaStack/middleware"
	"germinaStack/model"

	"github.com/gin-gonic/gin"
)

func TestRegisterMessageRoutesRegistersOnlyImmutableMessageSurface(t *testing.T) {
	router := gin.New()
	RegisterMessageRoutes(router.Group("/api"), handlers.NewMessageHandler(&routeMessageRepositoryFake{}, time.Second), middleware.APIAuthMiddleware("secret"))
	want := map[string]bool{
		"GET /api/posts": true, "POST /api/posts": true, "GET /api/posts/:id": true,
		"GET /api/posts/:id/comments": true, "POST /api/posts/:id/comments": true,
		"GET /api/comments/:id/replies": true, "POST /api/comments/:id/replies": true,
	}
	got := make(map[string]bool)
	for _, route := range router.Routes() {
		got[route.Method+" "+route.Path] = true
	}
	for route := range want {
		if !got[route] {
			t.Fatalf("missing route %q; got %#v", route, got)
		}
	}
	if len(got) != len(want) {
		t.Fatalf("routes = %#v, want exactly immutable message surface", got)
	}
	for _, forbidden := range []string{"PATCH /api/posts/:id", "PUT /api/posts/:id", "DELETE /api/posts/:id", "DELETE /api/posts/:id/comments"} {
		if got[forbidden] {
			t.Fatalf("forbidden route registered: %s", forbidden)
		}
	}
}

type routeMessageRepositoryFake struct{}

func (routeMessageRepositoryFake) CreatePost(context.Context, int64, int64, string, *string, *string, string) (model.Post, error) {
	return model.Post{}, nil
}
func (routeMessageRepositoryFake) CreateComment(context.Context, int64, int64, string) (model.Comment, error) {
	return model.Comment{}, nil
}
func (routeMessageRepositoryFake) CreateReply(context.Context, int64, int64, string) (model.CommentOnComment, error) {
	return model.CommentOnComment{}, nil
}
func (routeMessageRepositoryFake) GetPost(context.Context, int64) (model.Post, error) {
	return model.Post{}, nil
}
func (routeMessageRepositoryFake) ListPosts(context.Context, *int64) ([]model.Post, error) {
	return []model.Post{}, nil
}
func (routeMessageRepositoryFake) GetComment(context.Context, int64) (model.Comment, error) {
	return model.Comment{}, nil
}
func (routeMessageRepositoryFake) ListComments(context.Context, int64) ([]model.Comment, error) {
	return []model.Comment{}, nil
}
func (routeMessageRepositoryFake) GetReply(context.Context, int64) (model.CommentOnComment, error) {
	return model.CommentOnComment{}, nil
}
func (routeMessageRepositoryFake) ListReplies(context.Context, int64) ([]model.CommentOnComment, error) {
	return []model.CommentOnComment{}, nil
}

var _ database.MessageRepository = routeMessageRepositoryFake{}
