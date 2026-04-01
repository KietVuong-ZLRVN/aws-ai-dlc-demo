package commands

import "ai-styling-engine/domain/valueobjects"

type GenerateCombosCommand struct {
	Preferences    *valueobjects.StylePreferences
	ExcludedIds    valueobjects.ExcludedComboIds
	ShopperSession valueobjects.ShopperSession
}
