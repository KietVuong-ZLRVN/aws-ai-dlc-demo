package policies

import (
	"ai-styling-engine/domain/events"
	"log"
)

// ComboExclusionPolicy handles CombosGenerated.
// It is informational at this stage — the actual filtering of excluded combos
// is performed by StyleSession before raising the event. This policy logs
// the outcome and can be extended to enforce additional exclusion rules.
type ComboExclusionPolicy struct{}

func NewComboExclusionPolicy() *ComboExclusionPolicy {
	return &ComboExclusionPolicy{}
}

func (p *ComboExclusionPolicy) Handle(event events.DomainEvent) {
	e := event.(events.CombosGenerated)
	log.Printf("[ComboExclusionPolicy] session %s finalised with %d combo(s)", e.SessionId, e.ComboCount)
}
