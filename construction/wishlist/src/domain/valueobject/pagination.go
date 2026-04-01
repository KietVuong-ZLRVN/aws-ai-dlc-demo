package valueobject

import "errors"

type Pagination struct {
	Offset int
	Limit  int
}

func NewPagination(offset, limit int) (Pagination, error) {
	if offset < 0 {
		return Pagination{}, errors.New("pagination offset cannot be negative")
	}
	if limit <= 0 {
		return Pagination{}, errors.New("pagination limit must be greater than zero")
	}
	return Pagination{Offset: offset, Limit: limit}, nil
}
