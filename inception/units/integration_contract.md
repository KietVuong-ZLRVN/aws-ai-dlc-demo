# Integration Contract

This document defines the API surface each unit exposes, the contracts between units, and their dependencies on existing platform APIs.

## Architecture Philosophy

All units follow the **Backend for Frontend (BFF)** pattern as a philosophy: each backend service aggregates, transforms, and shapes data from upstream platform APIs into responses optimised for the client's needs. This reduces the number of round trips the client must make, removes data-shaping logic from the frontend, and keeps each service responsible for a clear domain.

The client calls each unit's service API directly — it does not call platform APIs.

## Unit Dependency Map

```
Client
  │
  ├── Unit 1: Product Discovery Service  ──► Platform Product API
  │
  ├── Unit 2: Wishlist Service  ──────────► Platform Wishlist API
  │             │                           Platform Auth API
  │             │
  │   (wishlist items read by)
  │             ▼
  ├── Unit 3: AI Styling Engine  ─────────► Platform Product API
  │             │                           Platform Complete-the-Look API
  │             │                           Unit 2: GET /api/v1/wishlist
  │             │
  │   (add suggested items)               (combo saved by)
  │             ▼                               ▼
  │         Unit 2 API                   Unit 4: Combo Portfolio
  │                                            │
  │                              (resolve saved combo items)
  │                                            ▼
  └── Unit 5: Cart Handoff  ─────────────► Platform Cart API (bulk)
                                              Unit 4: GET /api/v1/combos/{id}
```

---

## Unit 1: Product Discovery

### Service API

#### `GET /api/v1/products`

List and search products. Aggregates results from the platform product and filter APIs into a single response shaped for the listing UI.

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

**Response:**

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

**Upstream platform calls:** `GET /v1/products/list`, `GET /v1/products/filter`

---

#### `GET /api/v1/products/{configSku}`

Fetch full product detail, shaped for the product detail page.

**Path Parameters:**

| Param | Type | Required |
|---|---|---|
| `configSku` | string | Yes |

**Response:**

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

**Upstream platform call:** `GET /v1/products/{configSkuOrSlug}/details`

---

## Unit 2: Wishlist Management

### Service API

#### `GET /api/v1/wishlist`

Get the authenticated shopper's wishlist, enriched with product data in a single response.

**Auth:** Required (403 if unauthenticated)

**Query Parameters:**

| Param | Type | Required | Notes |
|---|---|---|---|
| `offset` | integer | No | Default 0 |
| `limit` | integer | No | Default 20 |

**Response:**

```json
{
  "totalCount": 12,
  "items": [
    {
      "itemId": "string",
      "simpleSku": "string",
      "configSku": "string",
      "name": "string",
      "brand": "string",
      "price": 99.90,
      "imageUrl": "string",
      "color": "black",
      "size": "M",
      "inStock": true
    }
  ]
}
```

**Upstream platform call:** `GET /v1/wishLists/wishlist`

---

#### `POST /api/v1/wishlist/items`

Add an item to the wishlist.

**Auth:** Required

**Request Body:**

```json
{
  "simpleSku": "string"
}
```

**Response:** `201 Created`

```json
{
  "itemId": "string",
  "simpleSku": "string",
  "configSku": "string"
}
```

**Upstream platform call:** `POST /v1/wishLists/items`

---

#### `DELETE /api/v1/wishlist/items/{configSku}`

Remove all variants of a product from the wishlist.

**Auth:** Required

**Response:** `200 OK` (empty body)

**Upstream platform call:** `DELETE /v1/wishLists/items/?configSku={configSku}`

---

## Unit 3: AI Styling Engine

### Service API

#### `GET /api/v1/style/preferences/options`

Returns the predefined options for the preference input form.

**Auth:** Required

**Response:**

```json
{
  "occasions": ["casual", "formal", "outdoor", "beach", "office", "party"],
  "styles": ["minimalist", "bold", "classic", "bohemian"],
  "colors": ["black", "white", "navy", "beige", "red", "green", "pastel"]
}
```

---

#### `POST /api/v1/style/preferences/confirm`

Submits preferences and returns an AI-generated natural-language summary for the shopper to review before triggering combo generation.

**Auth:** Required

**Request Body:**

```json
{
  "occasions": ["casual", "beach"],
  "styles": ["minimalist"],
  "budget": { "min": 50, "max": 200 },
  "colors": { "preferred": ["beige", "white"], "excluded": ["black"] },
  "freeText": "Something light and airy for a summer trip"
}
```

**Response:**

```json
{
  "summary": "You're looking for a minimalist, casual beach look between $50–$200 in light, airy tones — perfect for a summer trip.",
  "preferences": { "occasions": ["casual", "beach"], "styles": ["minimalist"], "budget": { "min": 50, "max": 200 }, "colors": { "preferred": ["beige", "white"], "excluded": ["black"] }, "freeText": "Something light and airy for a summer trip" }
}
```

---

#### `POST /api/v1/style/combos/generate`

Core AI endpoint. Fetches the shopper's wishlist server-side, runs AI inference, and returns ranked combo suggestions shaped for the client.

**Auth:** Required

**Request Body:**

```json
{
  "preferences": {
    "occasions": ["casual"],
    "styles": ["minimalist"],
    "budget": { "min": 0, "max": 200 },
    "colors": { "preferred": ["beige"], "excluded": [] },
    "freeText": "optional free text"
  },
  "excludeComboIds": ["combo-id-1", "combo-id-2"]
}
```

- `preferences` is fully optional. Omitting it triggers quick-generate (US-301).
- `excludeComboIds` is optional. Pass previously shown combo IDs to prevent repeats (US-405).
- Wishlist is resolved server-side using the shopper's session — the client does not need to send it.

**Response — success:**

```json
{
  "status": "ok",
  "combos": [
    {
      "id": "combo-id-3",
      "reasoning": "This linen blazer pairs well with the wide-leg trousers for a smart-casual summer look.",
      "items": [
        {
          "configSku": "string",
          "simpleSku": "string",
          "name": "string",
          "brand": "string",
          "price": 89.90,
          "imageUrl": "string",
          "source": "wishlist"
        },
        {
          "configSku": "string",
          "simpleSku": "string",
          "name": "string",
          "brand": "string",
          "price": 59.90,
          "imageUrl": "string",
          "source": "catalog"
        }
      ]
    }
  ]
}
```

- `source`: `"wishlist"` — item was on the shopper's wishlist; `"catalog"` — item was sourced from the product catalog to complete the combo.

**Response — fallback (no suitable combo can be formed):**

```json
{
  "status": "fallback",
  "message": "Your wishlist items don't have a color match for a formal occasion. Here are some suggestions:",
  "alternatives": [
    {
      "configSku": "string",
      "simpleSku": "string",
      "name": "string",
      "brand": "string",
      "price": 79.90,
      "imageUrl": "string",
      "reason": "Replaces your current blazer with a formal-friendly option"
    }
  ]
}
```

**Upstream calls:** `GET /api/v1/wishlist` (Unit 2), `GET /v1/products/list`, `GET /v1/recommendation/completethelook/{config_sku}`

---

## Unit 4: Combo Portfolio

### Service API

#### `POST /api/v1/combos`

Save an AI-generated combo.

**Auth:** Required

**Request Body:**

```json
{
  "name": "My Summer Look",
  "items": [
    { "configSku": "string", "simpleSku": "string", "name": "string", "imageUrl": "string", "price": 89.90 }
  ],
  "visibility": "private"
}
```

- `visibility`: `"public"` or `"private"` (default `"private"`)

**Response:** `201 Created`

```json
{
  "id": "combo-uuid",
  "name": "My Summer Look",
  "createdAt": "2026-03-31T10:00:00Z",
  "visibility": "private",
  "shareToken": null
}
```

---

#### `GET /api/v1/combos`

List all saved combos for the authenticated shopper.

**Auth:** Required

**Response:**

```json
{
  "combos": [
    {
      "id": "combo-uuid",
      "name": "My Summer Look",
      "createdAt": "2026-03-31T10:00:00Z",
      "visibility": "private",
      "itemThumbnails": ["url1", "url2", "url3"]
    }
  ]
}
```

---

#### `GET /api/v1/combos/{id}`

Get full detail of a saved combo.

**Auth:** Required

**Response:**

```json
{
  "id": "combo-uuid",
  "name": "My Summer Look",
  "createdAt": "2026-03-31T10:00:00Z",
  "visibility": "private",
  "shareToken": null,
  "items": [
    { "configSku": "string", "simpleSku": "string", "name": "string", "imageUrl": "string", "price": 89.90 }
  ]
}
```

---

#### `PUT /api/v1/combos/{id}`

Rename a combo or change its visibility.

**Auth:** Required

**Request Body:**

```json
{
  "name": "My Updated Look",
  "visibility": "public"
}
```

**Response:** `200 OK` — updated combo object (same shape as GET response)

---

#### `DELETE /api/v1/combos/{id}`

Delete a saved combo.

**Auth:** Required

**Response:** `204 No Content`

---

#### `POST /api/v1/combos/{id}/share`

Generate a public share token for a combo.

**Auth:** Required

**Response:**

```json
{
  "shareToken": "abc123xyz",
  "shareUrl": "https://app.example.com/combos/shared/abc123xyz"
}
```

---

#### `GET /api/v1/combos/shared/{token}`

Public endpoint — view a shared combo without authentication.

**Auth:** None required

**Response:**

```json
{
  "name": "My Summer Look",
  "items": [
    { "configSku": "string", "name": "string", "imageUrl": "string", "price": 89.90 }
  ]
}
```

Returns `404` if the token does not exist or the combo visibility is `"private"`.

---

## Unit 5: Cart Handoff

### Service API

#### `POST /api/v1/cart/combo`

Resolves combo items server-side and adds them to the cart in a single bulk call. Handles both unsaved combos (inline items from Unit 3) and saved combos (resolved by ID from Unit 4).

**Auth:** Required

**Request Body:**

```json
{
  "comboId": "combo-uuid"
}
```

Or for an unsaved combo from the AI suggestion view:

```json
{
  "items": [
    { "simpleSku": "string", "quantity": 1, "size": "M" }
  ]
}
```

- Provide either `comboId` or `items` — not both.

**Response — success:**

```json
{
  "status": "ok",
  "addedItems": ["simpleSku1", "simpleSku2"],
  "skippedItems": []
}
```

**Response — partial (some items out of stock):**

```json
{
  "status": "partial",
  "addedItems": ["simpleSku1"],
  "skippedItems": [
    { "simpleSku": "simpleSku2", "reason": "out_of_stock" }
  ]
}
```

**Upstream calls:** `GET /api/v1/combos/{id}` (Unit 4, when `comboId` is provided), `POST /v1/checkout/cart/bulk`

---

## Cross-Unit Dependency Summary

| Consumer | Depends On | Endpoint |
|---|---|---|
| Unit 2 (Wishlist) | Platform Auth | `POST /v1/customers/login` |
| Unit 3 (AI Engine) | Unit 2 | `GET /api/v1/wishlist` |
| Unit 3 (AI Engine) | Platform Product | `GET /v1/products/list` |
| Unit 3 (AI Engine) | Platform Recommendation | `GET /v1/recommendation/completethelook/{sku}` |
| Unit 3 client → Unit 2 | Unit 2 | `POST /api/v1/wishlist/items` (add suggested items) |
| Unit 5 (Cart Handoff) | Unit 4 | `GET /api/v1/combos/{id}` |
| Unit 5 (Cart Handoff) | Platform Cart | `POST /v1/checkout/cart/bulk` |
