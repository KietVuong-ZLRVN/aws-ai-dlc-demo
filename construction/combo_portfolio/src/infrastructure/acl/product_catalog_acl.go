package acl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/domain"
)

// ProductCatalogPort is the port for fetching product data from Unit 1.
type ProductCatalogPort interface {
	FetchProduct(ctx context.Context, configSku string) (*CatalogProduct, error)
}

// CatalogProduct is the translated product from Unit 1's API.
type CatalogProduct struct {
	ConfigSku string
	Name      string
	Price     float64
	ImageURL  string
	Variants  []CatalogVariant
}

type CatalogVariant struct {
	SimpleSku string
	InStock   bool
}

// HTTPProductCatalogACL calls Unit 1's product API and translates the response.
type HTTPProductCatalogACL struct {
	baseURL         string
	contentLanguage string
	httpClient      *http.Client
}

func NewHTTPProductCatalogACL(baseURL, contentLanguage string, timeoutMs int) *HTTPProductCatalogACL {
	return &HTTPProductCatalogACL{
		baseURL:         baseURL,
		contentLanguage: contentLanguage,
		httpClient:      &http.Client{Timeout: time.Duration(timeoutMs) * time.Millisecond},
	}
}

type catalogAPIResponse struct {
	ConfigSku string   `json:"configSku"`
	Name      string   `json:"name"`
	Price     float64  `json:"price"`
	Images    []string `json:"images"`
	Variants  []struct {
		SimpleSku string `json:"simpleSku"`
		InStock   bool   `json:"inStock"`
	} `json:"variants"`
}

func (a *HTTPProductCatalogACL) FetchProduct(ctx context.Context, configSku string) (*CatalogProduct, error) {
	url := fmt.Sprintf("%s/api/v1/products/%s", a.baseURL, configSku)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Language", a.contentLanguage)
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch product %s: %w", configSku, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product catalog returned %d for %s", resp.StatusCode, configSku)
	}

	var payload catalogAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode product response: %w", err)
	}

	product := &CatalogProduct{
		ConfigSku: payload.ConfigSku,
		Name:      payload.Name,
		Price:     payload.Price,
	}
	if len(payload.Images) > 0 {
		product.ImageURL = payload.Images[0]
	}
	for _, v := range payload.Variants {
		product.Variants = append(product.Variants, CatalogVariant{
			SimpleSku: v.SimpleSku,
			InStock:   v.InStock,
		})
	}
	return product, nil
}

// ComboEnrichmentService uses the ProductCatalogPort to enrich combo item snapshots.
type ComboEnrichmentService struct {
	catalogPort ProductCatalogPort
}

func NewComboEnrichmentService(port ProductCatalogPort) *ComboEnrichmentService {
	return &ComboEnrichmentService{catalogPort: port}
}

// Enrich returns an EnrichedCombo with live catalog data for each item.
// Falls back to snapshot data if the catalog is unavailable for a specific SKU.
func (s *ComboEnrichmentService) Enrich(ctx context.Context, combo *domain.Combo) (*domain.EnrichedCombo, error) {
	// Collect unique configSkus and fetch concurrently.
	type result struct {
		configSku string
		product   *CatalogProduct
		err       error
	}

	unique := map[string]struct{}{}
	for _, item := range combo.Items() {
		unique[item.ConfigSku] = struct{}{}
	}

	resultCh := make(chan result, len(unique))
	var wg sync.WaitGroup

	for sku := range unique {
		wg.Add(1)
		go func(sku string) {
			defer wg.Done()
			p, err := s.catalogPort.FetchProduct(ctx, sku)
			resultCh <- result{configSku: sku, product: p, err: err}
		}(sku)
	}
	wg.Wait()
	close(resultCh)

	products := make(map[string]*CatalogProduct)
	for r := range resultCh {
		if r.err == nil && r.product != nil {
			products[r.configSku] = r.product
		}
	}

	// Build variant index for quick inStock lookup.
	variantInStock := make(map[string]bool)
	for _, p := range products {
		for _, v := range p.Variants {
			variantInStock[v.SimpleSku] = v.InStock
		}
	}

	enriched := &domain.EnrichedCombo{
		ID:         combo.ID(),
		ShopperID:  combo.ShopperID(),
		Name:       combo.Name(),
		Visibility: combo.Visibility(),
		ShareToken: combo.ShareToken(),
		Items:      make([]domain.EnrichedComboItem, len(combo.Items())),
	}

	for i, item := range combo.Items() {
		ei := domain.EnrichedComboItem{
			ConfigSku: item.ConfigSku,
			SimpleSku: item.SimpleSku,
			ImageUrl:  item.ImageUrl,
		}
		if p, ok := products[item.ConfigSku]; ok {
			ei.Name = p.Name
			ei.Price = p.Price
			ei.ImageUrl = p.ImageURL
		} else {
			// Fallback to snapshot
			ei.Name = item.Name
			ei.Price = item.Price
			ei.CatalogUnavailable = true
		}
		ei.InStock = variantInStock[item.SimpleSku]
		enriched.Items[i] = ei
	}

	return enriched, nil
}
