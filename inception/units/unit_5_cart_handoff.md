# Unit 5: Cart Handoff

## Purpose

Bridges a confirmed combo (from the AI Styling Engine or a saved combo) to the existing in-app cart. This is a deliberately thin unit — it resolves combo items, calls the platform bulk cart API, and hands control to the existing checkout flow. It owns no storage and introduces no checkout logic.

## Implementation Note

This unit is implemented independently. All information needed to build it is contained in this file. It has one inter-unit dependency and one platform dependency at runtime:

1. **Unit 4 (Combo Portfolio)** — when the client provides a `comboId`, this service fetches the saved combo's items from Unit 4. The contract for this call is specified below. During development, mock this with a fixed combo response.
2. **Platform Cart API** — to perform the bulk add. Contract specified below.

When the client sends an inline `items` array (unsaved combo from Unit 3), Unit 4 is not called.

## Scope

| Layer | Responsibility |
|---|---|
| Service API | Resolves combo items and triggers bulk cart add |
| Client UI | "Add All to Cart" button on Unit 3 combo cards and Unit 4 saved combo detail pages; out-of-stock warning modal |

---

## Consumed Unit APIs (Mock This During Development)

### Unit 4 — `GET /api/v1/combos/{id}`

Called when the client provides a `comboId` to resolve the saved combo's item list.

**Auth:** Forwarded session cookie from the incoming client request.

**Response to expect:**

```json
{
  "id": "combo-uuid",
  "name": "My Summer Look",
  "createdAt": "2026-03-31T10:00:00Z",
  "visibility": "private",
  "shareToken": null,
  "items": [
    {
      "configSku": "string",
      "simpleSku": "string",
      "name": "string",
      "imageUrl": "string",
      "price": 89.90
    }
  ]
}
```

**Error cases:** `404` if combo not found, `403` if shopper does not own it.

---

## Platform APIs This Unit Calls

All platform requests require:
- Header: `Content-Language: en-SG`
- Header: `Accept: application/json`
- Auth: Session cookie forwarded from the client.

### `POST /v1/checkout/cart/bulk`

Adds multiple items to the cart in one request.

Form data: `products` — JSON-encoded array of items to add.

```json
[
  { "simpleSku": "string", "quantity": 1, "size": "M" }
]
```

Response: Updated `ZDTCart.Cart` object. If `checkLimit` is set and the cart limit is exceeded, no items are added.

---

## User Story

### US-601 — Add Confirmed Combo to Cart

**As a** registered shopper,
**I want to** add all items in a confirmed combo to my cart in a single action,
**So that** I can quickly proceed to the existing checkout without adding each item individually.

**API endpoint:** `POST /api/v1/cart/combo`

**Acceptance Criteria:**

- An "Add All to Cart" button is present on combo cards (Unit 3) and saved combo detail pages (Unit 4).
- Clicking it calls `POST /api/v1/cart/combo`, which resolves items and forwards them to `POST /v1/checkout/cart/bulk`.
- All in-stock items are added in a single bulk request.
- If any item is out of stock, the shopper is shown a warning listing the affected items and can choose to add the remaining items or cancel.
- The shopper can also add individual items one at a time (client calls platform `POST /v1/checkout/cart` directly).
- After a successful bulk add, the shopper is navigated to the existing in-app cart and checkout flow.

---

## Frontend Screens & Components

### Component: "Add All to Cart" Button

Reused in two places: the Combo Suggestion Page (Unit 3) and the Combo Detail Page (Unit 4). The button behaviour is identical in both contexts.

**States:**

- **Default**: "Add All to Cart" — enabled.
- **Loading**: spinner + "Adding…" — shown while `POST /api/v1/cart/combo` is in flight.
- **Success**: brief "Added to cart!" confirmation state (1–2 seconds), then navigates to the existing in-app cart page.
- **Disabled**: shown if all items in the combo are out of stock.

**Triggering the call:**

- From Unit 3 (unsaved combo): sends `{ "items": [...] }` with the inline item list from the combo response.
- From Unit 4 (saved combo): sends `{ "comboId": "..." }`.

---

### Component: Out-of-Stock Warning Modal

Shown when the response from `POST /api/v1/cart/combo` returns `"status": "partial"`.

**Layout:**

- Title: "Some items are unavailable"
- Body: list of skipped items (name + reason), e.g., *"Linen Blazer — Out of stock"*.
- Buttons:
  - **"Add available items anyway"** — proceeds with the partial add; navigates to the existing in-app cart page.
  - **"Cancel"** — dismisses the modal; no items are added.

---

## Service API This Unit Exposes

### `POST /api/v1/cart/combo`

Resolves combo items and adds them to the cart in one call.

**Auth:** Required

**Request Body — option A (saved combo by ID):**

```json
{
  "comboId": "combo-uuid"
}
```

Service fetches item list from Unit 4 `GET /api/v1/combos/{id}`, then calls `POST /v1/checkout/cart/bulk`.

**Request Body — option B (inline items, unsaved combo from Unit 3):**

```json
{
  "items": [
    { "simpleSku": "string", "quantity": 1, "size": "M" }
  ]
}
```

Service calls `POST /v1/checkout/cart/bulk` directly with the provided items.

- Provide either `comboId` **or** `items` — not both. Returns `400 Bad Request` if both or neither are provided.

**Response `200 OK` — all items added:**

```json
{
  "status": "ok",
  "addedItems": ["simpleSku1", "simpleSku2"],
  "skippedItems": []
}
```

**Response `200 OK` — partial (some items out of stock):**

```json
{
  "status": "partial",
  "addedItems": ["simpleSku1"],
  "skippedItems": [
    { "simpleSku": "simpleSku2", "reason": "out_of_stock" }
  ]
}
```

**Response `400 Bad Request`:** Both or neither of `comboId`/`items` provided.
**Response `403 Forbidden`:** Shopper is not authenticated, or does not own the referenced combo.
**Response `404 Not Found`:** `comboId` does not exist.
