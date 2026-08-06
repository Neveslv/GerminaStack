package model

import (
	"errors"
	"strings"
	"time"
)

type CommentOnComment struct {
	ID        int64      `db:"id" json:"id"`
	CommentID int64      `db:"id_comment" json:"id_comment"`
	UserID    int64      `db:"id_user" json:"id_user"`
	Content   string     `db:"content" json:"content"`
	Likes     int64      `db:"likes" json:"likes"`
	Dislikes  int64      `db:"dislikes" json:"dislikes"`
	CreatedAt *time.Time `db:"created_at" json:"created_at"`
}

func (reply CommentOnComment) ValidateForCreate() error {
	if reply.UserID <= 0 {
		return errors.New("reply user must be positive")
	}
	if reply.CommentID <= 0 {
		return errors.New("reply comment must be positive")
	}
	if strings.TrimSpace(reply.Content) == "" {
		return errors.New("reply content cannot be empty")
	}
	return nil
}
