# Transactions

The thing the product exists for. A transaction records money moving: out of an account, into
an account, or between two accounts. Everything else — categories, balances, charts — is built
on top of this one record.

## Requirements

- Three types: `expense` debits an account, `income` credits an account, `transfer` moves money
  between two accounts of the same user.
- `expense` requires an expense category, `income` requires an income category, `transfer` has
  no category and does not enter the expense or income charts at all.
- Transfer is not optional. Without it "withdrew cash from the card" is entered as an expense
  plus an income and every chart lies.
- A transaction may reference any category node — a parent as well as a subcategory.
- Amount is stored positive; the type determines the direction.
- A transaction belongs to the context it was created in and never moves between contexts. To
  change the context: delete and recreate.
- The author may edit and delete their own transactions. The context owner may delete anyone's.
- Deleting is real deletion, not a flag — except for transactions of a member who has left the
  context, which stay untouched.

## Data model

```
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
```

Depends on: [accounts](./accounts.md), [categories](./categories.md),
[currencies](./currencies.md), [contexts](./contexts.md).
