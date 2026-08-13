package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"germinaStack/auth"
	"germinaStack/model"

	"github.com/gin-gonic/gin"
)

func TestMessageHandlerCreatesPostUsingAuthenticatedUser(t *testing.T) {
	repository := &messageRepositoryFake{}
	recorder := performMessageHandlerRequest(NewMessageHandler(repository, time.Second).CreatePost, http.MethodPost, "/api/posts", `{"id_subject":7,"title":"Title","content":"Body"}`, 42)
	if recorder.Code != http.StatusCreated || repository.createdUserID != 42 || repository.createdSubjectID != 7 {
		t.Fatalf("status/user/subject = %d/%d/%d", recorder.Code, repository.createdUserID, repository.createdSubjectID)
	}
}

func TestMessageHandlerRejectsUnauthenticatedOrInvalidPost(t *testing.T) {
	repository := &messageRepositoryFake{}
	unauthenticated := performMessageHandlerRequest(NewMessageHandler(repository, time.Second).CreatePost, http.MethodPost, "/api/posts", `{"id_subject":7,"title":"Title","content":"Body"}`, 0)
	invalid := performMessageHandlerRequest(NewMessageHandler(repository, time.Second).CreatePost, http.MethodPost, "/api/posts", `{"id_subject":0,"title":"","content":""}`, 42)
	if unauthenticated.Code != http.StatusUnauthorized || invalid.Code != http.StatusBadRequest || repository.createPostCalls != 0 {
		t.Fatalf("statuses/calls = %d/%d/%d", unauthenticated.Code, invalid.Code, repository.createPostCalls)
	}
}

func TestMessageHandlerCreatesCommentAndReplyFromPathParents(t *testing.T) {
	repository := &messageRepositoryFake{}
	comment := performMessageHandlerRequest(NewMessageHandler(repository, time.Second).CreateComment, http.MethodPost, "/api/posts/7/comments", `{"content":"Comment"}`, 42)
	reply := performMessageHandlerRequest(NewMessageHandler(repository, time.Second).CreateReply, http.MethodPost, "/api/comments/9/replies", `{"content":"Reply"}`, 42)
	if comment.Code != http.StatusCreated || reply.Code != http.StatusCreated || repository.createdPostID != 7 || repository.createdCommentID != 9 {
		t.Fatalf("statuses/parents = %d/%d/%d/%d", comment.Code, reply.Code, repository.createdPostID, repository.createdCommentID)
	}
}

type messageRepositoryFake struct {
	createdUserID    int64
	createdSubjectID int64
	createdPostID    int64
	createdCommentID int64
	createPostCalls  int
}

func (f *messageRepositoryFake) CreatePost(_ context.Context, userID, subjectID int64, _ string, _, _ *string, _ string) (model.Post, error) {
	f.createPostCalls++
	f.createdUserID, f.createdSubjectID = userID, subjectID
	return model.Post{ID: 1, UserID: userID, SubjectID: subjectID, Title: "Title", Content: "Body"}, nil
}
func (f *messageRepositoryFake) CreateComment(_ context.Context, userID, postID int64, _ string) (model.Comment, error) {
	f.createdUserID, f.createdPostID = userID, postID
	return model.Comment{ID: 2, UserID: userID, PostID: postID, Content: "Comment"}, nil
}
func (f *messageRepositoryFake) CreateReply(_ context.Context, userID, commentID int64, _ string) (model.CommentOnComment, error) {
	f.createdUserID, f.createdCommentID = userID, commentID
	return model.CommentOnComment{ID: 3, UserID: userID, CommentID: commentID, Content: "Reply"}, nil
}
func (*messageRepositoryFake) GetPost(context.Context, int64) (model.Post, error) {
	return model.Post{}, nil
}
func (*messageRepositoryFake) ListPosts(context.Context, *int64) ([]model.Post, error) {
	return []model.Post{}, nil
}
func (*messageRepositoryFake) GetComment(context.Context, int64) (model.Comment, error) {
	return model.Comment{}, nil
}
func (*messageRepositoryFake) ListComments(context.Context, int64) ([]model.Comment, error) {
	return []model.Comment{}, nil
}
func (*messageRepositoryFake) GetReply(context.Context, int64) (model.CommentOnComment, error) {
	return model.CommentOnComment{}, nil
}
func (*messageRepositoryFake) ListReplies(context.Context, int64) ([]model.CommentOnComment, error) {
	return []model.CommentOnComment{}, nil
}

func performMessageHandlerRequest(handler gin.HandlerFunc, method, path, body string, userID int64) *httptest.ResponseRecorder {
	router := gin.New()
	router.Handle(method, path, func(c *gin.Context) {
		if userID > 0 {
			c.Set(auth.ContextUserID, userID)
		}
		if strings.Contains(path, "/posts/") {
			c.Params = gin.Params{{Key: "id", Value: "7"}}
		} else if strings.Contains(path, "/comments/") {
			c.Params = gin.Params{{Key: "id", Value: "9"}}
		}
		handler(c)
	})
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}
