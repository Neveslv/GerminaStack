package model

import (
	"errors"
	"strings"
	"time"
)

type Post struct {
	ID                     int64      `db:"id" json:"id"`
	UserID                 int64      `db:"id_user" json:"id_user"`
	SubjectID              int64      `db:"id_subject" json:"id_subject"`
	Title                  string     `db:"title" json:"title"`
	ImageURL               *string    `db:"image_url" json:"image_url"`
	ImageDescription       *string    `db:"image_description" json:"image_description"`
	Content                string     `db:"content" json:"content"`
	Likes                  int64      `db:"likes" json:"likes"`
	Dislikes               int64      `db:"dislikes" json:"dislikes"`
	CommentsCount          int64      `db:"comments_count" json:"comments_count"`
	AuthorName             string     `db:"author_name" json:"author_name"`
	AuthorUsername         string     `db:"author_username" json:"author_username"`
	AuthorImageURL         *string    `db:"author_image_url" json:"author_image_url"`
	AuthorImageDescription *string    `db:"author_image_description" json:"author_image_description"`
	CreatedAt              *time.Time `db:"created_at" json:"created_at"`
}

type PostPage struct {
	Items   []Post `json:"items"`
	HasMore bool   `json:"has_more"`
}

func (post Post) Validate() error {
	if (post.ImageURL == nil) != (post.ImageDescription == nil) {
		return errors.New("imagem do post e descrição devem ser informadas juntas")
	}

	return nil
}

func (post Post) ValidateForCreate() error {
	if post.UserID <= 0 || post.SubjectID <= 0 || strings.TrimSpace(post.Title) == "" || strings.TrimSpace(post.Content) == "" {
		return errors.New("post inválido")
	}
	return post.Validate()
}
