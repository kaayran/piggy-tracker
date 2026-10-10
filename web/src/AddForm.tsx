import { useEffect, useRef, useState } from "react";
import { api, ApiError, inTelegram, tg, type Category, type Me, type Transaction } from "./api";

const today = () => new Date().toISOString().slice(0, 10);

const recent = {
  read: (): string[] => {
    try {
      return JSON.parse(localStorage.getItem("recent-categories") ?? "[]");
    } catch {
      return [];
    }
  },
  push(id: string) {
    const next = [id, ...this.read().filter((x) => x !== id)].slice(0, 12);
    localStorage.setItem("recent-categories", JSON.stringify(next));
  },
};

function byRecent(categories: Category[], order: string[]) {
  const rank = (c: Category) => {
    const i = order.indexOf(c.id);
    return i === -1 ? order.length : i;
  };
  return [...categories].sort((a, b) => rank(a) - rank(b));
}

export function AddForm({ me, onCreated }: { me: Me; onCreated: (t: Transaction) => void }) {
  const [kind, setKind] = useState<"expense" | "income">("expense");
  const [amount, setAmount] = useState("");
  const [categoryId, setCategoryId] = useState<string | null>(null);
  const [accountId, setAccountId] = useState(me.accounts[0]?.id ?? "");
  const [date, setDate] = useState(today());
  const [note, setNote] = useState("");
  const [showNote, setShowNote] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const categories = byRecent(
    me.categories.filter((c) => c.kind === kind),
    recent.read(),
  );

  const ready = Number(amount) > 0 && categoryId !== null && accountId !== "" && !saving;

  const submit = useRef(() => {});
  submit.current = async () => {
    if (!ready) return;
    setSaving(true);
    setError(null);
    try {
      const created = await api.create({
        type: kind,
        account_id: accountId,
        category_id: categoryId!,
        amount: Number(amount).toFixed(2),
        occurred_on: date,
        note: note.trim() || undefined,
      });
      recent.push(categoryId!);
      onCreated(created);
      setAmount("");
      setNote("");
      setShowNote(false);
      setDate(today());
      tg?.HapticFeedback.notificationOccurred("success");
    } catch (e) {
      setError((e as ApiError).message);
      tg?.HapticFeedback.notificationOccurred("error");
    } finally {
      setSaving(false);
    }
  };

  useEffect(() => {
    const button = inTelegram ? tg!.MainButton : null;
    if (!button) return;
    const handler = () => submit.current();
    button.setText("Save");
    button.show();
    button.onClick(handler);
    return () => {
      button.offClick(handler);
      button.hide();
    };
  }, []);

  useEffect(() => {
    const button = inTelegram ? tg!.MainButton : null;
    if (!button) return;
    if (ready) button.enable();
    else button.disable();
  }, [ready]);

  return (
    <form
      className="add"
      onSubmit={(e) => {
        e.preventDefault();
        submit.current();
      }}
    >
      <div className="kinds" role="group" aria-label="Transaction type">
        {(["expense", "income"] as const).map((k) => (
          <button
            key={k}
            type="button"
            className={k === kind ? "kind on" : "kind"}
            aria-pressed={k === kind}
            onClick={() => {
              setKind(k);
              setCategoryId(null);
            }}
          >
            {k === "expense" ? "Expense" : "Income"}
          </button>
        ))}
      </div>

      <label className="amount">
        <span className="sr-only">Amount</span>
        <input
          autoFocus
          inputMode="decimal"
          placeholder="0"
          value={amount}
          onChange={(e) => setAmount(e.target.value.replace(/[^\d.]/g, "").replace(/(\..*)\./g, "$1"))}
        />
        <span aria-hidden="true">{me.accounts.find((a) => a.id === accountId)?.currency}</span>
      </label>

      <ul className="categories" aria-label="Category">
        {categories.map((c) => (
          <li key={c.id}>
            <button
              type="button"
              className={c.id === categoryId ? "category on" : "category"}
              aria-pressed={c.id === categoryId}
              onClick={() => setCategoryId(c.id)}
            >
              <span className="icon" aria-hidden="true">
                {c.icon}
              </span>
              <span className="label">{c.name}</span>
            </button>
          </li>
        ))}
      </ul>

      <div className="row">
        <label>
          <span className="sr-only">Date</span>
          <input type="date" value={date} max={today()} onChange={(e) => setDate(e.target.value)} />
        </label>
        {me.accounts.length > 1 && (
          <label>
            <span className="sr-only">Account</span>
            <select value={accountId} onChange={(e) => setAccountId(e.target.value)}>
              {me.accounts.map((a) => (
                <option key={a.id} value={a.id}>
                  {a.name}
                </option>
              ))}
            </select>
          </label>
        )}
        {showNote ? (
          <input
            className="note"
            placeholder="Note"
            maxLength={500}
            value={note}
            onChange={(e) => setNote(e.target.value)}
          />
        ) : (
          <button type="button" className="more" onClick={() => setShowNote(true)}>
            + Note
          </button>
        )}
      </div>

      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}

      {!inTelegram && (
        <button type="submit" className="save" disabled={!ready}>
          Save
        </button>
      )}
    </form>
  );
}
