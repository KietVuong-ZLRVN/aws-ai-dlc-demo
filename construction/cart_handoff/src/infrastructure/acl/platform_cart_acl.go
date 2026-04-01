package acl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/domain"
)

// CartSubmissionResult holds the outcome of a bulk cart add.
type CartSubmissionResult struct {
	AddedItems   []domain.CartItem
	SkippedItems []domain.SkippedItem
}

// PlatformCartPort is the interface for adding items to the platform cart.
// Also implemented by the demo stub.
type PlatformCartPort interface {
	BulkAddToCart(ctx context.Context, items []domain.CartItem, sessionCookie string) (CartSubmissionResult, error)
}

// platformCartBulkItem is the JSON shape expected by the Doraemon bulk cart API.
type platformCartBulkItem struct {
	SimpleSku string `json:"simpleSku"`
	Quantity  int    `json:"quantity"`
	Size      string `json:"size,omitempty"`
}

// platformCartResponse is a minimal mapping of the ZDTCart.Cart response.
// We extract which simpleSku values were accepted to determine added vs skipped.
type platformCartResponse struct {
	Products []struct {
		SimpleSku string `json:"simpleSku"`
	} `json:"products"`
}

// HTTPPlatformCartACL calls Doraemon's POST /v1/checkout/cart/bulk.
type HTTPPlatformCartACL struct {
	baseURL         string
	contentLanguage string
	httpClient      *http.Client
}

func NewHTTPPlatformCartACL(baseURL, contentLanguage string, timeoutMs int) *HTTPPlatformCartACL {
	return &HTTPPlatformCartACL{
		baseURL:         baseURL,
		contentLanguage: contentLanguage,
		httpClient:      &http.Client{Timeout: time.Duration(timeoutMs) * time.Millisecond},
	}
}

func (a *HTTPPlatformCartACL) BulkAddToCart(ctx context.Context, items []domain.CartItem, sessionCookie string) (CartSubmissionResult, error) {
	// Encode items as JSON for the "products" form field.
	bulk := make([]platformCartBulkItem, len(items))
	for i, it := range items {
		bulk[i] = platformCartBulkItem{SimpleSku: it.SimpleSku, Quantity: it.Quantity, Size: it.Size}
	}
	productsJSON, err := json.Marshal(bulk)
	if err != nil {
		return CartSubmissionResult{}, fmt.Errorf("marshal products: %w", err)
	}

	formData := url.Values{}
	formData.Set("products", string(productsJSON))

	reqURL := fmt.Sprintf("%s/v1/checkout/cart/bulk", a.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return CartSubmissionResult{}, fmt.Errorf("build cart request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Language", a.contentLanguage)
	req.Header.Set("Cookie", sessionCookie)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return CartSubmissionResult{}, domain.ErrPlatformCartUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return CartSubmissionResult{}, domain.ErrPlatformCartUnavailable
	}

	var cartResp platformCartResponse
	if err := json.NewDecoder(resp.Body).Decode(&cartResp); err != nil {
		return CartSubmissionResult{}, fmt.Errorf("decode cart response: %w", err)
	}

	// Classify: items in the response were added; items in the request but absent → skipped.
	addedSet := map[string]bool{}
	for _, p := range cartResp.Products {
		addedSet[p.SimpleSku] = true
	}

	var added []domain.CartItem
	var skipped []domain.SkippedItem
	for _, item := range items {
		if addedSet[item.SimpleSku] {
			added = append(added, item)
		} else {
			skipped = append(skipped, domain.SkippedItem{SimpleSku: item.SimpleSku, Reason: "out_of_stock"})
		}
	}

	return CartSubmissionResult{AddedItems: added, SkippedItems: skipped}, nil
}
