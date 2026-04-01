package entities

import "ai-styling-engine/domain/valueobjects"

// Combo is a single AI-generated outfit suggestion.
// It has identity (ComboId) so it can be tracked and excluded across multiple generate calls.
type Combo struct {
	Id        string
	Reasoning valueobjects.ComboReasoning
	Items     []valueobjects.ComboItem
	Rank      int
	rejected  bool
}

func NewCombo(id string, items []valueobjects.ComboItem, rank int) Combo {
	return Combo{Id: id, Items: items, Rank: rank}
}

func (c *Combo) AttachReasoning(r valueobjects.ComboReasoning) {
	c.Reasoning = r
}

// Reject marks this combo as rejected by the shopper.
// The caller is responsible for raising the ComboRejected domain event.
func (c *Combo) Reject() {
	c.rejected = true
}

func (c *Combo) IsRejected() bool {
	return c.rejected
}
