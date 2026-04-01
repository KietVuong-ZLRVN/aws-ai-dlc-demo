package response

type ComboItemDTO struct {
	ConfigSku string  `json:"configSku"`
	SimpleSku string  `json:"simpleSku"`
	Name      string  `json:"name"`
	Brand     string  `json:"brand"`
	Price     float64 `json:"price"`
	ImageUrl  string  `json:"imageUrl"`
	Source    string  `json:"source"`
}

type ComboDTO struct {
	Id        string         `json:"id"`
	Reasoning string         `json:"reasoning"`
	Items     []ComboItemDTO `json:"items"`
}

type AlternativeItemDTO struct {
	ConfigSku string  `json:"configSku"`
	SimpleSku string  `json:"simpleSku"`
	Name      string  `json:"name"`
	Brand     string  `json:"brand"`
	Price     float64 `json:"price"`
	ImageUrl  string  `json:"imageUrl"`
	Reason    string  `json:"reason"`
}

// ComboGenerationSuccessResponse is returned when status = "ok".
type ComboGenerationSuccessResponse struct {
	Status    string     `json:"status"`
	Combos    []ComboDTO `json:"combos"`
	Exhausted bool       `json:"exhausted,omitempty"`
}

// ComboGenerationFallbackResponse is returned when status = "fallback".
type ComboGenerationFallbackResponse struct {
	Status       string               `json:"status"`
	Message      string               `json:"message"`
	Alternatives []AlternativeItemDTO `json:"alternatives"`
}
