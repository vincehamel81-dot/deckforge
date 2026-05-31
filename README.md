# DeckForge

A card-game engine API and real-time frontend.  
Backend: Go (Gin, GORM, SQLite) · Frontend: React 18 + TypeScript + Vite

> **GoTo take-home assignment** — Senior Full-Stack Developer, WebVoice team (R26-1745)

---

## Live demo

- **Frontend:** https://deckforge-production.up.railway.app
- **Backend API:** https://content-optimism-production-bdb5.up.railway.app
- **Swagger UI:** https://content-optimism-production-bdb5.up.railway.app/swagger/index.html

---

## What it does

DeckForge manages poker-style card games. It provides the infrastructure for any multi-player card
game: shoe management, dealing, shuffling, player tracking, and scoring. The specific game rules
(poker hand evaluation, betting, blackjack strategy) are explicitly a layer above this engine.

**Implemented operations** (all 10 from the assignment spec):
- Create and delete a game
- Create a deck, add a deck to the game shoe (irreversible)
- Add and remove players from a game
- Deal cards to a player
- Get a player's hand
- Get leaderboard (players sorted by hand value descending)
- Get undealt card count per suit
- Get undealt card count per suit + face value (sorted)
- Shuffle the shoe (Fisher-Yates, no library shuffle)

---

## Prerequisites

- Go 1.26+ — [go.dev/dl](https://go.dev/dl/)
- Node.js 20+ — [nodejs.org](https://nodejs.org/)

No database setup required. SQLite is embedded (zero dependencies).

---

## Quickstart

```bash
# 1. Clone
git clone https://github.com/vincehamel81-dot/deckforge.git
cd deckforge

# 2. Backend
cd backend
cp .env.example .env
# Edit .env and set JWT_SECRET to any secret string
go run ./cmd/server
# → Server running on http://localhost:8080

# 3. Frontend (new terminal)
cd ../frontend
cp .env.example .env
npm install
npm run dev
# → App running on http://localhost:5173
```

Open [http://localhost:5173](http://localhost:5173) in two tabs to simulate two players.

---

## Running with Docker

```bash
# Copy and configure
cp backend/.env.example backend/.env
# Edit backend/.env — set JWT_SECRET to any secret string

# Build and run (SQLite, zero dependencies)
docker compose up --build

# With PostgreSQL instead (one env var swap)
DB_DRIVER=postgres \
DATABASE_URL="postgres://deckforge:deckforge@db:5432/deckforge?sslmode=disable" \
docker compose --profile postgres up --build
```

The backend image is ~10 MB (Alpine + statically linked Go binary). The frontend
has its own Dockerfile (Nginx) for containerised deployment — used in production on Railway.

---

## Environments

| | Local (native) | Docker Compose | Production (Railway) |
|---|---|---|---|
| Backend | `go run ./cmd/server` | `docker compose up` | Docker container (Railway) |
| Frontend | `npm run dev` (Vite HMR) | Run natively alongside Docker | Nginx container (Railway) |
| Database | SQLite (embedded) | SQLite or PostgreSQL (`--profile postgres`) | SQLite / PostgreSQL add-on |
| API base URL | `http://localhost:8080` | `http://localhost:8080` | `https://content-optimism-production-bdb5.up.railway.app` |
| Auth | JWT (dev secret) | JWT (dev secret) | JWT → OIDC (Entra ID, production path) |
| Logs | ConsoleWriter colored text (`LOG_FORMAT=text`) | Container stdout | JSON (`LOG_FORMAT=json`) |

**Frontend per-environment:** `VITE_API_URL` in a `.env` file. Vite resolves in order:
`.env.local` → `.env.development` / `.env.production` → `.env`. For production builds the URL
is inlined into the bundle at `npm run build` time — set it before building.

---

## Environment variables

### Backend (`backend/.env`)

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP server port |
| `DB_DRIVER` | `sqlite` | Storage backend: `sqlite` or `postgres` |
| `DATABASE_URL` | `deckforge.db` | SQLite file path or PostgreSQL DSN |
| `JWT_SECRET` | *(required)* | HMAC signing secret |
| `JWT_EXPIRY` | `24h` | Token lifetime |
| `CORS_ORIGIN` | `http://localhost:5173` | Allowed frontend origin |
| `MIN_PLAYERS` | `2` | Default min players per game |
| `MAX_PLAYERS` | `8` | Default max players per game |
| `ADMIN_SEED_USERNAMES` | *(empty)* | Comma-separated list of usernames to seed as admins on first boot (e.g. `admin,ops`) |
| `AUTO_END_GAME` | `true` | When `true`, game ends automatically when shoe cannot serve a full round or player count drops below minimum |
| `DISCONNECT_TIMEOUT_SECONDS` | `30` | Grace period before a player with a closed WebSocket is auto-removed from the game |
| `LOG_LEVEL` | `info` | Minimum log severity: `debug` `info` `warn` `error`. Change + restart — no deploy needed |
| `LOG_FORMAT` | `json` | `text` for coloured terminal (local dev); `json` for log aggregators (production) |
| `DEBUG_ENABLED` | `false` | Registers `/debug/error` and `/debug/warn` diagnostic endpoints |
| `DEBUG_TOKEN` | *(empty)* | When set, all `/debug/*` requests require `X-Debug-Token: <value>` header |

### Frontend (`frontend/.env`)

| Variable | Default | Description |
|---|---|---|
| `VITE_API_URL` | `http://localhost:8080` | Backend API base URL |

---

## Observability

Structured JSON logging is built in from the first request. Every log entry carries a
`correlationId` that ties together all log lines for a single HTTP request — the same
pattern as `Operation.Id` in Azure Application Insights or a trace ID in OpenTelemetry.

The local dev `.env` defaults enable two diagnostic endpoints to verify the logging
pipeline is working end-to-end without waiting for a real error:

```bash
# With the backend running locally (DEBUG_ENABLED=true is set in .env by default):

GET http://localhost:8080/debug/error   # triggers log.Error() with correlation ID
GET http://localhost:8080/debug/warn    # triggers log.Warn() with correlation ID
```

Terminal output (with `LOG_FORMAT=text`):
```
ERR simulated application error — triggered via /debug/error component=debug correlationId=b3d1a...
```

The same entry in production (`LOG_FORMAT=json`) becomes a structured record that any
log aggregator (Datadog, Loki, Railway) can index, alert on, and correlate with the
HTTP access log entry from the request that triggered it.

---

## Frontend design notes

### Undealt card count — manual refresh

The "Undealt cards per suit" panel in the table view shows a **manually-refreshed snapshot**
of `GET /games/:id/shoe/suits`, not a live auto-updating feed.

This is intentional: it demonstrates an explicit REST API call on demand. The `↻` button
(visible only after cards are dealt or the shoe is shuffled) lets the user trigger the request
themselves and see the result update in real time. Hovering over the button shows a tooltip
explaining this.

The live suit-count strip above it (♠ ♥ ♣ ♦ with totals) auto-updates via WebSocket events —
those two views together demonstrate both real-time push and on-demand REST in the same panel.

---

## Architecture

See [ARCHITECTURE.md](ARCHITECTURE.md) for the full data model, API contract, role model, and
folder structure.

**Clean architecture layers:**
```
Presentation (Gin handlers)
    ↓
Application (use cases — no framework deps)
    ↓
Domain (entities + repository interfaces — pure Go)
    ↑
Infrastructure (GORM repos, JWT, zerolog)
```

**Role model:**
- `user` — registered player (JWT claim)
- `dealer` — game creator; contextual, not in JWT; checked per-request from DB
- `admin` — platform operator (JWT claim); see [SETUP.md](SETUP.md) for creation

---

## Running tests

```bash
cd backend
go test ./...
```

17 tests across two layers:
- **Domain** (9): Fisher-Yates shuffle correctness, deck uniqueness, card numeric values, game state machine transitions
- **Application integration** (8): all key invariants from `ARCHITECTURE.md` — 52-unique-card deal, 53rd deal blocked after shoe exhaustion, player removal returns cards to shoe, `remainingCards` never negative, auto-end fires on `remaining < activeCount` (not `≤`), decks sealed after game starts, FINISHED game rejects deal/shuffle/join, leaderboard sorted descending with seat-order tie-break

---

## API documentation (Swagger UI)

With the backend running, open [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html).

All endpoints are documented with request/response schemas and a "Try it out" button. Use the **Authorize** button (top right) to paste a JWT and test protected routes directly from the browser.

**Importing into external tools:** The raw OpenAPI spec is served at
`http://localhost:8080/swagger/doc.json`. Import this URL directly into Postman, Insomnia, or
any OpenAPI-compatible client to get a fully-documented, runnable collection with no manual
setup. Useful for scripted testing or sharing the API with teammates who prefer a GUI client.

---

## API endpoints

```
POST   /auth/register          → register with username, receive JWT
POST   /auth/login             → login, receive JWT
GET    /health                 → health check

POST   /decks                  → generate deck UUID (no DB write)
POST   /games                  → create game
GET    /games                  → list games (filter by ?status=)
GET    /games/:id              → game detail + shoe status
DELETE /games/:id              → delete game (dealer only)
POST   /games/:id/start        → start game, optional initial deal
POST   /games/:id/end          → end game (dealer only)

POST   /games/:id/shoe/decks   → add deck to shoe (dealer, WAITING only)
POST   /games/:id/shoe/shuffle → shuffle undealt cards (dealer)
GET    /games/:id/shoe/suits   → undealt count per suit
GET    /games/:id/shoe/cards   → undealt count per card (sorted)

POST   /games/:id/players          → join game
DELETE /games/:id/players/:pid     → leave game
GET    /games/:id/players          → leaderboard (sorted by hand value)
POST   /games/:id/players/:pid/deal → deal N cards (dealer only)
GET    /games/:id/players/:pid/hand → player's hand (self or admin)
```

---

## Production path

- **Database:** Change `DB_DRIVER=sqlite` to `DB_DRIVER=postgres` and set `DATABASE_URL` to a
  PostgreSQL DSN. The Repository interface abstracts storage — no code changes needed.
- **Auth:** Swap the JWT issuer for an OIDC provider (Azure AD / Entra ID, Okta). The
  `auth_middleware.go` validates tokens — replace HMAC validation with OIDC token introspection.
  The `role` claim maps from IdP group membership.
- **Scale:** The WebSocket hub (Phase 2) is behind an interface — replace the in-process channel
  with Redis pub/sub for horizontal scaling.
- **Containers:** Docker multi-stage build is shipped — ~10 MB Alpine image. GitHub Actions CI
  builds and tests on every push. See [ROADMAP.md](ROADMAP.md) for Phase 2 items.

---

## Roadmap

Everything in the assignment spec is delivered. See [ROADMAP.md](ROADMAP.md) for what was
shipped vs. what comes next (turn-based game mechanic, OIDC, Redis pub/sub, mobile layout).

---

## AI usage disclosure

This project was developed with [Claude Code](https://claude.com/claude-code) (Anthropic) as an
AI pair-programming assistant. GoTo listed AI-assisted development as a hiring criterion — I'm
treating that as an invitation to be transparent, not something to hide.

**What the collaboration actually looked like:**

The pattern throughout: I raised a question or constraint, AI proposed an approach, I challenged
or validated it, we iterated. AI accelerated execution. The decisions were mine.

Specific decisions I initiated and drove:

- **Go over Java** — I asked whether Go was justified for GoTo's domain. The case (goroutine model,
  8 MB image, alignment with GoTo's real-time stack) came from that challenge.
- **Language-driven i18n from day one** — I pushed for proper i18n rather than hardcoded English,
  knowing GoTo serves fr-CA in production. Led to the namespace cascade architecture in ADR-002.
- **WebSockets vs polling** — I questioned whether polling was adequate for a real-time product demo.
  We landed on WS as primary with 15s silent fallback — a tradeoff I validated.
- **Feature-sliced frontend structure** — I pushed for component decomposition rather than monolithic
  page files, co-locating hooks, queries, and components per feature area.
- **Two independent auto-end triggers** — I identified shoe exhaustion and player-count drop as
  distinct conditions, each needing its own broadcast reason for the frontend.
- **Multiple admins via CSV env var** — I questioned why only one admin could be seeded and pushed
  for the comma-separated approach with no code changes.
- **Min/max player count by config** — I challenged hardcoded values; moved to per-game params
  with env-var defaults to reflect the "engine, not a game" framing.
- **Auth: passwordless but structurally correct** — I decided on username-only JWT for demo
  friction, but required the OIDC production path to be accurate and specific, not vague.
- **Dealer as contextual, not a JWT claim** — I understood that dealer is a per-game relationship
  and pushed back on encoding it globally in the token, which would break multi-game scenarios.
- **SQLite dev / PostgreSQL production** — I questioned the storage choice and validated that the
  repository interface makes the swap a one-env-var change with no business logic touched.
- **i18n cache busting via build hash** — I challenged the manual version-bump approach and
  recognised it was equivalent to a hardcoded query string. Pushed for the CI-injected git SHA.
- **Observability scope** — I identified LOG_LEVEL and DEBUG_ENABLED as separate, independent
  concerns and questioned whether the framing was over-engineering for the assignment audience.
- **Engine scope boundary** — I questioned whether building a game was over-engineering. That
  conversation produced ADR-014 (no rules engine) as an explicit, defensible boundary.

**Issues I caught during the process:**

- Shoe refresh button appeared for the dealer but not other players — I identified the symptom;
  root cause was a stale WS callback closure; fix replaced it with a `useMemo` derived from data.
- CI broke after a refactor — `setShoeStale` dangling reference, and `tsc --noEmit` was silently
  checking an empty file set instead of the full project graph (`tsc -b` vs `tsc --noEmit`).
- LocaleService test was asserting a side effect (same string value) not the invariant (cascade
  never returns undefined) — I caught this and we rewrote it as a parametric test across all
  locales × namespaces.
- Swagger UI was sending requests to `localhost:8080` on the Railway production URL.
- Questioned `useState` vs `useMemo` for staleness detection — understood and validated the
  derived-state argument before accepting the approach.

**What I didn't do:**

I didn't read every line of every generated file. I reviewed architecture, challenged key
decisions, read the files that mattered (auth flow, WebSocket hub, game lifecycle, i18n cascade),
ran the app, and tested behaviour in both local and production environments. Two days of focused
work with an AI accelerator is still two days of focused work — it covers more ground, not less
thinking.


---

## Key design decisions

See [DECISIONS.md](DECISIONS.md) for all 25 ADRs with full context, alternatives, and tradeoffs.
Notable decisions:

- **Go over Java:** GoTo's VoIP domain, goroutine model for WebSockets, 8 MB binary vs 200 MB JVM
- **SQLite default:** Zero setup for the reviewer; PostgreSQL swap = one env var
- **Dealer is contextual:** `dealer` is a relationship (game.dealerUserId == currentUser.id),
  not a JWT claim — prevents global privilege escalation
- **No state field on ShoeCard:** `held_by_player_id IS NULL` is the undealt state
- **POST /decks returns UUID only:** A deck is a batch of 52 inserts, not a persisted entity

---

## Scope control

The assignment says *"pretend this code will become a foundational part of a new product."*
That framing drove some decisions that go beyond the raw API requirements:

| Addition | Why it's here | Assignment scope? |
|---|---|---|
| JWT authentication | Identifies players; required for per-player hand privacy | Implied by multi-player spec |
| Game state machine (WAITING → IN_PROGRESS → FINISHED) | Enforces sequencing rules (can't deal before start, can't add decks after start) | Implied by "game" semantics |
| Admin role | Platform operator who can delete orphaned games | Reasonable extension, explicitly documented |
| Clean architecture layers | Makes the repository interfaces swappable (SQLite → PostgreSQL → Redis) | Required by "foundational product" framing |
| WebSocket / turn system (Phase 2+) | Documented in ROADMAP, not shipped — mentioned because GoTo is a real-time comms company | Out of scope for this submission |

**What was explicitly NOT added:** real-time WebSocket events, turn-based mechanics, chat,
SVG card graphics, OAuth/OIDC — all documented in [ROADMAP.md](ROADMAP.md) as Phase 2+.
Phase 1 delivers exactly the 10 operations in the spec, with auth and state machine as the
minimum viable foundation for a real product.
