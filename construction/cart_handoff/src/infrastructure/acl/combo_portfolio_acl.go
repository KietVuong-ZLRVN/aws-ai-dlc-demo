package acl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/domain"
)

// ComboPortfolioPort is the interface for fetching combo items from Unit 4.
// Also implemented by the demo stub.
type ComboPortfolioPort interface {
	FetchComboItems(ctx context.Context, comboId, sessionCookie string) ([]domain.CartItem, error)
}

// comboPortfolioResponse maps the relevant fields from Unit 4's GET /api/v1/combos/{id}.
type comboPortfolioResponse struct {
	ID    string `json:"id"`
	Items []struct {
		SimpleSku string `json:"simpleSku"`
	} `json:"items"`
}

// HTTPComboPortfolioACL calls Unit 4 and translates the combo item list into CartItems.
type HTTPComboPortfolioACL struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPComboPortfolioACL(baseURL string, timeoutMs int) *HTTPComboPortfolioACL {
	return &HTTPComboPortfolioACL{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: time.Duration(timeoutMs) * time.Millisecond},
	}
}

func (a *HTTPComboPortfolioACL) FetchComboItems(ctx context.Context, comboId, sessionCookie string) ([]domain.CartItem, error) {
	url := fmt.Sprintf("%s/api/v1/combos/%s", a.baseURL, comboId)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build combo request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cookie", sessionCookie)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, domain.ErrComboPortfolioUnavailable
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, domain.ErrComboNotFound
	case http.StatusForbidden:
		return nil, domain.ErrComboAccessDenied
	case http.StatusOK:
		// handled below
	default:
		if resp.StatusCode >= 500 {
			return nil, domain.ErrComboPortfolioUnavailable
		}
		return nil, fmt.Errorf("combo portfolio returned %d", resp.StatusCode)
	}

	var payload comboPortfolioResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode combo response: %w", err)
	}

	items := make([]domain.CartItem, len(payload.Items))
	for i, it := range payload.Items {
		items[i] = domain.CartItem{SimpleSku: it.SimpleSku, Quantity: 1}
	}
	return items, nil
}
