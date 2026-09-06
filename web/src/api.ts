import type { components } from "./types.gen";

export type Me = components["schemas"]["Me"];
export type Account = components["schemas"]["Account"];
export type Category = components["schemas"]["Category"];
export type Transaction = components["schemas"]["Transaction"];
export type NewTransaction = components["schemas"]["NewTransaction"];

export const tg = window.Telegram?.WebApp;

export const inTelegram = Boolean(tg?.initData);

// Outside Telegram there is no initData; a genuinely signed one from .env stands in for
// local work. The server verifies it exactly like any other request.
const initData = tg?.initData || import.meta.env.VITE_DEV_INIT_DATA || "";

export class ApiError extends Error {
  constructor(readonly code: string, message: string) {
    super(message);
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const response = await fetch(`/api${path}`, {
    ...init,
    headers: {
      "X-Telegram-Init-Data": initData,
      ...(init.body ? { "Content-Type": "application/json" } : {}),
    },
  });
  if (response.status === 204) return undefined as T;
  if (!response.ok) {
    const body = await response.json().catch(() => null);
    throw new ApiError(body?.code ?? "network", body?.message ?? response.statusText);
  }
  return response.json() as Promise<T>;
}

export const api = {
  me: () => request<Me>("/me"),
  transactions: () => request<Transaction[]>("/transactions"),
  create: (body: NewTransaction) =>
    request<Transaction>("/transactions", { method: "POST", body: JSON.stringify(body) }),
  remove: (id: string) => request<void>(`/transactions/${id}`, { method: "DELETE" }),
};
