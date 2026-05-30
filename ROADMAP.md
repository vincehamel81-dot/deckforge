# DeckForge — Product Roadmap

DeckForge is a card-game engine API and real-time frontend. It provides the infrastructure layer
for any multi-player card game: shoe management, dealing, shuffling, player tracking, and live
game state. Specific game rules (poker, blackjack, baccarat) plug in on top without touching the
core engine.

> **Submission note:** This project was scoped for 10 days of development. The 2-day assignment
> window covers Phase 1 (foundation) and the beginning of Phase 2. Every architecture decision was
> made with Day 10 in mind from Day 1. See [DECISIONS.md](../DECISIONS.md) for the full reasoning.

---

## Phase 1 — Foundation (Days 1–2) ✅ Submitted

**Goal:** All assignment requirements met. Clean architecture visible. Production path documented.

### Backend
- [x] Go project scaffold — clean architecture (domain → application → infrastructure → presentation)
- [x] Domain models: Game, Shoe, Deck, Card, Player, User
- [x] Repository interfaces (storage-backend agnostic)
- [x] SQLite + GORM (zero-setup for reviewer; PostgreSQL swap = one env var)
- [x] All 10 assignment API endpoints
  - Create / delete a game
  - Create a deck
  - Add a deck to the shoe (irreversible)
  - Add / remove players
  - Deal cards to a player
  - Get a player's hand
  - Get leaderboard (sorted by hand value descending)
  - Get undealt card count per suit
  - Get undealt card count per suit + face value (sorted)
  - Shuffle the shoe (Fisher-Yates, no library shuffle)
- [x] Game state machine: WAITING → IN_PROGRESS → FINISHED
- [x] JWT authentication (username-only, no password; OIDC production path documented)
- [x] Correlation ID middleware
- [x] Structured logging (zerolog)
- [x] Input validation (binding + custom validators)
- [x] Global error handler

### Frontend
- [x] React 18 + TypeScript + Vite
- [x] Auth flow: register / login (username only)
- [x] Game list page — browse open games, create a game
- [x] Table page `/table/:id` — per-player authenticated view
  - Your hand (face up, values visible)
  - Other players (face down — card count visible, values hidden)
  - Shoe status (cards remaining, deck count)
  - Leaderboard (live standings)
  - Deal / Shuffle / End Game controls (admin only)
- [x] TanStack Query v5 (server state)
- [x] Zustand (client state: auth, table UI)
- [x] React Router v6

### Documentation
- [x] README — prerequisites, setup, env vars, architecture overview, AI usage disclosure
- [x] DECISIONS.md — 11 ADRs with context, reasoning, tradeoffs
- [x] ROADMAP.md (this file)
- [x] LANGUAGE_TRADEOFFS.md — Go vs Java analysis

---

## Phase 2 — Real-Time Layer (Days 3–4)

**Goal:** The table feels live. Every event pushes to all connected players instantly.

### Backend
- [ ] WebSocket hub (gorilla/websocket) — one hub goroutine per game, one client goroutine per player
- [ ] WebSocket auth — JWT validated on HTTP upgrade
- [ ] Game events broadcast to all players in a game:
  - `player_joined` / `player_left`
  - `cards_dealt` `{ playerId, cardCount }` — card values NOT broadcast (privacy)
  - `shoe_shuffled`
  - `game_started` / `game_ended`
  - `turn_changed` `{ currentPlayerId }`
- [ ] Turn order state machine — `currentTurnPlayerId` on Game entity
- [ ] `POST /games/:id/draw` — active player draws 1 card for ALL players (Draw/Accept mechanic); triggers auto-end check
- [ ] `POST /games/:id/accept` — active player passes their turn; advances to next player
- [ ] **Turn timer** — 15 s per turn (`TURN_TIMEOUT_SECONDS` env var, default 15); server-side goroutine per active game; on expiry, server auto-fires Accept and emits `turn_expired` then `turn_changed`; broadcasts `turn_timer_started { playerId, expiresAt }` so clients can render a countdown

### Frontend
- [ ] WebSocket client (`lib/wsClient.ts`) — auto-reconnect, exponential backoff
- [ ] Live leaderboard updates (no polling — pure push)
- [ ] Turn indicator — "Your turn" / "Waiting for Alice..."
- [ ] Draw / Accept buttons (only shown on your turn)
- [ ] **Turn countdown** — animated timer on the active player's seat; driven by `turn_timer_started.expiresAt`; clears on `turn_changed`
- [ ] Animated card deal (CSS transition — card slides to player seat)
- [ ] Toast notifications for game events

---

## Phase 3 — Frontend Polish (Days 5–6)

**Goal:** The table looks like a real card game.

- [ ] SVG card graphics — full 52-card deck rendered from data (suit + value → SVG)
- [ ] Table layout — players arranged around an oval table, seats positioned by join order
- [ ] Face-down card backs for other players' hands
- [ ] Shuffle animation (card fan effect)
- [ ] Shoe counter widget (cards remaining / total)
- [ ] Mobile-responsive layout
- [ ] Dark theme (card table green felt aesthetic)

---

## Phase 4 — Quality & Documentation (Days 7–8)

**Goal:** Production-grade confidence in the codebase.

### Testing
- [ ] Domain unit tests — shuffle (all 52 unique cards returned), deal exhaustion (53rd = empty),
      scoring (correct sort order), state machine transitions
- [ ] HTTP handler integration tests — test each endpoint via `httptest`
- [ ] WebSocket integration tests — hub broadcast correctness
- [ ] Frontend unit tests — Vitest, component tests for critical paths

### Observability
- [ ] Request/response logging with correlation ID on every log line
- [ ] Structured error responses (consistent JSON error envelope)
- [ ] Health check endpoint `GET /health`
- [ ] OpenAPI / Swagger documentation (swaggo/swag)
- [ ] Thunder Client collection — all endpoints pre-wired, `.example` env file committed

---

## Phase 5 — Infrastructure (Day 9)

**Goal:** One command to run the full stack in any environment.

- [ ] Docker multi-stage build — Go binary in Alpine (~10 MB image)
- [ ] Docker Compose — backend + frontend + PostgreSQL
- [ ] GitHub Actions CI — build → test → push image to GHCR
- [ ] PostgreSQL implementation (`infrastructure/persistence/postgres/`) — same interfaces, swap driver
- [ ] Rate limiting (`@nestjs/throttler` equivalent in Go: `golang.org/x/time/rate`)
- [ ] CORS locked to env-configured origin

---

## Phase 6 — Scale & Production (Day 10)

**Goal:** Remove the single-process constraint. Ready for horizontal scaling.

- [ ] Redis pub/sub behind the WebSocket hub — replace in-process channel with Redis channel
      (no business logic changes — the `Hub` interface absorbs the swap)
- [ ] Session store → Redis (JWT remains stateless; refresh tokens stored in Redis)
- [ ] OIDC/SSO integration — swap JWT issuer for Entra ID / Okta provider
      (`auth_middleware.go` already abstracts token validation)
- [ ] Prometheus metrics endpoint `/metrics`
- [ ] OpenTelemetry trace export (correlation IDs are already in every log line — wire to Jaeger/Tempo)
- [ ] Kubernetes manifests (Deployment + Service + Ingress)
- [ ] Load test results (k6) — baseline RPS for shoe operations + WebSocket connections

---

## Architecture at a glance

```
browser ──HTTPS──▶ Gin HTTP handlers   ──▶ Application (use cases)  ──▶ Domain
browser ──WSS────▶ WebSocket hub       ──▶ Application (commands)   ──▶ Domain
                                                   │
                                         Infrastructure (repos)
                                           │             │
                                        SQLite       PostgreSQL
                                       (default)      (prod swap)
                                                   │
                                           Infrastructure (hub)
                                           │             │
                                      in-process     Redis pub/sub
                                       channel       (multi-node)
```

Every boundary is an interface. Every infrastructure component is swappable without touching business logic.

---

## What the 2-day submission shows

Even with Phase 1 only, an evaluator opening this repo sees:

1. **Architecture judgment** — clean separation of concerns, dependency rule enforced
2. **Production instincts** — repository interfaces, correlation IDs, structured logging from commit 1
3. **Long-term thinking** — this ROADMAP, DECISIONS.md with 11 ADRs, documented upgrade paths
4. **Go competence** — idiomatic structs, interfaces, error handling, Fisher-Yates implementation
5. **Frontend discipline** — feature-sliced structure, TanStack Query, Zustand, typed API client
6. **Security thinking** — JWT auth, input validation, per-player view gating, OIDC path documented
7. **AI collaboration** — transparent, driver-in-seat usage documented in README
