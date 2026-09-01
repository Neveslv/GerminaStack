package pagination

import (
	"fmt"
	"math"
	"strconv"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

type Pagination struct {
	Page     int
	PageSize int
}

func Parse(page, pageSize string) (Pagination, error) {
	parsedPage, err := positiveInt(page, 1)
	if err != nil {
		return Pagination{}, fmt.Errorf("parse page: %w", err)
	}
	parsedPageSize, err := positiveInt(pageSize, defaultPageSize)
	if err != nil {
		return Pagination{}, fmt.Errorf("parse page_size: %w", err)
	}
	if parsedPageSize > maxPageSize {
		return Pagination{}, fmt.Errorf("page_size must be at most %d", maxPageSize)
	}
	if parsedPage-1 > math.MaxInt/parsedPageSize {
		return Pagination{}, fmt.Errorf("page is too large")
	}
	return Pagination{Page: parsedPage, PageSize: parsedPageSize}, nil
}

func positiveInt(value string, fallback int) (int, error) {
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf("must be a positive integer")
	}
	return parsed, nil
}
