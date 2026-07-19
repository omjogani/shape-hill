"use client";

import { useHill, useMoveScope } from "@/lib/hooks";
import type { Scope } from "@/lib/api";
import { HillChart, type ChartDot } from "./HillChart";
import { TitleForm } from "../molecules/TitleForm";
import { AddScopeForm } from "../molecules/AddScopeForm";
import { ScopeRow } from "../molecules/ScopeRow";
import { CopyEmbed } from "../molecules/CopyEmbed";

// Mirrors the server's stalledAfter: a scope untouched for a week (and not done).
const STALLED_AFTER_MS = 7 * 24 * 60 * 60 * 1000;
const isStalled = (s: Scope) =>
  Date.now() - new Date(s.MovedAt).getTime() > STALLED_AFTER_MS && s.Position < 100;

export function HillEditor({ slug }: { slug: string }) {
  const { data, isLoading, isError, error } = useHill(slug);
  const move = useMoveScope(slug);

  if (isLoading) return <Centered>Loading hill…</Centered>;

  if (isError) {
    return (
      <Centered>
        <p className="font-serif text-lg text-alarm">Couldn&apos;t load “{slug}”.</p>
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
    position: s.Position,
    stalled: isStalled(s),
  }));
  const nextSortOrder = scopes.reduce((max, s) => Math.max(max, s.SortOrder), 0) + 1;

  return (
    <main className="mx-auto flex w-full max-w-3xl flex-col gap-8 px-6 py-12">
      <header>
        <p className="font-mono text-xs uppercase tracking-widest text-sage">shapehill</p>
        <TitleForm slug={slug} title={hill.Title} />
      </header>

      <HillChart
        dots={dots}
        onMove={(id, position) => move.mutate({ scopeId: id, position, note: "" })}
      />

      <section className="flex flex-col gap-3">
        <h2 className="font-mono text-xs uppercase tracking-widest text-sage">Scopes</h2>
        {scopes.length === 0 ? (
          <p className="text-sm text-sage">No scopes yet — add the first one below.</p>
        ) : (
          <ul className="divide-y divide-hill/60">
            {scopes.map((s, i) => (
              <ScopeRow key={s.ID} slug={slug} index={i} scope={s} stalled={isStalled(s)} />
            ))}
          </ul>
        )}
        <AddScopeForm slug={slug} nextSortOrder={nextSortOrder} />
      </section>

      <footer className="flex flex-col gap-2 border-t border-hill/60 pt-4">
        <p className="font-mono text-xs uppercase tracking-widest text-sage">Embed anywhere</p>
        <CopyEmbed slug={slug} title={hill.Title} />
      </footer>
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
