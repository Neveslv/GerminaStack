package catalog

import (
	"context"

	"germinaStack/model"
)

type Repository interface {
	ListYears(context.Context) ([]model.Year, error)
	ListSubjects(context.Context, int64) ([]model.Subject, error)
}
