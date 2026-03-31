# Unit 3: AI Styling Engine

## Purpose

The core intelligence unit. Accepts the shopper's wishlist and optional style preferences, orchestrates AI inference server-side, and returns ranked combo suggestions with explanations. The client sends inputs and renders results — all AI logic runs in the service layer following the Backend for Frontend philosophy.

## Implementation Note

This unit is implemented independently. All information needed to build it is contained in this file. It calls two external dependencies at runtime:

1. **Unit 2 (Wishlist Service)** — to read the shopper's wishlist server-side. The contract for this call is specified below under "Consumed Unit APIs". During development, mock this with a fixed wishlist response.
2. **Platform Product API** — to search for supplementary items. Contract specified below.

This unit has no persistent storage of its own.

## Scope

| Layer | Responsibility |
|---|---|
| Service API | Preference options, preference confirmation, combo generation with fallback |
| AI/ML | Orchestrates LLM or styling model: interprets preferences, scores item compatibility, generates combo explanations |
| Client UI | Preference input screen, AI summary confirmation screen, combo suggestion page, fallback/replacement screen |

---

## Consumed Unit APIs (Mock These During Development)

### Unit 2 — `GET /api/v1/wishlist`

Called server-side to fetch the shopper's wishlist before running AI inference.

**Auth:** Forwarded session cookie from the incoming client request.

**Response to expect:**

```json
{
  "totalCount": 5,
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

---

## Platform APIs This Unit Calls

All platform requests require:
- Header: `Content-Language: en-SG`
- Header: `Accept: application/json`

### `GET /v1/products/list`

Used to source supplementary catalog items when the wishlist alone cannot form a complete combo. Apply AI-derived filters (color, occasion, price range) to narrow results.

Key query params: `query`, `colors[]`, `occasion`, `price`, `categoryId`, `offset`, `limit`

Response shape (relevant fields):
```json
{
  "numProductFound": 40,
  "products": [
    {
      "config_sku": "string",
      "name": "string",
      "brand": "string",
      "price": 59.90,
      "image": "url",
      "attributes": { "color": ["beige"], "occasion": ["casual"] }
    }
  ]
}
```

### `GET /v1/recommendation/completethelook/{config_sku}`

Returns platform-computed "complete the look" recommendations for a given product. Use as a base styling signal, augmented by the AI model.

Response shape (relevant fields):
```json
{
  "products": [
    { "config_sku": "string", "name": "string", "price": 49.90, "image": "url" }
  ]
}
```

---

## User Stories

### US-301 — Generate Combo Directly from Wishlist (No Preferences)

**As a** registered shopper,
**I want to** trigger AI combo generation from my wishlist without providing any preferences,
**So that** I can quickly get a styled suggestion when I just want to shop without overthinking it.

**API endpoint:** `POST /api/v1/style/combos/generate` (omit `preferences` field)

**Acceptance Criteria:**

- A "Surprise me" or "Generate combo" one-tap action on the wishlist page requires no preference input.
- The AI generates the best possible combo from wishlist items using its own styling judgment.
- The shopper is not forced to fill in any preference fields.
- Both quick-generate (this story) and preference-guided (US-301b) entry points are accessible from the wishlist page.

---

### US-301b — Specify Style Preferences for Guided AI Suggestions

**As a** registered shopper,
**I want to** optionally specify my style preferences — occasion, styling direction, budget, and color palette — before the AI generates suggestions,
**So that** I can get more targeted recommendations when I have a specific look in mind.

**API endpoints:** `GET /api/v1/style/preferences/options`, then `POST /api/v1/style/combos/generate`

**Acceptance Criteria:**

- Preference form includes: occasion (multi-select), styling direction (predefined: minimalist, bold, classic, bohemian), budget (min/max range), preferred/excluded colors.
- A free-text field is available for additional details.
- All fields are optional; submitting empty falls back to quick-generate behavior (US-301).
- The shopper can save their preferences for reuse.

---

### US-302 — AI Confirms Understanding of Preferences

**As a** registered shopper,
**I want to** see the AI summarize its interpretation of my preferences before generating suggestions,
**So that** I can correct any misunderstanding before results are produced.

**API endpoint:** `POST /api/v1/style/preferences/confirm`

**Acceptance Criteria:**

- After submitting preferences, the service returns a natural-language summary (e.g., "You're looking for a casual beach look under $200 in warm tones.").
- The shopper can confirm, edit, or cancel before proceeding to combo generation.

---

### US-401 — Generate Combo from Wishlist

**As a** registered shopper,
**I want to** receive AI-generated outfit or product combo suggestions based on my wishlist items and stated preferences,
**So that** I can discover complete looks curated to my taste.

**API endpoint:** `POST /api/v1/style/combos/generate`

**Acceptance Criteria:**

- The AI prioritizes items already on the shopper's wishlist when building a combo.
- Each combo contains at minimum a primary item and one complementary item.
- Suggestions respect the shopper's budget, occasion, and color preferences when provided.
- At least 2, up to 5 combo options are presented when available.

---

### US-402 — AI Explains Combo Reasoning

**As a** registered shopper,
**I want to** see a short explanation for why the AI suggested a particular combo,
**So that** I can understand the styling logic and feel confident in the recommendation.

**Acceptance Criteria:**

- Each combo in the response includes a `reasoning` field with a 1–3 sentence explanation.
- Explanation references the shopper's stated preferences where applicable.

---

### US-403 — AI Supplements Wishlist with Catalog Items

**As a** registered shopper,
**I want the** AI to recommend products from the full catalog to complete a combo when my wishlist alone is insufficient,
**So that** I always receive a full, wearable suggestion even if my wishlist is sparse.

**Acceptance Criteria:**

- The service queries `GET /v1/products/list` with AI-derived filters to source supplementary items.
- Items sourced from the catalog are flagged with `"source": "catalog"` in the response; wishlist items with `"source": "wishlist"`.
- The client renders catalog-sourced items with a "Suggested for you" label.
- The shopper can add catalog-suggested items to their wishlist directly from the combo view (client calls Unit 2's `POST /api/v1/wishlist/items`).

---

### US-404 — AI Suggests Replacements When No Combo Can Be Formed

**As a** registered shopper,
**I want the** AI to suggest replacing or adding items to my wishlist when it cannot form a suitable combo,
**So that** I can still get a meaningful recommendation rather than a dead end.

**Acceptance Criteria:**

- When no combo can be formed, the response includes `"status": "fallback"` with an explanation and a list of alternative items.
- Each alternative includes a reason (e.g., "Replaces your blazer with a formal-friendly option").
- The client shows an "Add to Wishlist" action per alternative (calls Unit 2's `POST /api/v1/wishlist/items`).
- After updating the wishlist, the shopper can re-trigger combo generation.

---

### US-405 — Reject Suggestion and Refine

**As a** registered shopper,
**I want to** reject a combo suggestion and ask the AI to try again with adjusted preferences,
**So that** I can iterate until I find a combo I like.

**Acceptance Criteria:**

- Each combo has a "Not for me" / "Try again" action.
- The shopper can adjust any preference before regenerating.
- The client passes previously seen combo IDs in `excludeComboIds` so the service does not return them again.

---

## Frontend Screens & Components

### Screen: Preference Input Screen (optional path)

Reached when the shopper taps "Generate with preferences →" on the Wishlist Page (Unit 2). Skip this screen entirely for the "Surprise Me" quick-generate path.

**Layout:**

- Header: "Customise your combo" with a back button.
- Form sections (all optional):
  - **Occasion** — multi-select chips (values from `GET /api/v1/style/preferences/options`): e.g., Casual, Formal, Beach, Office, Party, Outdoor.
  - **Style direction** — single-select chips: e.g., Minimalist, Bold, Classic, Bohemian.
  - **Budget** — range slider with min/max input fields (e.g., $0–$500).
  - **Colours** — two colour picker groups: "I love" (preferred) and "Avoid" (excluded).
  - **Tell us more** — free-text field, placeholder: "e.g., something light for a summer trip".
- Footer: **"Generate Combo"** primary button (always enabled, even with no input). Tapping it calls `POST /api/v1/style/preferences/confirm` then navigates to the AI Summary Confirmation Screen.

---

### Screen: AI Summary Confirmation Screen

Shown after the preference form is submitted. Displays the AI's natural-language interpretation before generation begins.

**Layout:**

- Header: "Does this sound right?" with a back button.
- AI summary card: plain text, e.g., *"You're looking for a minimalist, casual beach look between $50–$200 in warm tones — perfect for a summer trip."*
- Buttons:
  - **"Looks good, generate!"** — calls `POST /api/v1/style/combos/generate` with confirmed preferences, navigates to the Combo Suggestion Page.
  - **"Edit preferences"** — navigates back to the Preference Input Screen with fields pre-filled.
  - **"Cancel"** — returns to the Wishlist Page.
- Skip this screen entirely for the quick-generate path (no preferences submitted).

---

### Screen: Combo Suggestion Page

The main output screen. Displays the AI-generated combo options.

**Layout:**

- Header: "Your combos" + a loading skeleton while `POST /api/v1/style/combos/generate` is in progress.
- **Combo cards** — vertically stacked, one per combo (2–5 cards):
  - Horizontal scroll row of product images (one per item in the combo).
  - Each product image has:
    - Product name and price below it.
    - **"Suggested for you"** label if `source: "catalog"`.
    - Heart (wishlist) icon — tapping it calls Unit 2's `POST /api/v1/wishlist/items` to add a catalog-sourced item to the wishlist.
  - **Reasoning text** — 1–3 sentences explaining the combo (from `reasoning` field in response).
  - **"Save Combo"** button — opens the Save Combo Modal (see Unit 4).
  - **"Add All to Cart"** button — calls Unit 5's `POST /api/v1/cart/combo` with the inline items array. See Unit 5 for out-of-stock modal behaviour.
  - **"Not for me"** link/button — marks this combo as rejected; client adds its `id` to `excludeComboIds` for the next generate call.
- **"Generate new combos"** button at the bottom — calls `POST /api/v1/style/combos/generate` again with `excludeComboIds` containing all previously shown combo IDs, and refreshes the page.
- If all possible combos are exhausted, replace the button with a message: "We've shown you all available combos. Try adjusting your preferences."

---

### Screen: Fallback Screen

Shown when `POST /api/v1/style/combos/generate` returns `"status": "fallback"`.

**Layout:**

- Header: "Let's adjust your wishlist"
- Explanation message (from `message` field in response), e.g., *"Your wishlist items don't match a formal occasion. Here are some suggestions:"*
- **Alternative item cards** — one per item in `alternatives` array:
  - Product image, name, price.
  - Replacement reason text (from `reason` field).
  - **"Add to Wishlist"** button — calls Unit 2's `POST /api/v1/wishlist/items`.
- **"Try generating again"** button — re-calls `POST /api/v1/style/combos/generate` after the shopper has added alternatives. Enabled once at least one alternative has been added to the wishlist.
- **"Edit preferences"** link — navigates back to the Preference Input Screen.

---

## Service API This Unit Exposes

### `GET /api/v1/style/preferences/options`

Returns predefined options to populate the preference input form.

**Auth:** Required

**Response `200 OK`:**

```json
{
  "occasions": ["casual", "formal", "outdoor", "beach", "office", "party"],
  "styles": ["minimalist", "bold", "classic", "bohemian"],
  "colors": ["black", "white", "navy", "beige", "red", "green", "pastel"]
}
```

---

### `POST /api/v1/style/preferences/confirm`

Returns a natural-language summary of the submitted preferences for the shopper to review.

**Auth:** Required

**Request Body:**

```json
{
  "occasions": ["casual", "beach"],
  "styles": ["minimalist"],
  "budget": { "min": 50, "max": 200 },
  "colors": { "preferred": ["beige", "white"], "excluded": ["black"] },
  "freeText": "Something light and airy for a summer trip"
}
```

**Response `200 OK`:**

```json
{
  "summary": "You're looking for a minimalist, casual beach look between $50–$200 in light, airy tones — perfect for a summer trip.",
  "preferences": {
    "occasions": ["casual", "beach"],
    "styles": ["minimalist"],
    "budget": { "min": 50, "max": 200 },
    "colors": { "preferred": ["beige", "white"], "excluded": ["black"] },
    "freeText": "Something light and airy for a summer trip"
  }
}
```

---

### `POST /api/v1/style/combos/generate`

Core AI endpoint. Fetches the shopper's wishlist server-side, runs AI inference, and returns ranked combos.

**Auth:** Required

**Request Body:**

```json
{
  "preferences": {
    "occasions": ["casual"],
    "styles": ["minimalist"],
    "budget": { "min": 0, "max": 200 },
    "colors": { "preferred": ["beige"], "excluded": [] },
    "freeText": "optional"
  },
  "excludeComboIds": ["combo-id-1", "combo-id-2"]
}
```

- `preferences` is fully optional. Omit it entirely for quick-generate (US-301).
- `excludeComboIds` is optional. Pass IDs of previously shown combos to prevent repeats (US-405).
- Wishlist is fetched server-side using the shopper's session — the client does not send it.

**Response `200 OK` — success:**

```json
{
  "status": "ok",
  "combos": [
    {
      "id": "combo-id-3",
      "reasoning": "This linen blazer pairs well with the wide-leg trousers for a smart-casual summer look.",
      "items": [
        {
          "configSku": "string",
          "simpleSku": "string",
          "name": "string",
          "brand": "string",
          "price": 89.90,
          "imageUrl": "string",
          "source": "wishlist"
        },
        {
          "configSku": "string",
          "simpleSku": "string",
          "name": "string",
          "brand": "string",
          "price": 59.90,
          "imageUrl": "string",
          "source": "catalog"
        }
      ]
    }
  ]
}
```

- `source`: `"wishlist"` — from the shopper's wishlist; `"catalog"` — sourced from the product catalog to complete the combo.

**Response `200 OK` — fallback (no suitable combo found):**

```json
{
  "status": "fallback",
  "message": "Your wishlist items don't match a formal occasion. Here are some suggestions:",
  "alternatives": [
    {
      "configSku": "string",
      "simpleSku": "string",
      "name": "string",
      "brand": "string",
      "price": 79.90,
      "imageUrl": "string",
      "reason": "Replaces your current blazer with a formal-friendly option"
    }
  ]
}
```

**Response `403 Forbidden`:** Shopper is not authenticated.
