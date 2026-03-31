# Unit 2: Wishlist Management

## Purpose

Manages the shopper's wishlist — the primary input surface for the AI Styling Engine. Handles the authentication gate when a guest tries to wishlist an item, and wraps the existing platform wishlist APIs into a single, client-ready service following the Backend for Frontend philosophy.

## Implementation Note

This unit is implemented independently. All information needed to build it is contained in this file. Authentication (login/register) is an **existing platform service** — do not rebuild it. This unit only invokes the existing auth endpoints to validate session state and redirect unauthenticated users.

## Scope

| Layer | Responsibility |
|---|---|
| Service API | Exposes wishlist read, add, and remove endpoints to the client |
| Client UI | Wishlist page (with Generate Combo CTA), add/remove heart actions on product cards and detail pages, login-gate modal |

---

## Platform APIs This Unit Calls

All platform requests require:
- Header: `Content-Language: en-SG` (or appropriate locale)
- Header: `Accept: application/json`
- Auth: Session cookie. Wishlist endpoints return `403` if the session is not authenticated.

### `GET /v1/wishLists/wishlist`

Fetches the authenticated shopper's wishlist.

Query params: `offset` (int, required), `limit` (int, required), `sort` (string, optional — `inStock`)

Response shape:
```json
{
  "id": "string",
  "name": "string",
  "totalCount": 12,
  "items": [
    {
      "itemId": "string",
      "simpleSku": "string",
      "configSku": "string",
      "inStock": true,
      "createdAt": 1700000000,
      "product": {
        "config_sku": "string",
        "name": "string",
        "brand": "string",
        "price": 99.90,
        "image": "url",
        "simples": [{ "sku": "string", "size": "M", "color": "black" }]
      }
    }
  ]
}
```

### `POST /v1/wishLists/items`

Add an item to the wishlist. Requires authentication.

Form data: `simpleSku` (string, required)

Response: `ZDTProduct.WishListItem` — the added item object.

### `DELETE /v1/wishLists/items/`

Remove items from the wishlist. Requires authentication.

Query params: `configSku` (string) — removes all variants of the product.

Response: `200 OK` with empty body.

---

## User Stories

### US-201 — Prompt Login When Adding to Wishlist

**As a** guest shopper,
**I want to** be prompted to log in or create an account when I attempt to add a product to my wishlist,
**So that** my wishlist can be saved and linked to my account.

**Acceptance Criteria:**

- Clicking "Add to Wishlist" while unauthenticated triggers a login/register modal or redirect.
- After successful authentication, the product is automatically added via `POST /api/v1/wishlist/items`.
- The shopper is returned to the product they were viewing after login.

---

### US-202 — Add Product to Wishlist

**As a** registered shopper,
**I want to** add products to my wishlist,
**So that** I can curate a collection of items to use as the basis for AI combo suggestions.

**API endpoint:** `POST /api/v1/wishlist/items`

**Acceptance Criteria:**

- A registered shopper can add any in-stock product to their wishlist.
- A visual indicator (e.g., filled heart icon) confirms the item has been added.
- Duplicate additions are prevented; re-clicking the heart toggles the item off the wishlist (calls `DELETE /api/v1/wishlist/items/{configSku}`).

---

### US-203 — View and Manage Wishlist

**As a** registered shopper,
**I want to** view my wishlist and remove items from it,
**So that** I can keep my curated list up to date before triggering AI suggestions.

**API endpoints:** `GET /api/v1/wishlist`, `DELETE /api/v1/wishlist/items/{configSku}`

**Acceptance Criteria:**

- The wishlist page shows all items with image, name, price, and in-stock status.
- Each item has a remove action.
- An empty wishlist displays a prompt encouraging the shopper to browse products.
- A "Generate Combo" / "Surprise me" entry point is visible on this page (navigates to the AI Styling Engine — Unit 3).

---

## Frontend Screens & Components

### Screen: Wishlist Page

The main screen for this unit. Displays all wishlisted items and serves as the entry point to the AI combo generation flow.

**Layout:**

- Header: "My Wishlist" title + item count (e.g., "12 items")
- **Generate Combo CTA** (prominent, always visible when wishlist has at least 1 item):
  - Primary button: **"Surprise Me"** — triggers combo generation immediately with no preferences (US-301). Navigates to the Combo Suggestion Page (Unit 3).
  - Secondary link/button: **"Generate with preferences →"** — navigates to the Preference Input Screen (Unit 3) before generating.
- Product grid/list: each item card shows:
  - Product image
  - Product name and brand
  - Price
  - In-stock / out-of-stock badge
  - Remove (trash/heart-toggle) icon — calls `DELETE /api/v1/wishlist/items/{configSku}`
- Empty state: illustration + message "Your wishlist is empty. Browse products to get started." with a Browse button.

**Behaviour:**

- If wishlist has 0 items, hide the Generate Combo CTA and show empty state.
- "Surprise Me" passes the shopper session to Unit 3's `POST /api/v1/style/combos/generate` (no preferences body) and navigates to the Combo Suggestion Page.

---

### Component: Add to Wishlist Heart Button

Used on every product card (Unit 1 listing page, Unit 1 detail page, Unit 3 combo suggestion cards).

**States:**

- **Unfilled heart**: item not in wishlist. On tap → if unauthenticated, show login modal. If authenticated, call `POST /api/v1/wishlist/items` and optimistically toggle to filled.
- **Filled heart**: item in wishlist. On tap → call `DELETE /api/v1/wishlist/items/{configSku}` and optimistically toggle to unfilled.

---

### Component: Login Gate Modal

Shown when a guest taps the heart button.

- Title: "Sign in to save items"
- Body: "Create an account or sign in to add items to your wishlist."
- Buttons: "Sign In", "Create Account", "Cancel"
- After successful auth, the pending wishlist add is automatically completed.

---

## Service API This Unit Exposes

### `GET /api/v1/wishlist`

Returns the authenticated shopper's wishlist, shaped for the client.

**Auth:** Required — `403` if unauthenticated

**Query Parameters:**

| Param | Type | Required | Notes |
|---|---|---|---|
| `offset` | integer | No | Default 0 |
| `limit` | integer | No | Default 20 |

**Response `200 OK`:**

```json
{
  "totalCount": 12,
  "items": [
    {
      "itemId": "string",
      "simpleSku": "string",
      "configSku": "string",
      "name": "string",
      "brand": "string",
      "price": 99.90,
      "imageUrl": "string",
      "color": "black",
      "size": "M",
      "inStock": true
    }
  ]
}
```

**Response `403 Forbidden`:** Shopper is not authenticated.

---

### `POST /api/v1/wishlist/items`

Add an item to the wishlist.

**Auth:** Required — `403` if unauthenticated

**Request Body:**

```json
{
  "simpleSku": "string"
}
```

**Response `201 Created`:**

```json
{
  "itemId": "string",
  "simpleSku": "string",
  "configSku": "string"
}
```

**Response `403 Forbidden`:** Shopper is not authenticated.

---

### `DELETE /api/v1/wishlist/items/{configSku}`

Remove all variants of a product from the wishlist.

**Auth:** Required — `403` if unauthenticated

**Path Parameters:**

| Param | Type | Required |
|---|---|---|
| `configSku` | string | Yes |

**Response `200 OK`:** Empty body.

**Response `403 Forbidden`:** Shopper is not authenticated.
