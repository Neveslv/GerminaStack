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
	if response.Code != http.StatusOK || strings.TrimSpace(response.Body.String()) != `{"items":[],"has_more":false}` || repository.postFilter.SubjectID == nil || repository.postFilter.AuthorID == nil || *repository.postFilter.SubjectID != 3 || *repository.postFilter.AuthorID != 7 || repository.postFilter.Pagination != (database.Pagination{Page: 2, PageSize: 10}) {
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

func TestDiscussionHandlerReturnsNotFoundForMissingThreadParents(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		handle gin.HandlerFunc
		path   string
	}{
		{name: "post", handle: (&DiscussionHandler{repository: &discussionRepositoryFake{err: database.ErrDiscussionNotFound}, operationTimeout: time.Second}).ListComments, path: "/api/posts/4/comments"},
		{name: "comment", handle: (&DiscussionHandler{repository: &discussionRepositoryFake{err: database.ErrDiscussionNotFound}, operationTimeout: time.Second}).ListReplies, path: "/api/comments/5/replies"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			response := performDiscussionRequest(tt.handle, http.MethodGet, tt.path, "", 42)
			if response.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
			}
		})
	}
}

func TestDiscussionHandlerListsThreadReads(t *testing.T) {
	t.Parallel()
	repository := &discussionRepositoryFake{post: model.Post{ID: 4}, comment: model.Comment{ID: 5}}
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
	if valid.Code != http.StatusOK || repository.reactionUserID != 42 || repository.reactionMessageID != 6 || repository.reactionMessageType != "comment" || repository.reactionType != model.ReactionTypeLike {
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

func TestDiscussionHandlerSeparatesNotificationHistoryAndClearsRead(t *testing.T) {
	t.Parallel()
	repository := &discussionRepositoryFake{}
	handler := NewDiscussionHandler(repository, time.Second)
	history := performDiscussionRequest(handler.ListNotificationHistory, http.MethodGet, "/api/notifications/history?page=2", "", 42)
	clear := performDiscussionRequest(handler.HideReadNotifications, http.MethodPatch, "/api/notifications/clear-read", "", 42)
	if history.Code != http.StatusOK || clear.Code != http.StatusNoContent || repository.notificationUserID != 42 || !repository.notificationFilter.History || repository.notificationFilter.Pagination.Page != 2 || repository.hiddenUserID != 42 {
		t.Fatalf("history/clear = %d/%d %#v", history.Code, clear.Code, repository)
	}
}

func TestDiscussionHandlerCreatesMessagesAndProtectsMutations(t *testing.T) {
	t.Parallel()
	repository := &discussionRepositoryFake{
		post:    model.Post{ID: 8, UserID: 42, SubjectID: 3, Title: "Title", Content: "Body"},
		comment: model.Comment{ID: 9, PostID: 8, UserID: 42, Content: "Comment"},
		reply:   model.CommentOnComment{ID: 10, CommentID: 9, UserID: 42, Content: "Reply"},
	}
	handler := NewDiscussionHandler(repository, time.Second)
	post := performDiscussionRequest(handler.CreatePost, http.MethodPost, "/api/posts", `{"subject_id":3,"title":"Title","content":"Body"}`, 42)
	comment := performDiscussionRequest(handler.CreateComment, http.MethodPost, "/api/posts/8/comments", `{"content":"Comment"}`, 42)
	reply := performDiscussionRequest(handler.CreateReply, http.MethodPost, "/api/comments/9/replies", `{"content":"Reply"}`, 42)
	if post.Code != http.StatusCreated || comment.Code != http.StatusCreated || reply.Code != http.StatusCreated || repository.createPostUserID != 42 || repository.createCommentPostID != 8 || repository.createReplyCommentID != 9 {
		t.Fatalf("creation status/inputs = %d/%d/%d/%#v", post.Code, comment.Code, reply.Code, repository)
	}

	for _, tt := range []struct {
		name   string
		handle gin.HandlerFunc
		method string
		path   string
		body   string
		userID int64
		want   int
	}{
		{name: "foreign post update", handle: handler.UpdatePost, method: http.MethodPatch, path: "/api/posts/8", body: `{"content":"changed"}`, userID: 7, want: http.StatusForbidden},
		{name: "unknown post field", handle: handler.CreatePost, method: http.MethodPost, path: "/api/posts", body: `{"subject_id":3,"title":"Title","content":"Body","user_id":7}`, userID: 42, want: http.StatusBadRequest},
		{name: "unpaired image", handle: handler.CreatePost, method: http.MethodPost, path: "/api/posts", body: `{"subject_id":3,"title":"Title","content":"Body","image_url":"https://image"}`, userID: 42, want: http.StatusBadRequest},
	} {
		t.Run(tt.name, func(t *testing.T) {
			response := performDiscussionRequest(tt.handle, tt.method, tt.path, tt.body, tt.userID)
			if response.Code != tt.want {
				t.Fatalf("status = %d, want %d; body=%s", response.Code, tt.want, response.Body.String())
			}
		})
	}
}

func TestDiscussionHandlerAllowsAdminToMutateAnotherUsersPost(t *testing.T) {
	t.Parallel()
	repository := &discussionRepositoryFake{post: model.Post{ID: 8, UserID: 7, SubjectID: 3, Title: "Title", Content: "Body"}}
	handler := NewDiscussionHandler(repository, time.Second)
	router := gin.New()
	router.PATCH("/api/posts/:id", func(c *gin.Context) {
		c.Set(auth.ContextUserID, int64(42))
		c.Set(auth.ContextIsAdmin, true)
		handler.UpdatePost(c)
	})
	request := httptest.NewRequest(http.MethodPatch, "/api/posts/8", strings.NewReader(`{"content":"changed"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("admin status = %d; body=%s", recorder.Code, recorder.Body.String())
	}
}

type discussionRepositoryFake struct {
	post                 model.Post
	comment              model.Comment
	reply                model.CommentOnComment
	createPostUserID     int64
	createCommentPostID  int64
	createReplyCommentID int64
	postFilter           database.PostFilter
	commentPostID        int64
	commentPagination    database.Pagination
	replyCommentID       int64
	reactionUserID       int64
	reactionMessageID    int64
	reactionMessageType  string
	reactionType         model.ReactionType
	notificationUserID   int64
	notificationFilter   database.NotificationFilter
	markedUserID         int64
	hiddenUserID         int64
	listPostsCalls       int
	err                  error
}

func (f *discussionRepositoryFake) GetPost(context.Context, int64) (model.Post, error) {
	return f.post, f.err
}
func (f *discussionRepositoryFake) CreatePost(_ context.Context, userID int64, _ database.PostInput) (model.Post, error) {
	f.createPostUserID = userID
	return f.post, f.err
}
func (f *discussionRepositoryFake) UpdatePost(context.Context, int64, database.PostInput) (model.Post, error) {
	return model.Post{}, f.err
}
func (f *discussionRepositoryFake) DeletePost(context.Context, int64) error { return f.err }
func (f *discussionRepositoryFake) GetComment(context.Context, int64) (model.Comment, error) {
	return f.comment, f.err
}
func (f *discussionRepositoryFake) CreateComment(_ context.Context, _ int64, postID int64, _ database.CommentInput) (model.Comment, error) {
	f.createCommentPostID = postID
	return f.comment, f.err
}
func (f *discussionRepositoryFake) UpdateComment(context.Context, int64, database.CommentInput) (model.Comment, error) {
	return model.Comment{}, f.err
}
func (f *discussionRepositoryFake) DeleteComment(context.Context, int64) error { return f.err }
func (f *discussionRepositoryFake) GetReply(context.Context, int64) (model.CommentOnComment, error) {
	return f.reply, f.err
}
func (f *discussionRepositoryFake) CreateReply(_ context.Context, _ int64, commentID int64, _ database.CommentInput) (model.CommentOnComment, error) {
	f.createReplyCommentID = commentID
	return f.reply, f.err
}
func (f *discussionRepositoryFake) UpdateReply(context.Context, int64, database.CommentInput) (model.CommentOnComment, error) {
	return model.CommentOnComment{}, f.err
}
func (f *discussionRepositoryFake) DeleteReply(context.Context, int64) error { return f.err }

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
func (f *discussionRepositoryFake) HideReadNotifications(_ context.Context, userID int64) error {
	f.hiddenUserID = userID
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
