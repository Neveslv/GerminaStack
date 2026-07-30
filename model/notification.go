package model

import "time"

type Notification struct {
	ID        int64     `db:"id" json:"id"`
	PostID    *int64    `db:"id_post" json:"id_post"`
	UserID    int64     `db:"id_user" json:"id_user"`
	TextShow  string    `db:"text_show" json:"text_show"`
	IsRead    bool      `db:"is_read" json:"is_read"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
