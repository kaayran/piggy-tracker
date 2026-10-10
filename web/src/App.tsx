import { useEffect, useState } from "react";
import { api, ApiError, type Me, type Transaction } from "./api";
import { AddForm } from "./AddForm";
import { TxList } from "./TxList";

export function App() {
  const [me, setMe] = useState<Me | null>(null);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([api.me(), api.transactions()])
      .then(([loadedMe, loaded]) => {
        setMe(loadedMe);
        setTransactions(loaded);
      })
      .catch((e: ApiError) => setError(e.message));
  }, []);

  if (error) {
    return (
      <main className="state" role="alert">
        <p>{error}</p>
      </main>
    );
  }
  if (!me) {
    return (
      <main className="state" aria-busy="true">
        <p>Loading…</p>
      </main>
    );
  }

  return (
    <main>
      <AddForm me={me} onCreated={(t) => setTransactions((all) => [t, ...all])} />
      <TxList
        me={me}
        transactions={transactions}
        onDeleted={(id) => setTransactions((all) => all.filter((t) => t.id !== id))}
      />
    </main>
  );
}
