import type { components } from "./types.gen";

export type Me = components["schemas"]["Me"];
export type Account = components["schemas"]["Account"];
export type Category = components["schemas"]["Category"];
export type Transaction = components["schemas"]["Transaction"];
export type NewTransaction = components["schemas"]["NewTransaction"];

export class ApiError extends Error {
  constructor(readonly code: string, message: string) {
    super(message);
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const response = await fetch(`/api${path}`, {
    ...init,
    headers: typeof init.body === "string" ? { "Content-Type": "application/json" } : {},
  });
  if (response.status === 204) return undefined as T;
  if (!response.ok) {
    const body = await response.json().catch(() => null);
    throw new ApiError(body?.code ?? "network", body?.message ?? response.statusText);
  }
  return response.json() as Promise<T>;
}

export const api = {
  signIn: (payload: URLSearchParams) =>
    request<void>("/auth/telegram", { method: "POST", body: payload }),
  signOut: () => request<void>("/auth/logout", { method: "POST" }),
  me: () => request<Me>("/me"),
  transactions: () => request<Transaction[]>("/transactions"),
  create: (body: NewTransaction) =>
    request<Transaction>("/transactions", { method: "POST", body: JSON.stringify(body) }),
  remove: (id: string) => request<void>(`/transactions/${id}`, { method: "DELETE" }),
};
