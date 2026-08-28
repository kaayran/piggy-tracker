# Quick input

Entering an expense has to be faster than deciding not to. This is the requirement the rest of
the form design answers to, and it is the reason no field gets added to the default path
casually.

## Requirements

- **A typical expense: 2 taps, no more than 5 seconds.** A number, not a wish — it is checked
  at review, and a change that breaks it does not ship.
- Field order: `date → amount → category → account → note`.
- Date defaults to today and is changed in one tap.
- The screen opens straight onto the numeric keypad, cursor in the amount.
- Category is a grid of icons, most recently used first.
- Account is preselected to the last used one.
- Note is optional and collapsed.
- Saving is Telegram's native `MainButton`.
- Anything that cannot fit in the budget goes behind an explicit "more" step, never into the
  default path.

Depends on: [transactions](./transactions.md), [accounts](./accounts.md),
[categories](./categories.md).
