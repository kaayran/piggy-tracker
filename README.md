# Piggy Tracker

A mobile-first website for tracking expenses and income — personal and shared. Transactions,
accounts, two-level categories, multi-currency, groups with a common budget.

Cross-cutting decisions are in [DESIGN.md](./docs/DESIGN.md); what we build — feature specs
with their slice of the data model — in [docs/features](./docs/features).

## Stack

| Layer    | What                                                                                                                          |
| -------- | ----------------------------------------------------------------------------------------------------------------------------- |
| Frontend | React + Vite + TypeScript, own CSS with light/dark tokens via `prefers-color-scheme`                                         |
| Backend  | Go, `net/http` + `pgx`                                                                                                        |
| DB       | PostgreSQL, migrations with `goose`                                                                                           |
| Contract | OpenAPI as the source of truth: `oapi-codegen` for Go, `openapi-typescript` for the frontend                                  |
| Auth     | Telegram Login Widget, HMAC-verified on the server; opaque session token in an HttpOnly cookie, stored hashed. No JWT        |
| Deploy   | A single Go binary: API + frontend static files via `embed`. One container, one domain, zero CORS                             |

## Layout

```
/api                    Go: HTTP API + migrations
/web                    React + Vite
/api/openapi.yaml       the contract; types for both sides are generated from it
```

## Running locally

```sh
cp .env.example .env      # fill in BOT_TOKEN and VITE_BOT_USERNAME of the dev bot
make dev                  # Postgres in Docker, Go on :8080, Vite on :5173 → http://localhost:5173
```

The login widget only renders on the domain set for the bot in BotFather (`/setdomain`). On
`localhost` the sign-in screen offers three dev profiles instead — Ann and Bob for collaboration,
Carl as a solo user (see [auth](./docs/features/auth.md#dev-profiles)). Their payloads are signed
with the real bot token by the Vite dev server and verified by the API like any other; there is
no auth bypass.

Cookies are per host, so two profiles can be signed in side by side: one on
`http://localhost:5173`, the other on `http://127.0.0.1:5173`, or in a private window.

Migrations run from the binary at startup.

`make gen` regenerates the types on both sides from `api/openapi.yaml`, `make test` runs the
Go tests, `make build` produces a single binary with the frontend embedded.
