# Piggy Tracker — design doc

A Telegram Web App for tracking personal and shared expenses/income. The model is a
"common pot" (like Money Manager): group members add their own transactions, the group
sees the whole picture.

Not Splitwise: there are no debts, no bill splitting and no "who owes whom" settlements
here, and there never will be.

This doc holds the decisions that cut across the whole product. What we build lives in
[`docs/features/`](./features) — one spec per feature, each with its own requirements and
its slice of the data model. A feature is in scope when it has a spec there.

## Key decisions

| Decision | Why |
|---|---|
| A Telegram Web App, not a standalone app | Auth, distribution and trust come for free. No passwords, no signup, no account recovery. |
| Our own groups on top of Telegram, not a binding to chats | The membership of a budget does not match the membership of a chat. We need explicit control over who is inside. |
| A common pot instead of debt splitting | Splitting is a different product with a different schema (shares, settlements, repayments). Others already do it well. |
| The exchange rate is frozen into the transaction | Otherwise yesterday's reports change on their own every time rates update. |
| One input path: a form in the WebApp | Held to a hard speed budget instead of widened. Every extra channel brings its own parser, its own errors and its own bugs. |
| A transaction never leaves the context it was created in | Otherwise joining a group can make a personal history visible to everyone in it. |

## Feature specs

Reading order, from the core mechanic outward — not a priority list.

| Spec | What it covers |
|---|---|
| [transactions](./features/transactions.md) | expense, income, transfer; who may edit and delete |
| [quick-input](./features/quick-input.md) | the 2 taps / 5 seconds budget and the input form |
| [accounts](./features/accounts.md) | personal accounts, balances, archiving |
| [categories](./features/categories.md) | two levels, expense and income, archiving |
| [currencies](./features/currencies.md) | account and transaction currency, frozen rate, rate source |
| [contexts](./features/contexts.md) | first launch, personal and group contexts, permissions, invites |
| [analytics](./features/analytics.md) | charts, periods, breakdowns |
| [budgets](./features/budgets.md) | per-category limits and progress against them |
| [recurring](./features/recurring.md) | templates that generate transactions on a schedule |
| [csv-export](./features/csv-export.md) | exporting a context's transactions |
