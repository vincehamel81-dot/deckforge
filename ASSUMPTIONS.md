# DeckForge — Design Assumptions

Assumptions that are intentional constraints, deferred decisions, or known simplifications made for
the scope of this assignment. Each entry documents what was assumed, why, and what the production
upgrade path looks like.

---

## A-001: Auth — Passwordless (username only)

Users register and log in with a username only — no password, no OAuth, no email verification.

**Why:** Eliminates auth complexity irrelevant to demonstrating real-time game mechanics.

**Production path:** Add `bcrypt`-hashed passwords to the `users` table, or replace entirely with
OAuth2 (e.g., MSAL / Entra ID). The JWT structure and `AuthMiddleware` remain unchanged — only the
credential validation step changes.

---

## A-002: Roles — Seeded via env var, not a management UI

Admin users are created at startup from `ADMIN_SEED_USERNAMES` (comma-separated list in `.env`).
There is no in-app role management screen.

**Why:** Role management UI is out of scope. The role field and guard logic are fully implemented;
only the assignment surface is simplified.

**Production path:** Add an admin panel endpoint (`PATCH /users/:id/role`) protected by admin role
guard.

---

## A-003: Game lifecycle — Auto-end when shoe cannot serve a full round

When `remaining cards < active player count` after a deal, the game transitions to `FINISHED`
automatically (controlled by `AUTO_END_GAME=true`).

**Why:** Prevents the dealer from needing to manually end the game after every shoe exhaustion.
The threshold is `<` not `<=` so that the last round can always be dealt (remaining == players
means one card each is still possible).

**Production path:** The `AUTO_END_GAME` env var already exists to disable this behaviour for
game variants where partial rounds are intentional.

---

## A-004: WebSocket — No reconnect / heartbeat logic

The WebSocket client (`useGameSocket.ts`) does not attempt to reconnect after a disconnect. If
the connection drops, the 15-second TanStack Query polling fallback keeps state eventually consistent.

**Why:** Reconnect logic adds complexity (exponential backoff, re-auth, duplicate-event deduplication)
that is out of scope for the assignment window.

**Production path:** Wrap the WebSocket in a reconnect loop with jitter, re-validate the JWT token
on reconnect (handle token refresh), and emit a client-side "reconnected" event to trigger a
full query invalidation.

---

## A-005: i18n — Language preference vs. locale cache storage

**Language preference** (e.g. `"fr-CA"`) is stored in `localStorage`.
This is intentional — it persists across browser sessions so the user does not have to re-select
their language on every visit.

**Pre-merged locale objects** (the result of merging a target locale over the `en-US` base) are
cached in `sessionStorage`. Each entry is keyed as `i18n:<locale>:<namespace>` and holds the
fully-resolved flat key map. The merge computation runs once per locale+namespace pair per session,
after which all `t()` calls are O(1) memory lookups.

**Long-term:** As locale files grow (more languages, more keys), the sessionStorage cache should
migrate to `localStorage` with a version-hash invalidation strategy: include a short hash of the
locale file bundle in the cache key (e.g. `i18n:fr-CA:common:a3f7b2`). On app update, the hash
changes, the stale cache entry is ignored, and the fresh merge result overwrites it. This avoids
users seeing stale translations after a deploy without requiring a manual cache clear.

---

## A-006: i18n — Backend error messages are not translated

Backend API responses carry short technical error strings (e.g. `"game already started"`) intended
for developer consumption, not end-user display. The frontend maps known error conditions to
translated `errors.json` keys. Unknown or unexpected errors fall back to a generic translated
message.

**Why:** Translating backend error strings inside the API would couple the backend to a locale
system it does not need. Frontend owns all user-facing text.

**Production path:** Already implemented via `errors.json` namespace. Extend the error-code map
as new backend error conditions are identified.

---

## A-007: i18n — Locale files are not lazy-loaded

All locale files (`en-US`, `fr-CA`, `es-MX`) are bundled into the Vite build output and loaded
synchronously at application startup. This avoids an async loading phase and the need for a
React Suspense boundary around translations.

**Why:** The total payload is small (three locales × four namespaces × ~50 keys ≈ < 10 KB).
Lazy loading would add code complexity with no measurable performance benefit at this scale.

**Production path:** If locale files grow significantly (hundreds of keys, many languages), switch
to dynamic `import()` per locale+namespace, add a `<Suspense>` boundary, and rely on the
`sessionStorage` merge cache (see A-005) to eliminate repeated loads.

---

## A-008: Language preference — per browser, not per user account

The selected language (`deckforge-locale` in localStorage) belongs to the browser instance, not to
the authenticated user. A user on two different devices or browsers will have independent language
settings and may need to reselect their language on each new browser.

Additionally, localStorage is shared across all tabs of the same origin within a single browser
session. Changing the language in one tab immediately takes effect in all other open tabs of the
same DeckForge session — this is intentional and provides consistent UX without tab-to-tab
coordination.

**Why:** Linking language preference to the user account requires a database column (`users.locale`),
an API endpoint to read/write it, and a login-time read before the first render. This is out of
scope for the demo.

**Production path:** Add a `locale` column to the `users` table (default `en-US`). Persist the
preference server-side on every language change via `PATCH /users/me`. Read it at login time and
hydrate the i18n instance before React renders. This ensures the correct language appears regardless
of browser or device.

---

## A-009: No IP-based session gating — multiple players per IP allowed

The backend does not track or restrict connections by IP address. Multiple users behind the same
NAT (home router, office network, mobile carrier) can all connect and participate in the same or
different games without any throttling or identity conflict.

**Why:** IP-based gating is not appropriate for a card game:
- Corporate and home networks routinely share a single public IP.
- Mobile carriers use CGNAT, meaning thousands of users share one IP.
- The correct identity boundary is the authenticated user (JWT), not the network address.

**Production path:** If abuse prevention is needed (e.g. bot farms or table collusion), the right
approach is account-level controls, not IP blocking:
- Rate limiting on the auth endpoint prevents account creation storms.
- With SSO (Entra ID / OIDC), each user is tied to an organizational identity with MFA — this is
  a far stronger anti-abuse signal than an IP address.
- Concurrent session limits (`max_sessions_per_user`) can be enforced at the token level: the
  server tracks active JWTs per user and rejects new logins once the limit is reached, regardless
  of IP. This prevents sharing accounts across a table without blocking legitimate multi-device use.
