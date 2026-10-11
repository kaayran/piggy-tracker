# Piggy Tracker — design doc

A mobile-first website for tracking personal and shared expenses/income. The model is a
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
| A website, not a native app | One codebase, no store review, opens from a link on any device. |
| Mobile-first layout | Expenses happen away from the desk; the phone screen is the one the input budget is measured on. Desktop works, it is not the target. |
| Sign-in through Telegram Login | Every user already has Telegram. No passwords to store, no signup form, no account recovery of our own. Telegram is the identity provider and nothing more. |
| Groups with explicit membership | Who is inside a budget is decided by its owner through invites, never inferred from anything else. |
| A common pot instead of debt splitting | Splitting is a different product with a different schema (shares, settlements, repayments). Others already do it well. |
| The exchange rate is frozen into the transaction | Otherwise yesterday's reports change on their own every time rates update. |
| One input path: a form on the site | Held to a hard speed budget instead of widened. Every extra channel brings its own parser, its own errors and its own bugs. |
| A transaction never leaves the context it was created in | Otherwise joining a group can make a personal history visible to everyone in it. |

## Feature specs

Reading order, from the core mechanic outward — not a priority list.

| Spec | What it covers |
|---|---|
| [auth](./features/auth.md) | Telegram Login sign-in, sessions, sign-out |
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
