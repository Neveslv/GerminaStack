package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"germinaStack/auth"
	"germinaStack/database"
	"germinaStack/model"

	"github.com/gin-gonic/gin"
)

type DiscussionRepository interface {
	ListPosts(context.Context, database.PostFilter) ([]model.Post, error)
	ListComments(context.Context, int64, database.Pagination) ([]model.Comment, error)
	ListReplies(context.Context, int64, database.Pagination) ([]model.CommentOnComment, error)
	React(context.Context, int64, int64, string, model.ReactionType) error
	ListNotifications(context.Context, int64, database.NotificationFilter) ([]model.Notification, error)
	MarkNotificationsRead(context.Context, int64) error
}

type DiscussionHandler struct {
	repository       DiscussionRepository
	operationTimeout time.Duration
}

func NewDiscussionHandler(repository DiscussionRepository, operationTimeout time.Duration) *DiscussionHandler {
	return &DiscussionHandler{repository: repository, operationTimeout: operationTimeout}
}

type reactionRequest struct {
	ReactionType model.ReactionType `json:"reaction_type"`
}

func (h *DiscussionHandler) ListPosts(c *gin.Context) {
	query, pagination, ok := discussionQuery(c, "subject_id", "author_id")
	if !ok {
		return
	}
	filter := database.PostFilter{Pagination: pagination}
	if filter.SubjectID, ok = discussionQueryID(query, "subject_id"); !ok {
		writeInvalidDiscussionRequest(c)
		return
	}
	if filter.AuthorID, ok = discussionQueryID(query, "author_id"); !ok {
		writeInvalidDiscussionRequest(c)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	posts, err := h.repository.ListPosts(ctx, filter)
	if err != nil {
		writeDiscussionError(c, err)
		return
	}
	if posts == nil {
		posts = []model.Post{}
	}
	c.JSON(http.StatusOK, posts)
}

func (h *DiscussionHandler) ListComments(c *gin.Context) {
	postID, ok := positivePathID(c)
	if !ok {
		return
	}
	_, pagination, ok := discussionQuery(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	comments, err := h.repository.ListComments(ctx, postID, pagination)
	if err != nil {
		writeDiscussionError(c, err)
		return
	}
	if comments == nil {
		comments = []model.Comment{}
	}
	c.JSON(http.StatusOK, comments)
}

func (h *DiscussionHandler) ListReplies(c *gin.Context) {
	commentID, ok := positivePathID(c)
	if !ok {
		return
	}
	_, pagination, ok := discussionQuery(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	replies, err := h.repository.ListReplies(ctx, commentID, pagination)
	if err != nil {
		writeDiscussionError(c, err)
		return
	}
	if replies == nil {
		replies = []model.CommentOnComment{}
	}
	c.JSON(http.StatusOK, replies)
}

func (h *DiscussionHandler) React(messageType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		messageID, ok := positivePathID(c)
		if !ok {
			return
		}
		userID, ok := discussionUserID(c)
		if !ok {
			return
		}
		var request reactionRequest
		if err := decodeJSON(c, &request); err != nil || !request.ReactionType.IsValid() {
			writeInvalidDiscussionRequest(c)
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
		defer cancel()
		if err := h.repository.React(ctx, userID, messageID, messageType, request.ReactionType); err != nil {
			writeDiscussionError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
		c.Writer.WriteHeaderNow()
	}
}

func (h *DiscussionHandler) ListNotifications(c *gin.Context) {
	userID, ok := discussionUserID(c)
	if !ok {
		return
	}
	query, pagination, ok := discussionQuery(c, "unread")
	if !ok {
		return
	}
	unread, ok := discussionQueryBool(query, "unread")
	if !ok {
		writeInvalidDiscussionRequest(c)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	notifications, err := h.repository.ListNotifications(ctx, userID, database.NotificationFilter{Unread: unread, Pagination: pagination})
	if err != nil {
		writeDiscussionError(c, err)
		return
	}
	if notifications == nil {
		notifications = []model.Notification{}
	}
	c.JSON(http.StatusOK, notifications)
}

func (h *DiscussionHandler) MarkNotificationsRead(c *gin.Context) {
	userID, ok := discussionUserID(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	if err := h.repository.MarkNotificationsRead(ctx, userID); err != nil {
		writeDiscussionError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
	c.Writer.WriteHeaderNow()
}

func discussionQuery(c *gin.Context, allowed ...string) (url.Values, database.Pagination, bool) {
	query, err := url.ParseQuery(c.Request.URL.RawQuery)
	if err != nil {
		writeInvalidDiscussionRequest(c)
		return nil, database.Pagination{}, false
	}
	for key := range query {
		if key == "page" || key == "page_size" {
			continue
		}
		valid := false
		for _, allowedKey := range allowed {
			if key == allowedKey {
				valid = true
				break
			}
		}
		if !valid || len(query[key]) != 1 || query[key][0] == "" {
			writeInvalidDiscussionRequest(c)
			return nil, database.Pagination{}, false
		}
	}
	for _, key := range []string{"page", "page_size"} {
		if values, supplied := query[key]; supplied && (len(values) != 1 || values[0] == "") {
			writeInvalidDiscussionRequest(c)
			return nil, database.Pagination{}, false
		}
	}
	pagination, err := database.ParsePagination(query.Get("page"), query.Get("page_size"))
	if err != nil {
		writeInvalidDiscussionRequest(c)
		return nil, database.Pagination{}, false
	}
	return query, pagination, true
}

func discussionQueryID(query url.Values, key string) (*int64, bool) {
	value, supplied := query[key]
	if !supplied {
		return nil, true
	}
	id, err := strconv.ParseInt(value[0], 10, 64)
	if err != nil || id <= 0 {
		return nil, false
	}
	return &id, true
}

func discussionQueryBool(query url.Values, key string) (bool, bool) {
	value, supplied := query[key]
	if !supplied {
		return false, true
	}
	parsed, err := strconv.ParseBool(value[0])
	return parsed, err == nil
}

func discussionUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get(auth.ContextUserID)
	id, ok := userID.(int64)
	if !exists || !ok || id <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autorizado"})
		return 0, false
	}
	return id, true
}

func writeDiscussionError(c *gin.Context, err error) {
	if errors.Is(err, database.ErrDiscussionNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "recurso não encontrado"})
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": "serviço indisponível"})
}

func writeInvalidDiscussionRequest(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"error": "requisição inválida"})
}
