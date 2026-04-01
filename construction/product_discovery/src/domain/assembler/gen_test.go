package assembler

import (
	"fmt"
	"product_discovery/domain/port"

	"pgregory.net/rapid"
)

func genPlatformSimple(t *rapid.T, label string) port.PlatformSimple {
	return port.PlatformSimple{
		SimpleSku: rapid.StringMatching(`[A-Z]{2}-[0-9]{3}-[A-Z]{1,2}`).Draw(t, label+".sku"),
		Size:      rapid.SampledFrom([]string{"XS", "S", "M", "L", "XL"}).Draw(t, label+".size"),
		Color:     rapid.SampledFrom([]string{"black", "white", "red"}).Draw(t, label+".color"),
		Quantity:  rapid.IntRange(-1, 10).Draw(t, label+".qty"),
	}
}

func genPlatformProduct(t *rapid.T, label string) port.PlatformProduct {
	n := rapid.IntRange(1, 5).Draw(t, label+".nSimples")
	simples := make([]port.PlatformSimple, n)
	for i := range simples {
		simples[i] = genPlatformSimple(t, fmt.Sprintf("%s.simple[%d]", label, i))
	}
	return port.PlatformProduct{
		ConfigSku: rapid.StringMatching(`[A-Z]{2}-[0-9]{3}`).Draw(t, label+".configSku"),
		Name:      rapid.StringMatching(`[A-Za-z ]{3,15}`).Draw(t, label+".name"),
		Brand:     rapid.StringMatching(`[A-Za-z]{3,10}`).Draw(t, label+".brand"),
		Price:     rapid.Float64Range(0, 999).Draw(t, label+".price"),
		ImageUrl:  "https://img.example.com/product.jpg",
		UrlKey:    rapid.StringMatching(`[a-z-]{3,15}`).Draw(t, label+".urlKey"),
		Currency:  "SGD",
		Occasions: []string{"casual"},
		Simples:   simples,
	}
}

func genRawProductListPayload(t *rapid.T) *port.RawProductListPayload {
	n := rapid.IntRange(0, 8).Draw(t, "nProducts")
	products := make([]port.PlatformProduct, n)
	for i := range products {
		products[i] = genPlatformProduct(t, fmt.Sprintf("product[%d]", i))
	}
	return &port.RawProductListPayload{
		Products:   products,
		TotalCount: rapid.IntRange(n, n+100).Draw(t, "totalCount"),
	}
}

func genRawFilterPayload(t *rapid.T) *port.RawFilterPayload {
	minPrice := rapid.Float64Range(0, 500).Draw(t, "filterMin")
	maxPrice := rapid.Float64Range(minPrice, 1000).Draw(t, "filterMax")
	nColors := rapid.IntRange(0, 5).Draw(t, "nColors")
	colors := make([]string, nColors)
	palette := []string{"black", "white", "red", "blue", "green"}
	for i := range colors {
		colors[i] = palette[i%len(palette)]
	}
	return &port.RawFilterPayload{
		Colors:   colors,
		MinPrice: minPrice,
		MaxPrice: maxPrice,
	}
}
