# Analytics

Where the money went, over a period, in the context's base currency. The bookkeeping side —
what is on each account right now — belongs to [accounts](./accounts.md).

## Requirements

- A pie chart of expenses by top-level category over a period. This is the primary view.
- A breakdown of one top-level category into its subcategories, opened from a pie slice.
- Income and expense by month, as a comparison over time.
- In a group context, a breakdown by member: who contributed which share of the spending.
- Filters: by category (including a single subcategory), by account, by member.
- Period: the calendar month by default, plus presets — month, 3 months, year, custom range.
- A configurable start-of-month day ("I count from payday"), applied to every period
  calculation, not just the current month.
- All amounts are converted with the rates frozen in the transactions, never with today's rate.

Depends on: [transactions](./transactions.md), [currencies](./currencies.md),
[categories](./categories.md).
