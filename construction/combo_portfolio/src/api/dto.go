package api

import "github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/domain"

// --- Request DTOs ---

type SaveComboRequest struct {
	Name       string               `json:"name"`
	Visibility string               `json:"visibility"`
	Items      []SaveComboItemInput `json:"items"`
}

type SaveComboItemInput struct {
	ConfigSku string  `json:"configSku"`
	SimpleSku string  `json:"simpleSku"`
	Name      string  `json:"name"`
	ImageUrl  string  `json:"imageUrl"`
	Price     float64 `json:"price"`
}

type UpdateComboRequest struct {
	Name       *string `json:"name"`
	Visibility *string `json:"visibility"`
}

// --- Response DTOs ---

type SaveComboResponse struct {
	ID string `json:"id"`
}

type ShareComboResponse struct {
	ShareToken string `json:"shareToken"`
	ShareURL   string `json:"shareUrl"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ComboResponse struct {
	ID         string              `json:"id"`
	ShopperID  string              `json:"shopperId"`
	Name       string              `json:"name"`
	Visibility string              `json:"visibility"`
	ShareToken *string             `json:"shareToken,omitempty"`
	Items      []ComboItemResponse `json:"items"`
}

type ComboItemResponse struct {
	ConfigSku string  `json:"configSku"`
	SimpleSku string  `json:"simpleSku"`
	Name      string  `json:"name"`
	ImageUrl  string  `json:"imageUrl"`
	Price     float64 `json:"price"`
	InStock   bool    `json:"inStock"`
}

// ToComboResponse maps an EnrichedCombo to the API response.
func ToComboResponse(c *domain.EnrichedCombo) ComboResponse {
	items := make([]ComboItemResponse, len(c.Items))
	for i, it := range c.Items {
		items[i] = ComboItemResponse{
			ConfigSku: it.ConfigSku,
			SimpleSku: it.SimpleSku,
			Name:      it.Name,
			ImageUrl:  it.ImageUrl,
			Price:     it.Price,
			InStock:   it.InStock,
		}
	}
	var shareToken *string
	if c.ShareToken != nil {
		s := c.ShareToken.String()
		shareToken = &s
	}
	return ComboResponse{
		ID:         c.ID.String(),
		ShopperID:  c.ShopperID.String(),
		Name:       c.Name.String(),
		Visibility: string(c.Visibility),
		ShareToken: shareToken,
		Items:      items,
	}
}
