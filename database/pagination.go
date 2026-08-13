package database

import "germinaStack/domain/pagination"

type Pagination = pagination.Pagination

func ParsePagination(page, pageSize string) (Pagination, error) {
	return pagination.Parse(page, pageSize)
}
