package database

import (
	"fmt"
	"strconv"
)

const defaultPageSize = 20
const maxPageSize = 100

type Pagination struct {
	Page     int
	PageSize int
}

func ParsePagination(page, pageSize string) (Pagination, error) {
	parsedPage, err := parsePositiveInt(page, 1)
	if err != nil {
		return Pagination{}, fmt.Errorf("parse page: %w", err)
	}
	parsedPageSize, err := parsePositiveInt(pageSize, defaultPageSize)
	if err != nil {
		return Pagination{}, fmt.Errorf("parse page_size: %w", err)
	}
	if parsedPageSize > maxPageSize {
		return Pagination{}, fmt.Errorf("page_size must be at most %d", maxPageSize)
	}
	return Pagination{Page: parsedPage, PageSize: parsedPageSize}, nil
}

func parsePositiveInt(value string, defaultValue int) (int, error) {
	if value == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf("must be a positive integer")
	}
	return parsed, nil
}
