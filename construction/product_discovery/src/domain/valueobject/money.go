package valueobject

import (
	"errors"
	"fmt"
)

// Money represents a monetary amount with a currency code.
type Money struct {
	Amount   float64
	Currency string
}

// NewMoney constructs a Money value object.
// Returns an error if the amount is negative.
func NewMoney(amount float64, currency string) (Money, error) {
	if amount < 0 {
		return Money{}, fmt.Errorf("money amount must not be negative, got %f", amount)
	}
	if currency == "" {
		return Money{}, errors.New("money currency must not be empty")
	}
	return Money{Amount: amount, Currency: currency}, nil
}
