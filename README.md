# DeckForge

A card-game engine API and real-time frontend.  
Backend: Go (Gin, GORM, SQLite) · Frontend: React 18 + TypeScript + Vite

> **GoTo take-home assignment** — Senior Full-Stack Developer, WebVoice team (R26-1745)

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

- Go 1.23+ — [go.dev/dl](https://go.dev/dl/)
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

## Environment variables

### Backend (`backend/.env`)

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP server port |
| `DATABASE_URL` | `deckforge.db` | SQLite file path |
| `JWT_SECRET` | *(required)* | HMAC signing secret |
| `JWT_EXPIRY` | `24h` | Token lifetime |
| `CORS_ORIGIN` | `http://localhost:5173` | Allowed frontend origin |
| `MIN_PLAYERS` | `2` | Default min players per game |
| `MAX_PLAYERS` | `8` | Default max players per game |
| `ADMIN_SEED_USERNAME` | *(empty)* | Creates admin user on first boot |
| `ENV` | `development` | Set to `production` for JSON logs |

### Frontend (`frontend/.env`)

| Variable | Default | Description |
|---|---|---|
| `VITE_API_URL` | `http://localhost:8080` | Backend API base URL |

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

9 tests cover: Fisher-Yates shuffle correctness, deck uniqueness, card numeric values,
game state machine transitions.

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
- **Containers:** Phase 5 adds Docker multi-stage build (~8 MB Alpine image) and
  GitHub Actions CI/CD → GHCR. See [ROADMAP.md](ROADMAP.md).

---

## Roadmap

This project was scoped for 10 days. Phase 1 (this submission) covers the full assignment spec.
See [ROADMAP.md](ROADMAP.md) for the complete plan:
- Phase 2: WebSocket real-time (goroutine hub, turn-based Draw/Accept mechanic)
- Phase 3: SVG card graphics, table layout animations
- Phase 4: Integration tests, Swagger/OpenAPI docs
- Phase 5: Docker, GitHub Actions CI/CD, PostgreSQL
- Phase 6: Redis pub/sub, OIDC, Prometheus metrics, Kubernetes manifests

---

## AI usage disclosure

This project was built with [Claude Code](https://claude.com/claude-code) (Anthropic) as an
AI pair-programming assistant.

**How AI was used:**
- Architecture planning: collaborative design sessions covering data model, role model, API
  contract, and Clean Architecture layer boundaries — all decisions documented in
  [DECISIONS.md](DECISIONS.md) and [ARCHITECTURE.md](ARCHITECTURE.md)
- Code generation: Go domain models, GORM repositories, Gin handlers, React components, and
  Zustand stores were generated iteratively with AI assistance
- Code review: AI reviewed each commit for correctness, security (input validation, auth gating,
  CORS), and adherence to the architecture

**What I drove:**
- All architectural decisions (language choice, role model, game rules, state machine, storage)
- Code review and validation of every generated file before committing
- Design of the Draw/Accept game mechanic (Phase 2)
- All commit messages and documentation

**Why this is on-brand for GoTo:** The job description explicitly lists AI-assisted development
as a hiring criterion. This project demonstrates that approach in practice.

---

## Key design decisions

See [DECISIONS.md](DECISIONS.md) for all 12 ADRs with full context, alternatives, and tradeoffs.
Notable decisions:

- **Go over Java:** GoTo's VoIP domain, goroutine model for WebSockets, 8 MB binary vs 200 MB JVM
- **SQLite default:** Zero setup for the reviewer; PostgreSQL swap = one env var
- **Dealer is contextual:** `dealer` is a relationship (game.dealerUserId == currentUser.id),
  not a JWT claim — prevents global privilege escalation
- **No state field on ShoeCard:** `held_by_player_id IS NULL` is the undealt state
- **POST /decks returns UUID only:** A deck is a batch of 52 inserts, not a persisted entity
