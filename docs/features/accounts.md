# Accounts

Where the money is. Accounts are personal even inside a group context: a group shares its
spending picture, not its wallets. Without a balances screen there is no way to tell how much
money there even is, so it ships as part of this feature and not as analytics.

## Requirements

- An account has a name, a currency and an initial balance. The initial balance exists so the
  first screen does not claim "0 on your card".
- Balance = `initial_balance + SUM(transactions)`, computed by a query.
  `ponytail: SUM on the fly, denormalize into a column when it gets slow on real volumes.`
- An account belongs to a context and to one user; other members of a group never post to it.
- A balances screen lists every non-archived account with its current balance, plus a total in
  the context's base currency.
- Accounts are archived, never deleted: archived ones disappear from pickers and history stays
  intact.

## Data model

```
accounts
  id               uuid PK
  context_id       uuid -> contexts
  owner_id         bigint -> users
  name             text
  currency         char(3)
  initial_balance  numeric(18,2)
  is_archived      bool
  created_at       timestamptz
```

Depends on: [contexts](./contexts.md), [currencies](./currencies.md).
