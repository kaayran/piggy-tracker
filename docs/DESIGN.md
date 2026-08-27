# Piggy Tracker — design doc

A Telegram Web App for tracking personal and shared expenses/income. The model is a
"common pot" (like Money Manager): group members add their own transactions, the group
sees the whole picture.

Not Splitwise: there are no debts, no bill splitting and no "who owes whom" settlements
here, and there never will be.

---

## 1. Key decisions

| Decision | Why |
|---|---|
| A Telegram Web App, not a standalone app | Auth, distribution and trust come for free. No passwords, no signup, no account recovery. |
| Our own groups on top of Telegram, not a binding to chats | The membership of a budget does not match the membership of a chat. We need explicit control over who is inside. |
| A common pot instead of debt splitting | Splitting is a different product with a different schema (shares, settlements, repayments). Others already do it well. |
| The exchange rate is frozen into the transaction | Otherwise yesterday's reports change on their own every time rates update. |
| Input is a form in the WebApp | A conscious choice. The risk is known and spelled out in section 8. |

## 2. Contexts

The core entity of the app. A user belongs to one **personal context** (created automatically
on first launch) and to **N group contexts**. The context is picked when entering the app.

A transaction belongs to the context it was created in and **never moves between contexts**.
To change it — delete and recreate. This guards against the scenario "joined a group → my
entire personal spending history instantly became visible to everyone else".

**Visibility inside a group.** All members see all transactions of one another: author,
amount, category, account, note. There are no hidden or private transactions — a deliberate
decision.

**Permissions:**

| Action | Owner | Member |
|---|---|---|
| Create own transactions | yes | yes |
| Edit/delete own | yes | yes |
| Delete someone else's | yes | no |
| Manage the context's categories | yes | no |
| Invite | yes | no |
| Leave the context | no, only delete the context entirely | yes |

**When a member leaves.** Their transactions **stay** in the context, the author is marked
as a "former member". They cannot be deleted: otherwise every group report for past months
shifts retroactively.

**Invites.** A single-use link with a 24-hour TTL. The token and its expiry live right in the
context row (`invite_token`, `invite_expires_at`), there is no separate table. The owner
generates the link, shares it via native Telegram share, and joining happens through the
deep link `t.me/<bot>?startapp=<token>`.

## 3. Data model

```
users
  id              bigint PK          -- telegram user id
  first_name      text
  username        text
  language_code   text
  base_currency   char(3)            -- picked on first launch, defaults from language_code
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

accounts                             -- accounts are personal, even inside a group
  id               uuid PK
  context_id       uuid -> contexts
  owner_id         bigint -> users
  name             text
  currency         char(3)
  initial_balance  numeric(18,2)     -- otherwise the first screen says "0 on your card"
  is_archived      bool
  created_at       timestamptz

categories                           -- exactly 2 levels, no arbitrary tree
  id          uuid PK
  context_id  uuid -> contexts
  parent_id   uuid NULL -> categories
  kind        text                   -- expense | income
  name        text
  icon        text
  is_archived bool

transactions
  id             uuid PK
  context_id     uuid -> contexts
  author_id      bigint -> users
  type           text                -- expense | income | transfer
  account_id     uuid -> accounts    -- from (expense/transfer), to (income)
  to_account_id  uuid NULL           -- transfer only
  category_id    uuid NULL           -- NULL for transfer
  amount         numeric(18,2)
  currency       char(3)
  rate           numeric(18,8)       -- frozen at creation time, 1 if the currencies matched
  occurred_on    date
  note           text
  created_at     timestamptz

exchange_rates
  date   date
  base   char(3)
  quote  char(3)
  rate   numeric(18,8)
  PK (date, base, quote)
```

**Categories are two-level.** A parent plus subcategories: `Food → Delivery, Groceries,
Restaurants`. Only the context owner creates and edits them, everyone else picks from what
already exists. A transaction may reference **any node** — a parent as well as a subcategory:
requiring a leaf means slowing input down, and input is already at risk (section 8).

Categories are split into expense and income ones, otherwise "Salary" ends up in the
expense pie.

**Categories are never deleted** — only `is_archived`: it disappears from the picker, history
stays intact.

**Account balance** = `initial_balance + SUM(transactions)`, computed by a query.
`ponytail: SUM on the fly, denormalize into a column when it gets slow on real volumes.`

## 4. Operation types

- `expense` — a debit from an account, an expense category is required.
- `income` — a credit to an account, an income category is required.
- `transfer` — a move between the user's **own** accounts. No category, and it does not enter
  the expense or income charts at all.

A transfer is not an optional feature. Without it "withdrew cash from the card" gets entered
as an expense plus an income, and all the analytics lie.

## 5. Currencies

Both an account and a transaction have a currency. If they differ — conversion by a rate that
is **written into the transaction itself and never recomputed**.

- Rate source: a daily fetch of a public API, cached in `exchange_rates`.
- API unavailable → the last known rate is used.
- **The rate can be overridden manually on input.** A real exchange office never works at the
  central bank rate; without this knob the numbers will not match real life.
- Context reports are computed in its `base_currency`.

A transfer between accounts in different currencies uses the same `rate`: `amount` is debited,
`amount * rate` is credited, and the rate is editable by hand.

## 6. Transaction input

**Requirement: 2 taps and no more than 5 seconds for a typical expense.** A number, not a
wish — it is checked at review.

Field order: `date → amount → category → account → note`.

- date — today by default, changed in one tap;
- amount — the screen opens straight onto the numeric keypad;
- category — a grid of icons, most recently used first;
- account — the last used one is preselected;
- note — optional, collapsed;
- saving — Telegram's native `MainButton`.

## 7. Analytics in v1

**One chart: a pie chart of expenses by top-level category over a period.**
Subcategories are not broken out in v1 — they exist in the data and in the filters.

Period: the calendar month by default plus presets (month / 3 months / year / custom).
A configurable start-of-month day ("I count from payday") is backlog, it drags a
recomputation of every period along with it.

Plus an **account balances** screen — that is not analytics but part of the bookkeeping:
without it there is no way to tell how much money there even is.

## 8. Known risk

Input happens only through a form in the WebApp. Expense trackers die not from bad charts but
from nobody entering anything two weeks in: open Telegram → find the bot → wait for the WebApp
to load → fill in a form. In a group context the risk multiplies — one member who stops
contributing makes the shared numbers untrustworthy, and therefore useless.

The v1 mitigation is the hard budget of 2 taps / 5 seconds (section 6). If retention drops
anyway, the first thing pulled from the backlog is text input to the bot (`500 coffee` as a
single chat message, without opening the app).

## 9. Out of scope for v1

Budgets and per-category limits · recurring payments · reminders and notifications ·
CSV import/export · offline mode · receipt photos and OCR · tags · search · debts and splits ·
shared group accounts · transferring group ownership · bank integrations · text input to the
bot · i18n (Russian only) · configurable start-of-month day · charts other than the pie.

## 10. Backlog, in the order it comes back

1. Budgets and per-category limits — the main retention driver.
2. Recurring payments — they remove the dullest manual input.
3. Text input to the bot — if the risk from section 8 materializes.
4. Charts: income/expense by month, a breakdown by group member.
5. CSV export.
6. Transferring group ownership.
