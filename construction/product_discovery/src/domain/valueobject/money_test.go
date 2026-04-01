package valueobject

import (
	"testing"

	"pgregory.net/rapid"
)

// --- PBT 4.1.9: valid construction ---

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
		if m.Currency != "SGD" {
			t.Fatalf("Currency: got %v, want SGD", m.Currency)
		}
	})
}

// --- PBT 4.1.10: rejects negative amount ---

func TestMoney_RejectsNegativeAmount(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		amount := rapid.Float64Range(-100000, -0.001).Draw(t, "negAmount")

		_, err := NewMoney(amount, "SGD")

		if err == nil {
			t.Fatalf("expected error for amount=%v, got nil", amount)
		}
	})
}

// --- Example: rejects empty currency ---

func TestMoney_RejectsEmptyCurrency(t *testing.T) {
	_, err := NewMoney(10.0, "")
	if err == nil {
		t.Fatal("expected error for empty currency, got nil")
	}
}
