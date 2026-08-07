package model

import "time"

type Comment struct {
	ID             int64      `db:"id" json:"id"`
	PostID         int64      `db:"id_post" json:"id_post"`
	UserID         int64      `db:"id_user" json:"id_user"`
	Content        string     `db:"content" json:"content"`
	Likes          int64      `db:"likes" json:"likes"`
	Dislikes       int64      `db:"dislikes" json:"dislikes"`
	AuthorName     string     `db:"author_name" json:"author_name"`
	AuthorUsername string     `db:"author_username" json:"author_username"`
	CreatedAt      *time.Time `db:"created_at" json:"created_at"`
}
