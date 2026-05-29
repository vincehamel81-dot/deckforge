# DeckForge — Setup Reference

Quick-start supplement to the README. Covers admin account creation and a full env var reference.

---

## Admin account

A platform admin can view all hands and delete any game. Regular users become dealers only within
their own games — `dealer` is a per-game relationship, not a JWT claim.

**Create an admin on first boot:**

```bash
# backend/.env
ADMIN_SEED_USERNAME=admin
```

On startup, DeckForge checks whether a user with that username exists. If not, it creates one with
`role=admin`. Subsequent restarts are idempotent — the check runs every boot, but no duplicate is
created.

To obtain an admin token, call `POST /auth/login` with `{ "username": "admin" }`. The returned JWT
contains `"role": "admin"`.

**No public admin signup.** Admin accounts are only created via this seed mechanism.

---

## Full environment variable reference

### Backend (`backend/.env`)

| Variable | Default | Required | Description |
|---|---|---|---|
| `JWT_SECRET` | — | **Yes** | HMAC-SHA256 signing secret. Any long random string. |
| `PORT` | `8080` | No | HTTP listener port. |
| `DATABASE_URL` | `deckforge.db` | No | SQLite file path, or PostgreSQL DSN when `DB_DRIVER=postgres`. |
| `JWT_EXPIRY` | `24h` | No | Token lifetime. Accepts Go duration strings: `1h`, `24h`, `7d`. |
| `CORS_ORIGIN` | `http://localhost:5173` | No | Allowed `Origin` for CORS. Lock to your frontend URL in production. |
| `MIN_PLAYERS` | `2` | No | Default minimum players when creating a game. |
| `MAX_PLAYERS` | `8` | No | Default maximum players when creating a game. |
| `ADMIN_SEED_USERNAME` | *(empty)* | No | Creates an admin user on first boot. Idempotent. |
| `ENV` | `development` | No | Set to `production` for JSON-structured logs (zerolog). |

### Frontend (`frontend/.env`)

| Variable | Default | Description |
|---|---|---|
| `VITE_API_URL` | `http://localhost:8080` | Backend base URL. Change when deploying. |

---

## Switching to PostgreSQL

```bash
# backend/.env
DATABASE_URL=postgres://user:pass@localhost:5432/deckforge?sslmode=disable
```

No code changes required. GORM's Repository interface abstracts the driver. Run `go run ./cmd/server`
and AutoMigrate creates the schema on first boot.

---

## Generating a JWT_SECRET

```bash
# macOS / Linux
openssl rand -hex 32

# Windows (PowerShell)
[System.Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Maximum 256 }))
```

Use any output as `JWT_SECRET`. Never commit this value — keep it in `.env` (gitignored).
