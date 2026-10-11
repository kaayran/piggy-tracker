# Auth

Who is on the other end of a request. Sign-in goes through the Telegram Login Widget: every user
of the service already has Telegram, so there are no passwords to store, no signup form and no
account recovery of our own. Telegram is only the identity provider — the product itself is the
website.

## Requirements

- **Sign in.** One button, the official "Log in with Telegram" widget. The widget hands the
  browser a signed payload (`id`, `first_name`, `last_name`, `username`, `photo_url`,
  `auth_date`, `hash`), the browser posts it to the API as-is.
- **Verification.** The server checks the HMAC as Telegram specifies — key `SHA256(bot_token)`
  over the sorted `key=value` lines — and rejects a payload older than 24 hours. Nothing about a
  request's identity comes from anywhere else.
- **First sign-in is signup.** The user row is created from the payload, keyed by the Telegram
  user id. Name and username are refreshed on every sign-in.
- **Session.** A random opaque token in an HttpOnly, `SameSite=Lax` cookie (`Secure` over
  HTTPS). Only its SHA-256 is stored, so a leaked database does not leak sessions. Lifetime is
  30 days, no sliding renewal.
- **CSRF.** Covered by `SameSite=Lax`: browsers do not attach the cookie to cross-site `POST`
  and `DELETE`, and no state-changing endpoint is a `GET`. The sign-in endpoint itself refuses
  cross-site requests (`Sec-Fetch-Site`), so nobody can sign a visitor into a foreign account.
- **Sign out** deletes the session row and clears the cookie.
- **Signed out** → every API call answers `401`, and the site shows the sign-in screen.
- The bot's domain is set in BotFather (`/setdomain`); the widget works only there. No auth
  bypass in development: work happens on the dev domain with the dev bot, or locally with a
  payload genuinely signed by the dev bot's token (`make devlogin`).

## Data model

```
users
  id              bigint PK          -- Telegram user id
  first_name      text
  username        text NULL
  base_currency   char(3)
  created_at      timestamptz

sessions
  token_hash  bytea PK               -- sha256 of the cookie value
  user_id     bigint -> users
  expires_at  timestamptz
```

`ponytail: expired sessions are swept on each sign-in, move to a scheduled job if the table grows.`
