# CSV export

Data the user entered is theirs and has to be able to leave. Also the answer to every "can it
also show me X" that does not deserve its own screen.

## Requirements

- Exports the transactions of one context for a chosen period.
- Columns: date, type, amount, currency, rate, amount in the context's base currency, account,
  category (parent and subcategory), author, note.
- The file is an ordinary browser download (`Content-Disposition: attachment`), named after the
  context and the period.
- Any member of a context can export it: everyone already sees everything in it.
- UTF-8 with a header row, ISO dates, a dot as the decimal separator.

Depends on: [transactions](./transactions.md), [currencies](./currencies.md).
