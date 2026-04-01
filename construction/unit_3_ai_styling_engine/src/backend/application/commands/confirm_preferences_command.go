package commands

import "ai-styling-engine/domain/valueobjects"

type ConfirmPreferencesCommand struct {
	Preferences    valueobjects.StylePreferences
	ShopperSession valueobjects.ShopperSession
}
