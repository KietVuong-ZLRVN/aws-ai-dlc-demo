package valueobject

import (
	"testing"

	"pgregory.net/rapid"
)

// 4.5.5: valid construction for any non-negative amount
func TestMoney_ValidConstruction(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		amount := rapid.Float64Range(0, 100000).Draw(t, "amount")

		m, err := NewMoney(amount, "SGD")

		if err != nil {
			t.Fatalf("expected no error for amount=%v, got: %v", amount, err)
		}
		if m.Amount != amount {
			t.Fatalf("Amount: got %v, want %v", m.Amount, amount)
		}
	})
}

// 4.5.6: rejects negative amount
func TestMoney_RejectsNegative(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		amount := rapid.Float64Range(-100000, -0.001).Draw(t, "negAmount")

		_, err := NewMoney(amount, "SGD")

		if err == nil {
			t.Fatalf("expected error for amount=%v, got nil", amount)
		}
	})
}
