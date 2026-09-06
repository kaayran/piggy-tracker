import { api, type Me, type Transaction } from "./api";

const dayFormat = new Intl.DateTimeFormat(undefined, { day: "numeric", month: "long" });

function money(amount: string, currency: string) {
  return new Intl.NumberFormat(undefined, {
    style: "currency",
    currency,
    maximumFractionDigits: 2,
  }).format(Number(amount));
}

function groupByDay(transactions: Transaction[]) {
  const sorted = [...transactions].sort(
    (a, b) =>
      b.occurred_on.localeCompare(a.occurred_on) || b.created_at.localeCompare(a.created_at),
  );
  const days = new Map<string, Transaction[]>();
  for (const t of sorted) {
    days.set(t.occurred_on, [...(days.get(t.occurred_on) ?? []), t]);
  }
  return [...days];
}

export function TxList({
  me,
  transactions,
  onDeleted,
}: {
  me: Me;
  transactions: Transaction[];
  onDeleted: (id: string) => void;
}) {
  if (transactions.length === 0) {
    return <p className="state">Nothing this month yet.</p>;
  }

  const categoryOf = (id: string | null | undefined) =>
    me.categories.find((c) => c.id === id);

  return (
    <section className="history" aria-label="This month">
      {groupByDay(transactions).map(([day, ofDay]) => (
        <div key={day}>
          <h2>{dayFormat.format(new Date(day))}</h2>
          <ul>
            {ofDay.map((t) => {
              const category = categoryOf(t.category_id);
              return (
                <li key={t.id}>
                  <span className="icon" aria-hidden="true">
                    {category?.icon ?? "•"}
                  </span>
                  <span className="what">
                    <span>{category?.name ?? "—"}</span>
                    {t.note && <small>{t.note}</small>}
                  </span>
                  <span className={t.type === "income" ? "sum in" : "sum"}>
                    {t.type === "income" ? "+" : "−"}
                    {money(t.amount, t.currency)}
                  </span>
                  <button
                    type="button"
                    className="delete"
                    aria-label={`Delete ${category?.name ?? "transaction"}`}
                    onClick={() => api.remove(t.id).then(() => onDeleted(t.id))}
                  >
                    ✕
                  </button>
                </li>
              );
            })}
          </ul>
        </div>
      ))}
    </section>
  );
}
