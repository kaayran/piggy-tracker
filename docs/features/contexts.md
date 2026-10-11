# Contexts

The core entity. A user belongs to one personal context and to N group contexts; the context is
picked when entering the app. A group context is our own object with explicit membership: who is
in a budget is decided by its owner, nothing else.

## Requirements

**First launch.** The user row comes from the first sign-in (see [auth](./auth.md)).
`base_currency` is guessed from the browser's `Accept-Language` and can be changed. A personal
context is created automatically — the user never sees an empty app asking them to set something
up first.

**Visibility.** All members of a group see all transactions of one another: author, amount,
category, account, note. There are no hidden or private transactions.

**Permissions.**

| Action | Owner | Member |
|---|---|---|
| Create own transactions | yes | yes |
| Edit/delete own | yes | yes |
| Delete someone else's | yes | no |
| Manage the context's categories | yes | no |
| Invite | yes | no |
| Leave the context | no, only delete the context entirely | yes |

**Leaving.** A member's transactions stay in the context and the author is shown as a former
member. They cannot be deleted: otherwise every group report for past months shifts
retroactively.

**Ownership.** Exactly one owner at all times. The owner can hand the context to another active
member; a context is never left ownerless.

**Invites.** A single-use link with a 24-hour TTL. The token and its expiry live in the context
row — no separate table. The owner generates the link `https://<host>/invite/<token>` and shares
it through the Web Share API, with copy-to-clipboard where sharing is unavailable. Opening the
link signed out goes through sign-in first and then joins.

## Data model

```
users -- see auth

contexts
  id                 uuid PK
  kind               text            -- personal | group
  name               text
  owner_id           bigint -> users
  base_currency      char(3)         -- the currency all reports of the context are computed in
  invite_token       text NULL
  invite_expires_at  timestamptz NULL
  created_at         timestamptz

context_members
  context_id  uuid -> contexts
  user_id     bigint -> users
  status      text                   -- active | left
  joined_at   timestamptz
  PK (context_id, user_id)
```

Depends on: [auth](./auth.md).
