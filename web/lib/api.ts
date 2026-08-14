import { supabase } from "./supabase";

// The API marshals Go structs with no json tags, so response keys are PascalCase.
export type User = {
  ID: string;
  Email: string;
  Username: string;
  Name: string;
};

export type Me = { onboarded: false; email: string } | { onboarded: true; user: User };

export type Hill = {
  ID: string;
  OwnerID: string;
  Slug: string;
  Title: string;
  Description: string;
  IsPublic: boolean;
  TrackStalled: boolean;
  CreatedAt: string;
  UpdatedAt: string;
};

export type Scope = {
  ID: string;
  Title: string;
  Description: string;
  Color: string;
  SortOrder: number;
  Position: number;
  Note: string;
  MovedAt: string;
};

export type Snapshot = {
  Position: number;
  Note: string;
  CreatedAt: string;
};

export type HillResponse = { hill: Hill; scopes: Scope[] };

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const {
    data: { session },
  } = await supabase.auth.getSession();

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(init?.headers as Record<string, string>),
  };
  if (session?.access_token) headers.Authorization = `Bearer ${session.access_token}`;

  const res = await fetch(path, { ...init, headers });

  if (!res.ok) {
    const raw = await res.text().catch(() => "");
    let message = raw || `${res.status} ${res.statusText}`;
    try {
      const parsed = JSON.parse(raw);
      if (parsed?.error) message = parsed.error;
    } catch {
      // raw wasn't JSON — keep it as-is
    }
    throw new ApiError(res.status, message);
  }

  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export const api = {
  me: () => request<Me>(`/api/me`),

  onboard: (username: string, name?: string) =>
    request<User>(`/api/onboard`, {
      method: "POST",
      body: JSON.stringify({ username, name }),
    }),

  listHills: () => request<Hill[]>(`/api/hills`),

  createHill: (title: string, slug: string) =>
    request<Hill>(`/api/hills`, {
      method: "POST",
      body: JSON.stringify({ title, slug }),
    }),

  getHill: (slug: string) => request<HillResponse>(`/api/hills/${slug}`),

  getPublicHill: (slug: string) => request<HillResponse>(`/api/public/hills/${slug}`),

  publicScopeSnapshots: (scopeId: string) =>
    request<Snapshot[]>(`/api/public/scopes/${scopeId}/positions`),

  updateTitle: (slug: string, title: string) =>
    request<Hill>(`/api/hills/${slug}`, {
      method: "PATCH",
      body: JSON.stringify({ title }),
    }),

  setVisibility: (slug: string, isPublic: boolean) =>
    request<Hill>(`/api/hills/${slug}`, {
      method: "PATCH",
      body: JSON.stringify({ is_public: isPublic }),
    }),

  setTrackStalled: (slug: string, trackStalled: boolean) =>
    request<Hill>(`/api/hills/${slug}`, {
      method: "PATCH",
      body: JSON.stringify({ track_stalled: trackStalled }),
    }),

  addScope: (slug: string, input: { title: string; color: string; sort_order: number }) =>
    request<Scope>(`/api/hills/${slug}/scopes`, {
      method: "POST",
      body: JSON.stringify(input),
    }),

  scopeSnapshots: (scopeId: string) => request<Snapshot[]>(`/api/scopes/${scopeId}/positions`),

  updateScope: (scopeId: string, input: { title: string; color: string }) =>
    request<void>(`/api/scopes/${scopeId}`, {
      method: "PATCH",
      body: JSON.stringify(input),
    }),

  deleteScope: (scopeId: string) => request<void>(`/api/scopes/${scopeId}`, { method: "DELETE" }),

  moveScope: (scopeId: string, position: number, note: string) =>
    request<void>(`/api/scopes/${scopeId}/positions`, {
      method: "POST",
      body: JSON.stringify({ position, note }),
    }),
};
