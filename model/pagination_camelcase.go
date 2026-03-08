//go:build camelcase_json

package model

type Pagination struct {
	Total       int64 `json:"total"`
	PerPage     int   `json:"perPage"`
	CurrentPage int   `json:"currentPage"`
	TotalPages  int64 `json:"totalPages"`
}
