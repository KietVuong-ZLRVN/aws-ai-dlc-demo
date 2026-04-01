# Domain Model: Cart Handoff

---

## 1. Bounded Context

**Context Name:** Cart Handoff Context

**Responsibility:** Bridges a confirmed combo (either a saved combo from the Combo Portfolio or an inline unsaved combo from the AI Styling Engine) to the platform's shopping cart in a single operation. It resolves the item list, submits a bulk cart request to the platform, and records an immutable audit log of every handoff attempt.

**Key architectural constraint:** This context owns exactly one type of persistent data — the `CartHandoffRecord` audit log. It introduces no cart logic, no pricing logic, and no checkout logic. All of those remain with the platform.

**Bounded Context Boundaries:**
- **Upstream dependency (combo resolution):** Combo Portfolio Context (Unit 4) — called when the handoff source is a saved combo identified by `ComboId`.
- **Upstream dependency (cart submission):** Platform Cart API — called to perform the bulk add-to-cart operation.
- **Downstream consumer:** None. The context returns a result to the client and records the outcome; it does not publish to other contexts.

---

## 2. Ubiquitous Language

| Term | Definition |
|---|---|
| **Cart Handoff** | The end-to-end operation of taking a combo's resolved item list and submitting it to the platform cart in one action. |
| **Handoff Source** | The origin of the items being added. Either a saved Combo (identified by `ComboId`) or an inline item list (unsaved AI suggestion passed directly by the client). |
| **Resolved Item** | A `CartItem` that has been confirmed to proceed to the cart add — either because it is in stock, or because the shopper chose to proceed with available items. |
| **Added Item** | A `CartItem` that was successfully accepted by the platform cart during a handoff. |
| **Skipped Item** | A `CartItem` excluded from the cart add due to being out of stock. |
| **Handoff Record** | An immutable audit log entry that captures the inputs, outcome, and timing of a single cart handoff attempt. Created once; never mutated. |
| **Partial Handoff** | A handoff where at least one item was added and at least one was skipped due to out-of-stock. |
| **Failed Handoff** | A handoff where the platform cart API rejected the entire request (e.g., system error, auth failure). |
| **Combo Resolution** | The process of fetching and translating a saved combo's item list from the Combo Portfolio Context into a set of `CartItem`s. |

---

## 3. Aggregates

### Aggregate Root: `CartHandoffRecord`

Represents a single, immutable record of a cart handoff attempt. It is created once (when the handoff completes or fails) and is never mutated. Its purpose is audit — it captures who requested what, what was resolved, and what the outcome was.

**Invariants enforced by the aggregate:**
- A `CartHandoffRecord` must have exactly one `HandoffSource` — either a `ComboId` or an inline `CartItem` list, never both and never neither.
- `addedItems` and `skippedItems` together must account for all items in the resolved item list.
- `HandoffStatus` must be consistent with `addedItems` and `skippedItems`:
  - `ok` → `skippedItems` is empty, `addedItems` is non-empty.
  - `partial` → both lists are non-empty.
  - `failed` → `addedItems` is empty; `skippedItems` may list all items with reason `platform_error`.
- `ShopperId` must be present and non-empty.

---

## 4. Entities

There are no subordinate entities within the `CartHandoffRecord` aggregate beyond the aggregate root itself. All inner data is modelled as Value Objects.

---

## 5. Value Objects

### `CartHandoffRecordId`
- **Type:** UUID string
- **Constraints:** Immutable; globally unique; generated at record creation.
- **Description:** The identity of a `CartHandoffRecord`.

---

### `ShopperId`
- **Type:** String
- **Constraints:** Non-empty; sourced from the authentication context. Immutable.
- **Description:** Identifies the shopper who initiated the handoff.

---

### `HandoffSource`
- **Type:** Discriminated union
- **Variants:**
  - `SavedComboSource { type: "saved_combo", comboId: string }` — the handoff was initiated from a saved combo in the Combo Portfolio.
  - `InlineItemsSource { type: "inline_items", items: CartItem[] }` — the handoff was initiated from an unsaved AI suggestion, with items passed inline by the client.
- **Constraints:** Exactly one variant must be present. An empty inline items list is rejected.
- **Description:** Captures the origin of the items for traceability in the audit log.

---

### `CartItem`
- **Fields:**
  - `simpleSku` — String; the specific product variant identifier submitted to the cart.
  - `quantity` — Integer; always 1 for combo items.
  - `size` — String; the selected size for the variant.
- **Constraints:** `simpleSku` and `size` are non-empty. `quantity` is ≥ 1.
- **Description:** A single item submitted (or attempted) in a bulk cart add. Used in both the resolved item list and the `addedItems` record.

---

### `SkippedItem`
- **Fields:**
  - `simpleSku` — String; the variant SKU that was excluded.
  - `reason` — Enum: `out_of_stock` | `platform_error`
- **Description:** Records why a specific item was not added to the cart during the handoff.

---

### `HandoffStatus`
- **Type:** Enum — `ok` | `partial` | `failed`
- **Semantics:**
  - `ok` — All resolved items were successfully added to the cart.
  - `partial` — Some items were added; others were skipped due to `out_of_stock`.
  - `failed` — The platform cart API call failed entirely; no items were added.

---

### `HandoffTimestamp`
- **Type:** ISO8601 datetime string (UTC)
- **Constraints:** Set at record creation; immutable.
- **Description:** The wall-clock time when the handoff attempt was recorded.

---

## 6. Domain Events

### `CartHandoffRecorded`
- **Trigger:** A `CartHandoffRecord` is successfully created after a `ok` or `partial` handoff outcome.
- **Payload:** `recordId`, `shopperId`, `handoffSource`, `handoffStatus`, `addedItemCount`, `skippedItemCount`, `timestamp`
- **Consumers:** Audit log store.

---

### `CartHandoffFailed`
- **Trigger:** A `CartHandoffRecord` is created with `HandoffStatus.failed` — the platform cart API returned an error for the entire bulk request.
- **Payload:** `recordId`, `shopperId`, `handoffSource`, `failureReason`, `timestamp`
- **Consumers:** Audit log store; may be used by operations/alerting.

---

## 7. Policies

### `HandoffSourceValidationPolicy`
- **Reacts to:** Any incoming cart handoff command
- **Rule:** The request must contain either a `comboId` or an inline `items` array — not both, and not neither. If both are present or neither is present, the command is rejected with a `400 Bad Request` before any processing begins.
- **Enforced by:** The `CartHandoffRecord` aggregate constructor and the application service layer.

---

### `HandoffOwnershipPolicy`
- **Reacts to:** A handoff command with a `comboId` source
- **Rule:** The shopper initiating the handoff must own the referenced saved combo. If the Combo Portfolio ACL returns a `403 Forbidden`, the handoff is rejected with `403 Forbidden` without creating a `CartHandoffRecord`.
- **Enforced by:** Application service layer, informed by the `ComboPortfolioACL` response.

---

## 8. Repositories

### `CartHandoffRecordRepository`

| Method | Description |
|---|---|
| `save(record: CartHandoffRecord)` | Persists a new (immutable) `CartHandoffRecord`. No updates — records are append-only. |
| `findById(id: CartHandoffRecordId): CartHandoffRecord \| null` | Retrieves a specific handoff record by ID. Used for audit lookups. |
| `findByShopperId(shopperId: ShopperId): CartHandoffRecord[]` | Returns the handoff history for a shopper, ordered by `timestamp` descending. Used for audit and support. |

---

## 9. Domain Services

### `ComboResolutionService`
- **Purpose:** Resolves the final `CartItem` list for a handoff based on the `HandoffSource`.
- **Why a Domain Service:** Resolution logic requires branching on the source type and, for `SavedComboSource`, calling an external bounded context. This logic spans infrastructure concerns and cannot live in the aggregate.
- **Inputs:** A `HandoffSource`, the shopper's session context
- **Outputs:** A list of `CartItem` value objects
- **Behaviour:**
  - If source is `SavedComboSource`: delegates to `ComboPortfolioACL` to fetch the combo's item list, then translates the result into `CartItem`s.
  - If source is `InlineItemsSource`: returns the provided `CartItem`s directly without any external call.
- **Error handling:** Propagates domain exceptions from the ACL — `ComboNotFound`, `ComboAccessDenied`.

---

### `CartSubmissionService`
- **Purpose:** Submits a resolved list of `CartItem`s to the platform cart and returns the outcome (added items and skipped items).
- **Why a Domain Service:** The platform cart API is an external system. Calling it, interpreting its response, and mapping the result into domain terms is not a responsibility of any aggregate.
- **Inputs:** A list of `CartItem` value objects, the shopper's session context
- **Outputs:** `{ addedItems: CartItem[], skippedItems: SkippedItem[] }`
- **Behaviour:** Delegates to `PlatformCartACL` to perform the bulk add. Maps the platform's response to domain terms. Classifies skipped items as `out_of_stock` or `platform_error` based on the platform response codes.

---

## 10. Anti-Corruption Layer (ACL) Adapters

ACL adapters protect the Cart Handoff domain model from being contaminated by the data structures and vocabulary of external systems.

### `ComboPortfolioACL`
- **Purpose:** Translates the Combo Portfolio Context's `GET /api/v1/combos/{id}` HTTP API response into the Cart Handoff domain model's `CartItem` value objects.
- **External contract consumed:** Combo Portfolio `GET /api/v1/combos/{id}` (see integration contract)
- **Translation mapping:**
  - Each item in the combo response `items[]` → one `CartItem { simpleSku, quantity: 1, size }`. Note: `size` is derived from the `simpleSku` variant data if not explicitly present in the combo item response.
- **Domain exceptions raised:**
  - `ComboNotFound` — when the external API returns `404`
  - `ComboAccessDenied` — when the external API returns `403`
  - `ComboPortfolioUnavailable` — when the external API is unreachable (network error, 5xx)

---

### `PlatformCartACL`
- **Purpose:** Encapsulates the platform `POST /v1/checkout/cart/bulk` interaction, translating the domain's `CartItem` list into the platform's expected format and translating the response back into domain terms.
- **External contract consumed:** Platform `POST /v1/checkout/cart/bulk` (form data with `products` as JSON-encoded array)
- **Request translation:** Maps `CartItem[]` → `[{ simpleSku, quantity, size }]` encoded as form data.
- **Platform-specific concerns handled internally:**
  - Sets `Content-Language: en-SG` header
  - Sets `Accept: application/json` header
  - Forwards session cookie from the original client request
- **Response translation:**
  - Platform accepts all items → all items mapped to `addedItems`
  - Platform signals individual item rejection (out of stock) → affected items mapped to `SkippedItem { simpleSku, reason: "out_of_stock" }`
  - Platform returns 5xx or network failure → raises `PlatformCartUnavailable` domain exception, which results in `HandoffStatus.failed`

---

## 11. Context Map

| Relationship | Direction | Pattern |
|---|---|---|
| Cart Handoff ↔ Combo Portfolio Context | Outbound (read combo items) | **Anticorruption Layer** — `ComboPortfolioACL` translates Combo Portfolio API responses into Cart Handoff domain objects. |
| Cart Handoff ↔ Platform Cart API | Outbound (bulk cart add) | **Anticorruption Layer** — `PlatformCartACL` encapsulates platform-specific request/response formats, headers, and error codes. |
| Cart Handoff ↔ Auth Context | Inbound (trusted) | **Conformist** — `ShopperId` and session tokens are accepted as-is from the auth layer; session cookie is forwarded to platform calls. |
