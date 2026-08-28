# Budgets

A limit on a category for a period, and how much of it is already spent. Charts say what
happened; a budget says whether it was too much — that is the difference between looking at a
tracker once and using it.

## Requirements

- A limit is set on a category — a parent or a subcategory — for a period, in the context's
  base currency.
- A limit on a parent counts spending in its subcategories too.
- Progress is shown as spent against the limit for the current period.
- Exceeding a limit is shown, never enforced: the app does not refuse to record money that was
  actually spent.
- In a group context a limit belongs to the context and covers all members together — the pot
  is common, so the limit is too.
- Limits carry over to the next period unchanged until edited.

Depends on: [categories](./categories.md), [transactions](./transactions.md),
[analytics](./analytics.md) (period boundaries, including the configurable start-of-month day).
