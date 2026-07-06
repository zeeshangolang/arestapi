package data

import (
	//"math"
	"main/internal/validator"
	"strings"
)

type Filters struct {
	Page         int
	Page_size    int
	Sort         string
	SortSafelist []string
}

type MetaData struct {
	CurrentPage int `json:"current_page,omitempty"`
	PageSize    int `json:"page_size,omitempty"`
	FirstPage   int `json:"first_page,omitempty"`
	LastPage    int `json:"last_page,omitempty"`
	TotalRecors int `json:"total_records,omitempty"`
}

func Calculatemetadata(totalRecords, page, PageSize int) MetaData {
	if totalRecords == 0 {
		return MetaData{}
	}

	return MetaData{
		CurrentPage: page,
		PageSize:    PageSize,
		FirstPage:   1,
		LastPage:    (totalRecords + PageSize - 1) / PageSize,
		TotalRecors: totalRecords,
	}
}

func (f Filters) limit() int {
	return f.Page_size
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.Page_size
}

func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter: " + f.Sort)
}

func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "value must be reater than zero")
	v.Check(f.Page <= 100, "page", "value must be smaller than 100")
	v.Check(f.Page_size > 0, "page_size", "value must be greater than 0")
	v.Check(f.Page_size < 20, "page_size", "value must be smaller than 20")
	v.Check(validator.PermittedValues(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}
