package model

import (
	"errors"
	"strings"
	"time"
)

type Comment struct {
	ID        int64      `db:"id" json:"id"`
	PostID    int64      `db:"id_post" json:"id_post"`
	UserID    int64      `db:"id_user" json:"id_user"`
	Content   string     `db:"content" json:"content"`
	Likes     int64      `db:"likes" json:"likes"`
	Dislikes  int64      `db:"dislikes" json:"dislikes"`
	CreatedAt *time.Time `db:"created_at" json:"created_at"`
}

func (comment Comment) ValidateForCreate() error {
	if comment.UserID <= 0 {
		return errors.New("comment user must be positive")
	}
	if comment.PostID <= 0 {
		return errors.New("comment post must be positive")
	}
	if strings.TrimSpace(comment.Content) == "" {
		return errors.New("comment content cannot be empty")
	}
	return nil
}
