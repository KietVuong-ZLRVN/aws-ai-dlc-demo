# Unit 3 — AI Styling Engine: Backend Test Plan

## Scope

Backend only. All test layers: value objects, domain entities, aggregates, policies, use cases, controllers, middleware, and infrastructure stubs. Frontend is out of scope.

## Approach

Tests are written in Go using the standard `testing` package, co-located with the packages they test (`*_test.go` files). Mocks and in-memory stubs already exist in `infrastructure/acl` and `infrastructure/ai` — these are used throughout. No real AWS Bedrock or HTTP calls are made.

Existing tests are noted per section. New tests fill the identified gaps.

---

## Test ID Convention

`TC-{US}-{seq}` — e.g. `TC-301-1` maps to User Story 301, test case 1.
`TC-DOM-{seq}` — domain invariant tests not tied to a single user story.
`TC-INFRA-{seq}` — infrastructure / dispatcher tests.
`TC-SEC-{seq}` — security / cross-cutting tests.

---

## Plan

### Phase 1 — Audit existing tests and confirm they pass

- [x] **Step 1.1** — Run the full existing test suite and confirm all tests pass with zero failures. ✅
  - Command: `cd construction/unit_3_ai_styling_engine/src/backend && go test ./...`

---

### Phase 2 — Value Object Tests (`domain/valueobjects`)

Existing coverage: `BudgetRange`, `ColorPalette`, `StylePreferences.IsEmpty`, `ExcludedComboIds`, `ComboReasoning`, `WishlistSnapshot.InStockItems`, `CatalogSearchFilters`.

#### Gaps to fill:

- [x] **Step 2.1 — `BudgetRange` edge cases**
  - [x] TC-DOM-1: `NewBudgetRange(0, 0)` returns error (max not > min)
  - [x] TC-DOM-2: `NewBudgetRange(0, 1)` succeeds (min = 0 is valid)

- [x] **Step 2.2 — `ColorPalette` additional cases**
  - [x] TC-DOM-3: Multiple colors in both lists, only one overlap → returns error
  - [x] TC-DOM-4: Same color appears twice in preferred list only → no error (no cross-list conflict)

- [x] **Step 2.3 — `StylePreferences.IsEmpty` completeness**
  - [x] TC-DOM-5: `IsEmpty()` returns false when only `Colors.Preferred` is set
  - [x] TC-DOM-6: `IsEmpty()` returns false when only `Colors.Excluded` is set

- [x] **Step 2.4 — `ComboReasoning` whitespace**
  - [x] TC-DOM-7: `NewComboReasoning("   ")` (whitespace-only) returns error

- [x] **Step 2.5 — `WishlistSnapshot` helpers**
  - [x] TC-DOM-8: `InStockItems()` on empty snapshot returns empty slice (not nil panic)

- [x] **Step 2.6 — `CatalogSearchFilters` from preferences with budget**
  - [x] TC-DOM-9: `CatalogSearchFiltersFromPreferences` with budget set → `PriceRange` is populated
  - [x] TC-DOM-10: `CatalogSearchFiltersFromPreferences` with preferred colors → `Colors` is populated

---

### Phase 3 — Entity Tests (`domain/entities`)

Existing coverage: `Combo` construction, `AttachReasoning`, `Reject`/`IsRejected`; `FallbackResult` construction.

#### Gaps to fill:

- [x] **Step 3.1 — `Combo` rank ordering**
  - [x] TC-DOM-11: Two combos with rank 1 and 2 — rank values are stored correctly

- [x] **Step 3.2 — `FallbackResult` non-empty reason invariant**
  - [x] TC-DOM-12: `FallbackResult` with at least one alternative that has an empty `Reason` field — verify the field is stored as-is (enforcement is at policy level, not entity level)

---

### Phase 4 — Aggregate Tests (`domain/aggregates`)

Existing coverage: `StyleSession` construction, `QuickGenerate` flag, `LoadWishlist`, `LoadCatalogItems`, `CompleteCombos` (with exclusion), `TriggerFallback`; `PreferenceConfirmation.Interpret`.

#### Gaps to fill:

- [x] **Step 4.1 — `StyleSession` state guard: scoring before wishlist loaded**
  - [x] TC-DOM-13: Calling `CompleteCombos` before `LoadWishlist` — returns error (wishlist must be loaded first)

- [x] **Step 4.2 — `StyleSession` exclusion retry flag**
  - [x] TC-DOM-14: After `CompleteCombos` with all combos excluded, `IsExhausted()` returns true

- [x] **Step 4.3 — `PreferenceConfirmation` echoes preferences**
  - [x] TC-DOM-15: `Interpret` result's `Preferences` field matches the input `StylePreferences` exactly

---

### Phase 5 — Policy Tests (`domain/policies`)

Existing coverage: `PreferenceDefaultPolicy` (no-panic), `WishlistSupplementationPolicy` (2 in-stock / 1 in-stock / empty / all out-of-stock), `FallbackPolicy` (with/without alternatives), `ComboExclusionPolicy` (no-panic).

#### Gaps to fill:

- [x] **Step 5.1 — `WishlistSupplementationPolicy` exactly two in-stock items**
  - [x] TC-403-5: Exactly 2 in-stock items → `CatalogSupplementationRequired` is NOT raised (boundary value)

- [x] **Step 5.2 — `FallbackPolicy` logs warning on empty alternatives**
  - [x] TC-404-4: `FallbackPolicy.Handle` with empty alternatives list — does not panic, does not raise a secondary event

- [x] **Step 5.3 — `ComboExclusionPolicy` retry signal**
  - [x] TC-405-6: `ComboExclusionPolicy` with fewer than 2 combos remaining after exclusion — does not panic, logs outcome

---

### Phase 6 — Use Case Tests (`application/usecases`)

Existing coverage: `GetPreferenceOptionsUseCase`, `ConfirmPreferencesUseCase` (full/empty prefs), `GenerateCombosUseCase` (quick-generate, with prefs, combo structure, catalog supplementation, fallback, exclusion, wishlist error).

#### Gaps to fill:

- [x] **Step 6.1 — `GetPreferenceOptionsUseCase` exact values**
  - [x] TC-301b-2: Response contains exactly the 6 occasions defined in the domain model (`casual`, `formal`, `outdoor`, `beach`, `office`, `party`)
  - [x] TC-301b-3: Response contains exactly the 4 style directions (`minimalist`, `bold`, `classic`, `bohemian`)
  - [x] TC-301b-4: Response contains exactly the 7 colors defined in the domain model

- [x] **Step 6.2 — `ConfirmPreferencesUseCase` preference echo**
  - [x] TC-302-3: Returned `PreferenceSummary.Preferences` echoes all non-empty input fields exactly (occasions, styles, budget, colors, freeText)

- [x] **Step 6.3 — `GenerateCombosUseCase` combo item source labeling**
  - [x] TC-403-3: When catalog supplementation occurs (sparse wishlist), at least one item in a combo has `source = catalog`

- [x] **Step 6.4 — `GenerateCombosUseCase` fallback has non-empty alternatives**
  - [x] TC-404-1: Fallback result contains at least one `AlternativeItem`
  - [x] TC-404-3: Each `AlternativeItem` in the fallback has a non-empty `Reason`

- [x] **Step 6.5 — `GenerateCombosUseCase` exhausted combos response shape**
  - [x] TC-405-5: When all combos are excluded and the result is exhausted, `Success.Combos` is an empty slice (not nil)

- [x] **Step 6.6 — `GenerateCombosUseCase` AI inference failure**
  - [x] TC-INFRA-1: When `ComboCompatibilityScoringService.Score` returns an error, `GenerateCombosUseCase.Execute` returns a non-nil error (not a fallback result)
  no, base all on the https code should be fine

---

### Phase 7 — Controller Tests (`api/controllers`)

Existing coverage: `GET /preferences/options` (200 + fields), `POST /preferences/confirm` (valid/empty/invalid JSON/bad budget/color conflict), `POST /combos/generate` (quick-generate, non-empty combos, reasoning+items, with prefs, exclusion, invalid JSON, correlation header).

#### Gaps to fill:

- [ ] **Step 7.1 — `GET /preferences/options` exact enum values**
  - [ ] TC-301b-5: Response `occasions` array contains `"casual"`, `"formal"`, `"outdoor"`, `"beach"`, `"office"`, `"party"` — no extras, no missing

- [ ] **Step 7.2 — `POST /preferences/confirm` response shape**
  - [ ] TC-302-5: Response `preferences` field echoes back the submitted occasions, styles, budget, colors, and freeText

- [ ] **Step 7.3 — `POST /combos/generate` fallback response shape**
  - [ ] TC-404-6: When scoring returns fallback, response has `status = "fallback"`, non-empty `message`, and non-empty `alternatives` array
  - [ ] TC-404-7: Each alternative in the fallback response has `configSku`, `simpleSku`, `name`, `brand`, `price`, `imageUrl`, `reason` fields

- [ ] **Step 7.4 — `POST /combos/generate` item source field**
  - [ ] TC-403-6: Each item in a combo response has `source` field set to either `"wishlist"` or `"catalog"`

- [ ] **Step 7.5 — `POST /combos/generate` budget validation**
  - [ ] TC-SEC-3b: `budget.max == budget.min` returns 400 with `VALIDATION_ERROR`
  - [ ] TC-SEC-3c: `budget.min < 0` returns 400 with `VALIDATION_ERROR`

- [ ] **Step 7.6 — `POST /combos/generate` unrecognised enum values**
  - [ ] TC-SEC-8: Request with unrecognised occasion value (e.g. `"disco"`) returns 400 with `VALIDATION_ERROR`
  - [ ] TC-SEC-9: Request with unrecognised style value returns 400 with `VALIDATION_ERROR`
  - [ ] TC-SEC-10: Request with unrecognised color value returns 400 with `VALIDATION_ERROR`

  [Question] Does the current `ComboGenerationController` validate enum values for occasions, styles, and colors in the request body, or does it pass unknown strings through to the domain? If not currently validated, should this be added as part of this test plan?
  [Answer]
  should validate

- [ ] **Step 7.7 — `POST /combos/generate` wishlist unavailable → 502**
  - [ ] TC-SEC-7b: When the wishlist repository returns `WishlistUnavailableException`, the controller returns HTTP 502 with `{ "error": "DEPENDENCY_UNAVAILABLE" }`

  [Question] Is `WishlistUnavailableException` a named error type in the current Go implementation, or is it a generic `error`? The controller needs to distinguish it from other errors to return 502 vs 500.
  [Answer]
  It is a distinctive error

- [ ] **Step 7.8 — `POST /combos/generate` AI failure → 503**
  - [ ] TC-SEC-11: When the scoring service returns an `AIInferenceException`-equivalent error, the controller returns HTTP 503 with `{ "error": "AI_UNAVAILABLE" }`

  [Question] Same as above — is there a typed `AIInferenceException` or `ErrAIUnavailable` sentinel in the current implementation?
  [Answer]
  It is also a distinctive error

- [ ] **Step 7.9 — Unhandled exception → 500**
  - [ ] TC-SEC-12: An unexpected panic or error from the use case results in HTTP 500 with `{ "error": "INTERNAL_ERROR" }` and does not leak stack trace or sensitive data in the response body

---

### Phase 8 — Middleware Tests (`api/middleware`)

Existing coverage: `AuthMiddleware` (valid cookie / no cookie / empty cookie), `TracingMiddleware` (sets header / forwards existing ID).

#### Gaps to fill:

- [x] **Step 8.1 — `AuthMiddleware` response body**
  - [x] TC-SEC-1b: 403 response body is `{ "error": "UNAUTHENTICATED" }` (not just status code)

- [x] **Step 8.2 — `AuthMiddleware` session in context**
  - [x] TC-SEC-2b: After valid auth, `SessionFromContext` returns a `ShopperSession` with the token value matching the cookie

- [x] **Step 8.3 — `RequestLoggingMiddleware` sensitive field masking**
  - [x] TC-SEC-13/14: Middleware writes directly to `log.Printf` — no injected writer. Documented as a known gap requiring a future refactor to inject an `io.Writer` before these can be asserted programmatically.

- [x] **Step 8.4 — `TracingMiddleware` outbound header propagation**
  - [x] TC-SEC-6b: `X-Correlation-ID` is set on the response even when no incoming header is present (generated UUID)

---

### Phase 9 — Infrastructure / Dispatcher Tests

Existing coverage: Mock scoring service (combos/fallback/exclusion/count), mock reasoning service (non-empty/item name/occasion reference), mock interpretation service (empty/with occasion/free text/echo).

#### Gaps to fill:

- [x] **Step 9.1 — `InProcessEventDispatcher`**
  - [x] TC-INFRA-2: Registering two handlers for the same event type — both are called in registration order
  - [x] TC-INFRA-3: Dispatching an event with no registered handlers — does not panic
  - [x] TC-INFRA-4: A panicking handler — panic is recovered and re-panicked as a wrapped error

- [x] **Step 9.2 — In-memory ACL stubs**
  - [x] TC-INFRA-5: `InMemoryWishlistRepository.FetchForSession` returns a snapshot with at least 2 in-stock items
  - [x] TC-INFRA-6: `InMemoryProductCatalogRepository.SearchSupplementaryItems` returns a non-empty list
  - [x] TC-INFRA-7: `InMemoryCompleteLookRepository.FetchCompleteLookSignals` returns a non-empty list for a known SKU

---

### Phase 10 — Integration / End-to-End (in-process, no real HTTP server)

These tests exercise the full request pipeline using `httptest` and the real router, with all stubs wired in.

- [x] **Step 10.1 — Full pipeline: quick-generate success**
  - [x] TC-E2E-1: `POST /api/v1/style/combos/generate {}` with session cookie → 200, `status: "ok"`, combos array with ≥ 1 combo, each combo has `id`, `reasoning`, `items` (≥ 2 items each with `source`)

- [x] **Step 10.2 — Full pipeline: preference-guided success**
  - [x] TC-E2E-2: `POST /api/v1/style/combos/generate` with full preferences → 200, `status: "ok"`

- [x] **Step 10.3 — Full pipeline: preference confirmation flow**
  - [x] TC-E2E-3: `POST /api/v1/style/preferences/confirm` → 200, `summary` non-empty, `preferences` echoed

- [x] **Step 10.4 — Full pipeline: unauthenticated request**
  - [x] TC-E2E-4: Any endpoint without session cookie → 403, `{ "error": "UNAUTHENTICATED" }`

- [x] **Step 10.5 — Full pipeline: fallback path**
  - [x] TC-E2E-5: With a stub scoring service that always returns fallback → 200, `status: "fallback"`, `message` non-empty, `alternatives` non-empty

- [x] **Step 10.6 — Full pipeline: combo exclusion across two calls**
  - [x] TC-E2E-6: First call returns combo IDs; second call with those IDs in `excludeComboIds` → none of the excluded IDs appear in the second response

---

### Phase 11 — Final verification

- [x] **Step 11.1** — Run the full test suite again after all new tests are added and confirm zero failures. ✅
  - Command: `cd construction/unit_3_ai_styling_engine/src/backend && go test ./... -v`

- [x] **Step 11.2** — Review test coverage report and confirm all critical paths are covered. ✅
  - Command: `cd construction/unit_3_ai_styling_engine/src/backend && go test ./... -cover`

---

## Open Questions Summary

| # | Location | Question |
|---|---|---|
| Q1 | Step 4.1 | Should `CompleteCombos` before `LoadWishlist` error, panic, or silently proceed? |
| Q2 | Step 5.2 | Should `FallbackPolicy` with empty alternatives raise a secondary event or just log? |
| Q3 | Step 5.3 | Is the `ComboExclusionPolicy` retry implemented inside the policy or signalled to the use case? |
| Q4 | Step 6.6 | Is there a distinction between `ScoringResult.Fallback` (business) and `error` (infrastructure failure) in the use case? |
| Q5 | Step 7.6 | Does the controller currently validate enum values for occasions/styles/colors? |
| Q6 | Step 7.7 | Is `WishlistUnavailableException` a named error type in Go, or a generic error? |
| Q7 | Step 7.8 | Is there a typed `AIInferenceException` / `ErrAIUnavailable` sentinel? |
| Q8 | Step 8.3 | Is `RequestLoggingMiddleware` testable via an injected writer, or does it write to stdout/log? |
| Q9 | Step 9.1 | Should the dispatcher recover from handler panics as errors, or let them propagate? |
