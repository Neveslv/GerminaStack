package moderation

import (
	"context"
	"errors"

	"germinaStack/domain/pagination"
	"germinaStack/model"
)

var ErrUserNotFound = errors.New("admin user not found")

type Filter struct {
	Search     string
	Pagination pagination.Pagination
}

type Post struct {
	ID             int64  `json:"id"`
	Title          string `json:"title"`
	AuthorName     string `json:"author_name"`
	AuthorUsername string `json:"author_username"`
}

type Repository interface {
	ListUsers(context.Context, Filter) ([]model.User, int, error)
	ListPosts(context.Context, Filter) ([]Post, int, error)
	GetUser(context.Context, int64) (model.User, error)
	SetBanned(context.Context, int64, bool) error
	SetAdmin(context.Context, int64, bool) error
	DeletePost(context.Context, int64) error
}
