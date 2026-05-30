# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

DeckForge is a card-game engine (shoe management, dealing, shuffling, scoring) built as a technical assignment for GoTo WebVoice. Stack: **Go 1.26 backend (Gin + GORM + gorilla/websocket)** + **React 18 / TypeScript / Vite frontend**. Database is SQLite by default, switchable to PostgreSQL via env var.

---

## Commands

### Backend (`backend/`)

```bash
go run ./cmd/server           # start server (SQLite, reads .env automatically)
go test ./...                 # all tests
go test ./internal/application/commands/...  # single package (adjust path)
go test -run TestDeal_52Unique ./internal/application/commands/  # single test
swag init -g cmd/server/main.go  # regenerate Swagger docs after changing @-annotations
go build ./cmd/server         # compile binary
```

Copy `.env.example` to `.env` before first run. Minimum required: `JWT_SECRET`.

### Frontend (`frontend/`)

```bash
npm run dev      # Vite dev server on :5173
npm run build    # production build
npm run lint     # ESLint
```

Copy `.env.example` to `.env` (sets `VITE_API_URL=http://localhost:8080`).

### Docker

```bash
docker compose up                       # backend + SQLite in container
docker compose --profile postgres up    # with PostgreSQL sidecar
```

Root `.env` drives all `${VAR}` references in `docker-compose.yml`. `JWT_SECRET` must be set there.

---

## Backend architecture

Clean Architecture with a strict inward dependency rule:

```
domain/          ← entities, repository interfaces, domain errors — zero external deps
application/     ← use-case functions (commands mutate, queries read)
infrastructure/  ← GORM repos, WebSocket hub, JWT
presentation/    ← Gin handlers, middleware, router
```

**Wiring:** No DI container. All repos and the WebSocket hub are constructed in `main.go` and injected as interfaces into `NewRouter`, which passes them to each handler constructor.

### Domain layer

- `game/game.go` — `Game` struct with `Start()` / `End()` methods enforcing status transitions; sentinel errors (`ErrAlreadyStarted`, `ErrShoeSealed`, …) returned from domain methods, not handlers.
- `shoe/card.go` — `NewDeck(gameID, positionOffset)` builds 52 cards; `shoe/shuffle.go` does Fisher-Yates in-place on `[]ShoeCard`.
- Repository interfaces live alongside the entity they own (`game/repository.go`, etc.).

### Application layer

Commands are **plain functions** — `func StartGame(cmd StartGameCommand, games game.Repository, …) (*game.Game, error)`. They:
1. Load aggregate via repo
2. Enforce authorization (dealer check)
3. Call domain method
4. Persist via repo

`deal_cards.go` has `checkAutoEnd`: after every deal, if `remaining < activePlayerCount` the game transitions to `FINISHED` automatically (controlled by `AUTO_END_GAME` env var).

### Infrastructure — persistence

`persistence/models.go` defines GORM models separate from domain entities. Each repo file has private `toModel` / `toDomain` mappers — ORM concerns never leak into domain structs.

**ShoeCard state:** `HeldByPlayerID IS NULL` = undealt (in shoe). `HeldByPlayerID IS NOT NULL` = dealt. No separate status column.

SQLite is forced to `MaxOpenConns(1)` + WAL mode + `busy_timeout=5000` to prevent lock contention. PostgreSQL has no such restrictions. `AutoMigrate` runs on every startup.

### Infrastructure — WebSocket

`ws/hub.go` is a single in-process `Hub` (map of `gameID → set<*Client>`) protected by `sync.RWMutex`. `Hub.Broadcast(gameID, Message)` fan-outs to all clients in that game. Each `Client` has a `send` buffered channel (size 32); slow clients are dropped without blocking the broadcaster.

**WS auth quirk:** browsers cannot send `Authorization` headers on WebSocket upgrades. The WS handler (`ws_handler.go`) reads `?token=` from the query string and validates it itself, outside the normal `AuthMiddleware`.

### Middleware chain

`CorrelationMiddleware → LoggerMiddleware → gin.Recovery()` on every request. JWT validation is a separate middleware applied to the `/games` group. `DealerMiddleware` is applied only to dealer-only routes (start, end, shuffle, deal) and short-circuits with 403 if the caller is not the game's dealer.

Auth claims are stored in the Gin context under key `"userClaims"` and retrieved via `middleware.ClaimsFromContext(c)`.

### Config

All config read from env (`config/config.go`). Notable vars:

| Var | Default | Purpose |
|-----|---------|---------|
| `DB_DRIVER` | `sqlite` | `sqlite` or `postgres` |
| `DATABASE_URL` | `deckforge.db` | file path (SQLite) or DSN (postgres) |
| `JWT_SECRET` | *(required)* | HMAC-SHA256 signing key |
| `CORS_ORIGIN` | `http://localhost:5173` | single allowed origin |
| `ADMIN_SEED_USERNAMES` | *(empty)* | CSV of usernames created as admin on startup |
| `AUTO_END_GAME` | `true` | auto-end when shoe can't serve a full round |

### Swagger

Annotations live in handler files (`// @Summary …`). After editing annotations, regenerate with `swag init -g cmd/server/main.go`. UI is at `/swagger/index.html`.

---

## Frontend architecture

### Routing

`main.tsx` registers three routes: `/login` → `LoginPage`, `/` → `GamesPage`, `/table/:id` → `TablePage`. Files under `pages/` are thin re-exports; all logic lives under `features/`.

### State management

- **Zustand** (`features/auth/authStore.ts`) — auth only: JWT token + decoded user. Token persisted in localStorage. `useRequireAuth()` redirects to `/login` if unauthenticated.
- **TanStack Query v5** — all server state. Each page's hooks are co-located in `useTable.ts` / `useGames.ts`. Queries refetch every 15 s as a slow fallback; WebSocket events are the primary real-time push.

### Real-time layer

`useGameSocket.ts` opens a native WebSocket to `/games/:id/ws?token=<jwt>`. On each event, it calls `queryClient.invalidateQueries(…)` for the affected query keys, which triggers re-fetches. The six event types are defined in `ws/hub.go` (Go) and mirrored as a TypeScript union in `useGameSocket.ts`.

### Feature structure (`features/table/`)

| File | Role |
|------|------|
| `TablePage.tsx` | Full table layout: sticky topbar, shoe status, player hand, dealer controls, leaderboard sidebar |
| `useTable.ts` | All TanStack Query hooks + mutations for the table page |
| `useGameSocket.ts` | WebSocket lifecycle, cache invalidation on events |
| `DealerControls.tsx` | Dealer-only action buttons (shuffle, start, deal, end, add deck) |
| `Leaderboard.tsx` | Player list with seat order, hand values, optional kick button |
| `GameResult.tsx` | Final standings screen (rendered when `game.status === 'FINISHED'`) |
| `CardBadge.tsx` | Single card display; exports `SUIT_SYMBOL`, `SUIT_COLOR`, `FACE_LABEL` maps |

### Auth flow

`POST /auth/register` (or `/auth/login`) with `{ username }` (no password — passwordless for demo). Returns a JWT. Stored in Zustand + localStorage. Passed as `Authorization: Bearer <token>` on all API requests via `lib/apiClient.ts`.

---

## Tests

Backend tests use **real in-memory SQLite** (no mocks). `invariants_test.go` exercises the full stack (commands + persistence + domain) and covers: 52 unique cards, auto-end threshold, player removal returns cards, decks sealed after start, FINISHED rejects mutations, leaderboard sort + tie-break.

Frontend tests: not yet implemented (Vitest is in the plan).

To run a single backend test by name:
```bash
go test -run TestAutoEnd_LessThan_NotLessOrEqual ./internal/application/commands/
```
