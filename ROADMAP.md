# DeckForge — Roadmap

DeckForge is a card-game engine: REST API + real-time WebSocket backend (Go) and a React TypeScript
frontend. It provides the infrastructure layer for any multi-player card game. Specific game rules
(poker, blackjack, baccarat) plug in on top without touching the core engine.

---

## Phase 1 — Delivered ✅

Everything below was built and shipped within the 2-day assignment window.

### Backend
- All 10 assignment API endpoints (Go · Gin · GORM)
- Clean Architecture — domain / application / infrastructure / presentation layers
- SQLite by default (zero setup); PostgreSQL via one env var (`DB_DRIVER=postgres`)
- JWT authentication — username-only for demo; OIDC/Entra ID upgrade path documented
- Three-tier identity: `user` (JWT claim) · `dealer` (contextual per-game DB lookup) · `admin` (JWT claim, seeded via env)
- Fisher-Yates shuffle (no library shuffle; crypto/rand seed)
- Game state machine: `WAITING → IN_PROGRESS → FINISHED`
- Auto-end when shoe cannot serve a full round (`remainingCards < activePlayerCount`)
- Auto-end when player count drops below configured minimum (e.g. admin kicks a player)
- Player removal returns cards to shoe
- Catch-up deal for players joining a game already in progress
- Join gated when remaining cards cannot cover a full catch-up hand
- WebSocket hub (gorilla/websocket) — one goroutine per game, one per connected player
- Six push event types: `game_started`, `game_ended`, `cards_dealt`, `player_joined`, `player_left`, `shoe_shuffled`
- Player disconnect detection — 30s grace period (`DISCONNECT_TIMEOUT_SECONDS`); auto-remove on expiry; reconnect within window cancels timer
- Structured JSON logging (zerolog) with correlation ID middleware on every request
- Swagger/OpenAPI — all endpoints annotated; UI at `/swagger/index.html`
- Health check: `GET /health`
- Debug observability endpoints (`/debug/error`, `/debug/warn`) — zerolog demo, non-production only
- Docker multi-stage build (~10 MB Alpine image)
- Docker Compose — SQLite default; `--profile postgres` for PostgreSQL sidecar
- GitHub Actions CI — backend tests + frontend TypeScript check on every push; image pushed to GHCR on `main`

### Frontend
- React 18 + TypeScript + Vite — feature-sliced structure
- All 10 assignment operations have a visual representation
- TanStack Query v5 — server state; WebSocket events as primary invalidation, 15s polling as fallback
- Zustand — auth state (JWT token + user)
- Real-time table: live leaderboard, shoe status, hand — all driven by socket events
- Lobby: create table, join table, admin delete, orphan detection
- Table view: dealer controls (deal, shuffle, add deck, start/end game), player hand, leaderboard sidebar, suit counts
- i18n: `i18next` + `react-i18next`; en-US (canonical), fr-CA (full), es-MX (core); namespace cascade with sessionStorage merge cache; TypeScript key safety
- Three locales, three namespaces, O(1) `t()` lookups after startup merge
- Admin kick with auto-close when player count drops below minimum
- Kicked player receives "removed from table" notification and returns to lobby
- Disconnected player (browser closed) auto-removed after 30s grace period
- Deployed: Railway (backend + frontend) with CI/CD on every push to `main`

### Tests
- 6 backend integration tests against real in-memory SQLite (no mocks)
- Covers all ARCHITECTURE.md invariants: 52 unique cards, 53rd deal blocked, player removal returns cards, `remainingCards` never negative, auto-end threshold (`<` not `≤`), leaderboard sort + tie-break, decks sealed after start, FINISHED rejects mutations

---

## Phase 2 — What's Next

Features scoped and designed but not built within the 2-day window. Architecture was built from
day one to support these without rework.

### Game mechanics
- **Turn-based Draw/Accept** — clockwise turn order; Draw deals 1 card to ALL players (strategic tension); Accept passes; 15s server-side timer with auto-accept on expiry
- **In-game chat** — ephemeral messages over the existing WebSocket hub
- **Save game results** — `completed_games` table: winner, draw winners, all player IDs, final scores
- **Full player history** — list all players (active + left) with join/leave timestamps per game

### Backend
- **Per-user language preference** — `locale` column on `users` table; `PATCH /users/me`; read at login before first render
- **Rate limiting** — `golang.org/x/time/rate` middleware on auth and public endpoints
- **Redis pub/sub** — replace in-process WebSocket hub for horizontal scaling (Hub interface already abstracted)
- **OIDC/SSO** — swap HMAC JWT validator for Entra ID / Okta; one function in `auth_middleware.go`
- **PostgreSQL as default** — already works via env var; make it the production-documented default

### Frontend
- **Tailwind CSS** — replace inline styles; responsive table layout for mobile and tablet
- **SVG card graphics** — full 52-card deck rendered from data (suit + value → SVG face)
- **Turn indicator + countdown** — animated timer on active player's seat
- **Auto-reconnect** — exponential backoff on WebSocket drop (currently falls back to 15s polling)
- **Frontend tests** — Vitest component tests for Leaderboard, DealerControls, GameResult

### Infrastructure
- **GitHub Actions → Azure Container Registry** — production image pipeline
- **Azure Container Apps** — swap Railway for production-grade hosting aligned with GoTo's Azure stack
- **Prometheus + OpenTelemetry** — metrics endpoint; trace export to Jaeger/Tempo
- **Load test** — k6 baseline for shoe operations and WebSocket connections at scale
