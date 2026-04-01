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
