package valueobjects

import "errors"

type BudgetRange struct {
	Min Money
	Max Money
}

func NewBudgetRange(min, max Money) (BudgetRange, error) {
	if min < 0 {
		return BudgetRange{}, errors.New("budget min must be >= 0")
	}
	if max <= min {
		return BudgetRange{}, errors.New("budget max must be greater than min")
	}
	return BudgetRange{Min: min, Max: max}, nil
}

func (b BudgetRange) Contains(price Money) bool {
	return price >= b.Min && price <= b.Max
}
