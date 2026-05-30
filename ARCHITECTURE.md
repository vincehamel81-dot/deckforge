# DeckForge — Architecture

> Read this before touching any code. It is the authoritative reference for structure, data model,
> API contract, and role model. See DECISIONS.md for the reasoning behind each choice.

---

## What DeckForge is

A card-game engine: REST API + real-time WebSocket backend (Go) and a React TypeScript frontend.
It manages shoes (multi-deck card pools), players, dealing, and scoring. Game rules (poker hands,
blackjack strategy, betting) are explicitly out of scope — DeckForge is the foundation they sit on.

---

## Clean architecture layers

```
Presentation  ← Gin HTTP handlers, WebSocket handler, DTOs
     ↓
Application   ← Use cases (commands + queries). No framework deps.
     ↓
Domain        ← Entities, value objects, repository interfaces. Pure Go, zero deps.
     ↑
Infrastructure← GORM repos, WebSocket hub, JWT, zerolog. Implements domain interfaces.
```

**Dependency rule:** outer layers depend inward. Domain knows nothing about Gin, GORM, or JWT.
**Testing rule:** domain and application are testable with no database, no HTTP server.

---

## Folder structure

```
DeckForge/
  backend/
    cmd/server/main.go                  # entry point — wires deps, starts Gin
    internal/
      domain/
        game/
          game.go                       # Game entity + state machine
          repository.go                 # GameRepository interface
        shoe/
          shoe.go                       # Shoe entity
          card.go                       # Card value object (suit + face + numeric value)
          shuffle.go                    # Fisher-Yates — no library shuffle
          repository.go                 # ShoeRepository interface
        player/
          player.go                     # Player entity (user's seat in a game)
          repository.go                 # PlayerRepository interface
        user/
          user.go                       # User entity
          repository.go                 # UserRepository interface
      application/
        commands/
          create_game.go
          start_game.go                 # WAITING → IN_PROGRESS
          end_game.go                   # IN_PROGRESS → FINISHED
          delete_game.go
          add_deck_to_shoe.go           # sealed after game starts
          shuffle_shoe.go
          deal_cards.go                 # dealer deals N cards to one player
          add_player.go                 # join game + catch-up deal
          remove_player.go              # leave game → cards returned to shoe
          draw_round.go                 # Phase 2: Draw = 1 card to all players
          accept_turn.go                # Phase 2: Accept = pass turn
        queries/
          list_games.go
          get_game.go
          get_leaderboard.go            # players sorted by hand value desc
          get_player_hand.go            # cards for one player
          get_shoe_suit_counts.go       # undealt per suit
          get_shoe_card_counts.go       # undealt per suit+value, sorted
      infrastructure/
        persistence/
          db.go                         # GORM setup — SQLite default, PostgreSQL via env
          models.go                     # GORM structs (≠ domain entities)
          game_repo.go
          shoe_repo.go
          player_repo.go
          user_repo.go
        websocket/
          hub.go                        # Phase 2: one hub goroutine per game
          client.go                     # Phase 2: one goroutine per WS connection
          events.go                     # Phase 2: event type definitions
        auth/
          jwt.go                        # issue + validate tokens
      presentation/
        http/
          router.go                     # all Gin routes registered here
          handlers/
            auth_handler.go
            game_handler.go
            shoe_handler.go
            player_handler.go
            leaderboard_handler.go
            ws_handler.go               # Phase 2: HTTP→WebSocket upgrade
          middleware/
            auth_middleware.go          # JWT validation, injects user into context
            dealer_middleware.go        # checks currentUser.id == game.dealerUserId
            correlation_middleware.go   # stamps X-Correlation-ID on every request
            logger_middleware.go        # structured request/response logging
    config/
      config.go                         # env var loading with defaults
  frontend/
    src/
      features/
        auth/
          LoginPage.tsx
          useAuth.ts
          authStore.ts                  # Zustand: { user, token, login, logout }
        games/
          GamesPage.tsx                 # list of open games + create button
          GameCard.tsx
          CreateGameModal.tsx           # deckCount, minPlayers, maxPlayers inputs
          useGames.ts                   # TanStack Query: GET /games
        table/
          TablePage.tsx                 # /table/:id — full game view
          PlayerSeat.tsx                # one seat around the table
          CardHand.tsx                  # this player's face-up cards
          CardBack.tsx                  # other players' face-down card stacks
          ShoeStatus.tsx                # remaining cards + draws remaining counter
          Leaderboard.tsx               # live standings
          DealerControls.tsx            # deal / shuffle / end game (dealer only)
          DrawControls.tsx              # Phase 2: Draw / Accept buttons (on your turn)
          useTable.ts                   # TanStack Query: game, players, hand, shoe
          useWebSocket.ts               # Phase 2: WS connection + event dispatch
          tableStore.ts                 # Zustand: turn state, reveal mode
      shared/
        components/
          Card.tsx                      # single card face (SVG, Phase 3)
          Button.tsx
          Badge.tsx
        hooks/
          useRequireAuth.ts             # redirects to login if no token
      lib/
        apiClient.ts                    # axios instance with JWT interceptor
        wsClient.ts                     # Phase 2: WebSocket wrapper
        queryClient.ts                  # TanStack Query client config
      pages/
        index.tsx                       # route: /  → GamesPage
        table.tsx                       # route: /table/:id → TablePage
        login.tsx                       # route: /login → LoginPage
  ARCHITECTURE.md                       # this file
  DECISIONS.md                          # 10 ADRs + assumptions log
  LANGUAGE_TRADEOFFS.md                 # Go vs Java analysis
  ROADMAP.md                            # 6-phase 10-day plan
  README.md                             # setup, env vars, AI usage disclosure
  SETUP.md                              # admin seed, env var reference
```

---

## Data model

```
User
  id            UUID  PK
  username      string  UNIQUE  (alphanumeric, min 3 chars)
  role          enum { user, admin }     ← PLATFORM role only. Orthogonal to dealer/player.
  created_at    timestamp                  An admin can also be a dealer and/or a player simultaneously.

Game
  id            UUID  PK
  dealer_user_id UUID  FK → User          ← contextual dealer; checked per-request, not in JWT
  status        enum { WAITING, IN_PROGRESS, FINISHED }
  deck_count    int   (fixed at creation; shoe = deck_count × 52 cards; never changes)
  min_players   int   (default 2, set at creation)
  max_players   int   (default 8, set at creation)
  current_turn_player_id  UUID?  FK → Player  (Phase 2)
  created_at    timestamp
  started_at    timestamp?
  finished_at   timestamp?

ShoeCard                              (one row per card instance in the shoe)
  id            UUID  PK
  game_id       UUID  FK → Game
  suit          enum { HEARTS, SPADES, CLUBS, DIAMONDS }
  value         enum { ACE, TWO, THREE, FOUR, FIVE, SIX, SEVEN,
                       EIGHT, NINE, TEN, JACK, QUEEN, KING }
  numeric_value int   (ACE=1, TWO=2, ..., TEN=10, JACK=11, QUEEN=12, KING=13)
  position      int   (order in shoe; shuffle updates undealt cards only)
  held_by_player_id  UUID?  FK → Player  ← null = undealt (in shoe); non-null = dealt to player
                                            No separate state field — the null check IS the state.

Player                                (a user's seat in a specific game)
  id            UUID  PK
  game_id       UUID  FK → Game
  user_id       UUID  FK → User
  seat_order    int   (join sequence; determines turn order)
  joined_at     timestamp
  left_at       timestamp?  (soft delete; null = still active)
```

**No Deck table.** `POST /decks` generates a UUID client-side and inserts 52 ShoeCard rows into the
game's shoe. `Game.deck_count` is incremented. There is nothing to persist about a deck itself —
it is a batch of 52 cards, not an entity.

**Key invariants:**
- `ShoeCard.position` values are a dense sequence per game; shuffle reassigns positions of undealt cards only
- `held_by_player_id IS NULL` ↔ card is undealt; `IS NOT NULL` ↔ card is dealt
- A User may have at most one active Player row (`left_at IS NULL`) across all games
- `Game.deck_count` never changes after creation; total cards = `deck_count × 52`

---

## REST API contract

All endpoints return `Content-Type: application/json`.  
Auth endpoints are public. All others require `Authorization: Bearer <jwt>`.  
Dealer-only endpoints additionally check `currentUser.id == game.dealer_user_id`.

```
POST   /auth/register              body: { username }
                                   201: { token, user }

POST   /auth/login                 body: { username }
                                   200: { token, user }

GET    /games                      query: ?status=WAITING|IN_PROGRESS
                                   200: [ GameSummary ]

POST   /games                      body: { deckCount, minPlayers?, maxPlayers? }
                                   201: Game  (creator becomes dealer + joins as player)

GET    /games/:id                  200: GameDetail (includes shoe status)

DELETE /games/:id                  dealer only
                                   204

POST   /games/:id/start            dealer only; WAITING → IN_PROGRESS
                                   body: { initialDealCount }  (cards dealt to each player)
                                   200: Game

POST   /games/:id/end              dealer only; IN_PROGRESS → FINISHED
                                   200: Game + final leaderboard

POST   /decks                      generate a deck UUID (no DB write)
                                   201: { id }   ← use this id in the next call

POST   /games/:id/shoe/decks       dealer only; WAITING state only
                                   body: { deckId }
                                   inserts 52 ShoeCard rows; increments Game.deck_count
                                   200: ShoeStatus { deckCount, totalCards, remainingCards }

POST   /games/:id/shoe/shuffle     dealer only
                                   204  (no body — as per assignment spec)

GET    /games/:id/shoe/suits       200: [ { suit, count } ]

GET    /games/:id/shoe/cards       200: [ { suit, value, count } ] sorted suit then value desc

POST   /games/:id/players          authenticated user joins game
                                   201: Player  (catch-up deal applied if IN_PROGRESS)

DELETE /games/:id/players/:pid     dealer or the player themselves
                                   204  (cards returned to shoe)

GET    /games/:id/players          200: [ { player, handValue, cardCount } ] sorted by handValue desc

GET    /games/:id/players/:pid/hand  self or admin only
                                   200: [ Card ]

POST   /games/:id/players/:pid/deal  dealer only
                                   body: { count }
                                   200: { dealtCount }  (may be < count if shoe exhausted)

# Phase 2 only
GET    /ws                         query: ?gameId=:id   Authorization header required
                                   101 Upgrade → WebSocket
POST   /games/:id/draw             active player's turn; deals 1 card to ALL players
                                   200: { remainingCards, nextTurnPlayerId }
POST   /games/:id/accept           active player passes turn
                                   200: { nextTurnPlayerId }
```

**REST tradeoff note:** `shuffle`, `start`, `end`, `deal`, `draw`, `accept` are
actions, not resources. We model them as POST to action sub-resources rather than forcing PATCH on
state fields. This is the pragmatic approach used by GitHub and Stripe APIs and is the most
readable representation of the intent.

---

## WebSocket events (Phase 2)

Connection: `wss://<host>/ws?gameId=<id>` with JWT in `Authorization` header.

All events are JSON: `{ "event": "<name>", "payload": { ... } }`.

```
player_joined        { playerId, username, seatOrder }
player_left          { playerId }
cards_dealt          { playerId, cardCount }          ← values NOT broadcast
shoe_shuffled        { remainingCards }
game_started         { initialHandSize, currentTurnPlayerId }
turn_changed         { currentTurnPlayerId }
game_ended           { winnerId, leaderboard: [...] } ← full hands revealed at end
auto_ended           { reason: "shoe_exhausted", winnerId, leaderboard: [...] }
```

**Privacy invariant:** `cards_dealt` never includes card values. Only the card holder learns
their values via `GET /games/:id/players/:pid/hand`.

### Turn timer (Phase 2)

Each player has **15 seconds** to Draw or Accept on their turn (configurable via `TURN_TIMEOUT_SECONDS`
env var). The server tracks the deadline server-side; the client renders an animated countdown.

On expiry the server auto-fires an Accept (pass) for that player and emits `turn_changed`.
The frontend animates the remaining seconds on the active player's seat during their turn.

```
turn_timer_started   { playerId, expiresAt }   ← ISO-8601 timestamp; client drives countdown
turn_expired         { playerId }               ← server auto-accepted; turn_changed follows
```

Client implementation: a `useEffect` in `DrawControls` starts a `setInterval` on `turn_timer_started`
and clears it on `turn_changed` or component unmount.

---

## Role model

```
role: user   → JWT claim. Any registered player.
              Can: join games, see own hand, take turns (Phase 2).

dealer       → NOT a JWT claim. Contextual per game.
              Resolved by: currentUser.id == game.dealer_user_id.
              Can: deal, shuffle, start/end their own game, add decks (WAITING only).
              Is also a player in their own game.

role: admin  → JWT claim. Platform operator.
              Can: view all games and all hands, delete any game, manage users.
              Created via seed script (ADMIN_SEED_USERNAME env var) — no public signup.
```

---

## Auth flow

```
1. User visits /login → enters username
2. POST /auth/register (or /auth/login if username exists)
3. Server issues JWT: { sub: userId, username, role, exp }
4. Frontend stores token in localStorage
5. axios interceptor attaches Authorization: Bearer <token> to every request
6. auth_middleware.go validates signature + expiry, injects user into Gin context
7. dealer_middleware.go (on dealer routes) checks game.dealer_user_id == ctx.user.id
```

---

## Configuration (env vars)

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP server port |
| `DB_DRIVER` | `sqlite` | `sqlite` or `postgres` |
| `DATABASE_URL` | `./deckforge.db` | SQLite path or PostgreSQL DSN |
| `JWT_SECRET` | *(required)* | HMAC signing secret |
| `JWT_EXPIRY` | `24h` | Token lifetime |
| `CORS_ORIGIN` | `http://localhost:5173` | Allowed frontend origin |
| `MIN_PLAYERS` | `2` | Default minimum players per game |
| `MAX_PLAYERS` | `8` | Default maximum players per game |
| `ADMIN_SEED_USERNAME` | *(optional)* | Creates admin user on first boot |
| `LOG_LEVEL` | `info` | zerolog level: debug/info/warn/error |

---

## Key invariants (test these)

1. Shuffle + 52 × dealCards(1) to same player → all 52 unique cards, no repeats
2. 53rd dealCards(1) → empty result, no error
3. Player removed → their card count returns to shoe's UNDEALT count
4. `remainingCards` never goes negative
5. Auto-end fires when `remainingCards < activePlayerCount` (not ≤)
6. Leaderboard is always sorted descending by hand value; ties broken by seat_order asc
7. Decks cannot be added once status = IN_PROGRESS
8. FINISHED games reject all mutations (deal, shuffle, join, draw)
