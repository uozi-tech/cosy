package model

type Pagination struct {
	Total       int64 `json:"total"`
	PerPage     int   `json:"per_page"`
	CurrentPage int   `json:"current_page"`
	TotalPages  int64 `json:"total_pages"`
}

type DataList struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination,omitempty"`
}

// TotalPage calculate total page
func TotalPage(total int64, pageSize int) int64 {
	// fix: divide by zero
	if pageSize == 0 {
		pageSize = 10
	}
	n := total / int64(pageSize)
	if total%int64(pageSize) > 0 {
		n++
	}
	return n
}
