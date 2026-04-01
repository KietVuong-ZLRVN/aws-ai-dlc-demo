package controllers

import (
	"ai-styling-engine/application/usecases"
	"ai-styling-engine/api/dto/response"
	"encoding/json"
	"net/http"
)

type StylePreferencesController struct {
	useCase *usecases.GetPreferenceOptionsUseCase
}

func NewStylePreferencesController(uc *usecases.GetPreferenceOptionsUseCase) *StylePreferencesController {
	return &StylePreferencesController{useCase: uc}
}

func (c *StylePreferencesController) GetOptions(w http.ResponseWriter, r *http.Request) {
	catalogue := c.useCase.Execute()

	occasions := make([]string, len(catalogue.Occasions))
	for i, o := range catalogue.Occasions {
		occasions[i] = string(o)
	}
	styles := make([]string, len(catalogue.Styles))
	for i, s := range catalogue.Styles {
		styles[i] = string(s)
	}
	colors := make([]string, len(catalogue.Colors))
	for i, c := range catalogue.Colors {
		colors[i] = string(c)
	}

	resp := response.PreferenceOptionsResponse{
		Occasions: occasions,
		Styles:    styles,
		Colors:    colors,
	}
	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
