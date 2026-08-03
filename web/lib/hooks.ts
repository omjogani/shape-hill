"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api, type Hill, type HillResponse } from "./api";

const key = (slug: string) => ["hill", slug] as const;

// Optimistic cache edits: run `fn` over the scopes list, leaving an empty cache untouched.
const patchScopes =
  (fn: (scopes: HillResponse["scopes"]) => HillResponse["scopes"]) =>
  (old: HillResponse | undefined) => (old ? { ...old, scopes: fn(old.scopes) } : old);

export function useHills() {
  return useQuery({ queryKey: ["hills"], queryFn: () => api.listHills() });
}

export function useHill(slug: string) {
  return useQuery({ queryKey: key(slug), queryFn: () => api.getHill(slug) });
}

export function useCreateHill() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (v: { title: string; slug: string }) => api.createHill(v.title, v.slug),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["hills"] }),
  });
}

export function useScopeSnapshots(scopeId: string, enabled: boolean) {
  return useQuery({
    queryKey: ["snapshots", scopeId],
    queryFn: () => api.scopeSnapshots(scopeId),
    enabled,
  });
}

export function useUpdateTitle(slug: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (title: string) => api.updateTitle(slug, title),
    onSuccess: (hill: Hill) =>
      qc.setQueryData<HillResponse>(key(slug), (prev) => (prev ? { ...prev, hill } : prev)),
  });
}

export function useSetVisibility(slug: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (isPublic: boolean) => api.setVisibility(slug, isPublic),
    onSuccess: (hill: Hill) =>
      qc.setQueryData<HillResponse>(key(slug), (prev) => (prev ? { ...prev, hill } : prev)),
  });
}

export function useAddScope(slug: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: { title: string; color: string; sort_order: number }) =>
      api.addScope(slug, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: key(slug) }),
  });
}

export function useUpdateScope(slug: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (v: { scopeId: string; title: string; color: string }) =>
      api.updateScope(v.scopeId, { title: v.title, color: v.color }),
    onMutate: async (v) => {
      await qc.cancelQueries({ queryKey: key(slug) });
      const prev = qc.getQueryData<HillResponse>(key(slug));
      qc.setQueryData<HillResponse>(
        key(slug),
        patchScopes((scopes) =>
          scopes.map((s) => (s.ID === v.scopeId ? { ...s, Title: v.title, Color: v.color } : s)),
        ),
      );
      return { prev };
    },
    onError: (_e, _v, ctx) => {
      if (ctx?.prev) qc.setQueryData(key(slug), ctx.prev);
    },
    onSettled: () => qc.invalidateQueries({ queryKey: key(slug) }),
  });
}

export function useDeleteScope(slug: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (scopeId: string) => api.deleteScope(scopeId),
    onMutate: async (scopeId) => {
      await qc.cancelQueries({ queryKey: key(slug) });
      const prev = qc.getQueryData<HillResponse>(key(slug));
      qc.setQueryData<HillResponse>(
        key(slug),
        patchScopes((scopes) => scopes.filter((s) => s.ID !== scopeId)),
      );
      return { prev };
    },
    onError: (_e, _v, ctx) => {
      if (ctx?.prev) qc.setQueryData(key(slug), ctx.prev);
    },
    onSettled: () => qc.invalidateQueries({ queryKey: key(slug) }),
  });
}

export type StagedMove = { scopeId: string; position: number; note: string };

// Commits staged dot moves. Each one appends a scope_positions row — the note is
// the snapshot of why it moved — so they're written together on Save.
export function useSaveMoves(slug: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (moves: StagedMove[]) =>
      Promise.all(moves.map((m) => api.moveScope(m.scopeId, m.position, m.note))),
    onSuccess: (_data, moves) => {
      qc.invalidateQueries({ queryKey: key(slug) });
      moves.forEach((m) => qc.invalidateQueries({ queryKey: ["snapshots", m.scopeId] }));
    },
  });
}
