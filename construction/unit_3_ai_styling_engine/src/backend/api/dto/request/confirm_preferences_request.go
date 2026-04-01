package request

type BudgetDTO struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type ColorsDTO struct {
	Preferred []string `json:"preferred"`
	Excluded  []string `json:"excluded"`
}

type ConfirmPreferencesRequest struct {
	Occasions []string   `json:"occasions"`
	Styles    []string   `json:"styles"`
	Budget    *BudgetDTO `json:"budget,omitempty"`
	Colors    *ColorsDTO `json:"colors,omitempty"`
	FreeText  string     `json:"freeText,omitempty"`
}
