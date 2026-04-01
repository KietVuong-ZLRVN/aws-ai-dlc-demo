package controllers

import (
	"ai-styling-engine/api/dto/request"
	"ai-styling-engine/api/dto/response"
	"ai-styling-engine/api/middleware"
	"ai-styling-engine/application/commands"
	"ai-styling-engine/application/usecases"
	"ai-styling-engine/domain/apperrors"
	"ai-styling-engine/domain/valueobjects"
	"encoding/json"
	"errors"
	"net/http"
)

type ComboGenerationController struct {
	useCase *usecases.GenerateCombosUseCase
}

func NewComboGenerationController(uc *usecases.GenerateCombosUseCase) *ComboGenerationController {
	return &ComboGenerationController{useCase: uc}
}

func (c *ComboGenerationController) Generate(w http.ResponseWriter, r *http.Request) {
	session, ok := middleware.SessionFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "UNAUTHENTICATED"})
		return
	}

	var req request.GenerateCombosRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "VALIDATION_ERROR", "detail": "invalid JSON body"})
			return
		}
	}

	var prefs *valueobjects.StylePreferences
	if req.Preferences != nil {
		p, err := mapToStylePreferences(req.Preferences)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "VALIDATION_ERROR", "detail": err.Error()})
			return
		}
		prefs = &p
	}

	cmd := commands.GenerateCombosCommand{
		Preferences:    prefs,
		ExcludedIds:    valueobjects.NewExcludedComboIds(req.ExcludeComboIds),
		ShopperSession: session,
	}

	result, err := c.useCase.Execute(cmd)
	if err != nil {
		if errors.Is(err, apperrors.ErrWishlistUnavailable) {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "DEPENDENCY_UNAVAILABLE", "detail": "Wishlist service is temporarily unavailable."})
			return
		}
		if errors.Is(err, apperrors.ErrAIUnavailable) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "AI_UNAVAILABLE", "detail": "Styling engine is temporarily unavailable. Please try again."})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "INTERNAL_ERROR"})
		return
	}

	if result.IsSuccess() {
		combos := make([]response.ComboDTO, len(result.Success.Combos))
		for i, combo := range result.Success.Combos {
			items := make([]response.ComboItemDTO, len(combo.Items))
			for j, item := range combo.Items {
				items[j] = response.ComboItemDTO{
					ConfigSku: string(item.ConfigSku),
					SimpleSku: string(item.SimpleSku),
					Name:      item.Name,
					Brand:     item.Brand,
					Price:     float64(item.Price),
					ImageUrl:  string(item.ImageUrl),
					Source:    string(item.Source),
				}
			}
			combos[i] = response.ComboDTO{
				Id:        combo.Id,
				Reasoning: combo.Reasoning.Text,
				Items:     items,
			}
		}
		writeJSON(w, http.StatusOK, response.ComboGenerationSuccessResponse{
			Status:    "ok",
			Combos:    combos,
			Exhausted: result.Success.Exhausted,
		})
		return
	}

	alternatives := make([]response.AlternativeItemDTO, len(result.Fallback.FallbackResult.Alternatives))
	for i, alt := range result.Fallback.FallbackResult.Alternatives {
		alternatives[i] = response.AlternativeItemDTO{
			ConfigSku: string(alt.ConfigSku),
			SimpleSku: string(alt.SimpleSku),
			Name:      alt.Name,
			Brand:     alt.Brand,
			Price:     float64(alt.Price),
			ImageUrl:  string(alt.ImageUrl),
			Reason:    alt.Reason,
		}
	}
	writeJSON(w, http.StatusOK, response.ComboGenerationFallbackResponse{
		Status:       "fallback",
		Message:      result.Fallback.FallbackResult.Message,
		Alternatives: alternatives,
	})
}
