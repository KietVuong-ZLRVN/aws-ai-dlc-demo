# Domain Model — Unit 1: Product Discovery

## Bounded Context

**Name:** Product Discovery

**Nature:** Read-only CQRS query context. This bounded context owns no state and performs no writes. All product data originates from the upstream Platform Product API. The domain model is therefore structured as a query-side / read model: Query objects, Read Models, Response Assemblers, and an Anti-Corruption Layer (ACL). There are no Aggregates, no Repositories with write contracts, and no local persistence.

**Context Map Position:**
- **Upstream:** Platform Product API (External System — Conformist relationship translated via ACL)
- **Downstream consumers:** Unit 3 AI Styling Engine (product catalogue lookups), Client UI

---

## Domain Concepts

### Query Objects

Query objects represent the intent of a specific read operation. They carry the input criteria and are handled by a corresponding Query Handler.

---

#### `ProductListQuery`

Encapsulates the parameters needed to retrieve a paginated, filtered list of products.

**Fields:**
- `keyword` — optional free-text search term
- `categoryId` — optional integer category identifier
- `colorFilters` — optional list of color identifiers
- `priceRange` — optional `PriceRange` value object (see below)
- `occasionFilters` — optional list of occasion identifiers
- `pagination` — `Pagination` value object (offset, limit)

**Handled by:** `ProductListQueryHandler`

---

#### `ProductDetailQuery`

Encapsulates the identifier needed to retrieve full detail for a single product.

**Fields:**
- `configSku` — the unique product configuration SKU

**Handled by:** `ProductDetailQueryHandler`

---

#### `ProductFilterQuery`

Encapsulates the same filtering parameters as `ProductListQuery`. Used when the client needs to retrieve available filter facets for the current search context. In practice, `ProductListQueryHandler` calls both the list and filter platform endpoints and merges the results, so this query object is used internally by the handler rather than as a separate public query.

---

### Read Models (Projections)

Read models are flat, immutable data structures returned to the client. They are shaped specifically for the UI and carry no behaviour.

---

#### `ProductSummaryReadModel`

Represents a single product in a listing result. Shaped for the product card in the listing UI.

**Fields:**
- `configSku`
- `name`
- `brand`
- `price` — `Money` value object
- `imageUrl`
- `inStock` — boolean, derived from variant stock data
- `colors` — list of color strings
- `occasions` — list of occasion strings

---

#### `ProductListReadModel`

The top-level response for a product listing query. Combines product items with filter facets in a single client-ready object.

**Fields:**
- `total` — total number of matching products
- `items` — list of `ProductSummaryReadModel`
- `filters` — `FilterFacetsReadModel`

---

#### `FilterFacetsReadModel`

Represents the available filter options for the current query context.

**Fields:**
- `colors` — list of `FilterOption` value objects
- `occasions` — list of `FilterOption` value objects
- `priceRange` — `PriceRange` value object

---

#### `ProductDetailReadModel`

Represents the full detail of a single product. Shaped for the product detail page.

**Fields:**
- `configSku`
- `slug`
- `name`
- `brand`
- `description`
- `price` — `Money` value object
- `images` — ordered list of image URLs
- `variants` — list of `ProductVariantReadModel`
- `occasions` — list of occasion strings
- `category`

---

#### `ProductVariantReadModel`

Represents a single purchasable variant (specific size/color combination) of a product.

**Fields:**
- `simpleSku`
- `size`
- `color`
- `inStock` — boolean

---

### Value Objects

Value objects are immutable, identity-less domain concepts used within queries and read models.

---

#### `PriceRange`

Represents a monetary price range used for filtering.

**Fields:**
- `min` — minimum price (non-negative decimal)
- `max` — maximum price (greater than or equal to `min`)

**Invariants:**
- `min` must be non-negative.
- `max` must be greater than or equal to `min`.

---

#### `Money`

Represents a price amount. Currency is implied by the service's locale context (SGD for `en-SG`).

**Fields:**
- `amount` — decimal, non-negative

---

#### `Pagination`

Represents paging parameters for list queries.

**Fields:**
- `offset` — integer, default 0, non-negative
- `limit` — integer, default 20, positive

---

#### `FilterOption`

Represents a single selectable filter facet returned alongside product results.

**Fields:**
- `id` — machine-readable identifier (e.g., `"black"`)
- `label` — human-readable display label (e.g., `"Black"`)
- `count` — number of products matching this facet in the current query

---

### Domain Services (Query Handlers)

Query Handlers orchestrate the retrieval and assembly of read models. They call the ACL to fetch platform data and delegate to Response Assemblers to shape the output.

---

#### `ProductListQueryHandler`

Handles `ProductListQuery`.

**Responsibilities:**
1. Translates the query into platform API parameters via the ACL.
2. Calls both `GET /v1/products/list` and `GET /v1/products/filter` concurrently via the ACL.
3. Delegates to `ProductListAssembler` to merge and shape the response into `ProductListReadModel`.

---

#### `ProductDetailQueryHandler`

Handles `ProductDetailQuery`.

**Responsibilities:**
1. Calls `GET /v1/products/{configSku}/details` via the ACL.
2. Returns `404` domain result if the platform returns no product.
3. Delegates to `ProductDetailAssembler` to shape the response into `ProductDetailReadModel`.

---

### Response Assemblers

Assemblers transform raw platform API responses (as received through the ACL) into read models. They contain the data-shaping logic and keep it out of the query handlers.

---

#### `ProductListAssembler`

Merges the platform list response and filter response into a single `ProductListReadModel`.

**Responsibilities:**
- Maps each platform product to a `ProductSummaryReadModel`.
- Derives `inStock` by checking whether at least one variant has stock greater than zero.
- Maps platform filter facets to `FilterFacetsReadModel` and `FilterOption` value objects.

---

#### `ProductDetailAssembler`

Transforms the platform product detail response into a `ProductDetailReadModel`.

**Responsibilities:**
- Maps platform `simples` array to a list of `ProductVariantReadModel`.
- Derives `inStock` per variant from the `quantity` field.
- Maps platform field names to the service's canonical field names (e.g., `config_sku` → `configSku`, `url_key` → `slug`).

---

### Anti-Corruption Layer (ACL)

The ACL is the boundary between this bounded context and the external Platform Product API. It translates between the platform's data model and the service's internal concepts, protecting the domain from changes in the upstream system.

---

#### `PlatformProductApiClient` (ACL interface)

The ACL interface defines the operations this bounded context needs from the platform, expressed in the service's own language.

**Operations:**
- `fetchProductList(params)` → raw platform list payload
- `fetchProductFilters(params)` → raw platform filter payload
- `fetchProductDetail(configSku)` → raw platform detail payload, or null if not found

**Responsibilities:**
- Attaches required platform headers (`Content-Language`, `Accept`).
- Handles platform-level HTTP errors and translates them to domain-level results (e.g., platform `404` → null return, allowing the query handler to produce a domain `ProductNotFound` result).
- Does not perform data shaping — raw platform payloads are returned to assemblers.

---

## Domain Events

This bounded context produces **no domain events**. It is a pure query context with no state changes. Downstream units (e.g., Unit 2 Wishlist) react to user intent, not to product discovery events.

---

## Policies

**No stateful policies** apply in this context. The following rules are enforced at the query/assembly layer:

- **Out-of-stock visibility rule:** Out-of-stock variants must be included in product detail responses but flagged with `inStock: false`. They are not filtered out — the UI is responsible for visual indication and preventing wishlist add.
- **Unauthenticated access rule:** All product listing and detail queries are permitted without authentication. No session check is performed in this context.

---

## Summary Table

| Concept | Type | Notes |
|---|---|---|
| `ProductListQuery` | Query Object | Input for product listing |
| `ProductDetailQuery` | Query Object | Input for product detail |
| `ProductListReadModel` | Read Model | Merged list + filters response |
| `ProductSummaryReadModel` | Read Model | Per-item card data |
| `FilterFacetsReadModel` | Read Model | Available filter options |
| `ProductDetailReadModel` | Read Model | Full product detail |
| `ProductVariantReadModel` | Read Model | Per-variant (size/color/stock) |
| `PriceRange` | Value Object | Filter range, also in detail |
| `Money` | Value Object | Price amount |
| `Pagination` | Value Object | Offset + limit |
| `FilterOption` | Value Object | Facet option with count |
| `ProductListQueryHandler` | Domain Service (Query Handler) | Orchestrates list + filter fetch |
| `ProductDetailQueryHandler` | Domain Service (Query Handler) | Orchestrates detail fetch |
| `ProductListAssembler` | Response Assembler | Shapes platform list response |
| `ProductDetailAssembler` | Response Assembler | Shapes platform detail response |
| `PlatformProductApiClient` | ACL Interface | Translates to/from platform API |
