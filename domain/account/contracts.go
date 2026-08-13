package account

import (
	"context"
	"errors"

	"germinaStack/model"
)

var (
	ErrNotFound            = errors.New("account not found")
	ErrPreferencesNotFound = errors.New("preferences not found")
)

type Repository interface {
	GetProfile(context.Context, int64) (model.User, error)
	GetPublicProfile(context.Context, string) (model.User, error)
	UpdateProfile(context.Context, int64, model.User) (model.User, error)
	GetPreferences(context.Context, int64) (model.Preference, error)
	UpsertPreferences(context.Context, int64, model.Preference) (model.Preference, error)
}
