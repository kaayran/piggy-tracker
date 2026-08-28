# Recurring transactions

Rent, subscriptions, salary — the entries that are known in advance and are the dullest to type.
A template describes one, and the app creates the transaction on schedule.

## Requirements

- A template holds everything a transaction needs — type, amount, currency, account, category,
  note — plus a period (monthly, weekly, yearly) and a start date.
- On the due date the template creates an ordinary transaction, indistinguishable from a
  hand-entered one and editable on its own.
- A template can be paused, edited and deleted. Deleting it never touches transactions it has
  already created.
- Editing a template affects future transactions only.
- Missed dates are caught up on the next launch, so a template does not silently skip a month.
- A template belongs to a context and to the user who created it; in a group it posts on their
  behalf.

Depends on: [transactions](./transactions.md), [accounts](./accounts.md),
[categories](./categories.md).
