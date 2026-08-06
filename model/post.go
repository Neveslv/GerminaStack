package model

import (
	"errors"
	"strings"
	"time"
)

type Post struct {
	ID               int64      `db:"id" json:"id"`
	UserID           int64      `db:"id_user" json:"id_user"`
	SubjectID        int64      `db:"id_subject" json:"id_subject"`
	Title            string     `db:"title" json:"title"`
	ImageURL         *string    `db:"image_url" json:"image_url"`
	ImageDescription *string    `db:"image_description" json:"image_description"`
	Content          string     `db:"content" json:"content"`
	Likes            int64      `db:"likes" json:"likes"`
	Dislikes         int64      `db:"dislikes" json:"dislikes"`
	CreatedAt        *time.Time `db:"created_at" json:"created_at"`
}

func (post Post) Validate() error {
	if (post.ImageURL == nil) != (post.ImageDescription == nil) {
		return errors.New("imagem do post e descrição devem ser informadas juntas")
	}

	return nil
}

func (post Post) ValidateForCreate() error {
	if err := post.Validate(); err != nil {
		return err
	}
	if post.UserID <= 0 {
		return errors.New("post user must be positive")
	}
	if post.SubjectID <= 0 {
		return errors.New("post subject must be positive")
	}
	if strings.TrimSpace(post.Title) == "" {
		return errors.New("post title cannot be empty")
	}
	if strings.TrimSpace(post.Content) == "" {
		return errors.New("post content cannot be empty")
	}
	return nil
}
