package policies

import (
	"ai-styling-engine/domain/events"
	"log"
)

// FallbackPolicy handles FallbackTriggered.
// It validates that the fallback result contains at least one alternative.
// A warning is logged if no alternatives are present — this is a degraded state.
type FallbackPolicy struct{}

func NewFallbackPolicy() *FallbackPolicy {
	return &FallbackPolicy{}
}

func (p *FallbackPolicy) Handle(event events.DomainEvent) {
	e := event.(events.FallbackTriggered)
	if len(e.Alternatives) == 0 {
		log.Printf("[FallbackPolicy] WARNING: session %s produced a fallback with no alternatives", e.SessionId)
	}
}
