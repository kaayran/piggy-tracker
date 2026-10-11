import { useEffect, useState } from "react";
import { api, ApiError, type Me, type Transaction } from "./api";
import { AddForm } from "./AddForm";
import { SignIn } from "./SignIn";
import { TxList } from "./TxList";

export function App() {
  const [me, setMe] = useState<Me | null>(null);
  const [signedOut, setSignedOut] = useState(false);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [error, setError] = useState<string | null>(null);

  const load = () => {
    setSignedOut(false);
    setError(null);
    Promise.all([api.me(), api.transactions()])
      .then(([loadedMe, loaded]) => {
        setMe(loadedMe);
        setTransactions(loaded);
      })
      .catch((e: ApiError) => (e.code === "unauthorized" ? setSignedOut(true) : setError(e.message)));
  };

  useEffect(load, []);

  const signOut = () =>
    api
      .signOut()
      .then(() => {
        setMe(null);
        setTransactions([]);
        setSignedOut(true);
      })
      .catch((e: ApiError) => setError(e.message));

  if (signedOut) {
    return <SignIn onSignedIn={load} />;
  }
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
      <header className="top">
        <span>{me.user.first_name}</span>
        <button type="button" className="more" onClick={signOut}>
          Sign out
        </button>
      </header>
      <AddForm me={me} onCreated={(t) => setTransactions((all) => [t, ...all])} />
      <TxList
        me={me}
        transactions={transactions}
        onDeleted={(id) => setTransactions((all) => all.filter((t) => t.id !== id))}
      />
    </main>
  );
}
