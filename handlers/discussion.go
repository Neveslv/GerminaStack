package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"germinaStack/auth"
	"germinaStack/domain/discussion"
	"germinaStack/domain/pagination"
	"germinaStack/frok"
	"germinaStack/model"

	"github.com/gin-gonic/gin"
)

type DiscussionHandler struct {
	repository       discussion.Repository
	operationTimeout time.Duration
	frok             *frok.Service
}

func NewDiscussionHandler(repository discussion.Repository, operationTimeout time.Duration, assistants ...*frok.Service) *DiscussionHandler {
	handler := &DiscussionHandler{repository: repository, operationTimeout: operationTimeout}
	if len(assistants) > 0 {
		handler.frok = assistants[0]
	}
	return handler
}

func (h *DiscussionHandler) GetPost(c *gin.Context) {
	postID, ok := positivePathID(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	post, err := h.repository.GetPost(ctx, postID)
	if err != nil {
		writeDiscussionError(c, err)
		return
	}
	c.JSON(http.StatusOK, post)
}

func (h *DiscussionHandler) CreatePost(c *gin.Context) {
	userID, ok := discussionUserID(c)
	if !ok {
		return
	}
	input, ok := decodePostInput(c, nil, true)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	post, err := h.repository.CreatePost(ctx, userID, input)
	if err != nil {
		writeDiscussionError(c, err)
		return
	}
	if h.frok != nil {
		h.frok.DispatchPost(userID, post)
	}
	c.JSON(http.StatusCreated, post)
}

func (h *DiscussionHandler) UpdatePost(c *gin.Context) {
	postID, ok := positivePathID(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	post, err := h.repository.GetPost(ctx, postID)
	if err != nil {
		writeDiscussionError(c, err)
		return
	}
	if !canMutateDiscussion(c, post.UserID) {
		return
	}
	input, ok := decodePostInput(c, &post, false)
	if !ok {
		return
	}
	updated, err := h.repository.UpdatePost(ctx, postID, input)
	if err != nil {
		writeDiscussionError(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *DiscussionHandler) DeletePost(c *gin.Context) {
	postID, ok := positivePathID(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	post, err := h.repository.GetPost(ctx, postID)
	if err != nil {
		writeDiscussionError(c, err)
		return
	}
	if !canMutateDiscussion(c, post.UserID) {
		return
	}
	if err := h.repository.DeletePost(ctx, postID); err != nil {
		writeDiscussionError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
	c.Writer.WriteHeaderNow()
}

func (h *DiscussionHandler) CreateComment(c *gin.Context) {
	postID, ok := positivePathID(c)
	if !ok {
		return
	}
	userID, ok := discussionUserID(c)
	if !ok {
		return
	}
	input, ok := decodeCommentInput(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	comment, err := h.repository.CreateComment(ctx, userID, postID, input)
	if err != nil {
		writeDiscussionError(c, err)
		return
	}
	if h.frok != nil {
		h.frok.DispatchComment(userID, comment)
	}
	c.JSON(http.StatusCreated, comment)
}

func (h *DiscussionHandler) UpdateComment(c *gin.Context) {
	commentID, ok := positivePathID(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	comment, err := h.repository.GetComment(ctx, commentID)
	if err != nil {
		writeDiscussionError(c, err)
		return
	}
	if !canMutateDiscussion(c, comment.UserID) {
		return
	}
	input, ok := decodeCommentInput(c)
	if !ok {
		return
	}
	updated, err := h.repository.UpdateComment(ctx, commentID, input)
	if err != nil {
		writeDiscussionError(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *DiscussionHandler) DeleteComment(c *gin.Context) {
	commentID, ok := positivePathID(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	comment, err := h.repository.GetComment(ctx, commentID)
	if err != nil {
		writeDiscussionError(c, err)
		return
	}
	if !canMutateDiscussion(c, comment.UserID) {
		return
	}
	if err := h.repository.DeleteComment(ctx, commentID); err != nil {
		writeDiscussionError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
	c.Writer.WriteHeaderNow()
}

func (h *DiscussionHandler) CreateReply(c *gin.Context) {
	commentID, ok := positivePathID(c)
	if !ok {
		return
	}
	userID, ok := discussionUserID(c)
	if !ok {
		return
	}
	input, ok := decodeCommentInput(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	reply, err := h.repository.CreateReply(ctx, userID, commentID, input)
	if err != nil {
		writeDiscussionError(c, err)
		return
	}
	if h.frok != nil {
		h.frok.DispatchReply(userID, reply)
	}
	c.JSON(http.StatusCreated, reply)
}

func (h *DiscussionHandler) UpdateReply(c *gin.Context) {
	replyID, ok := positivePathID(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	reply, err := h.repository.GetReply(ctx, replyID)
	if err != nil {
		writeDiscussionError(c, err)
		return
	}
	if !canMutateDiscussion(c, reply.UserID) {
		return
	}
	input, ok := decodeCommentInput(c)
	if !ok {
		return
	}
	updated, err := h.repository.UpdateReply(ctx, replyID, input)
	if err != nil {
		writeDiscussionError(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *DiscussionHandler) DeleteReply(c *gin.Context) {
	replyID, ok := positivePathID(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.operationTimeout)
	defer cancel()
	reply, err := h.repository.GetReply(ctx, replyID)
	if err != nil {
		writeDiscussionError(c, err)
		return
	}
	if !canMutateDiscussion(c, reply.UserID) {
		return
	}
	if err := h.repository.DeleteReply(ctx, replyID); err != nil {
		writeDiscussionError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
	c.Writer.WriteHeaderNow()
}

type reactionRequest struct {
	ReactionType model.ReactionType `json:"reaction_type"`
}

func (h *DiscussionHandler) ListPosts(c *gin.Context) {
	query, pagination, ok := discussionQuery(c, "subject_id", "author_id")
	if !ok {
		return
	}
	filter := discussion.PostFilter{Pagination: pagination}
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

func decodePostInput(c *gin.Context, existing *model.Post, requireCore bool) (discussion.PostInput, bool) {
	input := discussion.PostInput{}
	if existing != nil {
		input = discussion.PostInput{SubjectID: existing.SubjectID, Title: existing.Title, ImageURL: existing.ImageURL, ImageDescription: existing.ImageDescription, Content: existing.Content}
	}
	fields, ok := decodeObject(c)
	if !ok {
		writeInvalidDiscussionRequest(c)
		return discussion.PostInput{}, false
	}
	if requireCore {
		for _, key := range []string{"subject_id", "title", "content"} {
			if _, exists := fields[key]; !exists {
				writeInvalidDiscussionRequest(c)
				return discussion.PostInput{}, false
			}
		}
	}
	for key, raw := range fields {
		switch key {
		case "subject_id":
			var subjectID int64
			if json.Unmarshal(raw, &subjectID) != nil || subjectID <= 0 {
				writeInvalidDiscussionRequest(c)
				return discussion.PostInput{}, false
			}
			input.SubjectID = subjectID
		case "title":
			value, valid := optionalString(raw, false)
			if !valid || value == nil || strings.TrimSpace(*value) == "" || len([]byte(*value)) > 200 {
				writeInvalidDiscussionRequest(c)
				return discussion.PostInput{}, false
			}
			input.Title = strings.TrimSpace(*value)
		case "content":
			value, valid := optionalString(raw, false)
			if !valid || value == nil || strings.TrimSpace(*value) == "" || len([]byte(*value)) > 10000 {
				writeInvalidDiscussionRequest(c)
				return discussion.PostInput{}, false
			}
			input.Content = strings.TrimSpace(*value)
		case "image_url":
			value, valid := optionalString(raw, true)
			if !valid {
				writeInvalidDiscussionRequest(c)
				return discussion.PostInput{}, false
			}
			input.ImageURL = value
		case "image_description":
			value, valid := optionalString(raw, true)
			if !valid {
				writeInvalidDiscussionRequest(c)
				return discussion.PostInput{}, false
			}
			input.ImageDescription = value
		default:
			writeInvalidDiscussionRequest(c)
			return discussion.PostInput{}, false
		}
	}
	if input.SubjectID <= 0 || strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.Content) == "" || (input.ImageURL == nil) != (input.ImageDescription == nil) {
		writeInvalidDiscussionRequest(c)
		return discussion.PostInput{}, false
	}
	return input, true
}

func decodeCommentInput(c *gin.Context) (discussion.CommentInput, bool) {
	fields, ok := decodeObject(c)
	if !ok || len(fields) != 1 {
		writeInvalidDiscussionRequest(c)
		return discussion.CommentInput{}, false
	}
	raw, exists := fields["content"]
	if !exists {
		writeInvalidDiscussionRequest(c)
		return discussion.CommentInput{}, false
	}
	value, valid := optionalString(raw, false)
	if !valid || value == nil || strings.TrimSpace(*value) == "" || len([]byte(*value)) > 10000 {
		writeInvalidDiscussionRequest(c)
		return discussion.CommentInput{}, false
	}
	return discussion.CommentInput{Content: strings.TrimSpace(*value)}, true
}

func canMutateDiscussion(c *gin.Context, authorID int64) bool {
	userID, ok := discussionUserID(c)
	if !ok {
		return false
	}
	isAdmin, _ := c.Get(auth.ContextIsAdmin)
	if userID == authorID {
		return true
	}
	if admin, ok := isAdmin.(bool); ok && admin {
		return true
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "acesso proibido"})
	return false
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
	if _, err := h.repository.GetPost(ctx, postID); err != nil {
		writeDiscussionError(c, err)
		return
	}
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
	if _, err := h.repository.GetComment(ctx, commentID); err != nil {
		writeDiscussionError(c, err)
		return
	}
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
		var likes, dislikes int64
		switch messageType {
		case "post":
			item, err := h.repository.GetPost(ctx, messageID)
			if err != nil {
				writeDiscussionError(c, err)
				return
			}
			likes, dislikes = item.Likes, item.Dislikes
		case "comment":
			item, err := h.repository.GetComment(ctx, messageID)
			if err != nil {
				writeDiscussionError(c, err)
				return
			}
			likes, dislikes = item.Likes, item.Dislikes
		default:
			item, err := h.repository.GetReply(ctx, messageID)
			if err != nil {
				writeDiscussionError(c, err)
				return
			}
			likes, dislikes = item.Likes, item.Dislikes
		}
		c.JSON(http.StatusOK, gin.H{"likes": likes, "dislikes": dislikes})
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
	notifications, err := h.repository.ListNotifications(ctx, userID, discussion.NotificationFilter{Unread: unread, Pagination: pagination})
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

func discussionQuery(c *gin.Context, allowed ...string) (url.Values, pagination.Pagination, bool) {
	query, err := url.ParseQuery(c.Request.URL.RawQuery)
	if err != nil {
		writeInvalidDiscussionRequest(c)
		return nil, pagination.Pagination{}, false
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
			return nil, pagination.Pagination{}, false
		}
	}
	for _, key := range []string{"page", "page_size"} {
		if values, supplied := query[key]; supplied && (len(values) != 1 || values[0] == "") {
			writeInvalidDiscussionRequest(c)
			return nil, pagination.Pagination{}, false
		}
	}
	page, err := pagination.Parse(query.Get("page"), query.Get("page_size"))
	if err != nil {
		writeInvalidDiscussionRequest(c)
		return nil, pagination.Pagination{}, false
	}
	return query, page, true
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

func positivePathID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeInvalidDiscussionRequest(c)
		return 0, false
	}
	return id, true
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
	if errors.Is(err, discussion.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "recurso não encontrado"})
		return
	}
	if errors.Is(err, discussion.ErrInvalid) {
		writeInvalidDiscussionRequest(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": "serviço indisponível"})
}

func writeInvalidDiscussionRequest(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"error": "requisição inválida"})
}
