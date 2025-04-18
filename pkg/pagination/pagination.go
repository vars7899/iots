package pagination

type Pagination struct {
	Page         int
	PageSize     int
	SortBy       string
	SortOrder    string
	TotalEntries int64
	TotalPages   int64
}

func (p *Pagination) GetOffset() int {
	if p.Page < 1 {
		p.Page = 1
	}
	return (p.Page - 1) * p.PageSize
}

func (p *Pagination) GetLimit() int {
	if p.PageSize > 0 {
		return p.PageSize
	}
	return 0
}

func (p *Pagination) GetSortOrderClause() string {
	if p.SortBy == "" {
		return ""
	}
	order := "ASC"
	if p.SortOrder == "desc" || p.SortOrder == "DESC" {
		order = "DESC"
	}
	return p.SortBy + " " + order
}

func (p *Pagination) SetPage(n int64) {
	p.TotalPages = n
}

func (p *Pagination) SetTotalEntries(n int64) {
	p.TotalEntries = n
}
