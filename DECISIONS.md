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

## ADR-002: i18n — i18next with namespace cascade and pre-merge cache

**Status:** Accepted  
**Date:** 2026-05-30

**Context:**  
The frontend needed internationalisation to demonstrate scalability for the GoTo WebVoice assignment.
GoTo serves North America — en-US and fr-CA are both production-relevant. The architecture decision
was whether to use a simple single-file approach or a namespace-based system, and how to handle
the fallback cascade efficiently.

**Decision:** `i18next` + `react-i18next` + `i18next-browser-languagedetector`.  
Namespace-based source files (5 namespaces × 3 locales = 15 JSON files).  
Pre-merge strategy: at module load time, each `{locale, namespace}` pair is deep-merged over the
en-US base and cached in-memory + sessionStorage. All `t()` calls are O(1) memory lookups.

**Alternatives considered:**
- `react-intl` (FormatJS) — rejected: heavier API, ICU message syntax is verbose for a demo, no
  built-in cascade strategy.
- `LinguiJS` — rejected: requires a compile step (`lingui extract`/`compile`) which adds CI friction.
- Single file per locale (`en-US.json`) — rejected: one large flat file is harder to navigate and
  causes merge conflicts as the team grows.

**Why namespace files:**  
Each namespace maps to a feature area (`common`, `auth`, `lobby`, `table`, `errors`). A key's
namespace is always known by convention — there is no ambiguity about which file contains `logout`
(it's in `common`). Partial-override locales (fr-CA, es-MX) only need to define keys that differ
from en-US; everything else is filled in by the pre-merge step.

**Why pre-merge (not runtime fallback):**  
i18next's built-in `fallbackLng` performs a cascade on every missing key. Pre-merging runs the
cascade once per `{locale, namespace}` pair at startup and stores the fully-resolved object.
Subsequent `t()` calls have zero fallback logic — they're plain JS property access on the merged object.
The sessionStorage layer means a page reload for the same locale skips even the merge computation.

**Language preference vs. merged content storage:**  
Language preference (e.g. `"fr-CA"`) → `localStorage` (persists across sessions — UX intent).  
Pre-merged locale objects → `sessionStorage` (cleared on browser close — content cache, not preference).  
See ASSUMPTIONS A-005 for the long-term localStorage + version-hash migration path.

**TypeScript safety:**  
`src/types/i18n.d.ts` augments `i18next`'s `CustomTypeOptions` with resource types derived from the
en-US canonical files. Unknown keys fail at compile time. en-US is the single source of truth for
all key names.

**Tradeoffs accepted:**
- All locale files are bundled into the Vite build (no lazy-loading). Total payload is < 10 KB —
  acceptable at this scale. See ASSUMPTIONS A-007 for the lazy-load upgrade path.
- fr-CA and es-MX are partial translations — missing keys fall back to en-US silently. Documented
  in ASSUMPTIONS A-005 as intentional for the demo scope.

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

## ADR-013: Phase 1 frontend uses polling, not WebSocket

**Status:** Accepted — deliberately time-boxed  
**Date:** 2026-05-29

**Context:**  
The frontend needs to reflect live game state (players joining, cards dealt, shoe count changing).
Two options: 5-second polling via TanStack Query's `refetchInterval`, or WebSocket push.

**Decision:** Polling in Phase 1.

**Reasoning:**
- WebSocket requires a hub goroutine, JWT upgrade on the HTTP handshake, fan-out broadcast, and a
  persistent client connection — that's a full backend subsystem, not a one-day feature.
- 5-second polling is perfectly acceptable for a demo with 2–4 players on a LAN. The game moves
  slower than 5 seconds per action.
- The architecture explicitly plans for WebSocket in Phase 2. The `lib/wsClient.ts` stub is already
  in the frontend codebase. The upgrade path is clear and non-disruptive to the existing queries.

**Tradeoffs accepted:**
- Up to 5 seconds of lag between an action on one client and the other client seeing it.
- Minor unnecessary load from polling even when nothing changes. Acceptable at demo scale.

**Phase 2 upgrade path:** Replace `refetchInterval: 5000` queries with WebSocket event listeners.
Game state changes broadcast via the hub instantly — zero polling overhead in production.

---

## ADR-014: No game rules engine — DeckForge is infrastructure, not a ruleset

**Status:** Accepted  
**Date:** 2026-05-29

**Context:**  
The assignment says "a very basic game in which one or more decks are added to create a game deck...
along with a group of players getting cards from the game deck." It does not specify poker hand
evaluation, betting, blackjack strategy, or any win condition beyond "highest total hand value."

**Decision:** No game rules engine. DeckForge implements the card management infrastructure.
Specific game rules (poker, blackjack, baccarat) are explicitly a layer above this engine.

**Reasoning:**
- The assignment's own description uses "engine" language — shoe, dealing, shuffling, scoring.
- The stated win condition (highest numeric sum) maps directly to the leaderboard sort — no special
  rules logic needed.
- Adding a poker hand evaluator, betting round, or blind/ante system would be speculative scope
  expansion that the assignment does not request and that the interview team did not ask for.
- Clean seam: any game-specific rules can be implemented in a service layer above the engine API
  without touching a line of DeckForge's code.

**What this means in practice:** The engine deals cards and reports hand values. It does not know
about "flush," "straight," "blackjack," or "bust." Those interpretations live in a game adapter,
not in the engine.

---

## ADR-015: Consciously deferred — features not built in Phase 1

**Status:** Accepted  
**Date:** 2026-05-29

**Context:**  
The assignment window was 2 days. Many features that would belong in a complete product were
evaluated and explicitly scheduled for later phases rather than rushed into Phase 1. This ADR
records what was deferred and why, so the choices are transparent at the interview.

| Feature | Phase | Reason for deferral |
|---|---|---|
| Real-time WebSocket events *(shipped)* | ~~Phase 2~~ | Pulled forward. Hub goroutine + JWT-on-upgrade + 6 event types fully implemented. See ROADMAP Phase 2. |
| Per-player turn timer (15 s auto-accept) | Phase 2 | Server-side goroutine per active game; requires turn order state machine first. |
| In-game chat | Phase 2 | WebSocket hub already in place; needs ephemeral message table + chat event type. |
| SVG card graphics | Phase 3 | Cards render correctly as "A♥", "K♦" text. SVG assets are visual polish, not API correctness. |
| Frontend unit/component tests (Vitest) | Phase 4 | Backend invariants covered by 6 real-SQLite integration tests. Frontend tests most valuable on stable, finalized components. |
| Swagger/OpenAPI *(shipped in Phase 1)* | ~~Phase 4~~ | Pulled forward; took < 2 hours and improves reviewer experience significantly. |
| Docker *(shipped in Phase 1)* | ~~Phase 5~~ | Pulled forward for same reason. |
| CI/CD pipeline (GitHub Actions → GHCR) | Phase 5 | ~30 min value for reviewer; zero functional value before demo submission. |
| PostgreSQL *(shipped in Phase 1)* | ~~Phase 5~~ | `DB_DRIVER` env var switch took < 1 hour. |
| Redis pub/sub for horizontal scaling | Phase 6 | Relevant only with multiple backend instances behind a load balancer. |
| OIDC / Azure Entra ID authentication | Phase 6 | JWT with HMAC is structurally identical to JWT with OIDC — one function swap in `auth_middleware.go`. |
| Rate limiting on API endpoints | Phase 5 | Not in assignment spec. Documented as production requirement. |
| Spectator role | Post-Phase 2 | Straightforward RBAC addition once WebSocket turn order is live. |
| Multi-language / i18n *(shipped)* | ~~Phase 4~~ | Shipped. `i18next` + `react-i18next`; en-US (canonical), fr-CA (full), es-MX (core); namespace cascade with sessionStorage merge cache. See ADR-002. |

---

## ADR-016: Language-driven UI — i18n as a first-class feature, not a retrofit

**Status:** Accepted  
**Date:** 2026-05-30

**Context:**  
Most demo projects hardcode English strings. GoTo serves North America and their production
products ship in multiple languages. The question was whether to treat i18n as a Phase 4 polish
item or as a first-class architectural concern.

**Decision:** i18n from day one, namespace-first. Every user-facing string goes through a
`t('namespace:key')` call. No English literals in TSX files. en-US is the canonical source;
fr-CA and es-MX are partial overrides that cascade to English for any untranslated key.

**Reasoning:**
- GoTo evaluates architecture quality, not just feature count. A project that has i18n baked in
  demonstrates product thinking that a rushed hardcoded demo does not.
- The cascade design (pre-merge at startup, O(1) lookup thereafter) is defensible at interview:
  it's a deliberate performance choice, not a naive fallback chain.
- Partial translations (es-MX is intentionally sparse) show the cascade working. An evaluator
  switching to es-MX and seeing English fallback for untranslated keys demonstrates the system
  is behaving as designed, not broken.
- Language preference in `localStorage` (persists across sessions) vs merged namespace in
  `sessionStorage` (cleared on tab close, content cache) is a deliberate distinction. See ADR-002.

**Tradeoffs accepted:**  
All locale files are bundled — no lazy loading. Acceptable at < 15 KB total. See ROADMAP Phase 4
for the lazy-loading upgrade path.

---

## ADR-017: WebSocket push as primary, polling as fallback

**Status:** Accepted  
**Date:** 2026-05-30

**Context:**  
Phase 1 was originally planned as polling-only (ADR-013). The question was whether to pull
WebSocket forward given GoTo's real-time product context.

**Decision:** WebSocket push is the primary real-time mechanism, with TanStack Query's
`refetchInterval: 15000` as a slow fallback for resilience. The WS hub is goroutine-based
(one hub goroutine + one client goroutine per connected player). Six event types cover all
game state transitions.

**Reasoning:**
- GoTo's product is real-time voice/video. Demonstrating goroutine-based WebSocket competence
  is directly on-brand in a way that polling alone is not.
- The implementation cost was lower than expected: the gorilla/websocket hub is ~120 lines,
  the hub interface is already abstracted for Redis pub/sub scaling (Phase 6).
- Polling as a fallback (not the primary path) means the UI degrades gracefully if WS drops —
  players see updates within 15 seconds at worst.

**Tradeoffs accepted:**
- WS auth via `?token=` query string (browsers cannot send headers on WS upgrade). Token is
  short-lived (24h JWT); the risk is acceptable for a demo. Production upgrade: short-lived
  WS ticket endpoint.
- Horizontal scaling requires Redis pub/sub (Phase 6). In-process hub is correct for single-
  instance demo.

---

## ADR-018: Frontend decomposed into feature-sliced components, not monolithic pages

**Status:** Accepted  
**Date:** 2026-05-30

**Context:**  
A minimal assignment frontend could be a single large component per route. The question was
how much component decomposition to apply given the 2-day window.

**Decision:** Feature-sliced structure: `features/auth/`, `features/games/`, `features/table/`.
Within each feature, components, hooks, and queries are co-located. Shared primitives are in
`shared/`. No cross-feature imports.

**Key splits:**
- `TablePage` → `DealerControls`, `Leaderboard`, `GameResult`, `CardBadge`
- `useTable.ts` owns all TanStack Query hooks for the table route
- `useGameSocket.ts` owns the WS lifecycle; emits callbacks so TablePage stays stateless about WS internals

**Reasoning:**
- An evaluator reading the source should be able to navigate to any feature without reading the
  entire codebase. Co-location makes this possible.
- TanStack Query hooks isolated per feature means cache keys are owned by the file that defines
  them — no implicit dependencies across features.
- This structure is the React equivalent of feature modules in Angular or areas in ASP.NET MVC.
  It's a recognizable pattern that demonstrates production frontend experience.

---

## ADR-019: Auto-end strategies — two independent triggers

**Status:** Accepted  
**Date:** 2026-05-30

**Context:**  
The original auto-end only fired on shoe exhaustion (`remaining < activePlayerCount`). The
question was whether player removal below minimum players should also trigger auto-end.

**Decision:** Two independent auto-end conditions, both checked after every mutating operation:
1. `remainingCards < activePlayerCount` — shoe cannot serve a full round
2. `activePlayerCount < game.MinPlayers` — table falls below configured minimum

Both conditions call the same `checkAutoEnd` helper. The helper is idempotent (only acts on
IN_PROGRESS games). Condition 2 is the direct consequence of admin kicking a player.

**Reasoning:**
- A game with one player is not a multiplayer game. Auto-ending prevents a dealer from being
  left alone at the table with no path forward.
- Broadcasting `game_ended` with `reason: "not_enough_players"` gives clients a distinct message
  to display, separate from the normal game-over screen.
- The `AUTO_END_GAME` env var disables both conditions for QA testing.

---

## ADR-020: Multiple admin seeding via CSV env var

**Status:** Accepted  
**Date:** 2026-05-30

**Context:**  
The original design seeded a single admin via `ADMIN_SEED_USERNAME`. A real deployment needs
multiple operators (e.g., two developers both needing admin access during evaluation).

**Decision:** `ADMIN_SEED_USERNAMES` accepts a comma-separated list. Each username is created
as admin on startup if it does not already exist. Existing users are not modified (role elevation
would require an explicit admin API call).

**Reasoning:**
- Zero additional code complexity — a `strings.Split` on the config value.
- Reviewers can seed their own admin username alongside the documented default.
- Prevents the "who has the admin account?" problem during a live demo or pair evaluation.

---

## ADR-021: Game min/max player count configurable per-game and via env defaults

**Status:** Accepted  
**Date:** 2026-05-30

**Context:**  
The assignment does not specify player count limits. Hardcoding 2–8 is the simplest path.

**Decision:** `MinPlayers` and `MaxPlayers` are set per-game at creation time (body params) with
env-var defaults (`MIN_PLAYERS=2`, `MAX_PLAYERS=8`). The domain validates `2 ≤ min ≤ max ≤ 8`.

**Reasoning:**
- A card engine without configurable table sizes is not a reusable engine — it is a single
  hardcoded game. The per-game setting demonstrates the engine philosophy.
- Default env vars let operators change platform-wide limits without code changes.
- The frontend creation modal exposes both fields with sensible defaults.

---

## ADR-022: JWT authentication — simple foundation with explicit production upgrade path

**Status:** Accepted  
**Date:** 2026-05-30

**Context:**  
The assignment does not require authentication. But per-player hand visibility requires identity.
The question was how much auth infrastructure to build in a 2-day window.

**Decision:** Username-only JWT (no password). `POST /auth/register` or `/auth/login` returns a
token. HMAC-SHA256 signing. Stateless — no session store. Role claim embedded in the token.

**Reasoning:**
- No password reduces demo friction to near zero (evaluator can create accounts instantly) while
  keeping the security model structurally correct: tokens are signed, expiry is enforced,
  middleware validates on every request.
- The production upgrade is one function: replace HMAC validation in `auth_middleware.go` with
  OIDC token introspection. The `role` claim maps directly from IdP group membership. No business
  logic changes required.
- Separation of `user` role (JWT) from `dealer` role (per-game DB lookup) is a deliberate
  design choice, not a simplification. See ADR-004.

---

## ADR-023: Dealer as an application-layer role, not a JWT claim

**Status:** Accepted  
**Date:** 2026-05-30

**Context:**  
Many implementations conflate "game creator" with a JWT role (e.g., `role: dealer`). This
creates a global, permanent claim for what is actually a contextual, per-game relationship.

**Decision:** Dealer status is resolved from the database on every protected request:
`currentUser.id == game.dealer_user_id`. The `DealerMiddleware` performs this check and
short-circuits with 403 if the caller is not the game's dealer.

**Reasoning:**
- A user can be the dealer of game A and a regular player in game B simultaneously. A JWT claim
  would make this impossible or require re-issuing tokens on every game state change.
- The database is the authoritative source for who is the dealer of a given game. Trusting it
  on every request is correct — it cannot become stale the way a JWT claim can.
- This pattern is the exact equivalent of resource-level ownership checks in ASP.NET Core
  (e.g., `if (resource.OwnerId != currentUser.Id) return Forbid()`). It is the right pattern.

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
| A10 | Suit counts are auto-displayed and manually triggerable | In production only the automatic display is needed. For this assignment both are present: the live suit count strip satisfies real-time UX; the "Check Suits" button makes the REST call (op 8) explicitly user-initiated, demonstrating the endpoint is callable on demand. |
| A11 | Shoe is auto-shuffled when the dealer starts the game | In any real card game the shoe is pre-shuffled before cards leave the table. Requiring the dealer to manually shuffle before starting is friction with no gameplay value. The dealer may still shuffle manually at any point during WAITING or IN_PROGRESS. |
| A12 | Tied top score = draw; no tie-breaking rule applied | When two or more players share the highest hand value, the result is declared a draw. DeckForge is a card engine, not a game ruleset. Tie-breaking (dealer advantage in blackjack, pot split in poker, seat-order priority) belongs in the game-specific adapter layer above this engine. If money were involved, the split policy is similarly out of scope — a pot-splitting rule could be layered on by consuming the leaderboard and applying domain-specific logic. |
| A13 | Auto-end is configurable via `AUTO_END_GAME` env var (default: true) | When true, the game ends automatically when the shoe cannot serve a full round (`remaining < activePlayerCount`). When false, the game continues and deals return whatever cards remain — including empty on a fully exhausted shoe (see A7). Disabling auto-end lets operators and QA verify A7 directly without manual game management. |
