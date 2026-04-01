# Implementation Plan
## Focus: Unit 4 (Combo Portfolio) & Unit 5 (Cart Handoff)

**Stack:** Go + go-chi · Aurora MySQL (new dedicated cluster) · Two standalone ECS services · In-process domain events only · External HTTP ACLs for Unit 1 (enrichment) and Platform Cart API

---

## Clarifying Questions

**[Question]** For the demo script — external dependencies (Unit 1 Product API, Platform Cart API, Unit 4 Combo API) cannot be called from a local machine without real credentials and network access. Should the demo script use **hardcoded stub/mock implementations** of those ACL adapters so the full flow can be run locally with `go run`, or do you want the demo to hit real staging/dev endpoints?

**[Answer]**
can mock the API base on the swagger at the link https://docs.zalora.io/s/doraemon/index.html#/
---

**[Question]** For the Go module name — what should the Go module path be? (e.g., `github.com/KietVuong-ZLRVN/combo-portfolio` for Unit 4 and `github.com/KietVuong-ZLRVN/cart-handoff` for Unit 5, or an internal company path like `internal.company.com/combo-portfolio`?)

**[Answer]**
use the repo name
---

**[Question]** For the Aurora MySQL connection in the demo — should the demo script connect to a real local MySQL instance (e.g., running in Docker), or should it use an **in-memory repository stub** so no database setup is needed to run the demo?

**[Answer]**
can do it in local Docker for demo
---

## Plan Steps

### Phase 0 — Pre-implementation (logical design reference docs)
- [ ] **Step 0** — Write `/construction/combo_portfolio/logical_design.md` (derived from domain model + all approved decisions in the design plan)
- [ ] **Step 0b** — Write `/construction/cart_handoff/logical_design.md`

---

### Phase 1 — Unit 4: Combo Portfolio Service

- [ ] **Step 1** — Scaffold Go module and directory structure under `/construction/combo_portfolio/src/`
- [ ] **Step 2** — Implement domain layer
  - [ ] `domain/value_objects.go` — `ComboId`, `ShopperId`, `ComboName`, `ComboItem`, `ShareToken`, `Visibility`
  - [ ] `domain/events.go` — `ComboCreated`, `ComboRenamed`, `ComboDeleted`, `ComboShared`, `ComboMadePrivate`
  - [ ] `domain/combo.go` — `Combo` aggregate root with all command handlers and invariant enforcement
  - [ ] `domain/repository.go` — `ComboRepository` interface
- [ ] **Step 3** — Implement infrastructure layer
  - [ ] `infrastructure/persistence/mysql_combo_repository.go` — MySQL implementation of `ComboRepository`
  - [ ] `infrastructure/persistence/schema.sql` — `combos` and `combo_items` table definitions
  - [ ] `infrastructure/acl/product_catalog_acl.go` — HTTP client calling `GET /api/v1/products/{configSku}` on Unit 1
  - [ ] `infrastructure/services/combo_enrichment_service.go` — calls ACL, builds `EnrichedCombo` read model, fallback to snapshot
  - [ ] `infrastructure/services/share_token_service.go` — UUID generation with uniqueness check
- [ ] **Step 4** — Implement application layer
  - [ ] `application/save_combo_handler.go` — validates command, constructs aggregate, persists, emits event
  - [ ] `application/rename_combo_handler.go`
  - [ ] `application/delete_combo_handler.go`
  - [ ] `application/share_combo_handler.go`
  - [ ] `application/make_private_handler.go`
  - [ ] `application/get_combo_handler.go` — loads aggregate, calls enrichment, returns `EnrichedCombo`
  - [ ] `application/list_combos_handler.go` — loads all for shopper, enriches each
  - [ ] `application/get_shared_combo_handler.go` — unauthenticated path via share token
- [ ] **Step 5** — Implement API layer
  - [ ] `api/dto.go` — request/response structs matching integration contract JSON shapes
  - [ ] `api/middleware.go` — session auth middleware (validates session cookie, injects `ShopperId` into context)
  - [ ] `api/handlers.go` — HTTP handlers delegating to application layer
  - [ ] `api/router.go` — go-chi router wiring all routes and middleware
- [ ] **Step 6** — Implement entrypoint
  - [ ] `cmd/main.go` — wires dependencies (DB pool, ACL clients, services, handlers), starts HTTP server
- [ ] **Step 7** — Write Unit 4 demo script (`demo/main.go`) — runs a full scenario: save → list → share → make private → get by share token → delete, using stub ACL and in-memory or local DB

---

### Phase 2 — Unit 5: Cart Handoff Service

- [ ] **Step 8** — Scaffold Go module and directory structure under `/construction/cart_handoff/src/`
- [ ] **Step 9** — Implement domain layer
  - [ ] `domain/value_objects.go` — `CartHandoffRecordId`, `ShopperId`, `HandoffSource`, `CartItem`, `SkippedItem`, `HandoffStatus`, `HandoffTimestamp`
  - [ ] `domain/events.go` — `CartHandoffRecorded`, `CartHandoffFailed`
  - [ ] `domain/cart_handoff_record.go` — `CartHandoffRecord` aggregate root, invariant enforcement
  - [ ] `domain/repository.go` — `CartHandoffRecordRepository` interface
- [ ] **Step 10** — Implement infrastructure layer
  - [ ] `infrastructure/persistence/mysql_handoff_repository.go` — MySQL implementation
  - [ ] `infrastructure/persistence/schema.sql` — `cart_handoff_records` and `handoff_items` table definitions
  - [ ] `infrastructure/acl/combo_portfolio_acl.go` — HTTP client calling `GET /api/v1/combos/{id}` on Unit 4; raises `ComboNotFound`, `ComboAccessDenied`, `ComboPortfolioUnavailable`
  - [ ] `infrastructure/acl/platform_cart_acl.go` — HTTP client calling `POST /v1/checkout/cart/bulk`; handles out-of-stock item classification; raises `PlatformCartUnavailable`
  - [ ] `infrastructure/services/combo_resolution_service.go` — branches on `HandoffSource` type
  - [ ] `infrastructure/services/cart_submission_service.go` — delegates to `PlatformCartACL`, classifies response
- [ ] **Step 11** — Implement application layer
  - [ ] `application/add_combo_to_cart_handler.go` — validates source, resolves items, submits to cart, persists `CartHandoffRecord`, emits event
- [ ] **Step 12** — Implement API layer
  - [ ] `api/dto.go`
  - [ ] `api/middleware.go` — session auth middleware
  - [ ] `api/handlers.go`
  - [ ] `api/router.go` — go-chi router
- [ ] **Step 13** — Implement entrypoint
  - [ ] `cmd/main.go`
- [ ] **Step 14** — Write Unit 5 demo script (`demo/main.go`) — runs: add by `comboId` (stub Unit 4 ACL returns a combo), add by inline items, partial out-of-stock scenario, full failure scenario

---

> **Awaiting your review and approval before execution.**
> Please fill in the [Answer] tags above, then confirm. I will execute one step at a time and mark each checkbox as done upon completion.
