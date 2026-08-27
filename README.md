# Piggy Tracker

Telegram Web App for tracking expenses and income — personal and shared. Transactions,
accounts, two-level categories, multi-currency, groups with a common budget.

Product decisions, data model and v1 scope are in [DESIGN.md](./docs/DESIGN.md).

## Stack

| Layer | What |
|---|---|
| Frontend | React + Vite + TypeScript, `telegram-web-app.js` included via script tag, theming through native `--tg-theme-*` CSS variables |
| Backend | Go, `net/http` + `pgx` |
| DB | PostgreSQL, migrations with `goose` |
| Contract | OpenAPI as the source of truth: `oapi-codegen` for Go, `openapi-typescript` for the frontend |
| Auth | HMAC verification of Telegram `initData`. No passwords, no OAuth, no sessions, no JWT |
| Deploy | A single Go binary: API + frontend static files via `embed`. One container, one domain, zero CORS |

No UI kit on purpose: Telegram hands us the theme as CSS variables, and our own buttons are
cheaper than someone else's design system layered on top of someone else's theme.

## Layout

```
/api            Go: HTTP API + migrations
/web            React + Vite
/api/openapi.yaml   the contract; types for both sides are generated from it
```

## Running

TODO — arrives together with the code.

You will need: Go, Node, Docker (for Postgres) and a bot from @BotFather. Telegram will not
open `localhost`, so development goes through a real HTTPS host — see
`docs/INFRA.local.md` for the server layout, the dev tunnel and deploys.

## Status

No code yet. Decisions and scope are locked — see DESIGN.md, sections 9 and 10:
what is out of v1 and in which order it comes back.
