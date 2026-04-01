# Domain Model — Unit 2: Wishlist Management

## Bounded Context

**Name:** Wishlist Management

**Nature:** Write-capable domain facade. This bounded context models the shopper's wishlist as a domain aggregate, but owns no local persistence. All state is held by the upstream Platform Wishlist API. The repository implementations delegate to the platform via an Anti-Corruption Layer (ACL). The domain model exists to enforce invariants (duplicate prevention, auth gate), model intent, and emit domain events — not to maintain a local database.

**Context Map Position:**
- **Upstream (Platform):** Platform Wishlist API (External System — ACL-protected), Platform Auth API (External System — ACL-protected)
- **Downstream consumer:** Unit 3 AI Styling Engine reads `GET /api/v1/wishlist` from this context
- **Client:** Receives shaped wishlist responses; emits add/remove intents

---

## Aggregates

### `Wishlist` (Aggregate Root)

The central aggregate of this bounded context. Represents a single shopper's curated collection of product variants they intend to potentially purchase.

**Identity:** `WishlistId` — scoped to a `ShopperId`. Each authenticated shopper has exactly one wishlist in the platform.

**Fields:**
- `wishlistId` — `WishlistId` value object
- `shopperId` — `ShopperId` value object
- `items` — collection of `WishlistItem` entities
- `totalCount` — integer (total items in the platform, may exceed the current page)

**Invariants:**
- A shopper may not have the same `configSku` in the wishlist more than once (duplicate prevention).
- Only an authenticated shopper may mutate the wishlist.
- The wishlist may be empty (zero items is a valid state).

**Behaviour (Commands):**
- `addItem(simpleSku)` — checks for duplicate by `configSku`; if duplicate, raises `WishlistItemAlreadyPresent`; otherwise delegates to the repository and emits `WishlistItemAdded`.
- `removeItem(configSku)` — removes all variants of the product; delegates to repository and emits `WishlistItemRemoved`.
- `toggleItem(simpleSku, configSku)` — if the item is present, calls `removeItem`; if absent, calls `addItem`. Used for heart-toggle UX.

**Note on reconstitution:** Because the platform is the system of record, the `Wishlist` aggregate is reconstituted from the platform API response on each request via the repository. There is no local snapshot.

---

## Entities

### `WishlistItem`

A single product variant entry within the `Wishlist` aggregate. It has identity within the aggregate.

**Identity:** `WishlistItemId` (platform-assigned `itemId`)

**Fields:**
- `itemId` — `WishlistItemId` value object
- `simpleSku` — `SimpleSku` value object (identifies the specific size/colour variant)
- `configSku` — `ConfigSku` value object (identifies the product family; used for duplicate checking and removal)
- `name` — string
- `brand` — string
- `price` — `Money` value object
- `imageUrl` — string
- `color` — string
- `size` — string
- `inStock` — boolean

**Behaviour:**
- `WishlistItem` is created only through the `Wishlist` aggregate root — never instantiated directly by external callers.

---

## Value Objects

### `WishlistId`

Identifies the shopper's wishlist as returned by the platform.

**Fields:**
- `value` — string (platform-assigned identifier)

---

### `ShopperId`

Identifies the authenticated shopper. Derived from the session context.

**Fields:**
- `value` — string (platform shopper/customer ID, extracted from the auth session)

---

### `WishlistItemId`

Identifies a specific item entry in the wishlist, as assigned by the platform.

**Fields:**
- `value` — string (platform `itemId`)

---

### `SimpleSku`

Identifies a specific purchasable variant of a product (a unique size + colour combination).

**Fields:**
- `value` — string (e.g., `"SG-12345-M-BLK"`)

**Invariants:**
- Must be a non-empty string.

---

### `ConfigSku`

Identifies a product configuration (the parent product grouping all variants).

**Fields:**
- `value` — string (e.g., `"SG-12345"`)

**Invariants:**
- Must be a non-empty string.

---

### `Money`

Represents a price amount. Currency is implied by the service's locale context.

**Fields:**
- `amount` — decimal, non-negative

---

### `Pagination`

Represents paging parameters for wishlist list queries.

**Fields:**
- `offset` — integer, default 0, non-negative
- `limit` — integer, default 20, positive

---

### `PendingWishlistAdd`

Captures the intent of an unauthenticated guest to add a product to their wishlist, so that the add can be completed automatically after successful authentication. This is a transient value object held in client-side session state; it is not persisted by this service.

**Fields:**
- `simpleSku` — `SimpleSku` value object
- `returnPath` — the URL the shopper was viewing when they triggered the wishlist add, used to redirect them after login

---

## Domain Events

Domain events represent facts that have occurred within the Wishlist bounded context. They are emitted by the `Wishlist` aggregate and may be consumed by other parts of the system.

---

### `WishlistItemAdded`

Emitted when a product variant is successfully added to the shopper's wishlist.

**Payload:**
- `shopperId` — `ShopperId`
- `simpleSku` — `SimpleSku`
- `configSku` — `ConfigSku`
- `itemId` — `WishlistItemId` (as returned by the platform after the add)
- `occurredAt` — timestamp

**Consumers (within this bounded context):**
- Application layer updates the client response with the new `itemId`.

**Potential downstream consumers:**
- Unit 3 AI Styling Engine may react to this event if it maintains a local wishlist cache (out of scope for Unit 2).

---

### `WishlistItemRemoved`

Emitted when all variants of a product are removed from the shopper's wishlist.

**Payload:**
- `shopperId` — `ShopperId`
- `configSku` — `ConfigSku`
- `occurredAt` — timestamp

---

### `AuthenticationGateTriggered`

Emitted when an unauthenticated guest attempts to add an item to the wishlist. This event signals that the auth gate modal should be presented to the user and the pending add intent should be captured.

**Payload:**
- `simpleSku` — `SimpleSku`
- `returnPath` — string (the page the user was on)
- `occurredAt` — timestamp

**Consumer:**
- The application layer intercepts this event and returns a response instructing the client to show the login modal and store a `PendingWishlistAdd`.

---

## Policies

Policies are domain rules that respond to conditions or events and enforce invariants.

---

### Duplicate Prevention Policy

**Trigger:** `Wishlist.addItem(simpleSku)` is called.

**Rule:** Before delegating the add to the repository, the aggregate checks whether any existing `WishlistItem` shares the same `configSku` as the incoming `simpleSku`. If a match is found, the add is rejected and the application layer treats the operation as a no-op (toggle-off semantics are handled by `toggleItem`, not `addItem` directly).

**Why:** The platform's wishlist does not guarantee idempotency on duplicate adds. This rule ensures the client receives a consistent response without making an unnecessary platform API call.

---

### Authentication Gate Policy

**Trigger:** Any wishlist mutation command (`addItem`, `removeItem`, `toggleItem`) is attempted.

**Rule:** Before executing the command, the application layer checks the session for an authenticated `ShopperId`. If no valid session exists, the command is not executed; instead, `AuthenticationGateTriggered` is emitted and a `PendingWishlistAdd` is created with the shopper's intent.

**Why:** The platform wishlist API returns `403` for unauthenticated requests. The gate is enforced at the application layer rather than relying on the platform error, enabling the service to present a graceful login modal flow rather than a raw error.

---

### Post-Authentication Auto-Add Policy

**Trigger:** Shopper successfully authenticates with a `PendingWishlistAdd` present in their session.

**Rule:** After authentication completes (via the existing platform auth service), the application layer reads the pending intent and automatically calls `Wishlist.addItem(simpleSku)` on behalf of the newly authenticated shopper. The shopper is then redirected to the `returnPath`.

**Why:** User story US-201 requires that the product is added automatically after login, with no additional action from the shopper.

---

### Toggle Policy

**Trigger:** A heart-button tap on the client for a product the shopper has already wishlisted.

**Rule:** `Wishlist.toggleItem(simpleSku, configSku)` checks the in-memory item collection. If the `configSku` is already present, it calls `removeItem`; if absent, it calls `addItem`. This prevents the client from having to track wishlist state independently.

---

## Repositories

Repositories define the contract for retrieving and mutating the `Wishlist` aggregate. In this bounded context, all implementations delegate to the Platform Wishlist API via the ACL — there is no local database.

---

### `WishlistRepository` (interface)

**Operations:**
- `getByShopperId(shopperId, pagination)` → `Wishlist` — reconstitutes the aggregate from the platform API response
- `addItem(shopperId, simpleSku)` → `WishlistItemId` — calls the platform add endpoint and returns the platform-assigned item ID
- `removeItemByConfigSku(shopperId, configSku)` → void — calls the platform delete endpoint

**Implementation note:** The concrete implementation (`PlatformWishlistRepository`) uses `PlatformWishlistApiClient` (ACL) to make platform API calls. It maps platform response structures to domain objects using the `WishlistAssembler`.

---

## Domain Services

### `WishlistToggleService`

Orchestrates the toggle behaviour when the exact wishlist state of a `configSku` must be determined before deciding to add or remove.

**Responsibilities:**
- Loads the current `Wishlist` via the repository.
- Calls `toggleItem` on the aggregate.
- Returns the resulting event (`WishlistItemAdded` or `WishlistItemRemoved`) so the application layer can construct the appropriate response.

---

### `AuthSessionService` (ACL-backed)

Validates the current shopper session and resolves the authenticated `ShopperId`.

**Responsibilities:**
- Calls the Platform Auth API to verify the session cookie.
- Returns a `ShopperId` if authenticated, or raises an `UnauthenticatedShopperError`.
- Used by the application layer before any wishlist mutation command is executed.

---

## Anti-Corruption Layer (ACL)

### `PlatformWishlistApiClient` (ACL interface)

Translates between the service's domain concepts and the Platform Wishlist API.

**Operations:**
- `fetchWishlist(sessionCookie, offset, limit)` → raw platform wishlist payload
- `addItem(sessionCookie, simpleSku)` → raw platform wishlist item payload
- `removeItemsByConfigSku(sessionCookie, configSku)` → void

**Responsibilities:**
- Attaches required platform headers (`Content-Language`, `Accept`, session cookie).
- Translates platform `403` responses to `UnauthenticatedShopperError` (domain error), shielding the aggregate from HTTP concerns.
- Does not perform data shaping — raw platform payloads are passed to the `WishlistAssembler`.

---

### `WishlistAssembler`

Transforms raw platform API payloads into domain objects.

**Responsibilities:**
- Maps platform `wishlist` payload → `Wishlist` aggregate (reconstitution)
- Maps platform `wishListItem` payload → `WishlistItem` entity
- Maps platform field names to domain names (e.g., `createdAt` epoch → timestamp, `itemId` → `WishlistItemId`, etc.)

---

### `PlatformAuthApiClient` (ACL interface)

Used by `AuthSessionService` to validate the session.

**Operations:**
- `validateSession(sessionCookie)` → `ShopperId` or null

---

## Summary Table

| Concept | Type | Notes |
|---|---|---|
| `Wishlist` | Aggregate Root | Domain facade over platform wishlist; no local persistence |
| `WishlistItem` | Entity | Single variant entry; identity via `WishlistItemId` |
| `WishlistId` | Value Object | Platform-assigned wishlist identifier |
| `ShopperId` | Value Object | Authenticated shopper identifier from session |
| `WishlistItemId` | Value Object | Platform-assigned item entry identifier |
| `SimpleSku` | Value Object | Variant-level SKU (size + colour) |
| `ConfigSku` | Value Object | Product-level SKU (parent of all variants) |
| `Money` | Value Object | Price amount |
| `Pagination` | Value Object | Offset + limit |
| `PendingWishlistAdd` | Value Object | Transient guest intent; held in client session |
| `WishlistItemAdded` | Domain Event | Emitted after successful add |
| `WishlistItemRemoved` | Domain Event | Emitted after successful remove |
| `AuthenticationGateTriggered` | Domain Event | Emitted when unauthenticated add is attempted |
| Duplicate Prevention Policy | Policy | Prevents double-adding same configSku |
| Authentication Gate Policy | Policy | Blocks mutations if unauthenticated |
| Post-Authentication Auto-Add Policy | Policy | Completes pending add after login |
| Toggle Policy | Policy | Add if absent, remove if present |
| `WishlistRepository` | Repository Interface | Delegates to platform via ACL |
| `WishlistToggleService` | Domain Service | Orchestrates toggle add/remove logic |
| `AuthSessionService` | Domain Service (ACL-backed) | Resolves ShopperId from session |
| `PlatformWishlistApiClient` | ACL Interface | Translates to/from Platform Wishlist API |
| `WishlistAssembler` | ACL Assembler | Maps platform payloads → domain objects |
| `PlatformAuthApiClient` | ACL Interface | Validates session; returns ShopperId |

---

## Cross-Unit Integration

### Unit 2 → Unit 3 (AI Styling Engine)

Unit 3 is a downstream consumer of this bounded context. It reads the shopper's wishlist server-side via `GET /api/v1/wishlist` using the shopper's session. This endpoint is backed by `WishlistRepository.getByShopperId(...)` and returns the `Wishlist` aggregate's items shaped by the `ProductListReadModel`.

**Published Language:** The contract Unit 3 relies on is the `GET /api/v1/wishlist` response schema. Any changes to `WishlistItem` fields that affect that schema are breaking changes for Unit 3.

**Key fields Unit 3 depends on:** `configSku`, `simpleSku`, `inStock`, `name`, `brand`, `price`, `imageUrl`, `color`, `size`.

### Unit 3 Client → Unit 2 (Add Suggested Items)

The Unit 3 client may call `POST /api/v1/wishlist/items` to add AI-suggested catalog products to the shopper's wishlist. This flows through the standard `Wishlist.addItem` command, subject to the same Duplicate Prevention and Authentication Gate policies.

### Shared Kernel Candidates

- `ConfigSku` and `SimpleSku` value objects are referenced by both Unit 1 and Unit 2. If a shared kernel is introduced, these should live there. Until then, each bounded context defines its own copy — conforming to the DDD principle of avoiding coupling at the model level.
