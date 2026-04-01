package policies

import (
	"ai-styling-engine/domain/events"
)

// PreferenceDefaultPolicy handles ComboGenerationRequested.
// If no preferences are set, it signals quick-generate mode on the session
// via the QuickGenerate flag already set in the event payload.
// This policy is informational — the StyleSession reads the flag directly.
type PreferenceDefaultPolicy struct{}

func NewPreferenceDefaultPolicy() *PreferenceDefaultPolicy {
	return &PreferenceDefaultPolicy{}
}

func (p *PreferenceDefaultPolicy) Handle(event events.DomainEvent) {
	// The QuickGenerate flag is already set by StyleSession before raising the event.
	// This policy exists as an extension point for future rules on generation mode.
	_ = event.(events.ComboGenerationRequested)
}
