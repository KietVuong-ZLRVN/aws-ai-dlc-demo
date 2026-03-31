# Unit 4: Combo Portfolio

## Purpose

Allows shoppers to save, name, manage, and share AI-generated combos. This unit owns its own data store and is the only unit with a public-facing unauthenticated read endpoint (for shared combo links).

## Implementation Note

This unit is implemented independently. All information needed to build it is contained in this file. It has **no runtime dependencies on other units** — combo data is sent to it by the client at save time (the client passes the item list received from Unit 3). This unit does not call Unit 3 or any platform API.

## Scope

| Layer | Responsibility |
|---|---|
| Service API | CRUD for saved combos, share link generation, public share view |
| Storage | Combo records: item list, name, shopper ID, visibility, share token, timestamps |
| Client UI | Save/name modal on combo cards, "My Combos" profile section, combo detail view, share sheet |

## Data Model

```json
Combo {
  id: UUID,
  shopperId: string,
  name: string,
  items: [
    {
      configSku: string,
      simpleSku: string,
      name: string,
      imageUrl: string,
      price: number
    }
  ],
  visibility: "public" | "private",
  shareToken: string | null,
  createdAt: ISO8601,
  updatedAt: ISO8601
}
```

---

## User Stories

### US-501 — Save and Name a Combo

**As a** registered shopper,
**I want to** save an AI-generated combo and give it a custom name (e.g., "My Summer Look"),
**So that** I can revisit and act on it later.

**API endpoint:** `POST /api/v1/combos`

**Acceptance Criteria:**

- A "Save Combo" button is available on each combo suggestion card (in Unit 3's UI).
- Clicking it opens a name modal; a default name is pre-filled (e.g., "Combo – March 31, 2026").
- On confirmation, the combo is persisted via `POST /api/v1/combos`.
- Saved combos appear in "My Combos" in the shopper's profile.

---

### US-502 — View Saved Combos

**As a** registered shopper,
**I want to** view all my saved combos,
**So that** I can revisit previous AI recommendations and act on them at any time.

**API endpoints:** `GET /api/v1/combos`, `GET /api/v1/combos/{id}`, `PUT /api/v1/combos/{id}`, `DELETE /api/v1/combos/{id}`

**Acceptance Criteria:**

- "My Combos" page lists all saved combos with name, item thumbnails, and save date.
- Each saved combo can be expanded to view full item details.
- The shopper can rename (`PUT /api/v1/combos/{id}`) or delete (`DELETE /api/v1/combos/{id}`) a combo.

---

### US-503 — Share a Combo

**As a** registered shopper,
**I want to** share a saved combo via a shareable link or social media,
**So that** I can get feedback or inspire others with my style choices.

**API endpoints:** `POST /api/v1/combos/{id}/share`, `GET /api/v1/combos/shared/{token}`

**Acceptance Criteria:**

- Each saved combo has a "Share" action that calls `POST /api/v1/combos/{id}/share` to generate a unique share token.
- The public share URL renders combo items and name without requiring the viewer to log in.
- Social sharing shortcuts are provided (copy link, share to Instagram, share to WhatsApp).
- The shopper can set a combo to private via `PUT /api/v1/combos/{id}` (`visibility: "private"`), which disables the share link.

---

## Frontend Screens & Components

### Component: Save Combo Modal

Triggered from the "Save Combo" button on a combo card (Unit 3 Combo Suggestion Page). This is a modal overlay, not a full screen.

**Layout:**

- Title: "Save this combo"
- Text input: "Combo name" — pre-filled with today's date, e.g., "Combo – March 31, 2026". Editable.
- Buttons:
  - **"Save"** — calls `POST /api/v1/combos` with the combo's item list, the entered name, and `visibility: "private"`. On success, shows a toast "Combo saved!" and closes the modal.
  - **"Cancel"** — closes the modal without saving.

---

### Screen: My Combos Page

Accessible from the shopper's profile. Lists all saved combos.

**Layout:**

- Header: "My Combos" + item count.
- Grid or list of **Combo Summary Cards** (one per saved combo):
  - Row of up to 3 item thumbnail images.
  - Combo name (editable inline on long-press or tap-to-edit).
  - Save date (e.g., "Saved Mar 31, 2026").
  - Three-dot menu with: **"Rename"**, **"Share"**, **"Delete"**.
  - Tapping the card navigates to the Combo Detail Page.
- Empty state: illustration + "No combos saved yet. Generate your first combo from your wishlist." with a "Go to Wishlist" button.
- Pull-to-refresh calls `GET /api/v1/combos`.

---

### Screen: Combo Detail Page

Full view of a single saved combo. Reached by tapping a Combo Summary Card on the My Combos Page.

**Layout:**

- Header: combo name + back button + three-dot menu (Rename, Share, Delete).
- **Item list** — one item per row:
  - Product image (large), name, brand, price.
  - Heart (wishlist) icon — allows the viewer to wishlist individual items.
- **"Add All to Cart"** button (full-width, prominent at the bottom) — calls Unit 5's `POST /api/v1/cart/combo` with `comboId`. See Unit 5 for out-of-stock modal behaviour.
- **"Share"** button — calls `POST /api/v1/combos/{id}/share`, then shows the Share Sheet.
- Visibility toggle: "Public" / "Private" — calls `PUT /api/v1/combos/{id}` on change.

---

### Component: Share Sheet

Shown after tapping "Share" on a saved combo.

**Layout:**

- Displays the generated share URL.
- Action buttons: **"Copy link"**, **"Share to Instagram"**, **"Share to WhatsApp"**.
- Note: "Anyone with this link can view your combo." (only shown when visibility is public).

---

### Component: Rename Modal

Shown when tapping "Rename" from the three-dot menu.

- Text input pre-filled with the current combo name.
- **"Save"** button — calls `PUT /api/v1/combos/{id}` with the new name.
- **"Cancel"** button.

---

## Service API This Unit Exposes

### `POST /api/v1/combos`

Save a new combo.

**Auth:** Required

**Request Body:**

```json
{
  "name": "My Summer Look",
  "items": [
    {
      "configSku": "string",
      "simpleSku": "string",
      "name": "string",
      "imageUrl": "string",
      "price": 89.90
    }
  ],
  "visibility": "private"
}
```

- `visibility`: `"public"` or `"private"` — defaults to `"private"`.

**Response `201 Created`:**

```json
{
  "id": "combo-uuid",
  "name": "My Summer Look",
  "createdAt": "2026-03-31T10:00:00Z",
  "visibility": "private",
  "shareToken": null,
  "items": [
    { "configSku": "string", "simpleSku": "string", "name": "string", "imageUrl": "string", "price": 89.90 }
  ]
}
```

---

### `GET /api/v1/combos`

List all saved combos for the authenticated shopper.

**Auth:** Required

**Response `200 OK`:**

```json
{
  "combos": [
    {
      "id": "combo-uuid",
      "name": "My Summer Look",
      "createdAt": "2026-03-31T10:00:00Z",
      "visibility": "private",
      "itemThumbnails": ["url1", "url2", "url3"]
    }
  ]
}
```

---

### `GET /api/v1/combos/{id}`

Get full detail of a saved combo.

**Auth:** Required (shopper must own the combo)

**Response `200 OK`:**

```json
{
  "id": "combo-uuid",
  "name": "My Summer Look",
  "createdAt": "2026-03-31T10:00:00Z",
  "visibility": "private",
  "shareToken": null,
  "items": [
    { "configSku": "string", "simpleSku": "string", "name": "string", "imageUrl": "string", "price": 89.90 }
  ]
}
```

**Response `403 Forbidden`:** Shopper does not own this combo.
**Response `404 Not Found`:** Combo does not exist.

---

### `PUT /api/v1/combos/{id}`

Rename a combo or change its visibility.

**Auth:** Required (shopper must own the combo)

**Request Body** (all fields optional):

```json
{
  "name": "My Updated Look",
  "visibility": "public"
}
```

**Response `200 OK`:** Updated combo object (same shape as GET response).

---

### `DELETE /api/v1/combos/{id}`

Delete a saved combo.

**Auth:** Required (shopper must own the combo)

**Response `204 No Content`**

---

### `POST /api/v1/combos/{id}/share`

Generate or retrieve a public share token for a combo.

**Auth:** Required (shopper must own the combo)

**Response `200 OK`:**

```json
{
  "shareToken": "abc123xyz",
  "shareUrl": "https://app.example.com/combos/shared/abc123xyz"
}
```

- Sets the combo's `visibility` to `"public"` if it was `"private"`.

---

### `GET /api/v1/combos/shared/{token}`

View a shared combo. Public — no authentication required.

**Response `200 OK`:**

```json
{
  "name": "My Summer Look",
  "items": [
    { "configSku": "string", "name": "string", "imageUrl": "string", "price": 89.90 }
  ]
}
```

**Response `404 Not Found`:** Token does not exist or the combo visibility is `"private"`.
