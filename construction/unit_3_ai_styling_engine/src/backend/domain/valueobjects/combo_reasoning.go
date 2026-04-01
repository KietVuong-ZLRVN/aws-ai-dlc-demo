package valueobjects

import (
	"errors"
	"strings"
)

type ComboReasoning struct {
	Text string
}

func NewComboReasoning(text string) (ComboReasoning, error) {
	if strings.TrimSpace(text) == "" {
		return ComboReasoning{}, errors.New("combo reasoning text must not be empty")
	}
	return ComboReasoning{Text: text}, nil
}
