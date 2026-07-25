import { useState } from "react";
import { useHill, useSaveMoves } from "@/lib/hooks";
import type { Staged } from "../../molecules/PendingChanges";

export function useHillEditor(slug: string) {
  const query = useHill(slug);
  const save = useSaveMoves(slug);
  const [unsavedMoves, setUnsavedMoves] = useState<Record<string, Staged>>({});
  const [openScopeId, setOpenScopeId] = useState<string | null>(null);

  const stage = (scopeId: string, position: number) =>
    setUnsavedMoves((p) => ({ ...p, [scopeId]: { position, note: p[scopeId]?.note ?? "" } }));

  const setNote = (scopeId: string, note: string) =>
    setUnsavedMoves((p) => (p[scopeId] ? { ...p, [scopeId]: { ...p[scopeId], note } } : p));

  const toggle = (scopeId: string) => setOpenScopeId((cur) => (cur === scopeId ? null : scopeId));

  const discard = () => setUnsavedMoves({});

  const saveMoves = async () => {
    const moves = Object.entries(unsavedMoves).map(([scopeId, staged]) => ({
      scopeId,
      position: staged.position,
      note: staged.note,
    }));
    if (moves.length === 0) return;
    await save.mutateAsync(moves);
    setUnsavedMoves({});
  };

  return { query, save, unsavedMoves, openScopeId, stage, setNote, toggle, discard, saveMoves };
}
