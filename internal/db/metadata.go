package db

type PaginationMetadata struct {
	CurrentPage  int32 `json:"current_page,omitempty"`
	PageSize     int32 `json:"page_size,omitempty"`
	FirstPage    int32 `json:"first_page,omitempty"`
	LastPage     int32 `json:"last_page,omitempty"`
	TotalRecords int64 `json:"total_records,omitempty"`
}

func CalculatePaginationMetadata(totalRecords int64, page, pageSize int32) PaginationMetadata {
	if totalRecords == 0 {
		return PaginationMetadata{}
	}

	return PaginationMetadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int32((totalRecords + int64(pageSize) - 1) / int64(pageSize)),
		TotalRecords: totalRecords,
	}
}
