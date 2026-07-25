"use client";

import { useState } from "react";
import { useHill, useSaveMoves } from "@/lib/hooks";
import type { Scope } from "@/lib/api";
import { HillChart, type ChartDot } from "./HillChart";
import { TitleForm } from "../molecules/TitleForm";
import { AddScopeForm } from "../molecules/AddScopeForm";
import { ScopePanel } from "../molecules/ScopePanel";
import { EmbedMenu } from "../molecules/EmbedMenu";
import { PendingChanges, type Staged } from "../molecules/PendingChanges";

// Mirrors the server's stalledAfter: a scope untouched for a week (and not done).
const STALLED_AFTER_MS = 7 * 24 * 60 * 60 * 1000;
const isStalled = (s: Scope) =>
  Date.now() - new Date(s.MovedAt).getTime() > STALLED_AFTER_MS && s.Position < 100;

export function HillEditor({ slug }: { slug: string }) {
  const { data, isLoading, isError, error } = useHill(slug);
  const save = useSaveMoves(slug);
  // Dot moves live here until the user saves them, so nothing is written to the
  // database without a note the user had a chance to write.
  const [pending, setPending] = useState<Record<string, Staged>>({});
  // The open scope — shared by both panes, so the chart and the list stay in sync.
  const [selected, setSelected] = useState<string | null>(null);

  const stage = (scopeId: string, position: number) =>
    setPending((p) => ({ ...p, [scopeId]: { position, note: p[scopeId]?.note ?? "" } }));

  const setNote = (scopeId: string, note: string) =>
    setPending((p) => (p[scopeId] ? { ...p, [scopeId]: { ...p[scopeId], note } } : p));

  const toggle = (scopeId: string) => setSelected((cur) => (cur === scopeId ? null : scopeId));

  const saveMoves = async () => {
    const moves = Object.entries(pending).map(([scopeId, staged]) => ({
      scopeId,
      position: staged.position,
      note: staged.note,
    }));
    if (moves.length === 0) return;
    await save.mutateAsync(moves);
    setPending({});
  };

  if (isLoading) return <Centered>Loading hill…</Centered>;

  if (isError) {
    return (
      <Centered>
        <p className="font-display text-lg text-alarm">Couldn&apos;t load “{slug}”.</p>
        <p className="mt-2 text-sm text-sage">
          {(error as Error)?.message}. Is the API running on :8080 and seeded?
        </p>
      </Centered>
    );
  }

  if (!data) return null;

  const { hill, scopes } = data;
  const dots: ChartDot[] = scopes.map((s) => ({
    id: s.ID,
    label: s.Title,
    color: s.Color,
    position: pending[s.ID]?.position ?? s.Position,
    stalled: isStalled(s) && !pending[s.ID],
    pending: Boolean(pending[s.ID]),
  }));
  const nextSortOrder = scopes.reduce((max, s) => Math.max(max, s.SortOrder), 0) + 1;
  const stalledCount = scopes.filter(isStalled).length;

  return (
    <main className="mx-auto flex w-full max-w-[1700px] flex-col gap-8 px-8 py-8">
      <header className="flex items-start justify-between gap-6">
        <div className="min-w-0 flex-1">
          <p className="font-mono text-xs uppercase tracking-widest text-sage">shapehill</p>
          <TitleForm slug={slug} title={hill.Title} />
        </div>
        <div className="pt-6">
          <EmbedMenu slug={slug} title={hill.Title} />
        </div>
      </header>

      <div className="grid grid-cols-1 gap-10 xl:grid-cols-[minmax(0,1fr)_440px]">
        {/* Left: the hill, always whole. */}
        <div className="flex min-w-0 flex-col gap-4">
          <HillChart dots={dots} onStage={stage} selectedId={selected} onSelect={toggle} />

          <PendingChanges
            scopes={scopes}
            pending={pending}
            onNote={setNote}
            onSave={saveMoves}
            onDiscard={() => setPending({})}
            saving={save.isPending}
            failed={save.isError}
          />
        </div>

        {/* Right: scopes, each opening to its snapshot history. */}
        <div className="flex min-w-0 flex-col gap-3">
          <h2 className="font-mono text-xs uppercase tracking-widest text-sage">
            Scopes
            {stalledCount > 0 && (
              <span className="ml-2 text-alarm">· {stalledCount} not moving</span>
            )}
          </h2>

          {scopes.length === 0 ? (
            <p className="text-sm text-sage">No scopes yet — add the first one below.</p>
          ) : (
            <ul className="flex flex-col">
              {scopes.map((s, i) => (
                <ScopePanel
                  key={s.ID}
                  slug={slug}
                  index={i}
                  scope={s}
                  stalled={isStalled(s)}
                  open={selected === s.ID}
                  onToggle={() => toggle(s.ID)}
                />
              ))}
            </ul>
          )}

          <AddScopeForm slug={slug} nextSortOrder={nextSortOrder} />
        </div>
      </div>
    </main>
  );
}

function Centered({ children }: { children: React.ReactNode }) {
  return (
    <div className="mx-auto flex min-h-[60vh] max-w-md flex-col items-center justify-center px-6 text-center">
      {children}
    </div>
  );
}
