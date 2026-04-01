package response

type PreferenceOptionsResponse struct {
	Occasions []string `json:"occasions"`
	Styles    []string `json:"styles"`
	Colors    []string `json:"colors"`
}
