package assembler

import (
	"product_discovery/domain/port"
	"product_discovery/domain/readmodel"
)

// ProductListAssembler assembles a ProductListReadModel from raw platform payloads.
type ProductListAssembler struct{}

// NewProductListAssembler returns a new ProductListAssembler.
func NewProductListAssembler() *ProductListAssembler {
	return &ProductListAssembler{}
}

// Assemble maps a raw list payload and a raw filter payload into a ProductListReadModel.
func (a *ProductListAssembler) Assemble(
	listPayload *port.RawProductListPayload,
	filterPayload *port.RawFilterPayload,
) readmodel.ProductListReadModel {
	summaries := make([]readmodel.ProductSummaryReadModel, 0, len(listPayload.Products))
	for _, p := range listPayload.Products {
		summaries = append(summaries, assembleSummary(p))
	}

	facets := readmodel.FilterFacetsReadModel{
		Colors:     filterPayload.Colors,
		Categories: filterPayload.Categories,
		PriceRange: readmodel.PriceRangeFacet{
			Min: filterPayload.MinPrice,
			Max: filterPayload.MaxPrice,
		},
	}

	return readmodel.ProductListReadModel{
		Products: summaries,
		Total:    listPayload.TotalCount,
		Filters:  facets,
	}
}

// assembleSummary maps a single PlatformProduct to a ProductSummaryReadModel.
func assembleSummary(p port.PlatformProduct) readmodel.ProductSummaryReadModel {
	currency := p.Currency
	if currency == "" {
		currency = "SGD"
	}
	return readmodel.ProductSummaryReadModel{
		ConfigSku: p.ConfigSku,
		Name:      p.Name,
		Brand:     p.Brand,
		Price:     readmodel.Money{Amount: p.Price, Currency: currency},
		ImageUrl:  p.ImageUrl,
		Slug:      p.UrlKey,
		InStock:   anyInStock(p.Simples),
		Colors:    p.Colors,
	}
}

// anyInStock returns true if at least one simple has Quantity > 0.
func anyInStock(simples []port.PlatformSimple) bool {
	for _, s := range simples {
		if s.Quantity > 0 {
			return true
		}
	}
	return false
}
