# DeckForge — Architecture Decision Records

This document logs every significant decision made during the design and implementation of DeckForge,
including the reasoning, alternatives considered, and tradeoffs accepted. Maintained as a living
document throughout development. To be referenced during the GoTo live interview.

---

## ADR-001: Language — Go over Java

**Status:** Accepted  
**Date:** 2026-05-29

**Context:**  
The assignment offers Java or Go. Both are production languages at GoTo. The choice must be defensible
in a live interview.

**Decision:** Go.

**Reasoning:**
- GoTo builds VoIP and real-time communication infrastructure — Go is the dominant language in that
  space (used by Twilio, Discord, Cloudflare, and the Go standard library's `net/http` is production-grade).
- Go's goroutine model makes WebSocket connections trivially cheap (~2 KB per goroutine vs ~256 KB per
  Java thread). This matters for a card table where every player holds a persistent WebSocket connection.
- Single static binary. Docker image is ~8 MB vs ~200 MB for a Spring Boot JVM image.
- The domain logic (shuffle, deal, score) is pure functions — Go is cleaner than Java for this.
- "Foundational part of a new product" — if this becomes a GoTo product, it will live in the same
  ecosystem as their other Go services.

**Alternatives considered:** Spring Boot — rejected because the JVM overhead and framework magic are
unnecessary for a domain this size, and Go aligns better with GoTo's production stack.

**Tradeoffs accepted:**
- Go's ORM (GORM) is less mature than Hibernate/JPA. Acceptable for this domain size.
- No built-in DI container — manual wiring via constructors. Actually a feature: explicit dependencies.
- Unfamiliar syntax for a .NET developer — mitigated by 25yr experience reading any code.

---

## ADR-003: Storage — SQLite default, PostgreSQL production path

**Status:** Accepted  
**Date:** 2026-05-29

**Context:**  
The assignment does not specify persistence. Options: in-memory, SQLite, PostgreSQL.

**Decision:** SQLite as default (embedded, zero setup), with GORM Repository interface making
PostgreSQL a one-line swap.

**Reasoning:**
- Reviewer experience: `git clone && go run . && npm run dev` — no database setup required.
  Friction-free evaluation is a product quality signal.
- "If I hit F5, the game should still be there" — in-memory fails this. SQLite does not.
- The Repository interface (domain layer) means the storage backend is never referenced in business
  logic. Swapping to PostgreSQL requires changing one GORM driver import and one connection string.

**Production path documented in README:**  
Change `DB_DRIVER=sqlite` to `DB_DRIVER=postgres` and provide `DATABASE_URL`. No code changes.

**Tradeoffs accepted:**
- SQLite does not support concurrent writes well. For production scale, PostgreSQL is required.
  This is explicitly documented and does not affect correctness at demo scale.

---

## ADR-004: Identity — Three-tier model (user / dealer / admin), username-only JWT

**Status:** Accepted  
**Date:** 2026-05-29

**Context:**  
The assignment does not require authentication. But per-player views ("only the current player can
see their hand") cannot be secured without identity. Separately, a long-term product needs a
distinction between game-level control and platform-level administration — conflating the two
produces a role model that cannot scale.

**Decision:** Three distinct identity tiers:

| Tier | How it works | What they can do |
|---|---|---|
| `user` | JWT `role: user`. Any registered player. | Join/leave games, see own hand, take turns. |
| `dealer` | **Not a JWT role — contextual.** The user who created a game is its dealer. Checked per-request: `currentUser.id == game.dealerUserId`. | Deal cards to any player, add decks to shoe, shuffle, start game, end game — within their own game only. |
| `admin` | JWT `role: admin`. Platform-level operator. | View all games, all hands, delete any game, manage users. Never created through the public API. |

**Why dealer is not in the JWT:**  
Dealer is a *relationship* between a user and a specific game, not a permanent account attribute.
A user can be the dealer of game A and a regular player in game B simultaneously. Encoding this in
the JWT would make it global and permanent — wrong model. The backend resolves dealer status from
the database on every protected request.

**Why admin is separate from dealer:**  
A dealer controls one table. An admin oversees the platform. These are different jobs with different
trust levels and different access patterns. Conflating them (e.g. "game creator = admin") breaks
multi-tenancy and makes future role expansions (moderator, spectator, auditor) impossible without
rework.

**Admin creation:**  
No public registration path for admin. Documented in `SETUP.md`: seed script sets the first admin
via environment variable (`ADMIN_SEED_USERNAME`). Subsequent admins promoted via admin API
(admin-only endpoint).

**Reasoning:**
- Per-player gating requires identity — auth is implicitly required by the frontend spec.
- No password reduces friction for demo while keeping the security model structurally correct.
- Username uniqueness enforced at DB level — no impersonation possible within a session.
- JWT is stateless — scales horizontally without a session store.

**Production path documented in README:**  
Swap the JWT issuer for an OIDC provider (Azure AD / Entra ID, Okta). The `auth_middleware.go`
validates tokens — replace HMAC validation with OIDC token introspection. The `role` claim maps
directly from the IdP's group membership. No business logic changes required.

**Tradeoffs accepted:**
- No password means anyone who knows a username can claim it on a new device. Acceptable for demo.
  Unacceptable for production — OIDC fixes this entirely.
- No email — no account recovery in demo mode. Documented limitation.

---

## ADR-005: Shuffle — Fisher-Yates, implemented manually

**Status:** Accepted  
**Date:** 2026-05-29

**Context:**  
The assignment explicitly forbids using library-provided shuffle operations. Random number generators
are permitted.

**Decision:** Fisher-Yates (Knuth) shuffle using Go's `math/rand` (seeded with `crypto/rand` for
unpredictability).

**Reasoning:**
- Fisher-Yates is the standard algorithm for unbiased uniform random permutations.
- O(n) time, O(1) extra space.
- Operates only on the UNDEALT portion of the shoe — cards already dealt are not re-permuted.
- Using `crypto/rand` as the seed source prevents predictable sequences.

**Verification:** Unit test confirms that after 1 shuffle + 52 sequential `dealCards(1)` calls on
a single-deck shoe, all 52 unique cards are returned exactly once. The 53rd call returns nothing.

---

## ADR-006: Game state machine — WAITING → IN_PROGRESS → FINISHED

**Status:** Accepted  
**Date:** 2026-05-29

**Context:**  
The assignment does not define game lifecycle states. We need rules for when certain operations
are valid.

**Decision:**

| State | Description | Allowed operations |
|---|---|---|
| `WAITING` | Game created, shoe being built, players joining | Add decks, add/remove players, shuffle |
| `IN_PROGRESS` | Dealer started the game explicitly | Deal, shuffle, add/remove players, query |
| `FINISHED` | Dealer ended the game | Read-only queries only, no new joins |

Transition `WAITING → IN_PROGRESS`: explicit "start game" action by the dealer.  
Transition `IN_PROGRESS → FINISHED`: explicit "end game" by the dealer, or game deletion.

Players can join and leave during WAITING and IN_PROGRESS. Not during FINISHED.
Decks can only be added to the shoe during WAITING (shoe is locked once IN_PROGRESS).

**Reasoning:**
- Prevents dealing from an empty or uninitialized shoe.
- Shoe lock prevents mid-game deck injection (which would affect card counts mid-session).
- Principle of least surprise: a finished game should not accept new players.

---

## ADR-007: Player removal — cards returned to shoe as undealt

**Status:** Accepted  
**Date:** 2026-05-29

**Context:**  
When a player leaves or is removed, the assignment is silent on what happens to their cards.

**Decision:** Cards held by a removed player are returned to the shoe as undealt. They rejoin the
pool and can be dealt again on the next deal operation (or after a shuffle).

**Reasoning:**
- "Principle of least surprise" for a shoe-based game — cards don't disappear, they return to play.
- Prevents shoe depletion when players leave frequently.
- Consistent with casino shoe mechanics (mucked cards can re-enter the shoe).

**Tradeoffs accepted:**
- A card returned to the shoe and re-dealt could theoretically be "remembered" by a player who saw
  it. This is a known casino problem solved by shuffling — shuffle before resuming is the operator's
  responsibility, not the engine's.

---

## ADR-008: WebSocket for real-time game state push

**Status:** Accepted  
**Date:** 2026-05-29

**Context:**  
The assignment requires a frontend with visual representation of all operations. The question is
whether the UI updates via polling, manual refresh, or real-time push.

**Decision:** WebSocket push using gorilla/websocket. One hub goroutine per game. One client
goroutine per connected player. The hub broadcasts game state events to all clients in the game.

**Events pushed:**
- `player_joined`, `player_left`
- `cards_dealt` (notifies all players that X cards were dealt to player Y — not the card values)
- `shoe_shuffled`
- `game_started`, `game_ended`

**Reasoning:**
- Go's goroutine model makes this natural and cheap.
- GoTo's core product is real-time communication — demonstrating WebSocket competence is on-brand.
- A card table without live updates would feel static and incomplete.

**Privacy rule:** When `cards_dealt` is broadcast, the payload contains only `{ playerId, cardCount }`.
The actual card values are only visible to the card holder (requires authenticated request).
Platform admins can query any hand via the admin API — this is an operational tool, not a game view.

**Scalability note:** The hub currently runs in-process. To scale horizontally (multiple server
instances), replace the in-process channel with Redis pub/sub. The `Hub` interface is already
abstracted for this swap.

---

## ADR-009: Card face values for scoring

**Status:** Accepted  
**Date:** 2026-05-29

**Context:**  
The assignment says "use face values of cards only" and gives examples: 10 + King = 23, 7 + Queen = 19.
The sort order spec confirms: "King, Queen, Jack, 10….2, Ace with value of 1."

**Decision:**

| Card | Numeric value |
|---|---|
| Ace | 1 |
| 2–10 | Face value (2–10) |
| Jack | 11 |
| Queen | 12 |
| King | 13 |

**Reasoning:** Derived directly from the assignment's own examples. Not an assumption.

---

## ADR-010: POST /decks returns a UUID with no database write

**Status:** Accepted  
**Date:** 2026-05-29

**Context:**  
The assignment specifies "Create a deck" and "Add a deck to the game deck" as two distinct operations.
The naive implementation would create a Deck table and persist a row on `POST /decks`. But a Deck has
no meaningful attributes beyond an ID — it is defined entirely by its 52 cards, which only exist once
added to a shoe.

**Decision:** `POST /decks` generates and returns a UUID without writing to the database. That UUID is
passed as `deckId` to `POST /games/:id/shoe/decks`, which inserts 52 ShoeCard rows into the game's
shoe and increments `Game.deck_count`. There is no Deck table.

**Reasoning:**
- A deck is a batch operation (52 card inserts), not an entity worth persisting independently.
- Persisting a Deck row that has no attributes except an ID adds a table with zero query value.
- The two-step API contract from the assignment is preserved — callers still `POST /decks` first,
  then `POST /games/:id/shoe/decks` — the persistence model is simply optimized.
- `Game.deck_count` and the ShoeCard row count provide all the shoe capacity information needed.

**Tradeoffs accepted:**
- The UUID returned by `POST /decks` cannot be re-used across games (it is single-use by design).
  Documented in API response: `{ "id": "...", "note": "single-use — add to a game shoe immediately" }`.
- If a caller creates a deck UUID but never adds it to a shoe, nothing is leaked — no orphan rows.

---

## ADR-011: Shoe terminology and capacity (Deck, Shoe, Card)

**Status:** Accepted  
**Date:** 2026-05-29

**Context:**  
The assignment uses "deck" (52 cards) and "game deck" / "shoe" (combined decks) interchangeably.
We need consistent internal naming.

**Decision:**
- `Deck`: a transient concept. `POST /decks` generates a UUID but writes nothing to the database.
  That UUID is passed to `POST /games/:id/shoe/decks`, which inserts 52 ShoeCard rows and increments
  `Game.deck_count`. There is no Deck table — a deck is a batch operation, not an entity.
- `Shoe`: surfaced via `Game.deck_count` (total decks added) and a count query on ShoeCard rows
  where `held_by_player_id IS NULL` (remaining). No separate Shoe table needed.
- `undealt` state on ShoeCard is expressed as `held_by_player_id IS NULL`. No state enum — the
  null check is the state. This eliminates the risk of the state field diverging from the FK.
- No "packet" or "pack" concept — `Game.deck_count × 52` surfaces total capacity directly.

**Math:** 1 deck = 52 cards. 2 decks = 104 cards. 4 decks = 208 cards. 6 decks = 312 cards
(standard blackjack casino shoe).

---

## ADR-012: Turn-based Draw/Accept mechanic (Phase 2 extension)

**Status:** Accepted — planned for Phase 2  
**Date:** 2026-05-29

**Context:**  
The assignment defines `dealCards(playerId, count)` but specifies no game loop, turn order, or
strategy mechanic. Without a constraint, a player could draw indefinitely and always win — no
strategic tension.

**Decision:** Implement a clockwise Draw/Accept mechanic in Phase 2:
- Dealer sets deck count at game creation (shoe is sealed — no additions after game starts)
- Dealer deals the initial hand (N cards to every player) before the draw phase begins
- After the initial deal, play proceeds in turn order (clockwise by join sequence)
- On your turn: **Draw** = 1 card is dealt to every active player from the shoe; **Accept** = pass your turn
- Every player always holds the same number of cards — fairness is structural, not enforced by trust
- The mechanic creates genuine strategy: drawing helps you but also helps your opponents
- If you are winning, you Accept to lock your lead; if losing, you Draw and gamble on the distribution
- **Auto-end:** when `remainingCards < activePlayerCount`, the system cannot deal a full round — game ends automatically. All hands are revealed. Winner = highest total hand value.
- Dealer can also end the game manually at any time
- Maximum draw rounds = `(deckCount × 52 − initialDeal × playerCount) / playerCount` — predictable, displayable as a live counter ("X draws remaining")

**Why Draw triggers everyone:**  
A player-only draw has no strategic cost — you'd always draw. Dealing to everyone makes each draw
a calculated risk. This mirrors mechanics in games like "Go Fish" variants and creates observable
tension that makes the frontend interesting.

**Backend additions required (Phase 2):**
- `currentTurnPlayerId` field on Game entity
- `POST /games/:id/draw` — triggers 1-card deal to all active players, advances turn
- `POST /games/:id/accept` — passes current player's turn, advances to next player
- WebSocket event: `turn_changed { currentPlayerId }`

**Tradeoffs accepted:**
- Not in the assignment spec — clearly documented as a designed extension.
- Adds backend state (turn order). Complexity is isolated to the application command layer.
- If time runs short, Phase 2 is deferred and the assignment API spec is still 100% met in Phase 1.

---

## Assumptions log

Decisions where the assignment was silent and we applied "principle of least surprise":

| # | Assumption | Rationale |
|---|---|---|
| A1 | Max 8 players default | Standard poker table size; configurable via env var |
| A2 | Min 2 players default | A game with 1 player is not a multiplayer game; configurable |
| A3 | One player cannot be in two simultaneous games | Simplifies identity and prevents ambiguous hand queries |
| A4 | Deck count is fixed at game creation; shoe is sealed at game start | Set at creation form (e.g. "5 decks = 260 cards"). No additions after. Prevents mid-game manipulation and makes remaining-card counts predictable. Auto-end triggers when `remainingCards < activePlayerCount` — can't deal a full round. On auto-end, all hands are revealed and the leaderboard shows the winner. |
| A5 | The game creator is the dealer AND a player | Home game model: the person who sets up the table also plays. Dealer ≠ admin (see ADR-004). |
| A6 | Shuffle permutes only undealt cards | Dealt cards belong to players; shuffling them back in would be a bug |
| A7 | 53rd dealCards(1) on an exhausted shoe returns empty, not an error | Assignment specifies this explicitly |
| A8 | Player joining mid-game receives an automatic catch-up deal | Fairness rule: every active player always holds the same number of cards. On join, the system deals `currentHandSize` cards from the shoe to the new player. If the shoe has fewer cards than needed for a full catch-up, the player receives what remains, then the auto-end check runs immediately. |
| A9 | Win condition = highest total hand value when game ends | Derived from the leaderboard sort spec — descending value IS the ranking |
