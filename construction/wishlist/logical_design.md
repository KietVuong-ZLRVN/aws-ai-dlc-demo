# Logical Design — Unit 2: Wishlist Management

## Overview

Unit 2 is a write-capable BFF service built in Go. It exposes three HTTP endpoints: wishlist retrieval, add item, and remove item. All wishlist state is owned by the Platform Wishlist API; this service enforces domain invariants (duplicate prevention, authentication gate), orchestrates domain commands, emits domain events, and returns UI-shaped responses. The architecture follows a DDD layered approach with clear Dependency Inversion at each boundary.

---

## Technology Choices

| Concern | Choice | Rationale |
|---|---|---|
| Language | Go | Backend language specified for this unit |
| HTTP router | `go-chi/chi` v5 | Lightweight, idiomatic, composable middleware |
| HTTP client | `net/http` (standard library) | Upstream platform calls |
| Event bus | In-memory synchronous event bus | Suitable for demo; abstracted behind an interface |
| Session | HTTP cookie (passed through to platform) | Platform Auth API validates the session cookie |
| Configuration | Environment variables | Twelve-factor app |
| Logging | `log/slog` (standard library) | Structured logging |

---

## Layered Architecture

Dependencies flow strictly inward. The domain layer has no external dependencies.

```
┌──────────────────────────────────────┐
│           API Layer                  │  HTTP handlers, request/response, cookie extraction
├──────────────────────────────────────┤
│        Application Layer             │  Command/query orchestrators, session resolution, event dispatch
├──────────────────────────────────────┤
│          Domain Layer                │  Wishlist aggregate, WishlistItem entity, value objects,
│                                      │  domain events, policies, repository interface, assembler
├──────────────────────────────────────┤
│       Infrastructure Layer           │  PlatformWishlistApiClient (ACL), PlatformAuthApiClient (ACL),
│                                      │  PlatformWishlistRepository, in-memory event bus
└──────────────────────────────────────┘
```

---

## Package / Directory Structure

```
construction/wishlist/src/
├── cmd/
│   └── server/
│       └── main.go                         # Composition root: wire all components, start server
│
├── api/
│   ├── router.go                           # chi router, middleware, route mounting
│   ├── wishlist_get_handler.go             # Handler for GET /api/v1/wishlist
│   ├── wishlist_add_handler.go             # Handler for POST /api/v1/wishlist/items
│   └── wishlist_remove_handler.go         # Handler for DELETE /api/v1/wishlist/items/{configSku}
│
├── application/
│   ├── get_wishlist_service.go             # Orchestrates GET: resolves session → loads wishlist
│   ├── add_wishlist_item_service.go        # Orchestrates POST: auth gate → addItem command → events
│   ├── remove_wishlist_item_service.go     # Orchestrates DELETE: auth gate → removeItem command → events
│   └── event_dispatcher.go                # Dispatches domain events to registered handlers
│
├── domain/
│   ├── aggregate/
│   │   └── wishlist.go                     # Wishlist aggregate root + addItem/removeItem/toggleItem
│   ├── entity/
│   │   └── wishlist_item.go                # WishlistItem entity
│   ├── valueobject/
│   │   ├── wishlist_id.go                  # WishlistId VO
│   │   ├── shopper_id.go                   # ShopperId VO
│   │   ├── wishlist_item_id.go             # WishlistItemId VO
│   │   ├── simple_sku.go                   # SimpleSku VO
│   │   ├── config_sku.go                   # ConfigSku VO
│   │   ├── money.go                        # Money VO
│   │   ├── pagination.go                   # Pagination VO
│   │   └── pending_wishlist_add.go         # PendingWishlistAdd VO
│   ├── event/
│   │   ├── wishlist_item_added.go          # WishlistItemAdded event
│   │   ├── wishlist_item_removed.go        # WishlistItemRemoved event
│   │   └── authentication_gate_triggered.go # AuthenticationGateTriggered event
│   ├── repository/
│   │   └── wishlist_repository.go          # WishlistRepository interface
│   ├── service/
│   │   ├── auth_session_service.go         # AuthSessionService interface
│   │   └── wishlist_toggle_service.go      # WishlistToggleService
│   └── assembler/
│       └── wishlist_assembler.go           # Maps platform payloads → domain objects
│
├── infrastructure/
│   ├── platform/
│   │   ├── wishlist_api_client.go          # PlatformWishlistApiClient (ACL implementation)
│   │   ├── auth_api_client.go              # PlatformAuthApiClient (ACL implementation)
│   │   ├── wishlist_repository.go          # PlatformWishlistRepository (implements WishlistRepository)
│   │   └── platform_types.go              # Raw platform response structs (private to infra)
│   └── eventbus/
│       └── in_memory_event_bus.go          # In-memory synchronous EventBus implementation
│
└── config/
    └── config.go                           # Reads env vars
```

---

## Component Responsibilities

### `cmd/server/main.go`

The composition root. Wiring order:

1. `cfg := config.Load()` — reads env vars
2. `wishlistApiClient := platform.NewWishlistApiClient(cfg.PlatformBaseURL)` — infra
3. `authApiClient := platform.NewAuthApiClient(cfg.PlatformBaseURL)` — infra
4. `wishlistAssembler := assembler.NewWishlistAssembler()` — domain
5. `wishlistRepo := platform.NewPlatformWishlistRepository(wishlistApiClient, wishlistAssembler)` — infra
6. `authService := platform.NewPlatformAuthSessionService(authApiClient)` — infra (implements `AuthSessionService`)
7. `eventBus := eventbus.NewInMemoryEventBus()` — infra
8. Register event handlers onto `eventBus` (e.g., a logging handler)
9. `getService := application.NewGetWishlistService(wishlistRepo, authService)` — app
10. `addService := application.NewAddWishlistItemService(wishlistRepo, authService, eventBus)` — app
11. `removeService := application.NewRemoveWishlistItemService(wishlistRepo, authService, eventBus)` — app
12. Mount API handlers; start server.

---

### API Layer

#### `api/router.go`

- Registers middleware: `RequestID`, `Logger`, `Recoverer`.
- Mounts `WishlistGetHandler` on `GET /api/v1/wishlist`.
- Mounts `WishlistAddHandler` on `POST /api/v1/wishlist/items`.
- Mounts `WishlistRemoveHandler` on `DELETE /api/v1/wishlist/items/{configSku}`.

#### `api/wishlist_get_handler.go`

1. Extract session cookie from `Cookie` header.
2. Parse `offset` and `limit` query params; apply defaults (0 / 20); reject negative values with `400`.
3. Build `Pagination` VO.
4. Call `GetWishlistService.Execute(ctx, sessionCookie, pagination)`.
5. On `UnauthenticatedShopperError`: return `403 Forbidden`.
6. On success: JSON-encode the wishlist response with `200 OK`.

#### `api/wishlist_add_handler.go`

1. Extract session cookie from `Cookie` header.
2. Parse JSON body; extract `simpleSku` string.
3. Build `SimpleSku` VO; return `400` if invalid.
4. Call `AddWishlistItemService.Execute(ctx, sessionCookie, simpleSku)`.
5. On `UnauthenticatedShopperError`: return `403 Forbidden` with auth-gate payload (includes `returnPath` if provided via `Referer` header).
6. On `WishlistItemAlreadyPresent`: return `409 Conflict` with a descriptive message.
7. On success: JSON-encode `{ itemId, simpleSku, configSku }` with `201 Created`.

#### `api/wishlist_remove_handler.go`

1. Extract session cookie from `Cookie` header.
2. Extract `configSku` from chi URL params.
3. Build `ConfigSku` VO; return `400` if empty.
4. Call `RemoveWishlistItemService.Execute(ctx, sessionCookie, configSku)`.
5. On `UnauthenticatedShopperError`: return `403 Forbidden`.
6. On success: return `200 OK` (empty body).

---

### Application Layer

Application services orchestrate use cases. They resolve the shopper's identity via `AuthSessionService`, delegate domain commands to the aggregate (loaded via the repository), and dispatch the resulting domain events.

#### `application/get_wishlist_service.go`

**Dependencies:** `WishlistRepository`, `AuthSessionService`

**Execute flow:**
1. Call `AuthSessionService.ResolveShopperId(ctx, sessionCookie)`.
2. On failure: return `UnauthenticatedShopperError`.
3. Call `WishlistRepository.GetByShopperId(ctx, shopperId, pagination)` → `Wishlist` aggregate.
4. Map the aggregate's `Items` slice into the response shape (handled inline or via a small response mapper).
5. Return the shaped wishlist response.

#### `application/add_wishlist_item_service.go`

**Dependencies:** `WishlistRepository`, `AuthSessionService`, `EventBus`

**Execute flow:**
1. Call `AuthSessionService.ResolveShopperId(ctx, sessionCookie)`.
2. On failure: emit `AuthenticationGateTriggered` via `EventBus`; return `UnauthenticatedShopperError`.
3. Call `WishlistRepository.GetByShopperId(ctx, shopperId, Pagination{offset:0, limit:100})` to load current items for duplicate check.
4. Call `wishlist.AddItem(simpleSku)` on the loaded aggregate.
5. On `WishlistItemAlreadyPresent`: return the error (no platform call made).
6. On success (aggregate returned an `AddItemIntent`): call `WishlistRepository.AddItem(ctx, shopperId, simpleSku)` → `WishlistItemId`.
7. Construct `WishlistItemAdded` event with returned `itemId`; dispatch via `EventBus`.
8. Return `{ itemId, simpleSku, configSku }`.

**Note on aggregate state load:** The aggregate is reconstituted from the platform before every mutation to ensure duplicate-prevention logic operates on the current wishlist state. This avoids stale cache inconsistencies since the platform is the system of record.

#### `application/remove_wishlist_item_service.go`

**Dependencies:** `WishlistRepository`, `AuthSessionService`, `EventBus`

**Execute flow:**
1. Call `AuthSessionService.ResolveShopperId(ctx, sessionCookie)`.
2. On failure: return `UnauthenticatedShopperError`.
3. Call `wishlist.RemoveItem(configSku)` — this is a command expressed on the in-memory aggregate.
4. Call `WishlistRepository.RemoveItemByConfigSku(ctx, shopperId, configSku)`.
5. Construct `WishlistItemRemoved` event; dispatch via `EventBus`.
6. Return void.

#### `application/event_dispatcher.go`

A thin wrapper that synchronously calls all registered handlers for a given event type. Event handlers are registered at startup in `main.go`.

---

### Domain Layer

#### `domain/aggregate/wishlist.go` — `Wishlist` Aggregate Root

**State:**
- `WishlistId`, `ShopperId`, `items []WishlistItem`, `totalCount int`

**Behaviour:**

`AddItem(simpleSku SimpleSku) → (AddItemIntent, error)`
- Derives `configSku` from `simpleSku` (the domain knows the configSku from the loaded items or derives it by convention).
- Checks `items` for any existing entry with the same `configSku`.
- If duplicate found: return `WishlistItemAlreadyPresent` error.
- Otherwise: return an `AddItemIntent` struct (a value indicating the command was accepted; actual persistence is done by the application service via the repository).

`RemoveItem(configSku ConfigSku)`
- Records the removal intent. The application service calls the repository to execute it.

`ToggleItem(simpleSku SimpleSku, configSku ConfigSku) → (AddItemIntent or RemoveIntent, error)`
- Checks for existing `configSku` in `items`.
- If present: delegates to `RemoveItem`.
- If absent: delegates to `AddItem`.

**Note on intent pattern:** Rather than the aggregate calling the repository directly, it returns an intent value. This keeps the aggregate free of I/O and makes behaviour unit-testable without mocks.

---

#### `domain/entity/wishlist_item.go` — `WishlistItem` Entity

Holds `WishlistItemId`, `SimpleSku`, `ConfigSku`, `name`, `brand`, `Money price`, `imageUrl`, `color`, `size`, `inStock`. Constructed only by `WishlistAssembler` when reconstituting the aggregate.

---

#### Domain Value Objects (`domain/valueobject/`)

Each VO has a constructor that validates invariants and returns an error on invalid input. All VOs are immutable (no setters).

| VO | Key Invariant |
|---|---|
| `SimpleSku` | Non-empty string |
| `ConfigSku` | Non-empty string |
| `Money` | Non-negative decimal |
| `Pagination` | offset ≥ 0, limit > 0 |
| `PendingWishlistAdd` | Holds `SimpleSku` + `returnPath`; transient |

---

#### Domain Events (`domain/event/`)

| Event | When emitted | Key payload fields |
|---|---|---|
| `WishlistItemAdded` | After successful platform add | `shopperId`, `simpleSku`, `configSku`, `itemId`, `occurredAt` |
| `WishlistItemRemoved` | After successful platform delete | `shopperId`, `configSku`, `occurredAt` |
| `AuthenticationGateTriggered` | When unauthenticated add attempted | `simpleSku`, `returnPath`, `occurredAt` |

Events are plain Go structs. The `EventBus` interface accepts an `Event` interface (all events implement it via an `EventName() string` method).

---

#### `domain/repository/wishlist_repository.go` — Repository Interface

```
WishlistRepository interface:
  GetByShopperId(ctx, shopperId ShopperId, pagination Pagination) → (Wishlist, error)
  AddItem(ctx, shopperId ShopperId, simpleSku SimpleSku) → (WishlistItemId, error)
  RemoveItemByConfigSku(ctx, shopperId ShopperId, configSku ConfigSku) → error
```

Defined in the domain layer. Implemented in the infrastructure layer. The domain never imports from infrastructure.

---

#### `domain/service/auth_session_service.go` — AuthSessionService Interface

```
AuthSessionService interface:
  ResolveShopperId(ctx, sessionCookie string) → (ShopperId, error)
    — returns UnauthenticatedShopperError if session is invalid or absent
```

---

#### `domain/service/wishlist_toggle_service.go` — WishlistToggleService

**Dependencies:** `WishlistRepository`

**Execute(ctx, shopperId, simpleSku, configSku) → (event, error):**
1. Calls `WishlistRepository.GetByShopperId(...)` to load current state.
2. Calls `wishlist.ToggleItem(simpleSku, configSku)`.
3. Depending on the returned intent, calls `WishlistRepository.AddItem` or `RemoveItemByConfigSku`.
4. Returns the emitted domain event to the application service.

---

#### `domain/assembler/wishlist_assembler.go` — WishlistAssembler

**Responsibilities:**
- Maps raw platform wishlist payload → `Wishlist` aggregate (reconstitution).
- Maps raw platform item entries → `WishlistItem` entities.
- Translates platform field names to domain names.

---

### Infrastructure Layer

#### `infrastructure/platform/wishlist_api_client.go`

Implements `PlatformWishlistApiClient` (interface defined in domain).

**Operations:**
- `FetchWishlist(ctx, sessionCookie, offset, limit)` → raw platform wishlist payload
- `AddItem(ctx, sessionCookie, simpleSku)` → raw platform item payload
- `RemoveItemsByConfigSku(ctx, sessionCookie, configSku)` → void

**Responsibilities:**
- Builds platform URLs from `PLATFORM_BASE_URL`.
- Injects session cookie into requests (`Cookie` header).
- Sets `Content-Language`, `Accept: application/json`.
- Translates platform `403` → `UnauthenticatedShopperError`.
- Translates platform `5xx` / network errors → `PlatformUnavailableError`.

#### `infrastructure/platform/auth_api_client.go`

Implements `PlatformAuthApiClient`.

**Operations:**
- `ValidateSession(ctx, sessionCookie)` → `ShopperId` or `UnauthenticatedShopperError`

Calls `POST /v1/customers/login` (or the session-validation equivalent on the platform). Returns the resolved shopper ID from the platform response.

#### `infrastructure/platform/wishlist_repository.go`

Implements `WishlistRepository`.

- `GetByShopperId`: calls `WishlistApiClient.FetchWishlist(...)`, passes raw payload to `WishlistAssembler`, returns reconstituted `Wishlist` aggregate.
- `AddItem`: calls `WishlistApiClient.AddItem(...)`, returns `WishlistItemId`.
- `RemoveItemByConfigSku`: calls `WishlistApiClient.RemoveItemsByConfigSku(...)`.

The session cookie is threaded through the context (stored as a context value) so the repository can pass it to the ACL without adding it as an explicit parameter on the `WishlistRepository` interface.

#### `infrastructure/eventbus/in_memory_event_bus.go`

Implements the `EventBus` interface.

```
EventBus interface:
  Publish(ctx, event Event)
  Subscribe(eventName string, handler EventHandler)
```

- `Publish` iterates the list of handlers registered for `event.EventName()` and calls each synchronously.
- Handlers are registered at startup. Handler errors are logged but do not propagate (fire-and-observe semantics for the demo).

---

## Data Flow Diagrams

### `GET /api/v1/wishlist`

```
Client (with session cookie)
  │ GET /api/v1/wishlist?offset=0&limit=20
  ▼
WishlistGetHandler
  │ extract cookie, build Pagination
  ▼
GetWishlistService.Execute(ctx, cookie, pagination)
  │
  ├──► AuthSessionService.ResolveShopperId(ctx, cookie)
  │       └──► PlatformAuthApiClient.ValidateSession(ctx, cookie) ─► Platform Auth API
  │
  └──► WishlistRepository.GetByShopperId(ctx, shopperId, pagination)
          └──► PlatformWishlistApiClient.FetchWishlist(ctx, cookie, offset, limit) ─► Platform Wishlist API
          └──► WishlistAssembler.AssembleWishlist(rawPayload) → Wishlist aggregate
  │
  ▼
WishlistGetHandler
  │ shape aggregate items → JSON response → 200 OK
  ▼
Client
```

### `POST /api/v1/wishlist/items` — Add Item

```
Client (with session cookie)
  │ POST /api/v1/wishlist/items  { "simpleSku": "SG-12345-M-BLK" }
  ▼
WishlistAddHandler
  │ parse body → SimpleSku VO
  ▼
AddWishlistItemService.Execute(ctx, cookie, simpleSku)
  │
  ├──► AuthSessionService.ResolveShopperId(ctx, cookie)
  │       └── on failure: emit AuthenticationGateTriggered → EventBus
  │                        return UnauthenticatedShopperError → 403
  │
  ├──► WishlistRepository.GetByShopperId(ctx, shopperId, Pagination{0,100})
  │       └── reconstitute Wishlist aggregate (current items for duplicate check)
  │
  ├──► wishlist.AddItem(simpleSku)
  │       └── on duplicate configSku: return WishlistItemAlreadyPresent → 409
  │
  └──► WishlistRepository.AddItem(ctx, shopperId, simpleSku)
          └──► PlatformWishlistApiClient.AddItem(ctx, cookie, simpleSku) ─► Platform Wishlist API
          └── returns WishlistItemId
  │
  ├──► EventBus.Publish(WishlistItemAdded{shopperId, simpleSku, configSku, itemId})
  │
  ▼
WishlistAddHandler
  │ JSON encode { itemId, simpleSku, configSku } → 201 Created
  ▼
Client
```

### `DELETE /api/v1/wishlist/items/{configSku}` — Remove Item

```
Client (with session cookie)
  │ DELETE /api/v1/wishlist/items/SG-12345
  ▼
WishlistRemoveHandler
  │ extract configSku → ConfigSku VO
  ▼
RemoveWishlistItemService.Execute(ctx, cookie, configSku)
  │
  ├──► AuthSessionService.ResolveShopperId(ctx, cookie)
  │       └── on failure: return UnauthenticatedShopperError → 403
  │
  └──► WishlistRepository.RemoveItemByConfigSku(ctx, shopperId, configSku)
          └──► PlatformWishlistApiClient.RemoveItemsByConfigSku(ctx, cookie, configSku) ─► Platform Wishlist API
  │
  ├──► EventBus.Publish(WishlistItemRemoved{shopperId, configSku})
  │
  ▼
WishlistRemoveHandler
  │ 200 OK (empty body)
  ▼
Client
```

---

## Error Handling Strategy

| Error Source | Domain Error Type | HTTP Response |
|---|---|---|
| Missing / invalid `simpleSku` in body | `ValidationError` | `400 Bad Request` |
| Invalid `configSku` path param | `ValidationError` | `400 Bad Request` |
| Session absent or invalid | `UnauthenticatedShopperError` | `403 Forbidden` |
| Duplicate `configSku` already in wishlist | `WishlistItemAlreadyPresent` | `409 Conflict` |
| Platform wishlist API unavailable (5xx) | `PlatformUnavailableError` | `502 Bad Gateway` |
| Platform returns 403 for a platform auth issue | `UnauthenticatedShopperError` (re-wrapped) | `403 Forbidden` |
| Unhandled panic | Recovered by chi `Recoverer` | `500 Internal Server Error` |

Errors are typed Go error values. The API layer uses type-switching to map them to HTTP status codes. Sensitive platform error details are logged at the infrastructure layer only — client responses contain only safe, user-facing messages.

---

## Session Cookie Threading

The session cookie is extracted from the HTTP request by the handler and passed to the application service as a plain string. The application service passes it to `AuthSessionService`, which validates it and returns a `ShopperId`. The `ShopperId` (not the raw cookie) is then used for all domain operations.

When the repository needs to pass the cookie to the platform ACL (since the platform requires the raw session for wishlist calls), the cookie is stored as a context value at the handler level. The infrastructure layer reads it from context. This keeps the `WishlistRepository` interface free of HTTP concerns.

---

## Authentication Gate Flow

When an unauthenticated shopper attempts a wishlist add:

1. `AuthSessionService.ResolveShopperId` returns `UnauthenticatedShopperError`.
2. Application service emits `AuthenticationGateTriggered` with `simpleSku` and `returnPath` (from `Referer` header, extracted by handler and stored in context).
3. Application service returns `UnauthenticatedShopperError`.
4. Handler returns `403 Forbidden` with body `{ "requiresAuth": true, "returnPath": "..." }`.
5. Client stores `PendingWishlistAdd { simpleSku, returnPath }` in browser session storage.
6. Client shows login modal.
7. After successful login, client reads `PendingWishlistAdd` and re-calls `POST /api/v1/wishlist/items`.

`PendingWishlistAdd` is a client-side concern — it is not persisted by this service.

---

## Domain Event Bus Design

```
EventBus interface (domain layer):
  Publish(ctx context.Context, event Event)
  Subscribe(eventName string, handler EventHandler)

EventHandler type:
  func(ctx context.Context, event Event)

Event interface:
  EventName() string
```

The `InMemoryEventBus` (infrastructure layer) holds a `map[string][]EventHandler`. `Publish` looks up the event name and calls each handler in the order registered. Handlers run synchronously in the same goroutine as the caller. This is intentional for the demo — it keeps the execution model simple and eliminates concurrency complexity.

Event handlers registered at startup (in `main.go`):
- A structured log handler for each event type (logs event name, shopper ID, SKUs, timestamp).

Additional handlers (e.g., cache invalidation, analytics) can be registered without modifying existing code.

---

## Configuration

| Variable | Required | Default | Description |
|---|---|---|---|
| `PLATFORM_BASE_URL` | Yes | — | Base URL of the Platform Wishlist and Auth APIs |
| `PORT` | No | `8081` | HTTP listen port (Unit 2 listens on a separate port from Unit 1) |
| `LOG_LEVEL` | No | `info` | Log verbosity |

---

## Cross-Unit Integration

### Unit 2 as a provider to Unit 3

Unit 3 (AI Styling Engine) calls `GET /api/v1/wishlist` server-side using the shopper's forwarded session cookie. This is served by Unit 2's `WishlistGetHandler`. No special handling is required — Unit 3 is an authenticated downstream caller, indistinguishable from the client at the HTTP level.

**Published contract fields Unit 3 depends on:** `items[].configSku`, `items[].simpleSku`, `items[].name`, `items[].brand`, `items[].price`, `items[].imageUrl`, `items[].color`, `items[].size`, `items[].inStock`. These fields must not be renamed or removed without a coordinated contract change with Unit 3.

### Unit 3 client as a consumer of Unit 2

The Unit 3 UI may call `POST /api/v1/wishlist/items` to add AI-suggested catalog items to the shopper's wishlist. This flows through the standard `WishlistAddHandler` and is subject to the same duplicate prevention and authentication gate policies. No special path is needed.

### Shared Types

`ConfigSku` and `SimpleSku` value objects exist in both Unit 1 and Unit 2's domain layers as independent copies. They are intentionally not shared to avoid coupling between bounded contexts. If a shared kernel is introduced in a later phase, these would be the primary candidates.
