# Logical Design: Combo Portfolio Service

---

## 1. Service Overview

| Property | Value |
|---|---|
| Service name | `combo-portfolio` |
| Runtime | Go 1.22 |
| HTTP framework | go-chi v5 |
| Deployment target | AWS ECS Fargate — standalone service |
| Listen port | `8080` |
| Architecture pattern | Clean Architecture (API → Application → Domain → Infrastructure) |
| Persistence | Amazon Aurora MySQL (dedicated new cluster, `combo_portfolio` database) |
| Event publishing | In-process only (no external bus) |
| Auth | Session cookie forwarded from client; `ShopperId` injected into request context by middleware |
| External HTTP dependency | Unit 1 Product Discovery service — `GET /api/v1/products/{configSku}` for combo item enrichment |

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
│  Command handlers · Query handlers                        │
│  Orchestrates domain objects + infrastructure services    │
└────────────────────────┬─────────────────────────────────┘
                         │ uses
┌────────────────────────▼─────────────────────────────────┐
│  Domain Layer  (domain/)                                  │
│  Combo aggregate · Value Objects · Domain Events          │
│  Repository interface · Domain errors                     │
└────────────────────────┬─────────────────────────────────┘
                         │ implemented by
┌────────────────────────▼─────────────────────────────────┐
│  Infrastructure Layer  (infrastructure/)                  │
│  MySQL repo · Product Catalog ACL · Enrichment service    │
│  Share token service                                      │
└──────────────────────────────────────────────────────────┘
```

---

## 3. Directory and File Structure

```
construction/combo_portfolio/src/
├── go.mod                                          # module: github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio
├── go.sum
├── cmd/
│   └── main.go                                     # entrypoint: wires deps, starts HTTP server
├── domain/
│   ├── combo.go                                    # Combo aggregate root
│   ├── value_objects.go                            # ComboId, ShopperId, ComboName, ComboItem, ShareToken, Visibility
│   ├── events.go                                   # ComboCreated, ComboRenamed, ComboDeleted, ComboShared, ComboMadePrivate
│   ├── repository.go                               # ComboRepository interface
│   └── errors.go                                   # ComboNotFound, ComboAccessDenied, DomainValidationError
├── application/
│   ├── commands.go                                 # SaveComboCommand, RenameComboCommand, DeleteComboCommand, ShareComboCommand, MakePrivateCommand
│   ├── queries.go                                  # GetComboQuery, ListCombosQuery, GetSharedComboQuery
│   ├── save_combo_handler.go
│   ├── rename_combo_handler.go
│   ├── delete_combo_handler.go
│   ├── share_combo_handler.go
│   ├── make_private_handler.go
│   ├── get_combo_handler.go
│   ├── list_combos_handler.go
│   └── get_shared_combo_handler.go
├── infrastructure/
│   ├── persistence/
│   │   ├── mysql_combo_repository.go               # ComboRepository MySQL implementation
│   │   └── schema.sql                              # DDL for combos + combo_items tables
│   ├── acl/
│   │   └── product_catalog_acl.go                  # HTTP client to Unit 1 GET /api/v1/products/{configSku}
│   └── services/
│       ├── combo_enrichment_service.go             # builds EnrichedCombo, fallback to snapshot
│       └── share_token_service.go                  # UUID generation with uniqueness check
├── api/
│   ├── dto.go                                      # request/response structs
│   ├── middleware.go                               # session auth middleware
│   ├── handlers.go                                 # HTTP handlers
│   └── router.go                                   # go-chi router setup
└── demo/
    ├── main.go                                     # runnable demo scenario
    ├── stub_product_acl.go                         # stub for Unit 1 HTTP API
    └── docker-compose.yml                          # local MySQL for demo
```

---

## 4. API Layer

### Routes

| Method | Path | Auth | Handler | Description |
|---|---|---|---|---|
| `POST` | `/api/v1/combos` | Required | `SaveComboHandler` | Save a new combo |
| `GET` | `/api/v1/combos` | Required | `ListCombosHandler` | List authenticated shopper's combos |
| `GET` | `/api/v1/combos/{id}` | Required | `GetComboHandler` | Get a single combo (enriched) |
| `PUT` | `/api/v1/combos/{id}` | Required | `UpdateComboHandler` | Rename combo or change visibility |
| `DELETE` | `/api/v1/combos/{id}` | Required | `DeleteComboHandler` | Delete a combo |
| `POST` | `/api/v1/combos/{id}/share` | Required | `ShareComboHandler` | Generate share token |
| `GET` | `/api/v1/combos/shared/{token}` | None | `GetSharedComboHandler` | Public share view |

### Middleware Chain

```
Request → RecoverMiddleware → RequestIDMiddleware → AuthMiddleware → Handler
```

- **`AuthMiddleware`**: Reads session cookie from the request. Calls the existing session validation service (already implemented in the platform). On success, injects `ShopperId` string into the request context. On failure, returns `401 Unauthorized`. Skipped for the public `/shared/{token}` route.

### Request/Response DTO Design

**`POST /api/v1/combos` — Request**
```
{
  name: string           // max 100 chars
  items: [{
    configSku: string
    simpleSku: string
    name: string
    imageUrl: string
    price: number
  }]                     // 2–10 items
  visibility: "public" | "private"   // default: "private"
}
```

**`GET /api/v1/combos/{id}` — Response (Enriched)**
```
{
  id: string
  name: string
  shopperId: string
  visibility: "public" | "private"
  shareToken: string | null
  createdAt: string      // ISO8601
  updatedAt: string
  items: [{
    configSku: string
    simpleSku: string
    name: string         // refreshed from catalog
    imageUrl: string     // refreshed from catalog
    price: number        // refreshed from catalog
    inStock: bool        // from catalog
    catalogUnavailable: bool   // true if product no longer in catalog
  }]
}
```

**`PUT /api/v1/combos/{id}` — Request** (fields are optional; only provided fields are updated)
```
{
  name: string           // renamed
  visibility: "public" | "private"
}
```

---

## 5. Application Layer

Each handler receives a typed command or query struct and returns a result or error. Handlers do not know about HTTP; they receive plain structs and return plain structs.

### Command Handlers

| Handler | Input | Steps |
|---|---|---|
| `SaveComboHandler` | `SaveComboCommand { ShopperId, Name, Items[], Visibility }` | 1. Validate command. 2. Construct `Combo` aggregate (enforces item count + uniqueness). 3. Call `ComboRepository.Save`. 4. Dispatch `ComboCreated` event (in-process). 5. Return saved combo ID. |
| `RenameComboHandler` | `RenameComboCommand { ShopperId, ComboId, NewName }` | 1. Load combo via repo. 2. Assert ownership (`combo.ShopperId == ShopperId`). 3. Call `combo.Rename(newName)`. 4. Call `ComboRepository.Save`. 5. Dispatch `ComboRenamed`. |
| `DeleteComboHandler` | `DeleteComboCommand { ShopperId, ComboId }` | 1. Load combo. 2. Assert ownership. 3. Call `ComboRepository.Delete`. 4. Dispatch `ComboDeleted`. |
| `ShareComboHandler` | `ShareComboCommand { ShopperId, ComboId }` | 1. Load combo. 2. Assert ownership. 3. Generate unique `ShareToken` via `ShareTokenService`. 4. Call `combo.Share(token)` — sets visibility to `public`. 5. Save. 6. Dispatch `ComboShared`. Return share token. |
| `MakePrivateHandler` | `MakePrivateCommand { ShopperId, ComboId }` | 1. Load combo. 2. Assert ownership. 3. Call `combo.MakePrivate()` — nullifies share token atomically. 4. Save. 5. Dispatch `ComboMadePrivate`. |

### Query Handlers

| Handler | Input | Steps |
|---|---|---|
| `GetComboHandler` | `GetComboQuery { ShopperId, ComboId }` | 1. Load combo. 2. Assert ownership. 3. Call `ComboEnrichmentService.Enrich(combo)`. 4. Return `EnrichedComboDTO`. |
| `ListCombosHandler` | `ListCombosQuery { ShopperId }` | 1. Load all combos for shopper via `ComboRepository.FindByShopperId`. 2. Enrich each. 3. Return `[]EnrichedComboDTO`. |
| `GetSharedComboHandler` | `GetSharedComboQuery { ShareToken }` | 1. Load combo via `ComboRepository.FindByShareToken`. 2. Verify combo exists and visibility is `public`. 3. Enrich. 4. Return `EnrichedComboDTO`. |

### `UpdateComboHandler` routing logic

The `PUT /api/v1/combos/{id}` endpoint maps to different command handlers based on the fields present in the request:
- If `name` is present → delegates to `RenameComboHandler`
- If `visibility == "private"` → delegates to `MakePrivateHandler`
- Both fields can be applied in sequence within the same request

---

## 6. Domain Layer

### `Combo` Aggregate — Behaviour

| Method | Preconditions | State change | Event emitted |
|---|---|---|---|
| `NewCombo(id, shopperId, name, items, visibility)` | items len ∈ [2,10]; no duplicate simpleSku | Sets all fields; `updatedAt = createdAt` | `ComboCreated` |
| `Rename(newName)` | name non-empty, ≤100 chars | Updates `Name`, `UpdatedAt` | `ComboRenamed` |
| `Share(token)` | combo not already public (idempotent if already public with same token) | Sets `Visibility = public`, `ShareToken = token`, `UpdatedAt` | `ComboShared` |
| `MakePrivate()` | — | Sets `Visibility = private`, `ShareToken = nil`, `UpdatedAt` | `ComboMadePrivate` |

### Domain Errors

| Error | HTTP mapping |
|---|---|
| `ErrComboNotFound` | 404 |
| `ErrComboAccessDenied` | 403 |
| `ErrInvalidItemCount` | 400 |
| `ErrDuplicateItem` | 400 |
| `ErrInvalidComboName` | 400 |
| `ErrShareTokenConflict` | 500 (retry) |

---

## 7. Infrastructure Layer

### MySQL Repository — `MySQLComboRepository`

**Implements:** `domain.ComboRepository`

**Operations:**
- `Save`: Upserts the `combos` row and replaces all `combo_items` rows for that combo (delete + insert within a transaction).
- `FindById`: Single JOIN query across `combos` and `combo_items`.
- `FindByShopperId`: Queries `combos` filtered by `shopper_id`, JOINs `combo_items`, orders by `created_at DESC`.
- `FindByShareToken`: Queries `combos` where `share_token = ?` and `visibility = 'public'`.
- `Delete`: Deletes from `combos` (cascade deletes `combo_items` via FK).

**Connection:** Uses `database/sql` with `go-sql-driver/mysql`. Connection pool configured via env vars (`DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`).

### Product Catalog ACL — `HTTPProductCatalogACL`

**Purpose:** Calls Unit 1's `GET /api/v1/products/{configSku}` and translates the response into a `CatalogProduct` struct used by `ComboEnrichmentService`.

**Behaviour:**
- Accepts a list of `configSku` values and fans out HTTP calls (one per unique configSku in the combo).
- Parses the response to extract the variant (`simpleSku`) matching the stored `simpleSku`, extracting `name`, `imageUrl`, `price`, and `inStock` status.
- On HTTP 404 → marks item as `catalogUnavailable: true`; falls back to snapshot.
- On HTTP 5xx or timeout → falls back to snapshot for that item only; does not fail the whole request.
- Timeout: 2 seconds per call. Calls are made concurrently (one goroutine per configSku).

**Interface:**
```
ProductCatalogPort interface {
    FetchProduct(ctx, configSku) (CatalogProduct, error)
}
```

### `ComboEnrichmentService`

**Purpose:** Merges live catalog data into a combo's item snapshots to produce an `EnrichedCombo`.

**Algorithm:**
1. Collect all unique `configSku` values from the combo's items.
2. Concurrently call `ProductCatalogPort.FetchProduct` for each.
3. For each `ComboItem`, match on `simpleSku` within the catalog product's variants.
4. Build `EnrichedComboItem` with live data, or with snapshot data + `catalogUnavailable: true` if not found.
5. Return `EnrichedCombo` (read model only — never mutates the aggregate).

### `ShareTokenService`

**Purpose:** Generates a `ShareToken` that is globally unique within the combo portfolio.

**Algorithm:**
1. Generate a UUID v4.
2. Call `ComboRepository.FindByShareToken(token)`.
3. If nil returned → token is unique; return it.
4. If a combo is found → regenerate (collision, expected probability ≈ 0).

---

## 8. Database Schema

```sql
-- combos table
CREATE TABLE combos (
    id           CHAR(36)     NOT NULL PRIMARY KEY,
    shopper_id   VARCHAR(255) NOT NULL,
    name         VARCHAR(100) NOT NULL,
    visibility   ENUM('public','private') NOT NULL DEFAULT 'private',
    share_token  CHAR(36)     NULL,
    created_at   DATETIME(3)  NOT NULL,
    updated_at   DATETIME(3)  NOT NULL,
    INDEX idx_shopper_id (shopper_id),
    UNIQUE INDEX idx_share_token (share_token)
);

-- combo_items table
CREATE TABLE combo_items (
    id           BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
    combo_id     CHAR(36)     NOT NULL,
    config_sku   VARCHAR(255) NOT NULL,
    simple_sku   VARCHAR(255) NOT NULL,
    name         VARCHAR(500) NOT NULL,
    image_url    TEXT         NOT NULL,
    price        DECIMAL(10,2) NOT NULL,
    sort_order   TINYINT      NOT NULL DEFAULT 0,
    FOREIGN KEY (combo_id) REFERENCES combos(id) ON DELETE CASCADE,
    INDEX idx_combo_id (combo_id)
);
```

---

## 9. Configuration (Environment Variables)

| Variable | Description | Example |
|---|---|---|
| `PORT` | HTTP listen port | `8080` |
| `DB_DSN` | MySQL DSN | `user:pass@tcp(host:3306)/combo_portfolio?parseTime=true` |
| `DB_MAX_OPEN_CONNS` | Max open DB connections | `25` |
| `DB_MAX_IDLE_CONNS` | Max idle DB connections | `10` |
| `PRODUCT_CATALOG_BASE_URL` | Unit 1 service base URL | `http://product-discovery:8080` |
| `PRODUCT_CATALOG_TIMEOUT_MS` | Per-call timeout in ms | `2000` |
| `CONTENT_LANGUAGE` | Platform locale header | `en-SG` |

---

## 10. Key Dependencies (go.mod)

| Package | Purpose |
|---|---|
| `github.com/go-chi/chi/v5` | HTTP router |
| `github.com/go-sql-driver/mysql` | MySQL driver |
| `github.com/google/uuid` | UUID generation for IDs and share tokens |
