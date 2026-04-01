package main

import (
	"context"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/domain"
	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/infrastructure/acl"
)

// stubProductCatalogACL fakes Unit 1's product API for local demo purposes.
type stubProductCatalogACL struct{}

var stubProducts = map[string]*acl.CatalogProduct{
	"CONFIG-SKU-001": {
		ConfigSku: "CONFIG-SKU-001",
		Name:      "Classic White T-Shirt",
		Price:     19.99,
		ImageURL:  "https://example.com/images/config-sku-001.jpg",
		Variants: []acl.CatalogVariant{
			{SimpleSku: "SIMPLE-SKU-001-S", InStock: true},
			{SimpleSku: "SIMPLE-SKU-001-M", InStock: true},
		},
	},
	"CONFIG-SKU-002": {
		ConfigSku: "CONFIG-SKU-002",
		Name:      "Slim Fit Blue Jeans",
		Price:     59.99,
		ImageURL:  "https://example.com/images/config-sku-002.jpg",
		Variants: []acl.CatalogVariant{
			{SimpleSku: "SIMPLE-SKU-002-30", InStock: true},
			{SimpleSku: "SIMPLE-SKU-002-32", InStock: true},
		},
	},
	"CONFIG-SKU-003": {
		ConfigSku: "CONFIG-SKU-003",
		Name:      "White Canvas Sneakers",
		Price:     89.99,
		ImageURL:  "https://example.com/images/config-sku-003.jpg",
		Variants: []acl.CatalogVariant{
			{SimpleSku: "SIMPLE-SKU-003-42", InStock: true},
			{SimpleSku: "SIMPLE-SKU-003-43", InStock: false},
		},
	},
}

func (s *stubProductCatalogACL) FetchProduct(_ context.Context, configSku string) (*acl.CatalogProduct, error) {
	if p, ok := stubProducts[configSku]; ok {
		return p, nil
	}
	return nil, nil // unknown SKU — treated as unavailable
}

// stubItems returns the sample ComboItems used in the demo.
func stubItems() []domain.ComboItem {
	return []domain.ComboItem{
		{ConfigSku: "CONFIG-SKU-001", SimpleSku: "SIMPLE-SKU-001-M", Name: "Classic White T-Shirt", Price: 19.99, ImageUrl: "https://example.com/images/config-sku-001.jpg"},
		{ConfigSku: "CONFIG-SKU-002", SimpleSku: "SIMPLE-SKU-002-30", Name: "Slim Fit Blue Jeans", Price: 59.99, ImageUrl: "https://example.com/images/config-sku-002.jpg"},
		{ConfigSku: "CONFIG-SKU-003", SimpleSku: "SIMPLE-SKU-003-42", Name: "White Canvas Sneakers", Price: 89.99, ImageUrl: "https://example.com/images/config-sku-003.jpg"},
	}
}
