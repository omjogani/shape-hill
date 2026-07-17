"use client";

import {
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { api, type Hill, type HillResponse } from "./api";

const key = (slug: string) => ["hill", slug] as const;

export function useHill(slug: string) {
  return useQuery({ queryKey: key(slug), queryFn: () => api.getHill(slug) });
}

export function useUpdateTitle(slug: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (title: string) => api.updateTitle(slug, title),
    onSuccess: (hill: Hill) =>
      qc.setQueryData<HillResponse>(key(slug), (prev) =>
        prev ? { ...prev, hill } : prev,
      ),
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

export function useMoveScope(slug: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (v: { scopeId: string; position: number; note: string }) =>
      api.moveScope(v.scopeId, v.position, v.note),
    // Optimistic: the dot should stay where it was dropped, not snap back while
    // the request is in flight.
    onMutate: async (v) => {
      await qc.cancelQueries({ queryKey: key(slug) });
      const prev = qc.getQueryData<HillResponse>(key(slug));
      qc.setQueryData<HillResponse>(key(slug), (old) =>
        old
          ? {
              ...old,
              scopes: old.scopes.map((s) =>
                s.ID === v.scopeId ? { ...s, Position: v.position, Note: v.note } : s,
              ),
            }
          : old,
      );
      return { prev };
    },
    onError: (_e, _v, ctx) => {
      if (ctx?.prev) qc.setQueryData(key(slug), ctx.prev);
    },
    onSettled: () => qc.invalidateQueries({ queryKey: key(slug) }),
  });
}
