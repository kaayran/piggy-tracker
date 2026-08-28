# Piggy Tracker

Telegram Web App for tracking expenses and income — personal and shared. Transactions,
accounts, two-level categories, multi-currency, groups with a common budget.

Cross-cutting decisions are in [DESIGN.md](./docs/DESIGN.md); what we build — feature specs
with their slice of the data model — in [docs/features](./docs/features).

## Stack

| Layer    | What                                                                                                                          |
| -------- | ----------------------------------------------------------------------------------------------------------------------------- |
| Frontend | React + Vite + TypeScript, `telegram-web-app.js` included via script tag, theming through native `--tg-theme-*` CSS variables |
| Backend  | Go, `net/http` + `pgx`                                                                                                        |
| DB       | PostgreSQL, migrations with `goose`                                                                                           |
| Contract | OpenAPI as the source of truth: `oapi-codegen` for Go, `openapi-typescript` for the frontend                                  |
| Auth     | HMAC verification of Telegram `initData`. No passwords, no OAuth, no sessions, no JWT                                         |
| Deploy   | A single Go binary: API + frontend static files via `embed`. One container, one domain, zero CORS                             |

## Layout

```
/api                    Go: HTTP API + migrations
/web                    React + Vite
/api/openapi.yaml       the contract; types for both sides are generated from it
```
