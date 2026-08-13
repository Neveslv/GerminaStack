package model

import (
	"errors"
	"strings"
	"time"
)

type CommentOnComment struct {
	ID             int64      `db:"id" json:"id"`
	CommentID      int64      `db:"id_comment" json:"id_comment"`
	UserID         int64      `db:"id_user" json:"id_user"`
	Content        string     `db:"content" json:"content"`
	Likes          int64      `db:"likes" json:"likes"`
	Dislikes       int64      `db:"dislikes" json:"dislikes"`
	AuthorName     string     `db:"author_name" json:"author_name"`
	AuthorUsername string     `db:"author_username" json:"author_username"`
	CreatedAt      *time.Time `db:"created_at" json:"created_at"`
}

func (reply CommentOnComment) ValidateForCreate() error {
	if reply.UserID <= 0 || reply.CommentID <= 0 || strings.TrimSpace(reply.Content) == "" {
		return errors.New("resposta inválida")
	}
	return nil
}
