# Plan: User Stories for AI-Powered Product Matching Application

## Steps

- [x] **Step 1**: Clarify open questions with stakeholder (see Questions below)
- [x] **Step 2**: Create `/inception/` directory
- [x] **Step 3**: Write user stories for **Product Browsing & Wishlist** module
- [x] **Step 4**: Write user stories for **AI Preference Input** module
- [x] **Step 5**: Write user stories for **AI Combo Suggestion** module
- [x] **Step 6**: Write user stories for **Combo Confirmation & Cart/Checkout** module
- [x] **Step 7**: Review all user stories for completeness, acceptance criteria, and consistency
- [x] **Step 8**: Write final `user_stories.md` to `/inception/` directory

---

## Phase 2 — Architecture Units

- [x] **Step 9**: Clarify open questions for unit grouping (see Questions below)
- [x] **Step 10**: Create `/inception/units/` directory
- [x] **Step 11**: Group user stories into independent units and write individual unit `.md` files
- [x] **Step 12**: Write `integration_contract.md` in `/inception/units/`

---

## Open Questions (Phase 1 — User Stories)

**[Question 1]** What type of products does this application sell? (e.g., fashion/clothing, home decor, accessories, beauty products, or a mix?)
**[Answer 1]**
fashion/clothing, accessories, beauty products

**[Question 2]** Can customers browse and wishlist products without logging in, or is authentication required?
**[Answer 2]**
can browse without authentication but need to login to add to wishlist

**[Question 3]** Is the AI combo suggestion strictly limited to products already on the customer's wishlist, or can it recommend any product from the full database?
**[Answer 3]**
start with customer wishlist, for suggestion, it could be from any in the our database

**[Question 4]** When AI cannot find a suitable combo from the wishlist, should it suggest replacing wishlist items, adding new items to the wishlist first, or go directly to cart recommendation?
**[Answer 4]**
replacing or adding items to wishlist is sufficient

**[Question 5]** What does "checkout" flow look like — is it handled in-app (payment, shipping), or does it hand off to an existing e-commerce platform (e.g., Shopify, WooCommerce)?
**[Answer 5]**
in app

**[Question 6]** Are there any persona types to consider beyond a single "customer" role? (e.g., guest shopper, registered member, stylist, admin?)
**[Answer 6]**
just shopper

**[Question 7]** Should customers be able to save, name, or share AI-generated combos (e.g., "My summer look")?
**[Answer 7]**
of course

---

## Open Questions (Phase 2 — Unit Grouping)

**[Question A]** Is authentication (login/register) already built as an existing service, or does it need to be included as a unit in this build?
**[Answer A]**
Authentication already build, you can check in the API yaml file provided

**[Question B]** Do you have a preference for the number of teams? (e.g., 2–3 larger teams or 4–5 smaller, more focused ones?)
**[Answer B]**
i plan for an AI agent to implement each unit independently
**[Note]** Unit files have been made fully self-contained: each includes its full API spec, all consumed upstream contracts (platform + inter-unit), mock guidance for dependencies, and data models where applicable. An AI agent can implement any unit without reading other unit files.

**[Question C]** Should the AI Styling Engine be treated as one unit (preference input + combo generation together), or split into a separate frontend preferences unit and a backend AI inference unit?
**[Answer C]**
it should be an API processed by BE and send the result to client, we are built follow the backed for frontend

**[Question D]** The "Combo Portfolio" features (save, name, share — US-501 to US-503) and the "Add Combo to Cart" handoff (US-601) are closely related in the user journey. Should they be in the same unit, or kept separate?
**[Answer D]**
i think you can separate them
---

> Please fill in Answers A–D above and approve to proceed with Step 9 onwards.
