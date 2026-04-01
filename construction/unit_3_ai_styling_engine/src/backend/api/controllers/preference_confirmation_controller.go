package controllers

import (
	"ai-styling-engine/api/dto/request"
	"ai-styling-engine/api/dto/response"
	"ai-styling-engine/api/middleware"
	"ai-styling-engine/application/commands"
	"ai-styling-engine/application/usecases"
	"ai-styling-engine/domain/valueobjects"
	"encoding/json"
	"fmt"
	"net/http"
)

type PreferenceConfirmationController struct {
	useCase *usecases.ConfirmPreferencesUseCase
}

func NewPreferenceConfirmationController(uc *usecases.ConfirmPreferencesUseCase) *PreferenceConfirmationController {
	return &PreferenceConfirmationController{useCase: uc}
}

func (c *PreferenceConfirmationController) Confirm(w http.ResponseWriter, r *http.Request) {
	session, ok := middleware.SessionFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "UNAUTHENTICATED"})
		return
	}

	var req request.ConfirmPreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "VALIDATION_ERROR", "detail": "invalid JSON body"})
		return
	}

	prefs, err := mapToStylePreferences(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "VALIDATION_ERROR", "detail": err.Error()})
		return
	}

	cmd := commands.ConfirmPreferencesCommand{Preferences: prefs, ShopperSession: session}
	summary, err := c.useCase.Execute(cmd)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "INTERNAL_ERROR"})
		return
	}

	writeJSON(w, http.StatusOK, response.PreferenceConfirmationResponse{
		Summary:     summary.Text,
		Preferences: req,
	})
}

func mapToStylePreferences(req *request.ConfirmPreferencesRequest) (valueobjects.StylePreferences, error) {
	if req == nil {
		return valueobjects.StylePreferences{}, nil
	}

	// Validate and map occasions.
	validOccasions := make(map[valueobjects.Occasion]bool)
	for _, o := range valueobjects.ValidOccasions {
		validOccasions[o] = true
	}
	occasions := make([]valueobjects.Occasion, len(req.Occasions))
	for i, o := range req.Occasions {
		occ := valueobjects.Occasion(o)
		if !validOccasions[occ] {
			return valueobjects.StylePreferences{}, fmt.Errorf("unrecognised occasion: %q", o)
		}
		occasions[i] = occ
	}

	// Validate and map styles.
	validStyles := make(map[valueobjects.StyleDirection]bool)
	for _, s := range valueobjects.ValidStyleDirections {
		validStyles[s] = true
	}
	styles := make([]valueobjects.StyleDirection, len(req.Styles))
	for i, s := range req.Styles {
		sd := valueobjects.StyleDirection(s)
		if !validStyles[sd] {
			return valueobjects.StylePreferences{}, fmt.Errorf("unrecognised style: %q", s)
		}
		styles[i] = sd
	}

	var budget *valueobjects.BudgetRange
	if req.Budget != nil {
		b, err := valueobjects.NewBudgetRange(valueobjects.Money(req.Budget.Min), valueobjects.Money(req.Budget.Max))
		if err != nil {
			return valueobjects.StylePreferences{}, err
		}
		budget = &b
	}

	// Validate and map colors.
	validColors := make(map[valueobjects.Color]bool)
	for _, c := range valueobjects.ValidColors {
		validColors[c] = true
	}
	var palette valueobjects.ColorPalette
	if req.Colors != nil {
		preferred := make([]valueobjects.Color, len(req.Colors.Preferred))
		for i, c := range req.Colors.Preferred {
			col := valueobjects.Color(c)
			if !validColors[col] {
				return valueobjects.StylePreferences{}, fmt.Errorf("unrecognised color: %q", c)
			}
			preferred[i] = col
		}
		excluded := make([]valueobjects.Color, len(req.Colors.Excluded))
		for i, c := range req.Colors.Excluded {
			col := valueobjects.Color(c)
			if !validColors[col] {
				return valueobjects.StylePreferences{}, fmt.Errorf("unrecognised color: %q", c)
			}
			excluded[i] = col
		}
		var err error
		palette, err = valueobjects.NewColorPalette(preferred, excluded)
		if err != nil {
			return valueobjects.StylePreferences{}, err
		}
	}

	return valueobjects.StylePreferences{
		Occasions: occasions,
		Styles:    styles,
		Budget:    budget,
		Colors:    palette,
		FreeText:  req.FreeText,
	}, nil
}
