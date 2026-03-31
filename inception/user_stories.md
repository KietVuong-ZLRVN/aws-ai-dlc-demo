# User Stories: AI-Powered Style Combo Application

## Actors

- **Guest Shopper**: An unauthenticated visitor browsing the catalog.
- **Shopper**: An authenticated, registered user with a wishlist and account.

## Existing APIs (Available — No Build Required)

| Capability | API |
|---|---|
| Product search by keyword + filters | `GET /v1/products/list` |
| Product detail | `GET /v1/products/{configSkuOrSlug}/details` |
| Get wishlist | `GET /v1/wishLists/wishlist` |
| Add item to wishlist | `POST /v1/wishLists/items` (requires `simpleSku`, auth) |
| Remove item from wishlist | `DELETE /v1/wishLists/items/` (by `simpleSku` or `configSku`) |
| Add single item to cart | `POST /v1/checkout/cart` |
| Add multiple items to cart | `POST /v1/checkout/cart/bulk` |
| Cart review & full checkout | Handled in-app — out of scope |
| "Complete the Look" base recommendations | `GET /v1/recommendation/completethelook/{config_sku}` |

---

## Module 1: Product Browsing

### US-101 — Browse Catalog Without Login

**As a** guest shopper,
**I want to** browse products across fashion/clothing, accessories, and beauty categories without logging in,
**So that** I can freely explore the catalog before committing to an account.

**API:** `GET /v1/products/list`

**Acceptance Criteria:**

- All product listings are visible to unauthenticated users.
- No login prompt is shown during browsing.
- Browsing is available by category, search term, and filter.

---

### US-102 — Search and Filter Products

**As a** guest or registered shopper,
**I want to** search and filter products by keyword, category, color, price range, and occasion,
**So that** I can quickly narrow down products relevant to my needs.

**API:** `GET /v1/products/list` — using `query`, `colors[]`, `price`, `occasion`, `categoryId` params

**Acceptance Criteria:**

- Keyword search matches product name, description, and tags.
- Filters include: category (clothing, accessories, beauty), color, price range, and occasion (e.g., casual, formal, outdoor).
- Filters can be combined and cleared independently.
- Filter options (available values) are sourced from `GET /v1/products/filter`.

---

### US-103 — View Product Details

**As a** guest or registered shopper,
**I want to** view detailed information for a product, including images, description, price, available sizes, and color variants,
**So that** I can make an informed decision before adding it to my wishlist.

**API:** `GET /v1/products/{configSkuOrSlug}/details`

**Acceptance Criteria:**

- Product detail page shows: multiple images, full description, price, size options, and available colors.
- Out-of-stock variants are visually indicated and cannot be added to the wishlist.

---

## Module 2: Wishlist Management

### US-201 — Prompt Login When Adding to Wishlist

**As a** guest shopper,
**I want to** be prompted to log in or create an account when I attempt to add a product to my wishlist,
**So that** my wishlist can be saved and linked to my account.

**Acceptance Criteria:**

- Clicking "Add to Wishlist" while unauthenticated triggers a login/register modal or redirect.
- After successful authentication, the product is automatically added to the wishlist via `POST /v1/wishLists/items`.
- The shopper is returned to the product they were viewing after login.

---

### US-202 — Add Product to Wishlist

**As a** registered shopper,
**I want to** add products to my wishlist,
**So that** I can curate a collection of items to use as the basis for AI combo suggestions.

**API:** `POST /v1/wishLists/items` (body: `simpleSku`)

**Acceptance Criteria:**

- A registered shopper can add any in-stock product to their wishlist.
- A visual indicator (e.g., filled heart icon) confirms the item has been added.
- Duplicate additions are prevented; re-clicking toggles the item off the wishlist (calls `DELETE /v1/wishLists/items/`).

---

### US-203 — View and Manage Wishlist

**As a** registered shopper,
**I want to** view my wishlist and remove items from it,
**So that** I can keep my curated list up to date.

**API:** `GET /v1/wishLists/wishlist`, `DELETE /v1/wishLists/items/`

**Acceptance Criteria:**

- The wishlist page shows all items with image, name, and price; sourced from `GET /v1/wishLists/wishlist`.
- Each item has a remove action that calls `DELETE /v1/wishLists/items/` by `configSku`.
- An empty wishlist displays a message prompting the shopper to browse products.

---

## Module 3: AI Preference Input

### US-301 — Generate Combo Directly from Wishlist (No Preferences)

**As a** registered shopper,
**I want to** trigger AI combo generation from my wishlist without providing any preferences,
**So that** I can quickly get a styled suggestion when I just want to shop without overthinking it.

**Acceptance Criteria:**

- A "Surprise me" or "Generate combo" one-tap action is available on the wishlist page requiring no preference input.
- The AI generates the best possible combo from wishlist items using its own styling judgment.
- The shopper is not forced to fill in any preference fields to use this path.
- Both this quick-generate path and the preference-guided path (US-301b) are accessible from the wishlist.

---

### US-301b — Specify Style Preferences for Guided AI Suggestions

**As a** registered shopper,
**I want to** optionally specify my style preferences — including occasion, styling direction, budget, and color palette — before the AI generates suggestions,
**So that** I can get more targeted recommendations when I have a specific look in mind.

**Acceptance Criteria:**

- Preference input form includes: occasion (multi-select from predefined list), styling direction (e.g., minimalist, bold, classic), budget (min/max range), and preferred/excluded colors.
- A free-text field is available for additional details.
- All fields are optional; leaving all blank falls back to quick-generate behavior (US-301).
- The shopper can save preferences as a profile for reuse.

---

### US-302 — AI Confirms Understanding of Preferences

**As a** registered shopper,
**I want to** see the AI summarize its interpretation of my preferences before generating suggestions,
**So that** I can correct any misunderstanding before results are produced.

**Acceptance Criteria:**

- After preference submission, the AI displays a brief natural-language summary (e.g., "You're looking for a casual beach look under $200 in warm tones.").
- The shopper can confirm, edit, or cancel before the AI proceeds.

---

## Module 4: AI Combo Suggestion

### US-401 — Generate Combo from Wishlist

**As a** registered shopper,
**I want to** receive AI-generated outfit or product combo suggestions based on my wishlist items and stated preferences,
**So that** I can discover complete looks curated to my taste.

**Note:** The AI may leverage `GET /v1/recommendation/completethelook/{config_sku}` as a base signal, augmented with the shopper's wishlist and preferences.

**Acceptance Criteria:**

- The AI prioritizes items already on the shopper's wishlist when building a combo.
- Each combo contains at minimum a primary item and one complementary item.
- Suggestions respect the shopper's budget, occasion, and color preferences when provided.
- Multiple combo options (at least 2, up to 5) are presented when available.

---

### US-402 — AI Explains Combo Reasoning

**As a** registered shopper,
**I want to** see a short explanation for why the AI suggested a particular combo,
**So that** I can understand the styling logic and feel confident in the recommendation.

**Acceptance Criteria:**

- Each combo card includes a 1–3 sentence explanation (e.g., "This linen blazer complements the wide-leg trousers for a smart-casual summer office look.").
- Explanation references the shopper's stated preferences where applicable.

---

### US-403 — AI Supplements Wishlist with Catalog Items

**As a** registered shopper,
**I want the** AI to recommend products from the full catalog to complete a combo when my wishlist alone is insufficient,
**So that** I always receive a full, wearable suggestion even if my wishlist is sparse.

**API:** AI queries `GET /v1/products/list` with relevant filters to source supplementary items.

**Acceptance Criteria:**

- Items sourced from the catalog (not the wishlist) are clearly labeled (e.g., "Suggested for you").
- Catalog-sourced items still respect the shopper's preferences (budget, color, occasion).
- The shopper can add catalog-suggested items to their wishlist directly from the combo view via `POST /v1/wishLists/items`.

---

### US-404 — AI Suggests Wishlist Replacements or Additions When No Match Found

**As a** registered shopper,
**I want the** AI to suggest replacing or adding items to my wishlist when it cannot form a suitable combo from existing wishlist items,
**So that** I can still get a meaningful recommendation rather than a dead end.

**Acceptance Criteria:**

- When no combo can be formed, the AI presents an explanation and specific alternative items (e.g., "Your wishlist items don't match a formal occasion. Here are substitutions:").
- Each suggested alternative has an "Add to Wishlist" action (`POST /v1/wishLists/items`).
- After updating the wishlist, the shopper can re-trigger combo generation.

---

### US-405 — Reject Suggestion and Refine

**As a** registered shopper,
**I want to** reject a combo suggestion and ask the AI to try again with adjusted preferences,
**So that** I can iterate until I find a combo I like.

**Acceptance Criteria:**

- Each combo has a "Not for me" or "Try again" action.
- The shopper can adjust any preference field before the AI regenerates.
- Previously shown combos are not repeated in the same session unless no new options exist.

---

## Module 5: Combo Save, Name, and Share

### US-501 — Save and Name a Combo

**As a** registered shopper,
**I want to** save an AI-generated combo and give it a custom name (e.g., "My Summer Look"),
**So that** I can revisit and act on it later.

**Acceptance Criteria:**

- A "Save Combo" button is available on each combo suggestion card.
- The shopper is prompted to enter a name; a default name is pre-filled (e.g., "Combo – March 31, 2026").
- Saved combos are accessible from a dedicated "My Combos" section in the shopper's profile.

---

### US-502 — View Saved Combos

**As a** registered shopper,
**I want to** view all my saved combos,
**So that** I can revisit previous AI recommendations and act on them at any time.

**Acceptance Criteria:**

- "My Combos" page lists all saved combos with name, item thumbnails, and save date.
- Each saved combo can be expanded to view full item details.
- The shopper can rename or delete a saved combo.

---

### US-503 — Share a Combo

**As a** registered shopper,
**I want to** share a saved combo via a shareable link or social media,
**So that** I can get feedback or inspire others with my style choices.

**Acceptance Criteria:**

- Each saved combo has a "Share" action that generates a unique, publicly accessible link.
- The shared link displays the combo items and name without requiring the viewer to log in.
- Social sharing shortcuts are provided (e.g., copy link, share to Instagram, share to WhatsApp).
- The shopper can set a combo as "private" to disable public link sharing.

---

## Module 6: Add Combo to Cart

> Cart review, checkout flow, payment, and order confirmation are fully handled by the existing in-app checkout. This module covers only the handoff from a confirmed combo to the cart.

### US-601 — Add Confirmed Combo to Cart

**As a** registered shopper,
**I want to** add all items in a confirmed combo to my cart in a single action,
**So that** I can quickly proceed to the existing checkout without adding each item individually.

**API:** `POST /v1/checkout/cart/bulk` (body: JSON array of `{simpleSku, quantity, size}` for each combo item)

**Acceptance Criteria:**

- An "Add All to Cart" button is present on each combo card.
- All in-stock combo items are added to the cart in one call to `POST /v1/checkout/cart/bulk`.
- If any item in the combo is out of stock, the shopper is notified and can choose to add the remaining items or cancel.
- The shopper can also add individual items from a combo to the cart one at a time.
- After a successful bulk add, the shopper is navigated to the existing cart/checkout flow.
