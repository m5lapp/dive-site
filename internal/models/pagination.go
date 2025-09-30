package models

type Pager struct {
	page     int
	pageSize int
}

// NewPager returns a new Pager struct. The pageSize parameter is the size of
// the page requested by the user; if it's outside of the allowed range, then
// the defaultPageSize will be used instead.
func NewPager(page, pageSize, defaultPageSize int) Pager {
	if page < 1 || page > 10_000_000 {
		page = 1
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = defaultPageSize
	}

	return Pager{
		page:     page,
		pageSize: pageSize,
	}
}

func (lf Pager) limit() int {
	return lf.pageSize
}

func (lf Pager) offset() int {
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
