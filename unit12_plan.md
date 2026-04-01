# Plan: DDD Domain Model & Logical Design for Unit 1 (Product Discovery) & Unit 2 (Wishlist)

## Scope
Design DDD tactical domain models for:
- **Unit 1 — Product Discovery** (`/construction/product_discovery/domain_model.md`)
- **Unit 2 — Wishlist Management** (`/construction/wishlist/domain_model.md`)

No code snippets will be generated. Each domain model will cover: Aggregates, Entities, Value Objects, Domain Events, Policies, Repositories, and Domain Services.

---

## Questions Requiring Clarification

[Question] Unit 1 (Product Discovery) is a BFF that proxies/aggregates read-only data from the Platform Product API. It has no owned state of its own — no writes, no persistence. In a strict DDD sense, a pure read/query layer typically does not need a rich domain model (aggregates, repositories, etc.) and is better served by a thin Query/Read model. Should I:
  - (a) Model it with a full DDD tactical design treating `Product` and `ProductVariant` as domain concepts with value objects and domain services even though they are never written to?
  - (b) Apply a lighter CQRS-style read model — named Query objects, Response Assemblers, and Anti-Corruption Layer — that is more honest about the unit's nature?
  - (c) Do both: define the domain concepts (entities/VOs) that represent the product catalogue view, but acknowledge no repository writes and no aggregates own state?

[Answer] (b)

[Question] For Unit 2 (Wishlist), the service acts as a thin BFF over the Platform Wishlist API — wishlist state is owned and persisted by the platform, not this service. Should the Wishlist aggregate in the domain model represent:
  - (a) The local service's view of the wishlist (i.e., the service is the system of record and has its own persistence)?
  - (b) A domain facade that models intent and delegates all state changes to the platform API via an Anti-Corruption Layer, meaning repositories map to platform API calls rather than a local database?

[Answer] (b)

---

## Plan Steps

- [x] **Step 1 — Review & align on questions above** _(blocked until answers received)_

- [x] **Step 2 — Design Unit 1: Product Discovery domain model**
  - Identify bounded context and context map positioning
  - Define domain concepts: Product, ProductVariant, ProductSummary, SearchCriteria, FilterCatalog
  - Classify each as Aggregate, Entity, or Value Object
  - Identify domain services (e.g., ProductSearchService, FilterAssembler)
  - Identify domain events (if any)
  - Identify repositories and their contracts
  - Define anti-corruption layer for Platform Product API
  - Write to `/construction/product_discovery/domain_model.md`

- [x] **Step 3 — Design Unit 2: Wishlist Management domain model**
  - Identify bounded context and context map positioning
  - Define domain concepts: Wishlist, WishlistItem, ShopperId, AuthSession, PendingWishlistAdd
  - Classify each as Aggregate, Entity, or Value Object
  - Identify domain events: `WishlistItemAdded`, `WishlistItemRemoved`, `AuthenticationGateTriggered`
  - Identify policies: duplicate-prevention rule, auth-gate policy, toggle behaviour
  - Identify repositories and their contracts
  - Define anti-corruption layer for Platform Wishlist API
  - Write to `/construction/wishlist/domain_model.md`

- [x] **Step 4 — Cross-unit integration notes**
  - Document how Unit 2's Wishlist aggregate is consumed by Unit 3 (read contract)
  - Note published language / shared kernel considerations between units

- [x] **Step 5 — Review & finalise**
  - Self-review for completeness and DDD correctness
  - Ensure no code snippets have been included
  - Mark all steps done

---

## Phase 2: Logical Design

### Scope
Generate a logical design for software source code implementation for:
- **Unit 1 — Product Discovery** (`/construction/product_discovery/logical_design.md`)
- **Unit 2 — Wishlist Management** (`/construction/wishlist/logical_design.md`)

The logical design covers: layered architecture, package/module structure, component responsibilities, data flow, event-driven patterns, external integration design (ACL implementations), error handling strategy, and technology choices. No code snippets will be generated.

---

### Questions Requiring Clarification

[Question] Step 2.3 (Implement Source Code) specifies Go for the backend. For the logical design, I need to reference a specific HTTP router/framework. Which should I use?
  - (a) `net/http` (standard library only — zero dependencies)
  - (b) `gin-gonic/gin` (most widely used, middleware ecosystem)
  - (c) `go-chi/chi` (lightweight, idiomatic)
  - (d) `labstack/echo`

[Answer] (c)

[Question] For the event-driven layer (domain events emitted by the `Wishlist` aggregate), Step 2.3 says "assume event stores are in-memory". Should the logical design:
  - (a) Specify only an in-memory synchronous event bus (simple, suited for demo)
  - (b) Specify an in-memory event bus for the demo layer but describe the abstraction clearly so it can be swapped for an async broker (e.g., SNS/SQS) in a production deployment

[Answer] (a)

[Question] For dependency wiring between layers (HTTP handler → application service → domain → repository → ACL), should the design specify:
  - (a) Plain manual constructor injection (no framework)
  - (b) A wire/fx-style DI framework (e.g., Google Wire, Uber Fx)

[Answer] (a)

---

### Plan Steps

- [x] **Step 6 — Review & align on Phase 2 questions above** _(blocked until answers received)_

- [x] **Step 7 — Design logical design for Unit 1: Product Discovery**
  - Define layered architecture (API, Application, Domain, Infrastructure)
  - Specify package/directory structure
  - Map each HTTP endpoint → handler → query handler → ACL client
  - Define component responsibilities at each layer
  - Describe data flow for `GET /api/v1/products` and `GET /api/v1/products/{configSku}`
  - Define error handling strategy (platform errors → domain results → HTTP responses)
  - Define technology choices (HTTP framework, HTTP client for platform calls)
  - Write to `/construction/product_discovery/logical_design.md`

- [x] **Step 8 — Design logical design for Unit 2: Wishlist Management**
  - Define layered architecture (API, Application, Domain, Infrastructure)
  - Specify package/directory structure
  - Map each HTTP endpoint → handler → application service → domain aggregate → repository → ACL client
  - Define component responsibilities at each layer
  - Describe data flow for each of the 3 endpoints (`GET`, `POST`, `DELETE`)
  - Describe domain event dispatch: event bus interface, in-memory implementation, handlers
  - Describe auth gate flow: session extraction → `AuthSessionService` → `ShopperId` or gate trigger
  - Define error handling strategy
  - Define technology choices
  - Write to `/construction/wishlist/logical_design.md`

- [x] **Step 9 — Cross-unit integration design**
  - Describe how Unit 3 consumes Unit 2's `GET /api/v1/wishlist` (service-to-service call)
  - Note shared type candidates (`ConfigSku`, `SimpleSku`) and how they are handled without coupling

- [x] **Step 10 — Review & finalise**
  - Self-review both logical designs for completeness and consistency with domain models
  - Ensure no code snippets have been included
  - Mark all steps done

---

## Phase 3: Implementation (Source Code)

### Scope

Implement both bounded contexts as runnable Go backends + a shared React frontend:
- **Unit 1 — Product Discovery** backend in `construction/product_discovery/src/`
- **Unit 2 — Wishlist Management** backend in `construction/wishlist/src/`
- **Shared React frontend** demonstrating both systems together

All platform API clients, repositories, and event stores are in-memory (no real external calls).

---

### Questions Requiring Clarification

[Question] The two backend services run on separate ports per the logical design (Unit 1 on `:8080`, Unit 2 on `:8081`). For the demo, should they be kept as **two separate Go binaries** (closer to production topology), or **combined into one binary** that mounts both routers for simpler local startup?
  - (a) Two separate binaries — each `go run ./cmd/server` from its own directory
  - (b) One combined demo binary to simplify the `./demo.sh` startup script

[Answer] (1)

[Question] For auth simulation in the demo: the Wishlist service requires a session cookie to identify the shopper. Since there is no real auth platform, which approach should the in-memory `AuthSessionService` use?
  - (a) Always return a hardcoded `ShopperId` ("shopper-123") — effectively always authenticated; skip the auth-gate flow entirely
  - (b) Start unauthenticated; a "Login" button in the UI sets a cookie/header that the mock auth service recognises — demonstrates the auth-gate flow

[Answer] (b)

[Question] The React frontend needs to be placed somewhere. Should it live at:
  - (a) `construction/frontend/` — a sibling to the two backend units
  - (b) `construction/product_discovery/src/frontend/` — co-located with Unit 1 (product browsing is the primary view)

[Answer] (a)

[Question] For the in-memory product catalogue, should I seed a small set of **fictional** fashion products (e.g., 6–10 items with names, brands, prices, sizes, colours), or do you have a sample data fixture I should use?
  - (a) Generate simple fictional data (e.g., "Classic White Tee", "Slim Fit Jeans", etc.)
  - (b) Use a specific dataset — please provide or point to it

[Answer] (a)

---

### Plan Steps

#### Setup & Scaffolding
- [x] **Step 11 — Scaffold Go module for Product Discovery** (`construction/product_discovery/src/go.mod`)
- [x] **Step 12 — Scaffold Go module for Wishlist** (`construction/wishlist/src/go.mod`)
- [x] **Step 13 — Scaffold React app** (Vite + React at `construction/frontend/`)

#### Product Discovery Backend (Unit 1)
- [x] **Step 14 — Domain layer**: value objects (`PriceRange`, `Money`, `Pagination`), query structs (`ProductListQuery`, `ProductDetailQuery`), read models
- [x] **Step 15 — Domain assemblers**: `ProductListAssembler`, `ProductDetailAssembler`
- [x] **Step 16 — Infrastructure layer**: in-memory `PlatformProductApiClient` with seeded product data (platform types, mock client)
- [x] **Step 17 — Application layer**: `ProductListQueryHandler`, `ProductDetailQueryHandler`
- [x] **Step 18 — API layer**: `ProductListHandler`, `ProductDetailHandler`, `router.go`
- [x] **Step 19 — Entry point**: `cmd/server/main.go` wiring

#### Wishlist Backend (Unit 2)
- [x] **Step 20 — Domain layer**: value objects, `Wishlist` aggregate, `WishlistItem` entity, domain events, repository interface, `AuthSessionService` interface
- [x] **Step 21 — Domain assembler**: `WishlistAssembler`
- [x] **Step 22 — Infrastructure layer**: in-memory `WishlistRepository`, in-memory `AuthSessionService` (mock), `InMemoryEventBus`
- [x] **Step 23 — Application layer**: `GetWishlistService`, `AddWishlistItemService`, `RemoveWishlistItemService`, `EventDispatcher`
- [x] **Step 24 — API layer**: `WishlistGetHandler`, `WishlistAddHandler`, `WishlistRemoveHandler`, `router.go`
- [x] **Step 25 — Entry point**: `cmd/server/main.go` wiring

#### React Frontend
- [x] **Step 26 — Project structure & API client helpers** (`src/api.js`, `src/auth.js`, CORS allowed in both backends)
- [x] **Step 27 — Product list page**: search bar, filter panel, product grid with heart button
- [x] **Step 28 — Product detail page**: variant selector, "Add to Wishlist" button, back navigation
- [x] **Step 29 — Wishlist page**: list of saved items, remove button per item
- [x] **Step 30 — Auth gate UI**: `LoginModal` component; unauthenticated → 403 → modal → login → retry add

#### Demo & Docs
- [x] **Step 31 — Demo script** (`construction/demo.sh`): starts both backends and Vite dev server, opens browser
- [x] **Step 32 — Demo README**: see below

---

## Phase 4: Test Plan — Backend Systems (Unit 1 + Unit 2)

### Scope
Backend test coverage for:
- **Unit 1 — Product Discovery** (`construction/product_discovery/src/`)
- **Unit 2 — Wishlist Management** (`construction/wishlist/src/`)

Testing strategy: **Property-Based Testing (PBT)** as the primary method, supplemented by example-based tests for HTTP boundary behaviour. All tests are self-contained — no real platform calls.

---

### Questions Requiring Clarification

[Question] **Q1 — PBT library choice:** Standard Go PBT options are `pgregory.net/rapid` (idiomatic, built-in shrinking) and `github.com/leanovate/gopter` (richer generator combinators). This plan assumes `pgregory.net/rapid`. Is that acceptable, or do you have a preference?
[Answer] Acceptable

[Question] **Q2 — Test doubles for infrastructure:** Both units ship in-memory infrastructure stubs (`in_memory_product_client.go`, `in_memory_wishlist_repository.go`, `in_memory_auth_service.go`). The plan proposes using those stubs directly as test doubles. Do you want generated mocks instead (e.g., `github.com/vektra/mockery`), or are the existing stubs sufficient?
[Answer] Generated mocks

[Question] **Q3 — Coverage target:** Should we enforce a minimum line/branch coverage percentage (e.g., 80%)? If yes, what threshold?
[Answer] 75%

[Question] **Q4 — HTTP handler test depth:** The plan includes `net/http/httptest` handler tests that exercise the full handler → application service → in-memory infra path. Do you prefer this integration-style handler coverage, or pure unit tests for the domain layer only?
[Answer] pure unit tests

---

### Plan Steps

#### Phase 4.0 — Test Infrastructure Setup

- [x] **4.0.1** Add `pgregory.net/rapid` as a test-only dependency in `construction/product_discovery/src/go.mod` and run `go mod tidy`.
- [x] **4.0.2** Add `pgregory.net/rapid` as a test-only dependency in `construction/wishlist/src/go.mod` and run `go mod tidy`.
- [x] **4.0.3** Verify both modules build cleanly (`go build ./...`) after adding the dependency.
- [x] **4.0.4** Install `mockery` v2 (`go install github.com/vektra/mockery/v2@latest`) and add a `//go:generate` directive for each interface requiring a mock:
  - Unit 1: `PlatformProductApiClient` in `domain/port/platform_product_client.go`
  - Unit 2: `WishlistRepository` in `domain/repository/wishlist_repository.go`, `AuthSessionService` in `domain/service/auth_session_service.go`, `EventBus` in `domain/event/` (or wherever the interface is declared)
- [x] **4.0.5** Run `go generate ./...` in each module to produce mock files under `mocks/`. Confirm generated files compile.
- [x] **4.0.6** Create a shared test generator helpers file in each module (inline in `_test.go` files or a `testgen` package) defining reusable `rapid` generators: `genSimpleSku`, `genConfigSku`, `genWishlistItem`, `genWishlistWithN`, `genPlatformProduct`, `genPlatformDetailPayload`, `genFilterPayload`, `genValidPagination`.

---

#### Phase 4.1 — Unit 1: Value Object PBT Tests

**File:** `construction/product_discovery/src/domain/valueobject/price_range_test.go`

- [x] **4.1.1** PBT — **PriceRange valid construction**: For all `(min, max float64)` where `0 ≤ min ≤ max`, `NewPriceRange("min-max")` returns no error and the result has `Min == min` and `Max == max`.
- [x] **4.1.2** PBT — **PriceRange rejects invalid range**: For all `(min, max)` where `min > max`, `NewPriceRange` returns a non-nil error.
- [x] **4.1.3** PBT — **PriceRange rejects negative min**: For all `min < 0`, `NewPriceRange` with any `max` returns a non-nil error.
- [x] **4.1.4** Example — **PriceRange rejects malformed strings**: No `-` separator, empty string, single number, non-numeric values each return errors.
- [x] **4.1.5** PBT — **PriceRange round-trip**: `NewPriceRange(fmt.Sprintf("%v-%v", min, max))` produces `PriceRange{Min: min, Max: max}` for any valid `(min, max)` pair.

**File:** `construction/product_discovery/src/domain/valueobject/pagination_test.go`

- [x] **4.1.6** PBT — **Pagination valid construction**: For all `offset ≥ 0` and `limit ≥ 1`, `NewPagination` returns no error with correct field values.
- [x] **4.1.7** PBT — **Pagination rejects negative offset**: For all `offset < 0`, `NewPagination` returns error.
- [x] **4.1.8** PBT — **Pagination rejects zero/negative limit**: For all `limit ≤ 0`, `NewPagination` returns error.

**File:** `construction/product_discovery/src/domain/valueobject/money_test.go`

- [x] **4.1.9** PBT — **Money valid construction**: For all `amount ≥ 0.0`, `NewMoney` returns no error and `money.Amount == amount`.
- [x] **4.1.10** PBT — **Money rejects negative amount**: For all `amount < 0`, `NewMoney` returns error.

---

#### Phase 4.2 — Unit 1: Domain Assembler PBT Tests

**File:** `construction/product_discovery/src/domain/assembler/product_list_assembler_test.go`

- [x] **4.2.1** PBT — **inStock derivation**: For any generated platform product list, each assembled `ProductSummaryReadModel.InStock` is `true` if and only if at least one `simple` has `stock > 0`.
- [x] **4.2.2** PBT — **Total count passthrough**: For any raw platform list payload, `ProductListReadModel.Total` equals the payload's `numProductFound` field.
- [x] **4.2.3** PBT — **Item count preservation**: The number of `Items` in the assembled model equals the number of products in the raw payload.
- [x] **4.2.4** PBT — **FilterFacets color passthrough**: For any raw filter payload with `N` color facets, `FilterFacetsReadModel.Colors` has exactly `N` entries with matching `id`, `label`, and `count`.
- [x] **4.2.5** PBT — **FilterFacets occasion passthrough**: Same as 4.2.4 for `occasions`.
- [x] **4.2.6** PBT — **PriceRange facet passthrough**: `FilterFacetsReadModel.PriceRange` matches the raw `filters.price` `min`/`max` values exactly.

**File:** `construction/product_discovery/src/domain/assembler/product_detail_assembler_test.go`

- [x] **4.2.7** PBT — **Variant inStock derivation**: For any platform detail payload, each assembled `ProductVariantReadModel.InStock` equals `(variant.quantity > 0)`.
- [x] **4.2.8** PBT — **Field name mapping**: `config_sku → ConfigSku`, `url_key → Slug`, `name → Name`, `brand → Brand`, `description → Description` are all correctly mapped.
- [x] **4.2.9** PBT — **Variant count preservation**: Number of assembled `Variants` equals number of `simples` in the raw payload.
- [x] **4.2.10** PBT — **Images passthrough**: Assembled `Images` slice has the same length and identical elements as the raw `images` array.

---

#### Phase 4.3 — Unit 1: Application Layer Tests
*(Uses generated `MockPlatformProductApiClient` — no real HTTP calls)*

**File:** `construction/product_discovery/src/application/product_list_query_handler_test.go`

- [x] **4.3.1** Example — **Success path**: Mock returns valid list + filter payloads → `Handle` returns a `ProductListReadModel` with no error.
- [x] **4.3.2** Example — **List fetch failure**: Mock `FetchProductList` returns error → `Handle` returns `ProductListUnavailable`.
- [x] **4.3.3** Example — **Filter fetch failure**: Mock `FetchProductFilters` returns error → `Handle` returns `ProductListUnavailable`.
- [x] **4.3.4** PBT — **Payload routing correctness**: For any combination of list and filter payloads returned by the mock, assembled `Items` come from the list payload and assembled `Filters` come from the filter payload (no cross-contamination).

**File:** `construction/product_discovery/src/application/product_detail_query_handler_test.go`

- [x] **4.3.5** Example — **Success path**: Mock returns valid payload → `Handle` returns `ProductDetailReadModel`.
- [x] **4.3.6** Example — **Not found**: Mock returns `(nil, nil)` → `Handle` returns `ProductNotFound`.
- [x] **4.3.7** Example — **Platform error**: Mock returns transport error → `Handle` returns `ProductDetailUnavailable`.
- [x] **4.3.8** PBT — **Identity mapping**: For any valid platform detail payload returned by the mock, assembled `ConfigSku` equals the `config_sku` field of the input.

---

#### Phase 4.5 — Unit 2: Value Object PBT Tests

**File:** `construction/wishlist/src/domain/valueobject/simple_sku_test.go`

- [x] **4.5.1** PBT — **SimpleSku valid construction**: For any non-empty string, `NewSimpleSku` returns no error and `sku.String()` equals the input.
- [x] **4.5.2** Example — **SimpleSku rejects empty string**: `NewSimpleSku("")` returns error.

**File:** `construction/wishlist/src/domain/valueobject/config_sku_test.go`

- [x] **4.5.3** PBT — **ConfigSku valid construction**: For any non-empty string, `NewConfigSku` returns no error.
- [x] **4.5.4** Example — **ConfigSku rejects empty string**: `NewConfigSku("")` returns error.

**File:** `construction/wishlist/src/domain/valueobject/money_test.go`

- [x] **4.5.5** PBT — **Money valid construction**: For all `amount ≥ 0.0`, `NewMoney` returns no error.
- [x] **4.5.6** PBT — **Money rejects negative**: For all `amount < 0`, `NewMoney` returns error.

**File:** `construction/wishlist/src/domain/valueobject/pagination_test.go`

- [x] **4.5.7** PBT — **Pagination valid construction**: For `offset ≥ 0` and `limit ≥ 1`, `NewPagination` returns no error with correct fields.
- [x] **4.5.8** PBT — **Pagination rejects negative offset**: For all `offset < 0`, returns error.
- [x] **4.5.9** PBT — **Pagination rejects zero/negative limit**: For all `limit ≤ 0`, returns error.

---

#### Phase 4.6 — Unit 2: Wishlist Aggregate PBT Tests

**File:** `construction/wishlist/src/domain/aggregate/wishlist_test.go`

- [x] **4.6.1** PBT — **deriveConfigSku two-part extraction**: For any SimpleSku string with ≥2 dash-separated parts (e.g., `"PD-001-M-BLK"`), `deriveConfigSku` returns the first two parts joined by `-` (e.g., `"PD-001"`).
- [x] **4.6.2** PBT — **deriveConfigSku single-part fallback**: For any SimpleSku string with no dashes, `deriveConfigSku` returns the whole string unchanged.
- [x] **4.6.3** PBT — **AddItem duplicate prevention**: For any `Wishlist` containing an item whose `ConfigSku` matches the derived configSku of the incoming `SimpleSku`, `AddItem` returns `ErrWishlistItemAlreadyPresent`.
- [x] **4.6.4** PBT — **AddItem success on empty wishlist**: For any valid `SimpleSku`, `AddItem` on an empty `Wishlist` returns `AddItemIntent` with no error.
- [x] **4.6.5** PBT — **AddItem success on non-conflicting wishlist**: For any `Wishlist` whose items all have configSkus different from the new item's derived configSku, `AddItem` succeeds.
- [x] **4.6.6** PBT — **AddItemIntent carries correct configSku**: For any successful `AddItem`, `AddItemIntent.ConfigSku` equals `deriveConfigSku(simpleSku)`.
- [x] **4.6.7** PBT — **RemoveItem is always non-error**: For any `Wishlist` (empty, non-empty, item present or absent), `RemoveItem(configSku)` always returns a `RemoveItemIntent` with the given `ConfigSku` and no error.
- [x] **4.6.8** PBT — **ToggleItem returns RemoveIntent when item present**: For any `Wishlist` containing an item with `configSku == X`, `ToggleItem(simpleSku, X)` returns a `RemoveItemIntent`.
- [x] **4.6.9** PBT — **ToggleItem returns AddIntent when item absent**: For any `Wishlist` with no item having `configSku == X`, `ToggleItem(simpleSku, X)` returns an `AddItemIntent`.
- [x] **4.6.10** PBT — **ToggleItem idempotency**: Starting from a wishlist without item `X`, calling ToggleItem (absent → add intent) and then again (present → remove intent) returns the aggregate to a state where `AddItem` succeeds again — no invariant violation.

---

#### Phase 4.7 — Unit 2: WishlistAssembler PBT Tests

**File:** `construction/wishlist/src/domain/assembler/wishlist_assembler_test.go`

- [x] **4.7.1** PBT — **Item count preservation**: For any raw platform wishlist payload with `N` items, `Wishlist.Items` has exactly `N` elements.
- [x] **4.7.2** PBT — **TotalCount passthrough**: Assembled `Wishlist.TotalCount` equals the `totalCount` field in the raw payload.
- [x] **4.7.3** PBT — **Field mapping**: For any raw item, `itemId → ItemId.Value`, `simpleSku → SimpleSku.Value`, `configSku → ConfigSku.Value`, `product.name → Name`, `product.brand → Brand`, `product.price → Price.Amount`, `product.image → ImageUrl` are all correctly mapped.
- [x] **4.7.4** PBT — **inStock passthrough**: Assembled `WishlistItem.InStock` equals the `inStock` boolean on the raw item.
- [x] **4.7.5** Example — **Empty items array**: Assembling a payload with `items: []` produces `Wishlist.Items == []` with no error.

---

#### Phase 4.8 — Unit 2: Application Service Tests
*(Uses generated `MockWishlistRepository`, `MockAuthSessionService`, `MockEventBus` — no HTTP, no real I/O)*

**File:** `construction/wishlist/src/application/get_wishlist_service_test.go`

- [x] **4.8.1** Example — **Unauthenticated returns error**: Mock `AuthSessionService` returns `UnauthenticatedShopperError` → service returns that error without calling repo.
- [x] **4.8.2** Example — **Authenticated success**: Mock auth resolves shopperId; mock repo returns wishlist → service returns shaped items with no error.
- [x] **4.8.3** PBT — **Item shape completeness**: For any `Wishlist` with `N` items returned by the mock repo, the service response contains exactly `N` items, each with all required fields: `itemId`, `simpleSku`, `configSku`, `name`, `brand`, `price`, `imageUrl`, `color`, `size`, `inStock`.

**File:** `construction/wishlist/src/application/add_wishlist_item_service_test.go`

- [x] **4.8.4** Example — **Unauthenticated triggers auth gate event**: Mock auth fails → service emits `AuthenticationGateTriggered` on mock event bus and returns `UnauthenticatedShopperError`.
- [x] **4.8.5** Example — **Duplicate configSku returns conflict error**: Mock repo returns wishlist already containing the derived `configSku` → `ErrWishlistItemAlreadyPresent` returned; mock repo `AddItem` is never called.
- [x] **4.8.6** Example — **Success emits WishlistItemAdded**: Mock repo `AddItem` succeeds → `WishlistItemAdded` event published on mock event bus with correct `shopperId`, `simpleSku`, `configSku`, `itemId`.
- [x] **4.8.7** PBT — **No repo call on duplicate**: For any wishlist state containing configSku `X`, calling the service with a SimpleSku that derives to `X` never invokes `WishlistRepository.AddItem` on the mock.
- [x] **4.8.8** PBT — **Response fields match repo return**: For any `WishlistItemId` returned by the mock repo, the service response `itemId` field equals that value exactly.

**File:** `construction/wishlist/src/application/remove_wishlist_item_service_test.go`

- [x] **4.8.9** Example — **Unauthenticated returns error**: Mock auth fails → `UnauthenticatedShopperError` propagated; mock repo `RemoveItemByConfigSku` never called.
- [x] **4.8.10** Example — **Success emits WishlistItemRemoved**: Mock remove succeeds → `WishlistItemRemoved` event published on mock event bus with correct `shopperId` and `configSku`.
- [x] **4.8.11** PBT — **Event payload correctness**: For any `configSku` string, `WishlistItemRemoved.ConfigSku` in the event captured by the mock bus equals the input `configSku`.

---

#### Phase 4.10 — Unit 2: Event Bus PBT Tests

**File:** `construction/wishlist/src/infrastructure/eventbus/in_memory_event_bus_test.go`

- [x] **4.10.1** PBT — **All subscribers receive event**: For any event published with `N` registered handlers, all `N` handlers are called exactly once.
- [x] **4.10.2** PBT — **Subscriber isolation**: Handlers subscribed to event type A do not receive events of type B.
- [x] **4.10.3** PBT — **Handler call order**: Handlers are called in registration order.
- [x] **4.10.4** Example — **No subscribers is a no-op**: Publishing an event with no subscribers does not panic or error.
- [x] **4.10.5** Example — **Handler receives identical event**: The event argument passed to the handler equals (by value) the event that was published.

---

#### Phase 4.11 — Cross-Cutting: Unit 2 → Unit 3 Contract Tests
*(Pure unit tests — JSON-marshal the domain struct directly, no HTTP server needed)*

**File:** `construction/wishlist/src/domain/assembler/contract_test.go`

- [x] **4.11.1** Example — **All Unit-3-required fields present in WishlistItem JSON**: Marshal a `WishlistItem` entity to JSON and assert that all nine required fields are present: `configSku`, `simpleSku`, `inStock`, `name`, `brand`, `price`, `imageUrl`, `color`, `size`.
- [x] **4.11.2** PBT — **Required fields never omitted**: For any `WishlistItem` value generated by `genWishlistItem`, its JSON serialisation contains all nine Unit-3-required fields regardless of field values (no accidental `omitempty` on required fields).

---

#### Phase 4.12 — Test Execution and Reporting

- [x] **4.12.1** Run all Unit 1 tests: `go test ./... -v` in `construction/product_discovery/src/`. All tests pass.
- [x] **4.12.2** Run all Unit 2 tests: `go test ./... -v` in `construction/wishlist/src/`. All tests pass.
- [x] **4.12.3** Run with race detector: `go test -race ./...` in both modules. No data races reported.
- [x] **4.12.4** Generate coverage report and enforce 75% threshold: `go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out` for both units. Fail the step if total coverage falls below **75%**.
- [x] **4.12.5** Document any coverage gaps or failing properties; raise follow-up tasks if needed.

---

### PBT Strategy Notes

#### Library: `pgregory.net/rapid`

Each PBT test uses `rapid.Check` to run a property hundreds of times with generated inputs. Failing inputs are automatically shrunk to the minimal counterexample.

**Structure template:**
```go
func TestPriceRange_ValidConstruction(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        min := rapid.Float64Range(0, 1000).Draw(t, "min")
        max := rapid.Float64Range(min, 2000).Draw(t, "max")

        pr, err := valueobject.NewPriceRange(fmt.Sprintf("%v-%v", min, max))

        if err != nil {
            t.Fatalf("expected no error for valid (min=%v, max=%v), got: %v", min, max, err)
        }
        if pr.Min != min || pr.Max != max {
            t.Fatalf("round-trip failed: got {%v, %v}, want {%v, %v}", pr.Min, pr.Max, min, max)
        }
    })
}
```

#### Reusable Generators to Define

| Generator | Module | Description |
|---|---|---|
| `genValidSimpleSku` | Unit 2 | Non-empty string; optionally shaped as `XX-NNN-S-CCC` |
| `genValidConfigSku` | Unit 2 | Non-empty string; optionally shaped as `XX-NNN` |
| `genWishlistItem` | Unit 2 | Fully populated `entity.WishlistItem` with random fields |
| `genWishlistWithN` | Unit 2 | `Wishlist` aggregate with `N` random items |
| `genPlatformProduct` | Unit 1 | Raw platform product struct with random simples |
| `genPlatformDetailPayload` | Unit 1 | Raw platform product detail struct |
| `genFilterPayload` | Unit 1 | Raw filter facets payload |
| `genValidPagination` | Both | `(offset ≥ 0, limit ≥ 1)` pair |

---

### Test File Layout Summary

```
construction/product_discovery/src/
├── mocks/                                    # Step 4.0.4–4.0.5 (mockery generated)
│   └── mock_platform_product_client.go
├── domain/valueobject/
│   ├── price_range_test.go                   # Steps 4.1.1–4.1.5
│   ├── pagination_test.go                    # Steps 4.1.6–4.1.8
│   └── money_test.go                         # Steps 4.1.9–4.1.10
├── domain/assembler/
│   ├── product_list_assembler_test.go        # Steps 4.2.1–4.2.6
│   └── product_detail_assembler_test.go      # Steps 4.2.7–4.2.10
└── application/
    ├── product_list_query_handler_test.go    # Steps 4.3.1–4.3.4
    └── product_detail_query_handler_test.go  # Steps 4.3.5–4.3.8

construction/wishlist/src/
├── mocks/                                    # Step 4.0.4–4.0.5 (mockery generated)
│   ├── mock_wishlist_repository.go
│   ├── mock_auth_session_service.go
│   └── mock_event_bus.go
├── domain/valueobject/
│   ├── simple_sku_test.go                    # Steps 4.5.1–4.5.2
│   ├── config_sku_test.go                    # Steps 4.5.3–4.5.4
│   ├── money_test.go                         # Steps 4.5.5–4.5.6
│   └── pagination_test.go                    # Steps 4.5.7–4.5.9
├── domain/aggregate/
│   └── wishlist_test.go                      # Steps 4.6.1–4.6.10
├── domain/assembler/
│   ├── wishlist_assembler_test.go            # Steps 4.7.1–4.7.5
│   └── contract_test.go                      # Steps 4.11.1–4.11.2
├── application/
│   ├── get_wishlist_service_test.go          # Steps 4.8.1–4.8.3
│   ├── add_wishlist_item_service_test.go     # Steps 4.8.4–4.8.8
│   └── remove_wishlist_item_service_test.go  # Steps 4.8.9–4.8.11
└── infrastructure/eventbus/
    └── in_memory_event_bus_test.go           # Steps 4.10.1–4.10.5
```
