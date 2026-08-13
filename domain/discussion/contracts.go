package discussion

import (
	"context"
	"errors"

	"germinaStack/domain/pagination"
	"germinaStack/model"
)

var (
	ErrNotFound = errors.New("discussion item not found")
	ErrInvalid  = errors.New("invalid discussion item")
)

type PostInput struct {
	SubjectID        int64
	Title            string
	ImageURL         *string
	ImageDescription *string
	Content          string
}

type CommentInput struct{ Content string }

type PostFilter struct {
	SubjectID  *int64
	AuthorID   *int64
	Pagination pagination.Pagination
}

type NotificationFilter struct {
	Unread     bool
	Pagination pagination.Pagination
}

type Repository interface {
	GetPost(context.Context, int64) (model.Post, error)
	CreatePost(context.Context, int64, PostInput) (model.Post, error)
	UpdatePost(context.Context, int64, PostInput) (model.Post, error)
	DeletePost(context.Context, int64) error
	GetComment(context.Context, int64) (model.Comment, error)
	CreateComment(context.Context, int64, int64, CommentInput) (model.Comment, error)
	UpdateComment(context.Context, int64, CommentInput) (model.Comment, error)
	DeleteComment(context.Context, int64) error
	GetReply(context.Context, int64) (model.CommentOnComment, error)
	CreateReply(context.Context, int64, int64, CommentInput) (model.CommentOnComment, error)
	UpdateReply(context.Context, int64, CommentInput) (model.CommentOnComment, error)
	DeleteReply(context.Context, int64) error
	ListPosts(context.Context, PostFilter) ([]model.Post, error)
	ListComments(context.Context, int64, pagination.Pagination) ([]model.Comment, error)
	ListReplies(context.Context, int64, pagination.Pagination) ([]model.CommentOnComment, error)
	React(context.Context, int64, int64, string, model.ReactionType) error
	ListNotifications(context.Context, int64, NotificationFilter) ([]model.Notification, error)
	MarkNotificationsRead(context.Context, int64) error
}
