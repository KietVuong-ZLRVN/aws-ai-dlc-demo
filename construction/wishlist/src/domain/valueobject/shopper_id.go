package valueobject

import "errors"

type ShopperId struct {
	value string
}

func NewShopperId(v string) (ShopperId, error) {
	if v == "" {
		return ShopperId{}, errors.New("shopper id cannot be empty")
	}
	return ShopperId{value: v}, nil
}

func (s ShopperId) String() string {
	return s.value
}
