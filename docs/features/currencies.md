# Currencies

An account has a currency and so does a transaction. When they differ the transaction is
converted — by a rate that is written into the transaction itself and never recomputed.
Otherwise last year's report changes every morning.

## Requirements

- The rate is frozen into the transaction at creation, `1` when the currencies matched.
- **The rate can be overridden by hand on input.** A real exchange office never works at the
  central bank rate; without this knob the numbers do not match real life.
- Rate source: a daily fetch of a public API, cached in `exchange_rates`.
- API unavailable → the last known rate is used, never a failed save.
- A transfer between accounts in different currencies uses the same rate: `amount` is debited,
  `amount * rate` is credited, and the rate is editable by hand.
- Reports of a context are computed in its `base_currency`, using the frozen rates.

## Data model

```
exchange_rates
  date   date
  base   char(3)
  quote  char(3)
  rate   numeric(18,8)
  PK (date, base, quote)
```
