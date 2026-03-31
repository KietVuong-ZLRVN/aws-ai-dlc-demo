# Unit 1: Product Discovery

## Purpose

Enables any user — authenticated or not — to browse, search, filter, and view products. This is the entry point to the application. The service aggregates data from multiple platform APIs into single, client-ready responses following the Backend for Frontend philosophy.

## Implementation Note

This unit is implemented independently. All information needed to build it is contained in this file. Do not depend on other unit files being built first — Unit 1 has no inter-unit dependencies.

## Scope

| Layer | Responsibility |
|---|---|
| Service API | Exposes product listing and detail endpoints to the client |
| Client UI | Product listing page, search bar, filter panel, product detail page |

---

## Platform APIs This Unit Calls

All platform requests require:
- Header: `Content-Language: en-SG` (or appropriate locale)
- Header: `Accept: application/json`
- Auth: Not required for these endpoints

### `GET /v1/products/list`

Fetches product results.

Key query params: `query` (string), `categoryId` (int), `colors[]` (multi), `price` (range string e.g. `0-200`), `occasion` (multi), `offset` (int), `limit` (int)

Response shape:
```json
{
  "numProductFound": 150,
  "products": [
    {
      "config_sku": "string",
      "name": "string",
      "brand": "string",
      "price": 99.90,
      "image": "url",
      "sizes": [{ "sku": "string", "size": "M", "stock": 5 }],
      "attributes": { "color": ["black"], "occasion": ["casual"] }
    }
  ]
}
```

### `GET /v1/products/filter`

Same query params as `/list`. Returns available filter facets instead of products.

Response shape:
```json
{
  "filters": {
    "colors": [{ "id": "black", "label": "Black", "count": 50 }],
    "occasions": [{ "id": "casual", "label": "Casual", "count": 120 }],
    "price": { "min": 10, "max": 500 }
  }
}
```

### `GET /v1/products/{configSkuOrSlug}/details`

Fetches full product detail.

Response shape:
```json
{
  "config_sku": "string",
  "url_key": "string",
  "name": "string",
  "brand": "string",
  "description": "string",
  "price": 99.90,
  "images": ["url1", "url2"],
  "simples": [
    { "sku": "string", "size": "M", "color": "black", "quantity": 5 }
  ],
  "attributes": { "occasion": ["casual"], "category": "clothing" }
}
```

---

## User Stories

### US-101 — Browse Catalog Without Login

**As a** guest shopper,
**I want to** browse products across fashion/clothing, accessories, and beauty categories without logging in,
**So that** I can freely explore the catalog before committing to an account.

**API endpoint:** `GET /api/v1/products`

**Acceptance Criteria:**

- All product listings are visible to unauthenticated users.
- No login prompt is shown during browsing.
- Browsing is available by category, search term, and filter.

---

### US-102 — Search and Filter Products

**As a** guest or registered shopper,
**I want to** search and filter products by keyword, category, color, price range, and occasion,
**So that** I can quickly narrow down products relevant to my needs.

**API endpoint:** `GET /api/v1/products`

**Acceptance Criteria:**

- Keyword search matches product name, description, and tags.
- Filters include: category (clothing, accessories, beauty), color, price range, and occasion (e.g., casual, formal, outdoor).
- Filters can be combined and cleared independently.
- Available filter facets are returned alongside results in the same response (service merges `/list` + `/filter` calls).

---

### US-103 — View Product Details

**As a** guest or registered shopper,
**I want to** view detailed product information including images, description, price, available sizes, and color variants,
**So that** I can make an informed decision before adding it to my wishlist.

**API endpoint:** `GET /api/v1/products/{configSku}`

**Acceptance Criteria:**

- Product detail page shows: multiple images, full description, price, size options, and available colors.
- Out-of-stock variants are visually indicated and cannot be added to the wishlist.

---

## Service API This Unit Exposes

### `GET /api/v1/products`

List and search products. Merges `/v1/products/list` and `/v1/products/filter` into one response.

**Auth:** Not required

**Query Parameters:**

| Param | Type | Required | Notes |
|---|---|---|---|
| `query` | string | No | Free-text keyword search |
| `categoryId` | integer | No | Category filter |
| `colors[]` | string (multi) | No | Color filter |
| `price` | string | No | Range, e.g. `0-200` |
| `occasion` | string (multi) | No | e.g. `casual`, `formal` |
| `offset` | integer | No | Default 0 |
| `limit` | integer | No | Default 20 |

**Response `200 OK`:**

```json
{
  "total": 150,
  "items": [
    {
      "configSku": "string",
      "name": "string",
      "brand": "string",
      "price": 99.90,
      "imageUrl": "string",
      "inStock": true,
      "colors": ["black", "white"],
      "occasions": ["casual"]
    }
  ],
  "filters": {
    "colors": ["black", "white", "red"],
    "occasions": ["casual", "formal"],
    "priceRange": { "min": 10, "max": 500 }
  }
}
```

---

### `GET /api/v1/products/{configSku}`

Fetch full product detail.

**Auth:** Not required

**Path Parameters:**

| Param | Type | Required |
|---|---|---|
| `configSku` | string | Yes |

**Response `200 OK`:**

```json
{
  "configSku": "string",
  "slug": "string",
  "name": "string",
  "brand": "string",
  "description": "string",
  "price": 99.90,
  "images": ["url1", "url2"],
  "variants": [
    {
      "simpleSku": "string",
      "size": "M",
      "color": "black",
      "inStock": true
    }
  ],
  "occasions": ["casual"],
  "category": "clothing"
}
```

**Response `404 Not Found`:** Product does not exist.
