# Unit 3 — AI Styling Engine: Local Development Guide

## Prerequisites

| Tool | Version | Install |
|---|---|---|
| Go | 1.21+ | https://go.dev/dl/ |
| Node.js | 18+ | https://nodejs.org |

---

## Backend (Go + Go-Chi)

### 1. Install dependencies

```bash
cd backend
go mod tidy
```

This generates `go.sum` and downloads `chi` and `uuid`.

### 2. Run the server

```bash
go run main.go
```

Server starts on **http://localhost:8080**.

### Local auth

The `AuthMiddleware` accepts **any non-empty cookie** named `session` as valid.
Set a test cookie in your browser or with curl:

```bash
# curl — set a test session cookie
curl -s -b "session=test-token" http://localhost:8080/api/v1/style/preferences/options
```

### Test the endpoints

```bash
# 1. Get preference options
curl -s -b "session=test" http://localhost:8080/api/v1/style/preferences/options | jq

# 2. Confirm preferences
curl -s -b "session=test" -X POST http://localhost:8080/api/v1/style/preferences/confirm \
  -H "Content-Type: application/json" \
  -d '{"occasions":["casual","beach"],"styles":["minimalist"],"budget":{"min":50,"max":200},"colors":{"preferred":["beige","white"],"excluded":["black"]},"freeText":"Something light for summer"}' | jq

# 3. Generate combos (with preferences)
curl -s -b "session=test" -X POST http://localhost:8080/api/v1/style/combos/generate \
  -H "Content-Type: application/json" \
  -d '{"preferences":{"occasions":["casual"],"styles":["minimalist"]}}' | jq

# 4. Quick-generate (no preferences)
curl -s -b "session=test" -X POST http://localhost:8080/api/v1/style/combos/generate \
  -H "Content-Type: application/json" \
  -d '{}' | jq

# 5. Generate with excluded combo IDs
curl -s -b "session=test" -X POST http://localhost:8080/api/v1/style/combos/generate \
  -H "Content-Type: application/json" \
  -d '{"excludeComboIds":["combo-CFG-BLAZER-BLK-CFG-TROUSERS-BGE"]}' | jq
```

---

## Frontend (React + Vite)

### 1. Install dependencies

```bash
cd frontend
npm install
```

### 2. Run the dev server

```bash
npm run dev
```

Dev server starts on **http://localhost:5173**.

The Vite dev server proxies all `/api/*` requests to `http://localhost:8080`, so the backend must be running first.

### 3. Set a session cookie

The backend requires a `session` cookie. Set one in your browser's DevTools before using the app:

```
Application → Cookies → http://localhost:5173 → Name: session, Value: test-token
```

Or run this in the browser console:

```js
document.cookie = "session=test-token; path=/"
```

### 4. Navigate the flow

```
http://localhost:5173/style/preferences  → Preference Input Screen
                                         → click "Surprise me" to quick-generate
                                         → click "Generate Combo →" to go through confirmation
http://localhost:5173/style/confirm      → AI Summary Confirmation Screen
http://localhost:5173/style/combos       → Combo Suggestion Page (success or fallback)
```

---

## Project Structure

```
src/
├── backend/          Go service (DDD layered architecture)
│   ├── domain/       Value objects, entities, aggregates, events, policies, services, repositories
│   ├── infrastructure/ ACL stubs, mock AI services, event dispatcher
│   ├── application/  Use cases and commands
│   ├── api/          Controllers, middleware, DTOs, router
│   └── main.go       Entry point — manual dependency wiring
└── frontend/         React + Vite app
    └── src/
        ├── api/      Typed API client (styleApi.js)
        └── screens/  PreferenceInputScreen, ConfirmationScreen, ComboSuggestionScreen
```

---

## Replacing stubs for production

| Stub | File | Replace with |
|---|---|---|
| Wishlist repository | `infrastructure/acl/in_memory_wishlist_repository.go` | `HttpWishlistRepository` calling Unit 2 |
| Product catalog repository | `infrastructure/acl/in_memory_product_catalog_repository.go` | `HttpProductCatalogRepository` calling Platform API |
| Complete-the-Look repository | `infrastructure/acl/in_memory_complete_look_repository.go` | `HttpCompleteLookRepository` calling Platform API |
| Scoring service | `infrastructure/ai/mock_combo_compatibility_scoring_service.go` | `BedrockComboCompatibilityScoringService` |
| Reasoning service | `infrastructure/ai/mock_combo_reasoning_generation_service.go` | `BedrockComboReasoningGenerationService` |
| Interpretation service | `infrastructure/ai/mock_preference_interpretation_service.go` | `BedrockPreferenceInterpretationService` |
| Auth middleware | `api/middleware/auth_middleware.go` | JWT validation with platform signing key |
