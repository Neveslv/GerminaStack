package model

import "time"

type Year struct {
	ID        int64      `db:"id" json:"id"`
	Year      string     `db:"year" json:"year"`
	CreatedAt *time.Time `db:"created_at" json:"created_at"`
}
