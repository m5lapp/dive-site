package models

type ListFilters struct {
	page     int
	pageSize int
}

func NewListFilters(page, pageSize, defaultPageSize int) ListFilters {
	if page < 1 || page > 10_000_000 {
		page = 1
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = defaultPageSize
	}

	return ListFilters{
		page:     page,
		pageSize: pageSize,
	}
}

func (lf ListFilters) limit() int {
	return lf.pageSize
}

func (lf ListFilters) offset() int {
	return (lf.page - 1) * lf.pageSize
}

type PageData struct {
	FirstPage    int
	LastPage     int
	CurrentPage  int
	PageSize     int
	TotalRecords int
}

func newPaginationData(totalRecods, page, pageSize int) PageData {
	if totalRecods == 0 {
		return PageData{}
	}

	return PageData{
		FirstPage:    1,
		LastPage:     (totalRecods + pageSize - 1) / pageSize,
		CurrentPage:  page,
		PageSize:     pageSize,
		TotalRecords: totalRecods,
	}
}
