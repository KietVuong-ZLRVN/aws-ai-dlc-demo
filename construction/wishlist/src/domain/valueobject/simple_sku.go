package valueobject

import "errors"

type SimpleSku struct {
	value string
}

func NewSimpleSku(v string) (SimpleSku, error) {
	if v == "" {
		return SimpleSku{}, errors.New("simple sku cannot be empty")
	}
	return SimpleSku{value: v}, nil
}

func (s SimpleSku) String() string {
	return s.value
}
