package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"germinaStack/auth"
	"germinaStack/database"
	"germinaStack/model"

	"github.com/gin-gonic/gin"
)

type MessageRepository interface {
	CreatePost(context.Context, int64, int64, string, *string, *string, string) (model.Post, error)
	CreateComment(context.Context, int64, int64, string) (model.Comment, error)
	CreateReply(context.Context, int64, int64, string) (model.CommentOnComment, error)
	GetPost(context.Context, int64) (model.Post, error)
	ListPosts(context.Context, *int64) ([]model.Post, error)
	GetComment(context.Context, int64) (model.Comment, error)
	ListComments(context.Context, int64) ([]model.Comment, error)
	GetReply(context.Context, int64) (model.CommentOnComment, error)
	ListReplies(context.Context, int64) ([]model.CommentOnComment, error)
}

type MessageHandler struct {
	repository       MessageRepository
	operationTimeout time.Duration
}

func NewMessageHandler(repository MessageRepository, operationTimeout time.Duration) *MessageHandler {
	return &MessageHandler{repository: repository, operationTimeout: operationTimeout}
}

type createPostRequest struct {
	SubjectID        int64   `json:"id_subject"`
	Title            string  `json:"title"`
	Content          string  `json:"content"`
	ImageURL         *string `json:"image_url"`
	ImageDescription *string `json:"image_description"`
}

type createMessageRequest struct {
	Content string `json:"content"`
}

func (h *MessageHandler) ListPosts(c *gin.Context) {
	var subjectID *int64
	query, err := url.ParseQuery(c.Request.URL.RawQuery)
	if err != nil {
		writeInvalidMessageRequest(c)
		return
	}
	if values, supplied := query["id_subject"]; supplied {
		if len(values) != 1 {
			writeInvalidMessageRequest(c)
			return
		}
		parsed, err := strconv.ParseInt(values[0], 10, 64)
		if err != nil || parsed <= 0 {
			writeInvalidMessageRequest(c)
			return
		}
		subjectID = &parsed
	}

	ctx, cancel := h.operationContext(c)
	defer cancel()
	posts, err := h.repository.ListPosts(ctx, subjectID)
	if err != nil {
		writeMessageError(c, err)
		return
	}
	c.JSON(http.StatusOK, posts)
}

func (h *MessageHandler) CreatePost(c *gin.Context) {
	userID, ok := messageUserID(c)
	if !ok {
		writeMessageUnauthorized(c)
		return
	}
	var request createPostRequest
	if err := decodeJSON(c, &request); err != nil {
		writeInvalidMessageRequest(c)
		return
	}
	post := model.Post{UserID: userID, SubjectID: request.SubjectID, Title: request.Title, Content: request.Content, ImageURL: request.ImageURL, ImageDescription: request.ImageDescription}
	if err := post.ValidateForCreate(); err != nil {
		writeInvalidMessageRequest(c)
		return
	}

	ctx, cancel := h.operationContext(c)
	defer cancel()
	created, err := h.repository.CreatePost(ctx, userID, request.SubjectID, request.Content, request.ImageURL, request.ImageDescription, request.Title)
	if err != nil {
		writeMessageError(c, err)
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *MessageHandler) GetPost(c *gin.Context) {
	id, ok := positiveMessagePathID(c)
	if !ok {
		return
	}
	ctx, cancel := h.operationContext(c)
	defer cancel()
	post, err := h.repository.GetPost(ctx, id)
	if err != nil {
		writeMessageError(c, err)
		return
	}
	c.JSON(http.StatusOK, post)
}

func (h *MessageHandler) ListComments(c *gin.Context) {
	postID, ok := positiveMessagePathID(c)
	if !ok {
		return
	}
	ctx, cancel := h.operationContext(c)
	defer cancel()
	comments, err := h.repository.ListComments(ctx, postID)
	if err != nil {
		writeMessageError(c, err)
		return
	}
	c.JSON(http.StatusOK, comments)
}

func (h *MessageHandler) CreateComment(c *gin.Context) {
	userID, ok := messageUserID(c)
	if !ok {
		writeMessageUnauthorized(c)
		return
	}
	postID, ok := positiveMessagePathID(c)
	if !ok {
		return
	}
	request, ok := decodeCreateMessageRequest(c)
	if !ok {
		return
	}
	ctx, cancel := h.operationContext(c)
	defer cancel()
	comment, err := h.repository.CreateComment(ctx, userID, postID, request.Content)
	if err != nil {
		writeMessageError(c, err)
		return
	}
	c.JSON(http.StatusCreated, comment)
}

func (h *MessageHandler) ListReplies(c *gin.Context) {
	commentID, ok := positiveMessagePathID(c)
	if !ok {
		return
	}
	ctx, cancel := h.operationContext(c)
	defer cancel()
	replies, err := h.repository.ListReplies(ctx, commentID)
	if err != nil {
		writeMessageError(c, err)
		return
	}
	c.JSON(http.StatusOK, replies)
}

func (h *MessageHandler) CreateReply(c *gin.Context) {
	userID, ok := messageUserID(c)
	if !ok {
		writeMessageUnauthorized(c)
		return
	}
	commentID, ok := positiveMessagePathID(c)
	if !ok {
		return
	}
	request, ok := decodeCreateMessageRequest(c)
	if !ok {
		return
	}
	ctx, cancel := h.operationContext(c)
	defer cancel()
	reply, err := h.repository.CreateReply(ctx, userID, commentID, request.Content)
	if err != nil {
		writeMessageError(c, err)
		return
	}
	c.JSON(http.StatusCreated, reply)
}

func (h *MessageHandler) operationContext(c *gin.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Request.Context(), h.operationTimeout)
}

func decodeCreateMessageRequest(c *gin.Context) (createMessageRequest, bool) {
	var request createMessageRequest
	if err := decodeJSON(c, &request); err != nil || strings.TrimSpace(request.Content) == "" {
		writeInvalidMessageRequest(c)
		return createMessageRequest{}, false
	}
	return request, true
}

func messageUserID(c *gin.Context) (int64, bool) {
	value, exists := c.Get(auth.ContextUserID)
	userID, ok := value.(int64)
	return userID, exists && ok && userID > 0
}

func positiveMessagePathID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeInvalidMessageRequest(c)
		return 0, false
	}
	return id, true
}

func writeMessageError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, database.ErrMessageNotFound), errors.Is(err, database.ErrMessageParentNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "recurso não encontrado"})
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "serviço indisponível"})
	}
}

func writeInvalidMessageRequest(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"error": "requisição inválida"})
}

func writeMessageUnauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{"error": "não autorizado"})
}
