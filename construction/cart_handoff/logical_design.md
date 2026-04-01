# Logical Design: Cart Handoff Service

---

## 1. Service Overview

| Property | Value |
|---|---|
| Service name | `cart-handoff` |
| Runtime | Go 1.22 |
| HTTP framework | go-chi v5 |
| Deployment target | AWS ECS Fargate — standalone service |
| Listen port | `8080` |
| Architecture pattern | Clean Architecture (API → Application → Domain → Infrastructure) |
| Persistence | Amazon Aurora MySQL (same dedicated cluster as Combo Portfolio, `cart_handoff` database) |
| Event publishing | In-process only (no external bus) |
| Auth | Session cookie forwarded from client; `ShopperId` injected into request context by middleware |
| External HTTP dependencies | Unit 4 Combo Portfolio — `GET /api/v1/combos/{id}`; Platform Doraemon Cart API — `POST /v1/checkout/cart/bulk` |

---

## 2. Layered Architecture

```
┌──────────────────────────────────────────────────────────┐
│  API Layer  (api/)                                        │
│  go-chi router · HTTP handlers · request/response DTOs   │
│  session auth middleware                                  │
└────────────────────────┬─────────────────────────────────┘
                         │ calls
┌────────────────────────▼─────────────────────────────────┐
│  Application Layer  (application/)                        │
│  AddComboToCartHandler                                    │
│  Orchestrates resolution → submission → audit record      │
└────────────────────────┬─────────────────────────────────┘
                         │ uses
┌────────────────────────▼─────────────────────────────────┐
│  Domain Layer  (domain/)                                  │
│  CartHandoffRecord aggregate · Value Objects              │
│  Domain Events · Repository interface · Domain errors     │
└────────────────────────┬─────────────────────────────────┘
                         │ implemented by
┌────────────────────────▼─────────────────────────────────┐
│  Infrastructure Layer  (infrastructure/)                  │
│  MySQL repo · ComboPortfolioACL · PlatformCartACL         │
│  ComboResolutionService · CartSubmissionService           │
└──────────────────────────────────────────────────────────┘
```

---

## 3. Directory and File Structure

```
construction/cart_handoff/src/
├── go.mod                                          # module: github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff
├── go.sum
├── cmd/
│   └── main.go                                     # entrypoint: wires deps, starts HTTP server
├── domain/
│   ├── cart_handoff_record.go                      # CartHandoffRecord aggregate root
│   ├── value_objects.go                            # CartHandoffRecordId, ShopperId, HandoffSource, CartItem, SkippedItem, HandoffStatus, HandoffTimestamp
│   ├── events.go                                   # CartHandoffRecorded, CartHandoffFailed
│   ├── repository.go                               # CartHandoffRecordRepository interface
│   └── errors.go                                   # ComboNotFound, ComboAccessDenied, ComboPortfolioUnavailable, PlatformCartUnavailable, ErrInvalidHandoffSource
├── application/
│   ├── commands.go                                 # AddComboToCartCommand
│   └── add_combo_to_cart_handler.go                # orchestrates resolution → submission → persist
├── infrastructure/
│   ├── persistence/
│   │   ├── mysql_handoff_repository.go             # CartHandoffRecordRepository MySQL implementation
│   │   └── schema.sql                              # DDL for cart_handoff_records + handoff_record_items tables
│   ├── acl/
│   │   ├── combo_portfolio_acl.go                  # HTTP client calling Unit 4 GET /api/v1/combos/{id}
│   │   └── platform_cart_acl.go                    # HTTP client calling POST /v1/checkout/cart/bulk
│   └── services/
│       ├── combo_resolution_service.go             # branches on HandoffSource type
│       └── cart_submission_service.go              # delegates to PlatformCartACL, classifies result
├── api/
│   ├── dto.go                                      # request/response structs
│   ├── middleware.go                               # session auth middleware
│   ├── handlers.go                                 # HTTP handlers
│   └── router.go                                   # go-chi router setup
└── demo/
    ├── main.go                                     # runnable demo: 4 scenarios
    ├── stub_combo_portfolio_acl.go                 # stub for Unit 4 HTTP API
    ├── stub_platform_cart_acl.go                   # stub for Doraemon cart bulk API
    └── docker-compose.yml                          # local MySQL for demo
```

---

## 4. API Layer

### Routes

| Method | Path | Auth | Handler | Description |
|---|---|---|---|---|
| `POST` | `/api/v1/cart/combo` | Required | `AddComboToCartHandler` | Resolve combo and add all items to cart |

### Middleware Chain

```
Request → RecoverMiddleware → RequestIDMiddleware → AuthMiddleware → Handler
```

- **`AuthMiddleware`**: Same pattern as Combo Portfolio — reads session cookie, validates it via the existing session service, injects `ShopperId` and raw session cookie into request context. The raw cookie is forwarded to the Platform Cart API by `PlatformCartACL`.

### Request DTO

**`POST /api/v1/cart/combo` — Request (option A: saved combo)**
```
{
  "comboId": "combo-uuid"
}
```

**`POST /api/v1/cart/combo` — Request (option B: inline items)**
```
{
  "items": [
    { "simpleSku": "string", "quantity": 1, "size": "M" }
  ]
}
```

Exactly one of `comboId` or `items` must be present. Both present or neither → `400 Bad Request`.

### Response DTOs

**`200 OK` — all items added**
```
{
  "status": "ok",
  "addedItems": ["simpleSku1", "simpleSku2"],
  "skippedItems": []
}
```

**`200 OK` — partial**
```
{
  "status": "partial",
  "addedItems": ["simpleSku1"],
  "skippedItems": [
    { "simpleSku": "simpleSku2", "reason": "out_of_stock" }
  ]
}
```

**Error responses:** `400`, `403`, `404` mapped from domain errors (see Domain Errors table).

---

## 5. Application Layer

### `AddComboToCartHandler`

**Input:** `AddComboToCartCommand { ShopperId string, SessionCookie string, ComboId *string, Items []CartItemInput }`

**Steps:**
1. **Validate source**: Exactly one of `ComboId` or `Items` must be set. Return `ErrInvalidHandoffSource` if both or neither.
2. **Build `HandoffSource`**: Construct either `SavedComboSource` or `InlineItemsSource`.
3. **Resolve items**: Call `ComboResolutionService.Resolve(ctx, source, sessionCookie)` → returns `[]CartItem` or domain error.
   - If `ComboNotFound` → return 404.
   - If `ComboAccessDenied` → return 403.
   - If `ComboPortfolioUnavailable` → return 503.
4. **Submit to cart**: Call `CartSubmissionService.Submit(ctx, cartItems, sessionCookie)` → returns `CartSubmissionResult { AddedItems, SkippedItems }` or error.
   - If `PlatformCartUnavailable` → build `CartHandoffRecord` with status `failed`, persist, emit `CartHandoffFailed`, return 503.
5. **Determine status**: If `SkippedItems` is empty → `ok`; if both lists non-empty → `partial`.
6. **Persist record**: Construct `CartHandoffRecord` aggregate and call `CartHandoffRecordRepository.Save`.
7. **Emit event**: `CartHandoffRecorded` or `CartHandoffFailed` (in-process only at this stage).
8. **Return response DTO**.

---

## 6. Domain Layer

### `CartHandoffRecord` Aggregate — Behaviour

| Constructor | Preconditions | Result |
|---|---|---|
| `NewCartHandoffRecord(id, shopperId, source, addedItems, skippedItems, timestamp)` | source is exactly one variant; status consistency (ok/partial/failed) | Immutable record; emits `CartHandoffRecorded` or `CartHandoffFailed` |

The aggregate is **create-only** — it has no mutation methods. Once constructed and persisted, it is never modified.

**Status derivation logic (within constructor):**
- `len(skippedItems) == 0 && len(addedItems) > 0` → `ok`
- `len(skippedItems) > 0 && len(addedItems) > 0` → `partial`
- `len(addedItems) == 0` → `failed`

### Domain Errors

| Error | HTTP mapping | Cause |
|---|---|---|
| `ErrInvalidHandoffSource` | 400 | Both or neither of `comboId`/`items` provided |
| `ErrComboNotFound` | 404 | Unit 4 returns 404 |
| `ErrComboAccessDenied` | 403 | Unit 4 returns 403 (shopper doesn't own combo) |
| `ErrComboPortfolioUnavailable` | 503 | Unit 4 unreachable or 5xx |
| `ErrPlatformCartUnavailable` | 503 | Doraemon cart API unreachable or 5xx |

---

## 7. Infrastructure Layer

### MySQL Repository — `MySQLCartHandoffRepository`

**Implements:** `domain.CartHandoffRecordRepository`

**Operations:**
- `Save`: Inserts into `cart_handoff_records` and bulk-inserts into `handoff_record_items` within a single transaction. Append-only — no updates.
- `FindById`: JOIN query across both tables.
- `FindByShopperId`: Queries `cart_handoff_records` by `shopper_id`, JOINs items, orders by `recorded_at DESC`.

### Combo Portfolio ACL — `HTTPComboPortfolioACL`

**Purpose:** Calls Unit 4's `GET /api/v1/combos/{id}` with the shopper's session cookie and translates the combo item list into `[]domain.CartItem`.

**Translation mapping:** Each combo item → `CartItem { SimpleSku, Quantity: 1, Size }`. Note: `size` is not directly stored in a combo item snapshot — the `simpleSku` encodes the variant. For the purposes of this service the `size` field in the `CartItem` is set to an empty string and the platform cart API receives `simpleSku` as the primary identifier; the platform resolves size internally from the SKU.

**Error mapping:**
- HTTP 404 → `ErrComboNotFound`
- HTTP 403 → `ErrComboAccessDenied`
- HTTP 5xx / timeout / network error → `ErrComboPortfolioUnavailable`

**Interface:**
```
ComboPortfolioPort interface {
    FetchComboItems(ctx, comboId, sessionCookie) ([]CartItem, error)
}
```

### Platform Cart ACL — `HTTPPlatformCartACL`

**Purpose:** Calls Doraemon's `POST /v1/checkout/cart/bulk` and classifies the response into `AddedItems` and `SkippedItems`.

**Request construction:**
- Encodes `items` as JSON array in the `products` form field.
- Sets `Content-Language: en-SG` header.
- Sets `Accept: application/json` header.
- Forwards session cookie from the original client request.

**Response classification:**
- Iterates the platform's updated `ZDTCart.Cart` response.
- Items successfully reflected in the cart → `AddedItems`.
- Items in the request but absent from the response (out of stock) → `SkippedItem { simpleSku, reason: "out_of_stock" }`.
- HTTP 5xx or network failure → returns `ErrPlatformCartUnavailable`.

**Interface:**
```
PlatformCartPort interface {
    BulkAddToCart(ctx, items []CartItem, sessionCookie) (CartSubmissionResult, error)
}
```

### `ComboResolutionService`

**Branches on `HandoffSource` type:**
- `SavedComboSource` → calls `ComboPortfolioPort.FetchComboItems`; returns the translated `[]CartItem`.
- `InlineItemsSource` → returns the provided `[]CartItem` directly with no external call.

### `CartSubmissionService`

**Delegates to `PlatformCartPort.BulkAddToCart`.** Returns `CartSubmissionResult { AddedItems []CartItem, SkippedItems []SkippedItem }`. Does not contain business logic — pure orchestration between domain and the ACL adapter.

---

## 8. Database Schema

```sql
-- cart_handoff_records table
CREATE TABLE cart_handoff_records (
    id                  CHAR(36)        NOT NULL PRIMARY KEY,
    shopper_id          VARCHAR(255)    NOT NULL,
    source_type         ENUM('saved_combo','inline_items') NOT NULL,
    source_combo_id     CHAR(36)        NULL,      -- set when source_type = 'saved_combo'
    status              ENUM('ok','partial','failed') NOT NULL,
    recorded_at         DATETIME(3)     NOT NULL,
    INDEX idx_shopper_id (shopper_id),
    INDEX idx_recorded_at (recorded_at)
);

-- handoff_record_items table  (one row per item attempted; outcome tracks whether added or skipped)
CREATE TABLE handoff_record_items (
    id              BIGINT          NOT NULL AUTO_INCREMENT PRIMARY KEY,
    record_id       CHAR(36)        NOT NULL,
    simple_sku      VARCHAR(255)    NOT NULL,
    outcome         ENUM('added','skipped') NOT NULL,
    skip_reason     ENUM('out_of_stock','platform_error') NULL,
    FOREIGN KEY (record_id) REFERENCES cart_handoff_records(id) ON DELETE CASCADE,
    INDEX idx_record_id (record_id)
);
```

---

## 9. Configuration (Environment Variables)

| Variable | Description | Example |
|---|---|---|
| `PORT` | HTTP listen port | `8080` |
| `DB_DSN` | MySQL DSN | `user:pass@tcp(host:3306)/cart_handoff?parseTime=true` |
| `DB_MAX_OPEN_CONNS` | Max open DB connections | `25` |
| `DB_MAX_IDLE_CONNS` | Max idle DB connections | `10` |
| `COMBO_PORTFOLIO_BASE_URL` | Unit 4 service base URL | `http://combo-portfolio:8080` |
| `PLATFORM_CART_BASE_URL` | Doraemon base URL | `https://api.zalora.com` |
| `PLATFORM_CART_TIMEOUT_MS` | Per-call timeout in ms | `5000` |
| `CONTENT_LANGUAGE` | Platform locale header | `en-SG` |

---

## 10. Key Dependencies (go.mod)

| Package | Purpose |
|---|---|
| `github.com/go-chi/chi/v5` | HTTP router |
| `github.com/go-sql-driver/mysql` | MySQL driver |
| `github.com/google/uuid` | UUID generation for record IDs |
