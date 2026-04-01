# Logical Design: Unit 3 — AI Styling Engine

## Overview

This document describes the logical software design for Unit 3, derived from the [domain model](./domain_model.md) and the [integration contract](../../inception/units/integration_contract.md). It specifies how the domain model maps to a layered source code structure, how each layer is organised, how components interact at runtime, and how cross-cutting concerns are handled.

**AI Platform:** AWS Bedrock (Claude model family)
**Domain events:** In-process synchronous dispatch
**External signals:** Complete-the-Look signals fetched as an explicit separate step before AI scoring

---

## 1. Layered Architecture

The service follows a strict four-layer architecture aligned with DDD conventions. Each layer has a single direction of dependency: outer layers depend on inner layers; inner layers never reference outer layers.

```
┌─────────────────────────────────────────────┐
│                  API Layer                  │  HTTP in / HTTP out
│  Controllers · Middleware · Request/Response│
├─────────────────────────────────────────────┤
│              Application Layer              │  Orchestration
│         Use Cases · Commands · Queries      │
├─────────────────────────────────────────────┤
│               Domain Layer                  │  Business logic
│  Aggregates · Entities · Value Objects      │
│  Events · Policies · Domain Services        │
│  Repository Interfaces (Ports)              │
├─────────────────────────────────────────────┤
│            Infrastructure Layer             │  I/O + AI
│  ACL Implementations · Bedrock Clients      │
│  In-Process Event Dispatcher                │
└─────────────────────────────────────────────┘
```

---

## 2. Source Folder Structure

```
src/
├── api/
│   ├── controllers/
│   │   ├── StylePreferencesController
│   │   ├── PreferenceConfirmationController
│   │   └── ComboGenerationController
│   ├── middleware/
│   │   ├── AuthenticationMiddleware
│   │   ├── RequestLoggingMiddleware
│   │   └── TracingMiddleware
│   └── dto/
│       ├── request/
│       │   ├── ConfirmPreferencesRequest
│       │   └── GenerateCombosRequest
│       └── response/
│           ├── PreferenceOptionsResponse
│           ├── PreferenceConfirmationResponse
│           ├── ComboGenerationSuccessResponse
│           └── ComboGenerationFallbackResponse
│
├── application/
│   ├── usecases/
│   │   ├── GetPreferenceOptionsUseCase
│   │   ├── ConfirmPreferencesUseCase
│   │   └── GenerateCombosUseCase
│   └── commands/
│       ├── ConfirmPreferencesCommand
│       └── GenerateCombosCommand
│
├── domain/
│   ├── aggregates/
│   │   ├── StyleSession
│   │   └── PreferenceConfirmation
│   ├── entities/
│   │   ├── Combo
│   │   └── FallbackResult
│   ├── valueobjects/
│   │   ├── StylePreferences
│   │   ├── BudgetRange
│   │   ├── ColorPalette
│   │   ├── ComboItem
│   │   ├── ComboReasoning
│   │   ├── WishlistItem
│   │   ├── WishlistSnapshot
│   │   ├── AlternativeItem
│   │   ├── PreferenceSummary
│   │   ├── ExcludedComboIds
│   │   ├── StyleOptionsCatalogue
│   │   ├── CatalogSearchFilters
│   │   └── ShopperSession
│   ├── events/
│   │   ├── ComboGenerationRequested
│   │   ├── WishlistFetchCompleted
│   │   ├── CatalogSupplementationRequired
│   │   ├── CatalogItemsFetched
│   │   ├── CombosGenerated
│   │   ├── FallbackTriggered
│   │   ├── ComboRejected
│   │   └── PreferencesInterpreted
│   ├── policies/
│   │   ├── PreferenceDefaultPolicy
│   │   ├── WishlistSupplementationPolicy
│   │   ├── FallbackPolicy
│   │   └── ComboExclusionPolicy
│   ├── services/
│   │   ├── ComboCompatibilityScoringService  (interface)
│   │   ├── ComboReasoningGenerationService   (interface)
│   │   └── PreferenceInterpretationService   (interface)
│   └── repositories/
│       ├── WishlistRepository                (interface)
│       ├── ProductCatalogRepository          (interface)
│       └── CompleteLookRepository            (interface)
│
└── infrastructure/
    ├── acl/
    │   ├── HttpWishlistRepository
    │   ├── HttpProductCatalogRepository
    │   └── HttpCompleteLookRepository
    ├── ai/
    │   ├── BedrockScoringClient
    │   ├── BedrockLLMClient
    │   ├── BedrockComboCompatibilityScoringService
    │   ├── BedrockComboReasoningGenerationService
    │   └── BedrockPreferenceInterpretationService
    └── events/
        ├── InProcessEventDispatcher
        └── EventHandlerRegistry
```

---

## 3. API Layer

### 3.1 Controllers

#### `StylePreferencesController`

Handles `GET /api/v1/style/preferences/options`.

- Receives the request after `AuthenticationMiddleware` and `TracingMiddleware` have run
- Delegates to `GetPreferenceOptionsUseCase` with no command parameters (the options are static configuration)
- Maps the returned `StyleOptionsCatalogue` value object to `PreferenceOptionsResponse`
- Returns HTTP 200

#### `PreferenceConfirmationController`

Handles `POST /api/v1/style/preferences/confirm`.

- Deserialises request body into `ConfirmPreferencesRequest`
- Validates request shape (all fields optional; body must be valid JSON)
- Constructs `ConfirmPreferencesCommand` with parsed `StylePreferences` and `ShopperSession` from middleware context
- Delegates to `ConfirmPreferencesUseCase`
- Maps returned `PreferenceSummary` to `PreferenceConfirmationResponse`
- Returns HTTP 200

#### `ComboGenerationController`

Handles `POST /api/v1/style/combos/generate`.

- Deserialises request body into `GenerateCombosRequest`
- Validates: `excludeComboIds` contains only non-empty strings; `budget.max > budget.min` if both present; no color appears in both preferred and excluded lists
- Constructs `GenerateCombosCommand` with optional `StylePreferences`, optional `ExcludedComboIds`, and `ShopperSession`
- Delegates to `GenerateCombosUseCase`
- Inspects result type:
  - `ComboGenerationResult.Success` → maps to `ComboGenerationSuccessResponse`, HTTP 200 with `"status": "ok"`
  - `ComboGenerationResult.Fallback` → maps to `ComboGenerationFallbackResponse`, HTTP 200 with `"status": "fallback"`
- Returns HTTP 403 if authentication is missing (handled by middleware before controller is reached)

---

### 3.2 Request DTOs

#### `ConfirmPreferencesRequest`

| Field | Type | Required |
|---|---|---|
| `occasions` | `string[]` | No |
| `styles` | `string[]` | No |
| `budget` | `{ min: number, max: number }` | No |
| `colors` | `{ preferred: string[], excluded: string[] }` | No |
| `freeText` | `string` | No |

#### `GenerateCombosRequest`

| Field | Type | Required |
|---|---|---|
| `preferences` | `ConfirmPreferencesRequest` (nested) | No |
| `excludeComboIds` | `string[]` | No |

---

### 3.3 Response DTOs

#### `PreferenceOptionsResponse`

| Field | Type |
|---|---|
| `occasions` | `string[]` |
| `styles` | `string[]` |
| `colors` | `string[]` |

#### `PreferenceConfirmationResponse`

| Field | Type |
|---|---|
| `summary` | `string` |
| `preferences` | `ConfirmPreferencesRequest` (echo) |

#### `ComboGenerationSuccessResponse`

| Field | Type |
|---|---|
| `status` | `"ok"` |
| `combos` | `ComboDTO[]` |

**`ComboDTO`:**

| Field | Type |
|---|---|
| `id` | `string` |
| `reasoning` | `string` |
| `items` | `ComboItemDTO[]` |

**`ComboItemDTO`:**

| Field | Type |
|---|---|
| `configSku` | `string` |
| `simpleSku` | `string` |
| `name` | `string` |
| `brand` | `string` |
| `price` | `number` |
| `imageUrl` | `string` |
| `source` | `"wishlist" \| "catalog"` |

#### `ComboGenerationFallbackResponse`

| Field | Type |
|---|---|
| `status` | `"fallback"` |
| `message` | `string` |
| `alternatives` | `AlternativeItemDTO[]` |

**`AlternativeItemDTO`:**

| Field | Type |
|---|---|
| `configSku` | `string` |
| `simpleSku` | `string` |
| `name` | `string` |
| `brand` | `string` |
| `price` | `number` |
| `imageUrl` | `string` |
| `reason` | `string` |

---

## 4. Application Layer

### 4.1 Use Cases

#### `GetPreferenceOptionsUseCase`

**Input:** None (no command; options are static)

**Steps:**
1. Returns the `StyleOptionsCatalogue` value object constructed from service configuration

**Output:** `StyleOptionsCatalogue`

**Note:** No domain events, no external I/O. The catalogue is a static configuration value, not fetched from a database or external API.

---

#### `ConfirmPreferencesUseCase`

**Input:** `ConfirmPreferencesCommand` { `preferences: StylePreferences`, `shopperSession: ShopperSession` }

**Steps:**
1. Construct `PreferenceConfirmation` aggregate with the provided `StylePreferences`
2. Call `PreferenceConfirmation.interpret()` — internally delegates to `PreferenceInterpretationService`
3. `PreferenceInterpretationService` sends preferences to AWS Bedrock; returns `PreferenceSummary`
4. `PreferenceConfirmation` raises `PreferencesInterpreted` (in-process; no handler registered for this event in this use case — it is informational/trace only)
5. Return `PreferenceSummary`

**Output:** `PreferenceSummary`

---

#### `GenerateCombosUseCase`

**Input:** `GenerateCombosCommand` { `preferences?: StylePreferences`, `excludeComboIds?: ExcludedComboIds`, `shopperSession: ShopperSession` }

This is the primary orchestration use case. Its steps are described in detail in Section 5 (Domain Layer — Pipeline Flow).

**Output:** `ComboGenerationResult` (either `Success` with combos, or `Fallback` with alternatives)

---

### 4.2 Commands

#### `ConfirmPreferencesCommand`

| Field | Type |
|---|---|
| `preferences` | `StylePreferences` |
| `shopperSession` | `ShopperSession` |

#### `GenerateCombosCommand`

| Field | Type |
|---|---|
| `preferences` | `StylePreferences` (optional) |
| `excludeComboIds` | `ExcludedComboIds` (optional) |
| `shopperSession` | `ShopperSession` |

---

## 5. Domain Layer

### 5.1 Aggregate Lifecycle: `StyleSession`

`StyleSession` is a transient aggregate that exists only for the duration of the `GenerateCombosUseCase` execution. It drives the combo generation pipeline through explicit state transitions, raising domain events at each stage that registered policies react to.

**State machine:**

```
INITIATED
   │
   ▼ (on construction)
   raises ComboGenerationRequested
   ──► PreferenceDefaultPolicy reacts:
       sets session.quickGenerate = true if preferences.isEmpty()
   │
   ▼ (after wishlist fetch)
   raises WishlistFetchCompleted
   ──► WishlistSupplementationPolicy reacts:
       if wishlist sufficient → no-op (proceed to scoring)
       if wishlist insufficient → raises CatalogSupplementationRequired
   │
   ├─ [if CatalogSupplementationRequired was raised]
   │     ▼ (after catalog fetch)
   │     raises CatalogItemsFetched
   │
   ▼ (after AI scoring and reasoning)
   raises CombosGenerated
   ──► ComboExclusionPolicy reacts:
       filters excluded combos
       if count falls below minimum → triggers one internal retry cycle
   │
   ├─ [if no combos remain after scoring]
   │     raises FallbackTriggered
   │     ──► FallbackPolicy reacts:
   │         ensures alternatives are non-empty
   │
   └─ COMPLETED (either combos or fallback result is ready)
```

---

### 5.2 Full Pipeline Flow — `GenerateCombosUseCase`

The following describes the step-by-step orchestration performed by the use case, including how it interacts with the domain and infrastructure layers.

**Step 1 — Initialise session**

The use case constructs a `StyleSession` with a new `StyleSessionId`, the shopper's `ShopperSession`, the optional `StylePreferences`, and the optional `ExcludedComboIds`.

`StyleSession` raises `ComboGenerationRequested`. The `InProcessEventDispatcher` synchronously invokes `PreferenceDefaultPolicy`, which checks `StylePreferences.isEmpty()`. If true, the session's `quickGenerate` flag is set to `true` — this flag is passed to the scoring service to signal that no preference constraints should be applied.

**Step 2 — Fetch wishlist**

The use case calls `WishlistRepository.fetchForSession(shopperSession)`. The `HttpWishlistRepository` implementation makes an authenticated HTTP call to Unit 2's `GET /api/v1/wishlist`, forwarding the shopper's session cookie. The response is translated into a `WishlistSnapshot`.

`StyleSession.loadWishlist(snapshot)` is called. The aggregate raises `WishlistFetchCompleted`. The `InProcessEventDispatcher` invokes `WishlistSupplementationPolicy`.

`WishlistSupplementationPolicy` evaluates: if fewer than two in-stock `WishlistItem`s are present, or if no valid pair (primary + complementary) can be identified from the wishlist alone, it raises `CatalogSupplementationRequired`.

**Step 3 — Fetch supplementary signals (conditional)**

This step only runs if `CatalogSupplementationRequired` was raised in Step 2.

*Step 3a — Fetch Complete-the-Look signals (explicit separate step):*
The use case iterates over each in-stock `WishlistItem` and calls `CompleteLookRepository.fetchCompleteLookSignals(configSku)` for each. The collected results form the base styling signals. These are `ComboItem` objects with `source = catalog`.

*Step 3b — Fetch catalog supplementary items:*
The use case constructs `CatalogSearchFilters` from the session's `StylePreferences` (preferred colors, first occasion, budget range). It calls `ProductCatalogRepository.searchSupplementaryItems(filters)`, which queries the Platform Product API and returns additional `ComboItem` objects.

Both signal sets (Complete-the-Look results and catalog search results) are merged and deduplicated by `configSku`. The merged set is stored on the session via `StyleSession.loadCatalogItems(mergedItems)`, which raises `CatalogItemsFetched`.

**Step 4 — AI compatibility scoring**

The use case calls `ComboCompatibilityScoringService.score(scoringInput)`.

`ScoringInput` contains:
- `wishlistItems` from `WishlistSnapshot`
- `supplementaryItems` from catalog (empty list if no supplementation occurred)
- `completeLookSignals` from Step 3a (empty list if no supplementation occurred)
- `preferences` from the session (null if quick-generate)
- `excludedComboIds` from the session

The `BedrockComboCompatibilityScoringService` implementation sends a structured prompt to AWS Bedrock (Claude). The prompt provides item attributes and instructs the model to:
- Group items into valid combos (at minimum: one wishlist item + one complementary item)
- Rank combos by fashion compatibility
- Return up to 5 combos as structured JSON
- If no valid combo can be formed: return a fallback list of alternative items with replacement reasons

The Bedrock response is parsed into a `ScoringResult`, which is either:
- `ScoringResult.Combos(rankedCombos: List<ComboCandidate>)` — each candidate has item list and internal compatibility score
- `ScoringResult.Fallback(alternatives: List<AlternativeItem>)` — when no combo can be formed

**Step 5 — Handle fallback (conditional)**

If `ScoringResult.Fallback` is returned:

`StyleSession.triggerFallback(alternatives)` is called. The aggregate creates the `FallbackResult` entity and raises `FallbackTriggered`. The `InProcessEventDispatcher` invokes `FallbackPolicy`, which verifies the alternatives list is non-empty (if empty, it logs a warning and the use case returns a minimal fallback response).

The use case returns `ComboGenerationResult.Fallback(fallbackResult)`. The pipeline ends here.

**Step 6 — Generate reasoning**

For each `ComboCandidate` from the scoring result, the use case calls `ComboReasoningGenerationService.generateReasoning(candidate, preferences)`.

The `BedrockComboReasoningGenerationService` implementation sends a prompt to AWS Bedrock (Claude) containing the combo's item attributes and the shopper's stated preferences (if any). The model returns a 1–3 sentence natural-language explanation. The result is wrapped in a `ComboReasoning` value object and attached to the `Combo` entity via `Combo.attachReasoning(reasoning)`.

These calls are made in parallel across all combo candidates to minimise latency.

**Step 7 — Finalise combos**

`StyleSession.completeCombos(combos)` is called with the fully reasoned `Combo` entities. The aggregate raises `CombosGenerated`. The `InProcessEventDispatcher` invokes `ComboExclusionPolicy`.

`ComboExclusionPolicy` removes any combo whose `ComboId` is in `ExcludedComboIds`. If the remaining count is below 2, the policy triggers one internal retry: it calls back into the scoring service with the full `ExcludedComboIds` set (now including the just-excluded IDs). If the retry still yields fewer than 2 combos, the policy accepts the result as-is and sets an `exhausted` flag on the session.

The use case returns `ComboGenerationResult.Success(combos, exhausted)`. The `exhausted` flag is mapped to a hint in the response for the client to display the "We've shown you all available combos" message.

---

### 5.3 Policy Wiring

Policies are registered with the `InProcessEventDispatcher` at application startup (dependency injection wiring). Each policy is a singleton that holds references to the interfaces it needs.

| Policy | Handles Event | Requires |
|---|---|---|
| `PreferenceDefaultPolicy` | `ComboGenerationRequested` | None (pure logic) |
| `WishlistSupplementationPolicy` | `WishlistFetchCompleted` | `InProcessEventDispatcher` (to raise `CatalogSupplementationRequired`) |
| `FallbackPolicy` | `FallbackTriggered` | None (pure validation) |
| `ComboExclusionPolicy` | `CombosGenerated` | `ComboCompatibilityScoringService` (for retry), `InProcessEventDispatcher` |

---

### 5.4 Domain Service Interfaces (Ports)

#### `ComboCompatibilityScoringService` (Port)

```
score(input: ScoringInput): ScoringResult
```

`ScoringInput`:
- `wishlistItems: List<WishlistItem>`
- `supplementaryItems: List<ComboItem>`
- `completeLookSignals: List<ComboItem>`
- `preferences: StylePreferences?`
- `excludedComboIds: ExcludedComboIds`
- `quickGenerate: boolean`

`ScoringResult`: sealed type, either `Combos(List<ComboCandidate>)` or `Fallback(List<AlternativeItem>)`

---

#### `ComboReasoningGenerationService` (Port)

```
generateReasoning(candidate: ComboCandidate, preferences: StylePreferences?): ComboReasoning
```

---

#### `PreferenceInterpretationService` (Port)

```
interpret(preferences: StylePreferences): PreferenceSummary
```

---

#### `WishlistRepository` (Port)

```
fetchForSession(session: ShopperSession): WishlistSnapshot
```

Throws `WishlistUnavailableException` if Unit 2 is unreachable or returns a non-2xx response.

---

#### `ProductCatalogRepository` (Port)

```
searchSupplementaryItems(filters: CatalogSearchFilters): List<ComboItem>
```

Returns empty list on platform API failure (graceful degradation — scoring proceeds with wishlist only).

---

#### `CompleteLookRepository` (Port)

```
fetchCompleteLookSignals(configSku: Sku): List<ComboItem>
```

Returns empty list on platform API failure (graceful degradation — scoring proceeds without these signals).

---

## 6. Infrastructure Layer

### 6.1 ACL Implementations

#### `HttpWishlistRepository`

Implements `WishlistRepository`. Makes a GET request to Unit 2's `GET /api/v1/wishlist` with the shopper's session cookie forwarded. Maps each item in the response array to a `WishlistItem` value object. If Unit 2 returns 403, propagates as authentication failure. If Unit 2 returns 5xx or is unreachable, throws `WishlistUnavailableException`.

---

#### `HttpProductCatalogRepository`

Implements `ProductCatalogRepository`. Makes a GET request to the Platform Product API at `GET /v1/products/list`. Constructs query parameters from `CatalogSearchFilters` (colors, occasion, price range). Applies required platform headers (`Content-Language: en-SG`, `Accept: application/json`). Maps `products[]` to `ComboItem` value objects with `source = catalog`. Returns empty list on any non-2xx response to enable graceful degradation.

---

#### `HttpCompleteLookRepository`

Implements `CompleteLookRepository`. Makes a GET request to the Platform Recommendation API at `GET /v1/recommendation/completethelook/{config_sku}`. Maps `products[]` to `ComboItem` value objects with `source = catalog`. Returns empty list on any non-2xx response.

---

### 6.2 In-Process Event Dispatcher

#### `InProcessEventDispatcher`

A synchronous, in-memory event dispatcher. It maintains a registry mapping each event type to zero or more ordered handler functions (policies).

**Dispatch behaviour:**
- When an aggregate raises an event, it calls `dispatcher.dispatch(event)`
- The dispatcher looks up all registered handlers for that event type
- Handlers are invoked synchronously, in registration order, within the same call stack
- If a handler raises an exception, it propagates immediately and halts the pipeline; the use case catches it as an application error

**`EventHandlerRegistry`:** Constructed at application startup and injected into all aggregates and policies that need to dispatch events. Handlers are registered once at startup and never modified at runtime.

---

### 6.3 AWS Bedrock Integration

All Bedrock calls go through a shared `BedrockLLMClient` or `BedrockScoringClient` that wraps the AWS Bedrock Runtime SDK. These clients handle:
- AWS credential resolution (IAM role / environment credentials)
- Model ID configuration (injected from environment, e.g. `anthropic.claude-3-5-sonnet-20241022-v2:0`)
- Request serialisation and response deserialisation
- Retry on transient throttling (one retry with exponential backoff)
- Timeout enforcement (configurable per service; recommended: 10s for scoring, 5s for reasoning and interpretation)

#### `BedrockComboCompatibilityScoringService`

Implements `ComboCompatibilityScoringService`. Uses `BedrockScoringClient`.

**Prompt strategy:**
- System prompt: Establishes the model as a fashion styling expert with knowledge of colour theory, occasion dressing, and item category balance
- User prompt: Provides a structured list of available items (name, brand, colour, occasion tags, price, category) grouped as wishlist items and supplementary items. Includes any Complete-the-Look signals as styling hints. Includes the shopper's preferences if not quick-generate. Instructs the model to return a structured JSON response
- Output format: Instructs the model to return either a `combos` array (ranked, 2–5 entries, each with an `id`, ranked `items` array, and internal `score`) or a `fallback` object (with an `alternatives` array, each item with a `reason`)
- Uses Bedrock's tool use / structured JSON output feature to enforce the response schema

#### `BedrockComboReasoningGenerationService`

Implements `ComboReasoningGenerationService`. Uses `BedrockLLMClient`.

**Prompt strategy:**
- System prompt: Fashion copywriter persona; concise, friendly tone
- User prompt: Lists the combo's items (name, brand, colour) and the shopper's preferences (if any). Asks for a 1–3 sentence explanation of why these items work together, referencing preferences where relevant
- Output: Plain text; trimmed to 3 sentences maximum before wrapping in `ComboReasoning`

#### `BedrockPreferenceInterpretationService`

Implements `PreferenceInterpretationService`. Uses `BedrockLLMClient`.

**Prompt strategy:**
- System prompt: Friendly personal stylist persona
- User prompt: Provides the structured preference fields (occasions, styles, budget, colors, freeText). Asks for a single natural-language sentence summarising the shopper's intent (e.g., "You're looking for a minimalist, casual beach look between $50–$200 in light tones.")
- Output: Single sentence; wrapped in `PreferenceSummary` alongside the echoed `StylePreferences`

---

## 7. Cross-Cutting Concerns

### 7.1 Authentication Middleware (`AuthenticationMiddleware`)

**Applies to:** All three endpoints — `GET /preferences/options`, `POST /preferences/confirm`, `POST /combos/generate`

**Behaviour:**
- Reads the shopper's session token from the request cookie
- Validates the session with the platform Auth API (or validates a signed session JWT locally, depending on platform auth design)
- If valid: extracts `ShopperSession` and attaches it to the request context for downstream use
- If invalid or missing: immediately returns HTTP 403 with a standard error body; the controller is never reached

**`ShopperSession`** is retrieved from middleware context (not from the request body) in all controllers. Controllers must never trust a shopper identity supplied by the client.

---

### 7.2 Request Logging Middleware (`RequestLoggingMiddleware`)

**Applies to:** All endpoints

**Logs on request arrival:**
- HTTP method and path
- `StyleSessionId` (generated here and attached to context, later used for all downstream log lines)
- Timestamp

**Logs on response dispatch:**
- HTTP status code
- End-to-end latency (ms)
- `StyleSessionId`
- For generate endpoint: result type (`ok` or `fallback`) and number of combos returned

**Sensitive fields masked in logs:**
- Session token value (cookie)
- `freeText` preference field (may contain personal information)

---

### 7.3 Distributed Tracing Middleware (`TracingMiddleware`)

**Applies to:** All endpoints

**Behaviour:**
- Generates a `StyleSessionId` (UUID) per request if not already present in an incoming trace header
- Propagates this ID as a correlation header on all outbound HTTP calls made by ACL implementations (`X-Correlation-ID` or `X-B3-TraceId` depending on platform convention)
- This ensures that a single generate request can be traced across Unit 2 logs and Platform API logs

**`StyleSessionId`** serves double duty as both the domain aggregate identity and the distributed trace correlation ID.

---

### 7.4 Error Handling

#### Validation errors (400 Bad Request)

Returned by controllers when request DTO validation fails. Standard error body:
```
{ "error": "VALIDATION_ERROR", "detail": "<field-level message>" }
```

Examples: budget max ≤ min; a color in both preferred and excluded; unrecognised occasion or style value.

---

#### Authentication failure (403 Forbidden)

Returned by `AuthenticationMiddleware` when the session is absent or invalid.
```
{ "error": "UNAUTHENTICATED" }
```

---

#### Wishlist unavailable (502 Bad Gateway)

Raised when `HttpWishlistRepository` cannot reach Unit 2 or receives a 5xx response. The use case catches `WishlistUnavailableException` and the controller returns:
```
{ "error": "DEPENDENCY_UNAVAILABLE", "detail": "Wishlist service is temporarily unavailable." }
```

This is not a graceful degradation — the wishlist is required for combo generation. The error surfaces to the client.

---

#### AI inference failure (503 Service Unavailable)

Raised when a Bedrock call fails after one retry. The use case catches `AIInferenceException` and the controller returns:
```
{ "error": "AI_UNAVAILABLE", "detail": "Styling engine is temporarily unavailable. Please try again." }
```

---

#### Catalog / Complete-the-Look fetch failure (graceful degradation)

`HttpProductCatalogRepository` and `HttpCompleteLookRepository` return empty lists on failure rather than throwing. The pipeline continues with wishlist items only. The degradation is logged with the `StyleSessionId` for monitoring, but is transparent to the client.

---

#### Unhandled exceptions (500 Internal Server Error)

A global error handler catches any unhandled exception, logs the full stack trace with `StyleSessionId`, and returns:
```
{ "error": "INTERNAL_ERROR" }
```

Sensitive data (preferences, session tokens) is never included in error responses.

---

## 8. Component Interaction Diagrams

### 8.1 `GET /api/v1/style/preferences/options`

```
Client
  │
  ▼ GET /api/v1/style/preferences/options
AuthenticationMiddleware → validates session
TracingMiddleware        → attaches StyleSessionId
RequestLoggingMiddleware → logs arrival
  │
  ▼
StylePreferencesController
  │
  ▼ delegates to
GetPreferenceOptionsUseCase
  │ returns StyleOptionsCatalogue (static config)
  ▼
StylePreferencesController → maps to PreferenceOptionsResponse
  │
  ▼ HTTP 200
```

---

### 8.2 `POST /api/v1/style/preferences/confirm`

```
Client
  │
  ▼ POST /api/v1/style/preferences/confirm
[Middleware chain]
  │
  ▼
PreferenceConfirmationController
  │ constructs ConfirmPreferencesCommand
  ▼
ConfirmPreferencesUseCase
  │ constructs PreferenceConfirmation aggregate
  │
  ▼
PreferenceInterpretationService.interpret(preferences)
  │
  ▼
BedrockPreferenceInterpretationService
  │ calls AWS Bedrock (Claude)
  │ returns PreferenceSummary
  ▼
PreferenceConfirmation raises PreferencesInterpreted (informational)
  │
  ▼
ConfirmPreferencesUseCase returns PreferenceSummary
  │
  ▼
PreferenceConfirmationController → maps to PreferenceConfirmationResponse
  │
  ▼ HTTP 200
```

---

### 8.3 `POST /api/v1/style/combos/generate` (success path)

```
Client
  │
  ▼ POST /api/v1/style/combos/generate
[Middleware chain]
  │
  ▼
ComboGenerationController
  │ constructs GenerateCombosCommand
  ▼
GenerateCombosUseCase
  │
  ├─ constructs StyleSession
  │    └─ raises ComboGenerationRequested
  │         └─ PreferenceDefaultPolicy: sets quickGenerate flag if needed
  │
  ├─ WishlistRepository.fetchForSession()
  │    └─ HttpWishlistRepository → Unit 2 GET /api/v1/wishlist
  │
  ├─ StyleSession.loadWishlist(snapshot)
  │    └─ raises WishlistFetchCompleted
  │         └─ WishlistSupplementationPolicy: evaluates sufficiency
  │              └─ [if insufficient] raises CatalogSupplementationRequired
  │
  ├─ [if CatalogSupplementationRequired]
  │    ├─ CompleteLookRepository.fetchCompleteLookSignals() per wishlist item
  │    │    └─ HttpCompleteLookRepository → Platform /v1/recommendation/completethelook/{sku}
  │    ├─ ProductCatalogRepository.searchSupplementaryItems()
  │    │    └─ HttpProductCatalogRepository → Platform /v1/products/list
  │    └─ StyleSession.loadCatalogItems(merged signals)
  │         └─ raises CatalogItemsFetched
  │
  ├─ ComboCompatibilityScoringService.score(scoringInput)
  │    └─ BedrockComboCompatibilityScoringService → AWS Bedrock (Claude)
  │         returns ScoringResult.Combos
  │
  ├─ [parallel] ComboReasoningGenerationService.generateReasoning() per combo
  │    └─ BedrockComboReasoningGenerationService → AWS Bedrock (Claude) per combo
  │
  ├─ StyleSession.completeCombos(combos)
  │    └─ raises CombosGenerated
  │         └─ ComboExclusionPolicy: filters excluded IDs; retries once if needed
  │
  └─ returns ComboGenerationResult.Success
       │
       ▼
ComboGenerationController → maps to ComboGenerationSuccessResponse
  │
  ▼ HTTP 200 { "status": "ok", "combos": [...] }
```

---

### 8.4 `POST /api/v1/style/combos/generate` (fallback path)

```
...same as 8.3 up through scoring...
  │
  ├─ ComboCompatibilityScoringService.score(scoringInput)
  │    └─ BedrockComboCompatibilityScoringService → AWS Bedrock (Claude)
  │         returns ScoringResult.Fallback(alternatives)
  │
  ├─ StyleSession.triggerFallback(alternatives)
  │    └─ raises FallbackTriggered
  │         └─ FallbackPolicy: validates alternatives non-empty
  │
  └─ returns ComboGenerationResult.Fallback
       │
       ▼
ComboGenerationController → maps to ComboGenerationFallbackResponse
  │
  ▼ HTTP 200 { "status": "fallback", "message": "...", "alternatives": [...] }
```
