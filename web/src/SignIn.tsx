import { useEffect, useRef, useState } from "react";
import { api, ApiError } from "./api";

declare global {
  interface Window {
    onTelegramAuth?: (user: Record<string, string | number>) => void;
  }
}

const botUsername: string | undefined = import.meta.env.VITE_BOT_USERNAME;
const devLogin: string | undefined = import.meta.env.DEV ? import.meta.env.VITE_DEV_LOGIN : undefined;

export function SignIn({ onSignedIn }: { onSignedIn: () => void }) {
  const widget = useRef<HTMLDivElement>(null);
  const [error, setError] = useState<string | null>(null);

  const signIn = (payload: URLSearchParams) => {
    setError(null);
    api
      .signIn(payload)
      .then(onSignedIn)
      .catch((e: ApiError) => setError(e.message));
  };

  useEffect(() => {
    const host = widget.current;
    if (!botUsername || !host) return;
    window.onTelegramAuth = (user) =>
      signIn(new URLSearchParams(Object.entries(user).map(([k, v]) => [k, String(v)])));
    const script = document.createElement("script");
    script.src = "https://telegram.org/js/telegram-widget.js?22";
    script.async = true;
    script.dataset.telegramLogin = botUsername;
    script.dataset.size = "large";
    script.dataset.onauth = "onTelegramAuth(user)";
    host.appendChild(script);
    return () => {
      host.replaceChildren();
      delete window.onTelegramAuth;
    };
  }, []);

  return (
    <main className="state signin">
      <h1>Piggy Tracker</h1>
      <p>Sign in to track your expenses.</p>
      <div ref={widget} />
      {devLogin && (
        <button type="button" className="save" onClick={() => signIn(new URLSearchParams(devLogin))}>
          Dev sign-in
        </button>
      )}
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
    </main>
  );
}
