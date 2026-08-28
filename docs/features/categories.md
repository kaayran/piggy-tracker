# Categories

What the money went on. Exactly two levels — a parent plus subcategories, `Food → Delivery,
Groceries, Restaurants`. Not an arbitrary tree: a tree needs a tree picker, and a tree picker
does not fit in the input budget.

## Requirements

- Two levels and no more. A subcategory cannot have children.
- A category is either an expense or an income one, otherwise "Salary" lands in the expense pie.
- A transaction may reference any node, parent or subcategory. Requiring a leaf slows input down.
- Only the context owner creates and edits categories; everyone else picks from what exists.
- Categories are archived, never deleted: gone from the picker, intact in history.
- Every context starts with a default set, so the first transaction does not begin with setup.

## Data model

```
categories
  id          uuid PK
  context_id  uuid -> contexts
  parent_id   uuid NULL -> categories
  kind        text                   -- expense | income
  name        text
  icon        text
  is_archived bool
```

Depends on: [contexts](./contexts.md).
