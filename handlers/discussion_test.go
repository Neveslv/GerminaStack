package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"germinaStack/auth"
	"germinaStack/database"
	"germinaStack/model"

	"github.com/gin-gonic/gin"
)

func TestDiscussionHandlerListsPostsValidatesFiltersAndReturnsEmptyList(t *testing.T) {
	t.Parallel()
	repository := &discussionRepositoryFake{}
	handler := NewDiscussionHandler(repository, time.Second)
	response := performDiscussionRequest(handler.ListPosts, http.MethodGet, "/api/posts?subject_id=3&author_id=7&page=2&page_size=10", "", 0)
	if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != "[]" || repository.postFilter.SubjectID == nil || repository.postFilter.AuthorID == nil || *repository.postFilter.SubjectID != 3 || *repository.postFilter.AuthorID != 7 || repository.postFilter.Pagination != (database.Pagination{Page: 2, PageSize: 10}) {
		t.Fatalf("status/body/filter = %d/%s/%#v", response.Code, response.Body.String(), repository.postFilter)
	}

	for _, query := range []string{"?subject_id=0", "?author_id=one", "?page=", "?page=1&page=2", "?unknown=1"} {
		invalid := &discussionRepositoryFake{}
		response := performDiscussionRequest(NewDiscussionHandler(invalid, time.Second).ListPosts, http.MethodGet, "/api/posts"+query, "", 0)
		if response.Code != http.StatusBadRequest || invalid.listPostsCalls != 0 {
			t.Fatalf("query %q status/calls = %d/%d", query, response.Code, invalid.listPostsCalls)
		}
	}
}

func TestDiscussionHandlerListsThreadReads(t *testing.T) {
	t.Parallel()
	repository := &discussionRepositoryFake{}
	handler := NewDiscussionHandler(repository, time.Second)
	comments := performDiscussionRequest(handler.ListComments, http.MethodGet, "/api/posts/4/comments?page=2", "", 0)
	replies := performDiscussionRequest(handler.ListReplies, http.MethodGet, "/api/comments/5/replies", "", 0)
	if comments.Code != http.StatusOK || replies.Code != http.StatusOK || repository.commentPostID != 4 || repository.replyCommentID != 5 || repository.commentPagination.Page != 2 || strings.TrimSpace(comments.Body.String()) != "[]" || strings.TrimSpace(replies.Body.String()) != "[]" {
		t.Fatalf("reads = %d/%d; repository=%#v", comments.Code, replies.Code, repository)
	}
}

func TestDiscussionHandlerReactsOnlyForAuthenticatedUserWithValidPayload(t *testing.T) {
	t.Parallel()
	repository := &discussionRepositoryFake{}
	handler := NewDiscussionHandler(repository, time.Second)
	valid := performDiscussionRequest(handler.React("comment"), http.MethodPut, "/api/comments/6/reaction", `{"reaction_type":"like"}`, 42)
	if valid.Code != http.StatusNoContent || repository.reactionUserID != 42 || repository.reactionMessageID != 6 || repository.reactionMessageType != "comment" || repository.reactionType != model.ReactionTypeLike {
		t.Fatalf("reaction = %d/%#v", valid.Code, repository)
	}
	for _, tt := range []struct {
		name string
		body string
		user int64
		want int
	}{
		{name: "missing user", body: `{"reaction_type":"like"}`, want: http.StatusUnauthorized},
		{name: "invalid type", body: `{"reaction_type":"heart"}`, user: 42, want: http.StatusBadRequest},
		{name: "invalid json", body: `{"reaction_type":"like","extra":true}`, user: 42, want: http.StatusBadRequest},
	} {
		t.Run(tt.name, func(t *testing.T) {
			response := performDiscussionRequest(handler.React("post"), http.MethodPut, "/api/posts/3/reaction", tt.body, tt.user)
			if response.Code != tt.want {
				t.Fatalf("status = %d, want %d", response.Code, tt.want)
			}
		})
	}
}

func TestDiscussionHandlerScopesNotificationsToJWTUser(t *testing.T) {
	t.Parallel()
	repository := &discussionRepositoryFake{}
	handler := NewDiscussionHandler(repository, time.Second)
	list := performDiscussionRequest(handler.ListNotifications, http.MethodGet, "/api/notifications?unread=true", "", 42)
	mark := performDiscussionRequest(handler.MarkNotificationsRead, http.MethodPatch, "/api/notifications/read-all", "", 42)
	if list.Code != http.StatusOK || mark.Code != http.StatusNoContent || repository.notificationUserID != 42 || repository.markedUserID != 42 || !repository.notificationFilter.Unread {
		t.Fatalf("notifications = %d/%d/%#v", list.Code, mark.Code, repository)
	}

	repository.err = database.ErrDiscussionNotFound
	response := performDiscussionRequest(handler.ListNotifications, http.MethodGet, "/api/notifications", "", 42)
	if response.Code != http.StatusNotFound || strings.Contains(response.Body.String(), "private") {
		t.Fatalf("error response = %d/%s", response.Code, response.Body.String())
	}

	invalid := performDiscussionRequest(handler.ListNotifications, http.MethodGet, "/api/notifications?unread=yes", "", 42)
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid unread status = %d", invalid.Code)
	}
}

type discussionRepositoryFake struct {
	postFilter          database.PostFilter
	commentPostID       int64
	commentPagination   database.Pagination
	replyCommentID      int64
	reactionUserID      int64
	reactionMessageID   int64
	reactionMessageType string
	reactionType        model.ReactionType
	notificationUserID  int64
	notificationFilter  database.NotificationFilter
	markedUserID        int64
	listPostsCalls      int
	err                 error
}

func (f *discussionRepositoryFake) ListPosts(_ context.Context, filter database.PostFilter) ([]model.Post, error) {
	f.listPostsCalls++
	f.postFilter = filter
	return nil, f.err
}
func (f *discussionRepositoryFake) ListComments(_ context.Context, postID int64, pagination database.Pagination) ([]model.Comment, error) {
	f.commentPostID, f.commentPagination = postID, pagination
	return nil, f.err
}
func (f *discussionRepositoryFake) ListReplies(_ context.Context, commentID int64, _ database.Pagination) ([]model.CommentOnComment, error) {
	f.replyCommentID = commentID
	return nil, f.err
}
func (f *discussionRepositoryFake) React(_ context.Context, userID, messageID int64, messageType string, reactionType model.ReactionType) error {
	f.reactionUserID, f.reactionMessageID, f.reactionMessageType, f.reactionType = userID, messageID, messageType, reactionType
	return f.err
}
func (f *discussionRepositoryFake) ListNotifications(_ context.Context, userID int64, filter database.NotificationFilter) ([]model.Notification, error) {
	f.notificationUserID, f.notificationFilter = userID, filter
	return nil, f.err
}
func (f *discussionRepositoryFake) MarkNotificationsRead(_ context.Context, userID int64) error {
	f.markedUserID = userID
	return f.err
}

func performDiscussionRequest(handler gin.HandlerFunc, method, path, body string, userID int64) *httptest.ResponseRecorder {
	router := gin.New()
	router.Handle(method, "/*path", func(c *gin.Context) {
		parts := strings.Split(strings.Trim(c.Request.URL.Path, "/"), "/")
		for index, part := range parts[:len(parts)-1] {
			if (part == "posts" || part == "comments" || part == "replies") && index+1 < len(parts) {
				if _, err := strconv.ParseInt(parts[index+1], 10, 64); err == nil {
					c.Params = append(c.Params, gin.Param{Key: "id", Value: parts[index+1]})
				}
				break
			}
		}
		if userID > 0 {
			c.Set(auth.ContextUserID, userID)
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
