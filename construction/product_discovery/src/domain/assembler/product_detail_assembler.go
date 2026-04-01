package assembler

import (
	"product_discovery/domain/port"
	"product_discovery/domain/readmodel"
)

// ProductDetailAssembler assembles a ProductDetailReadModel from a raw platform product.
type ProductDetailAssembler struct{}

// NewProductDetailAssembler returns a new ProductDetailAssembler.
func NewProductDetailAssembler() *ProductDetailAssembler {
	return &ProductDetailAssembler{}
}

// Assemble maps a PlatformProduct into a ProductDetailReadModel.
func (a *ProductDetailAssembler) Assemble(p port.PlatformProduct) readmodel.ProductDetailReadModel {
	currency := p.Currency
	if currency == "" {
		currency = "SGD"
	}

	variants := make([]readmodel.ProductVariantReadModel, 0, len(p.Simples))
	productInStock := false

	for _, s := range p.Simples {
		inStock := s.Quantity > 0
		if inStock {
			productInStock = true
		}
		variants = append(variants, readmodel.ProductVariantReadModel{
			SimpleSku: s.SimpleSku,
			Size:      s.Size,
			Color:     s.Color,
			InStock:   inStock,
		})
	}

	return readmodel.ProductDetailReadModel{
		ConfigSku: p.ConfigSku,
		Name:      p.Name,
		Brand:     p.Brand,
		Price:     readmodel.Money{Amount: p.Price, Currency: currency},
		ImageUrl:  p.ImageUrl,
		Slug:      p.UrlKey,
		InStock:   productInStock,
		Colors:    p.Colors,
		Variants:  variants,
		Occasions: p.Occasions,
	}
}
