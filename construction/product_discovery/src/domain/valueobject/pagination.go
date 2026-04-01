package valueobject

import "fmt"

const defaultLimit = 20

// Pagination represents offset-based pagination parameters.
type Pagination struct {
	Offset int
	Limit  int
}

// NewPagination constructs a Pagination value object.
// Negative offset or limit returns an error.
// Zero values fall back to defaults: offset=0, limit=20.
func NewPagination(offset, limit int) (Pagination, error) {
	if offset < 0 {
		return Pagination{}, fmt.Errorf("pagination offset must not be negative, got %d", offset)
	}
	if limit < 0 {
		return Pagination{}, fmt.Errorf("pagination limit must not be negative, got %d", limit)
	}
	if limit == 0 {
		limit = defaultLimit
	}
	return Pagination{Offset: offset, Limit: limit}, nil
}
