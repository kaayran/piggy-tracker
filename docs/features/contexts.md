# Contexts

The core entity. A user belongs to one personal context and to N group contexts; the context is
picked when entering the app. A group context is our own object, not a binding to a Telegram
chat — who is in a budget is not who is in a chat.

## Requirements

**First launch.** A user row is created from Telegram's `initData`. `base_currency` is guessed
from `language_code` and can be changed. A personal context is created automatically — the user
never sees an empty app asking them to set something up first.

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
row — no separate table. The owner generates the link, shares it through native Telegram share,
and joining happens via the deep link `t.me/<bot>?startapp=<token>`.

## Data model

```
users
  id              bigint PK          -- telegram user id
  first_name      text
  username        text
  language_code   text
  base_currency   char(3)
  created_at      timestamptz

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
