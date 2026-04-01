# DDD Domain Model Design Plan
## Focus: Unit 4 (Combo Portfolio) & Unit 5 (Cart Handoff)

---

## Clarifying Questions

> Please fill in the `[Answer]` tags before approving the plan. These decisions directly affect the domain model design.

---

**[Question]** For Unit 4 — Do share tokens expire? If yes, what is the expiry policy (e.g., 30 days, never, revoked only on manual visibility change to private)?
for simplicity, just need to revoked only on manual visibility change to private
**[Answer]**

---

**[Question]** For Unit 4 — When a shopper views a saved combo, should item data (price, availability) reflect the **snapshot at save time** (current data model) or should it be **refreshed from the product catalog**? This determines whether `ComboItem` is a pure snapshot Value Object or needs enrichment.

**[Answer]**
refreshed from the product catalog
---

**[Question]** For Unit 4 — Is there a business rule on the **maximum number of items** per combo (e.g., must a combo have at least 2 items, or at most N items)?

**[Answer]**
most 10 items, least 2 items for combo
---

**[Question]** For Unit 5 — The unit description states it "owns no storage." Should the cart handoff service record any **audit log or history** of handoff attempts (e.g., for debugging, analytics, or retry purposes), or is it truly stateless with no persistence at all?

**[Answer]**
the cart handoff should record the history add to cart for audit
---

**[Question]** For Unit 5 — When items are out-of-stock, the user can choose to "add available items anyway." Should this partial-add decision be **recorded as a domain event** (e.g., for analytics — how often users proceed with partial combos vs cancel), or is no tracking needed?

**[Answer]**
dont need to track this one at the moment
---

## Plan Steps

- [x] **Step 1** — Review Unit 4 (Combo Portfolio) user stories, acceptance criteria, and data model in detail
- [x] **Step 2** — Review Unit 5 (Cart Handoff) user stories, acceptance criteria, and service contract in detail
- [x] **Step 3** — Review the integration contract to map inter-unit dependencies and ACL boundaries
- [x] **Step 4** — Design DDD domain model for **Unit 4: Combo Portfolio**
  - [x] Define the Bounded Context and Ubiquitous Language glossary
  - [x] Identify Aggregates and Aggregate Roots
  - [x] Identify Entities within each Aggregate
  - [x] Identify Value Objects
  - [x] Identify Domain Events
  - [x] Identify Policies (business rules that react to domain events)
  - [x] Identify Repositories
  - [x] Identify Domain Services
- [x] **Step 5** — Design DDD domain model for **Unit 5: Cart Handoff**
  - [x] Define the Bounded Context and Ubiquitous Language glossary
  - [x] Identify Aggregates (or confirm stateless nature and model accordingly)
  - [x] Identify Value Objects
  - [x] Identify Domain Events
  - [x] Identify Policies
  - [x] Identify Domain Services
  - [x] Identify Anti-Corruption Layer (ACL) adapters for Unit 4 and Platform Cart API
- [x] **Step 6** — Create `/construction/combo_portfolio/domain_model.md` with full DDD design for Unit 4
- [x] **Step 7** — Create `/construction/cart_handoff/domain_model.md` with full DDD design for Unit 5

---

> **Awaiting your review and approval before execution.**
> Please fill in the [Answer] tags above and confirm the plan is correct. Once approved, I will execute one step at a time and mark each checkbox as done upon completion.
