package database

import (
	"math"
	"strconv"
	"testing"
)

func TestParsePagination(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		page     string
		pageSize string
		want     Pagination
		wantErr  bool
	}{
		{name: "defaults omitted values", want: Pagination{Page: 1, PageSize: 20}},
		{name: "accepts minimum values", page: "1", pageSize: "1", want: Pagination{Page: 1, PageSize: 1}},
		{name: "accepts maximum page size", page: "42", pageSize: "100", want: Pagination{Page: 42, PageSize: 100}},
		{name: "rejects malformed page", page: "one", pageSize: "20", wantErr: true},
		{name: "rejects malformed page size", page: "1", pageSize: "many", wantErr: true},
		{name: "rejects zero page", page: "0", pageSize: "20", wantErr: true},
		{name: "rejects negative page", page: "-1", pageSize: "20", wantErr: true},
		{name: "rejects zero page size", page: "1", pageSize: "0", wantErr: true},
		{name: "rejects negative page size", page: "1", pageSize: "-1", wantErr: true},
		{name: "rejects page size above maximum", page: "1", pageSize: "101", wantErr: true},
		{name: "rejects offset overflow", page: strconv.Itoa(math.MaxInt), pageSize: "100", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePagination(tt.page, tt.pageSize)
			if tt.wantErr {
				if err == nil {
					t.Fatal("ParsePagination() error = nil, want non-nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsePagination() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ParsePagination() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
