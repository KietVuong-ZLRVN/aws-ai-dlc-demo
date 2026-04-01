# Logical Design — Unit 1: Product Discovery

## Overview

Unit 1 is a read-only BFF service built in Go. It exposes two HTTP endpoints to the client, aggregates data from the Platform Product API, and returns UI-shaped responses. There is no local persistence, no domain events, and no write operations. The architecture follows a CQRS read-model pattern layered as a clean, dependency-inverted stack.

---

## Technology Choices

| Concern | Choice | Rationale |
|---|---|---|
| Language | Go | Backend language specified for this unit |
| HTTP router | `go-chi/chi` v5 | Lightweight, idiomatic, composable middleware, no code generation |
| HTTP client | `net/http` (standard library) | No external dependency needed for upstream API calls |
| Configuration | Environment variables | Twelve-factor app; no configuration library required |
| Logging | `log/slog` (standard library) | Structured logging without added dependencies |

---

## Layered Architecture

The service is divided into four layers. Dependencies flow strictly inward: outer layers depend on inner layers; inner layers know nothing of outer layers.

```
┌──────────────────────────────────┐
│         API Layer                │  HTTP handlers, request parsing, response writing
├──────────────────────────────────┤
│       Application Layer          │  Query handlers, orchestration, no domain logic
├──────────────────────────────────┤
│         Domain Layer             │  Query objects, read models, value objects, assemblers
├──────────────────────────────────┤
│      Infrastructure Layer        │  ACL (PlatformProductApiClient), HTTP client impl
└──────────────────────────────────┘
```

Each layer communicates through interfaces defined by the inner layer and implemented by the outer layer (Dependency Inversion Principle).

---

## Package / Directory Structure

```
construction/product_discovery/src/
├── cmd/
│   └── server/
│       └── main.go                   # Entry point: wires all components, starts HTTP server
│
├── api/
│   ├── router.go                     # chi router setup, middleware registration, route mounting
│   ├── product_list_handler.go       # Handler for GET /api/v1/products
│   └── product_detail_handler.go    # Handler for GET /api/v1/products/{configSku}
│
├── application/
│   ├── product_list_query_handler.go   # Handles ProductListQuery → ProductListReadModel
│   └── product_detail_query_handler.go # Handles ProductDetailQuery → ProductDetailReadModel
│
├── domain/
│   ├── query/
│   │   ├── product_list_query.go        # ProductListQuery struct
│   │   └── product_detail_query.go      # ProductDetailQuery struct
│   ├── readmodel/
│   │   ├── product_list_read_model.go   # ProductListReadModel, ProductSummaryReadModel
│   │   ├── product_detail_read_model.go # ProductDetailReadModel, ProductVariantReadModel
│   │   └── filter_facets_read_model.go  # FilterFacetsReadModel, FilterOption
│   ├── valueobject/
│   │   ├── price_range.go               # PriceRange VO
│   │   ├── money.go                     # Money VO
│   │   ├── pagination.go                # Pagination VO
│   │   └── filter_option.go             # FilterOption VO
│   └── assembler/
│       ├── product_list_assembler.go    # Transforms platform list + filter payload → ProductListReadModel
│       └── product_detail_assembler.go  # Transforms platform detail payload → ProductDetailReadModel
│
├── infrastructure/
│   └── platform/
│       ├── product_api_client.go        # Concrete PlatformProductApiClient (ACL implementation)
│       └── platform_types.go            # Raw platform response structs (internal to infra layer)
│
└── config/
    └── config.go                        # Reads env vars: PLATFORM_BASE_URL, PORT, etc.
```

---

## Component Responsibilities

### `cmd/server/main.go`

The composition root. Responsibilities:
- Reads configuration from environment variables via `config.Config`.
- Instantiates the infrastructure client (`PlatformProductApiClient`).
- Instantiates assemblers (`ProductListAssembler`, `ProductDetailAssembler`).
- Instantiates application-layer query handlers, injecting the ACL client and assemblers.
- Instantiates API-layer handlers, injecting the query handlers.
- Mounts the chi router and starts the HTTP server.

No business logic lives here. This is the only place where concrete types are named across layer boundaries.

---

### API Layer

#### `api/router.go`

- Creates the chi `Router`.
- Registers shared middleware: request ID injection, structured request logging, panic recovery.
- Mounts `ProductListHandler` on `GET /api/v1/products`.
- Mounts `ProductDetailHandler` on `GET /api/v1/products/{configSku}`.

#### `api/product_list_handler.go`

**Responsibility:** Parse the HTTP request into a `ProductListQuery`, delegate to `ProductListQueryHandler`, and write the HTTP response.

**Steps:**
1. Extract and validate query parameters (`query`, `categoryId`, `colors[]`, `price`, `occasion`, `offset`, `limit`).
2. Parse `price` string (`"0-200"`) into a `PriceRange` value object; return `400` if malformed.
3. Construct `Pagination` value object (apply defaults: offset=0, limit=20; reject negative values with `400`).
4. Build `ProductListQuery` and pass to `ProductListQueryHandler.Handle(ctx, query)`.
5. Serialize the returned `ProductListReadModel` as JSON with status `200`.
6. On domain error `ProductListUnavailable`: return `502 Bad Gateway` with a structured error body.

#### `api/product_detail_handler.go`

**Responsibility:** Parse the `configSku` path parameter, delegate to `ProductDetailQueryHandler`, and write the HTTP response.

**Steps:**
1. Extract `configSku` from the chi URL parameter.
2. Reject empty or obviously malformed SKUs with `400`.
3. Build `ProductDetailQuery` and pass to `ProductDetailQueryHandler.Handle(ctx, query)`.
4. On domain result `ProductNotFound`: return `404`.
5. On domain result `ProductDetailUnavailable`: return `502`.
6. On success: serialize `ProductDetailReadModel` as JSON with status `200`.

---

### Application Layer

The application layer contains query handlers. A query handler orchestrates one use case: it calls the ACL via an interface, passes the raw result to an assembler, and returns a read model. It contains no domain logic.

#### `application/product_list_query_handler.go`

**Dependencies (injected via constructor):**
- `PlatformProductApiClient` interface
- `ProductListAssembler`

**Handle flow:**
1. Translate `ProductListQuery` into platform API parameters.
2. Issue two concurrent calls via `PlatformProductApiClient`:
   - `FetchProductList(ctx, params)` → raw list payload
   - `FetchProductFilters(ctx, params)` → raw filter payload
3. If either call fails: return `ProductListUnavailable` domain error.
4. Pass both raw payloads to `ProductListAssembler.Assemble(listPayload, filterPayload)`.
5. Return `ProductListReadModel`.

**Concurrency note:** The two platform calls are issued using Go goroutines with a shared context and collected via channels or `errgroup`. If either fails, the other is cancelled via context cancellation.

#### `application/product_detail_query_handler.go`

**Dependencies (injected via constructor):**
- `PlatformProductApiClient` interface
- `ProductDetailAssembler`

**Handle flow:**
1. Call `PlatformProductApiClient.FetchProductDetail(ctx, configSku)`.
2. If the client returns `nil` (product not found on platform): return `ProductNotFound` domain result.
3. If the client returns a transport error: return `ProductDetailUnavailable` domain error.
4. Pass raw payload to `ProductDetailAssembler.Assemble(payload)`.
5. Return `ProductDetailReadModel`.

---

### Domain Layer

The domain layer contains only pure Go structs and value constructors. It has zero external dependencies and no I/O.

#### Query Objects (`domain/query/`)

`ProductListQuery` holds all optional filter fields as pointer types (nil = not set). `ProductDetailQuery` holds a single `configSku` string.

#### Read Models (`domain/readmodel/`)

Flat, immutable structs. All fields are exported for JSON serialisation. Read models are only constructed by assemblers — never by handlers directly.

#### Value Objects (`domain/valueobject/`)

Immutable structs with validation constructors (e.g., `NewPriceRange(min, max float64) (PriceRange, error)`). Validation errors bubble up to the API layer, which maps them to `400` responses.

#### Assemblers (`domain/assembler/`)

##### `ProductListAssembler`

**Input:** raw platform list payload (slice of platform product structs) + raw filter payload

**Output:** `ProductListReadModel`

**Logic:**
- Iterates platform product slice; maps each to `ProductSummaryReadModel`.
- Derives `InStock` by checking whether any `simple` in the product has `quantity > 0`.
- Maps platform filter facets to `FilterFacetsReadModel` and `FilterOption` value objects.
- Populates `Total` from the platform list's `totalCount` field.

##### `ProductDetailAssembler`

**Input:** raw platform product detail payload

**Output:** `ProductDetailReadModel`

**Logic:**
- Maps top-level platform fields to `ProductDetailReadModel` fields (e.g., `config_sku` → `ConfigSku`, `url_key` → `Slug`).
- Iterates platform `simples` array; maps each to `ProductVariantReadModel`.
- Derives `InStock` per variant from `quantity > 0`.

---

### Infrastructure Layer

#### `infrastructure/platform/product_api_client.go`

Concrete implementation of the `PlatformProductApiClient` interface.

**Interface (defined in the domain layer to support Dependency Inversion):**

```
PlatformProductApiClient interface:
  FetchProductList(ctx, params) → (rawListPayload, error)
  FetchProductFilters(ctx, params) → (rawFilterPayload, error)
  FetchProductDetail(ctx, configSku) → (rawDetailPayload, error)
    — returns (nil, nil) when the platform returns 404
```

**Implementation responsibilities:**
- Builds platform API URLs using `PLATFORM_BASE_URL` from config.
- Sets required headers: `Content-Language`, `Accept: application/json`.
- Executes `net/http` requests with the provided context (enabling cancellation).
- Parses platform HTTP status:
  - `200` → decode body into `platform_types.go` structs; return raw payload.
  - `404` → return `(nil, nil)` (signals "not found" to the query handler).
  - `5xx` / network error → return a `PlatformUnavailableError`.
- Does not transform data — raw structs are returned to assemblers.

#### `infrastructure/platform/platform_types.go`

Internal structs matching the exact JSON shape of the Platform Product API responses. These structs are private to the infrastructure package and never leak into the domain or application layers.

---

## Data Flow Diagrams

### `GET /api/v1/products` — Product List

```
Client
  │ GET /api/v1/products?query=dress&colors[]=black&price=0-200
  ▼
ProductListHandler
  │ parse params → ProductListQuery
  ▼
ProductListQueryHandler.Handle(ctx, query)
  │
  ├──► PlatformProductApiClient.FetchProductList(ctx, params)   ─► GET /v1/products/list
  └──► PlatformProductApiClient.FetchProductFilters(ctx, params) ─► GET /v1/products/filter
  │    (concurrent; context-cancelled if either fails)
  ▼
ProductListAssembler.Assemble(listPayload, filterPayload)
  │ → ProductListReadModel
  ▼
ProductListHandler
  │ JSON encode → 200 OK
  ▼
Client
```

### `GET /api/v1/products/{configSku}` — Product Detail

```
Client
  │ GET /api/v1/products/SG-12345
  ▼
ProductDetailHandler
  │ extract configSku → ProductDetailQuery
  ▼
ProductDetailQueryHandler.Handle(ctx, query)
  │
  └──► PlatformProductApiClient.FetchProductDetail(ctx, "SG-12345") ─► GET /v1/products/SG-12345/details
  ▼
ProductDetailAssembler.Assemble(payload)
  │ → ProductDetailReadModel
  ▼
ProductDetailHandler
  │ JSON encode → 200 OK
  ▼
Client
```

---

## Error Handling Strategy

| Error Source | Domain Result | HTTP Response |
|---|---|---|
| Invalid query param (e.g., malformed price range) | `ValidationError` (VO constructor) | `400 Bad Request` with field message |
| Platform product not found (platform 404) | `ProductNotFound` | `404 Not Found` |
| Platform list/filter unavailable (5xx / timeout) | `ProductListUnavailable` | `502 Bad Gateway` |
| Platform detail unavailable (5xx / timeout) | `ProductDetailUnavailable` | `502 Bad Gateway` |
| Panic / unhandled error | Recovered by chi middleware | `500 Internal Server Error` |

Errors are represented as typed Go errors (distinct types, not string matching). The API layer switches on error type to determine the HTTP status code. Stack traces are logged at the infrastructure layer; only safe, user-facing messages are included in HTTP responses.

---

## Dependency Wiring (Constructor Injection)

All wiring happens in `cmd/server/main.go`. The wiring order is:

1. `cfg := config.Load()` — reads env vars
2. `platformClient := platform.NewProductApiClient(cfg.PlatformBaseURL)` — infra
3. `listAssembler := assembler.NewProductListAssembler()` — domain
4. `detailAssembler := assembler.NewProductDetailAssembler()` — domain
5. `listQH := application.NewProductListQueryHandler(platformClient, listAssembler)` — app
6. `detailQH := application.NewProductDetailQueryHandler(platformClient, detailAssembler)` — app
7. `listHandler := api.NewProductListHandler(listQH)` — api
8. `detailHandler := api.NewProductDetailHandler(detailQH)` — api
9. `router := api.NewRouter(listHandler, detailHandler)` — api
10. `http.ListenAndServe(cfg.Port, router)` — entry point

No global state, no package-level singletons.

---

## Middleware

All middleware is registered on the chi root router:

| Middleware | Responsibility |
|---|---|
| `middleware.RequestID` | Attaches a unique request ID to context and response header |
| `middleware.Logger` (slog-backed) | Logs method, path, status code, latency per request |
| `middleware.Recoverer` | Catches panics, logs stack trace, returns `500` |

No authentication middleware — all product discovery endpoints are public.

---

## Configuration

All configuration is read from environment variables at startup. The service fails fast if required variables are missing.

| Variable | Required | Default | Description |
|---|---|---|---|
| `PLATFORM_BASE_URL` | Yes | — | Base URL of the Platform Product API |
| `PORT` | No | `8080` | HTTP listen port |
| `LOG_LEVEL` | No | `info` | Log verbosity (`debug`, `info`, `warn`, `error`) |

---

## Cross-Unit Notes

- Unit 1 is consumed by the client UI only. It does not have server-to-server consumers in the integration contract.
- Unit 3 (AI Styling Engine) calls the Platform Product API directly for catalogue lookups — it does not route through Unit 1.
- The `PlatformProductApiClient` interface is defined in the domain layer and implemented in the infrastructure layer. This means the application and domain layers can be tested with a mock implementation without any HTTP calls.
