package model

import "time"

type Subject struct {
	ID        int64     `db:"id" json:"id"`
	YearID    *int64    `db:"id_year" json:"id_year"`
	Subject   string    `db:"subject" json:"subject"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
