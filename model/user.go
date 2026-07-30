package model

import (
	"errors"
	"time"
)

type User struct {
	ID                      int64     `db:"id" json:"id"`
	YearID                  int64     `db:"id_year" json:"id_year"`
	Name                    string    `db:"name" json:"name"`
	ProfileImageURL         *string   `db:"profile_image_url" json:"profile_image_url"`
	ProfileImageDescription *string   `db:"profile_image_description" json:"profile_image_description"`
	Username                string    `db:"username" json:"username"`
	Email                   string    `db:"email" json:"email"`
	Password                string    `db:"password" json:"-"`
	CreatedAt               time.Time `db:"created_at" json:"created_at"`
}

func (user User) Validate() error {
	if (user.ProfileImageURL == nil) != (user.ProfileImageDescription == nil) {
		return errors.New("imagem de perfil e descrição devem ser informadas juntas")
	}

	return nil
}
