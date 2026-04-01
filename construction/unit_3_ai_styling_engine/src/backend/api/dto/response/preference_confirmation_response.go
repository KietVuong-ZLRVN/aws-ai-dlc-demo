package response

type PreferenceConfirmationResponse struct {
	Summary     string      `json:"summary"`
	Preferences interface{} `json:"preferences"`
}
