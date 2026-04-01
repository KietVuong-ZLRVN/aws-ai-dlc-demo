package request

type GenerateCombosRequest struct {
	Preferences     *ConfirmPreferencesRequest `json:"preferences,omitempty"`
	ExcludeComboIds []string                   `json:"excludeComboIds,omitempty"`
}
