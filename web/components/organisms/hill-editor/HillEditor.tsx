"use client";

import Link from "next/link";
import type { Scope } from "@/lib/api";
import { HillChart, type ChartDot } from "../hill-chart/HillChart";
import { TitleForm } from "../../molecules/TitleForm";
import { EmbedMenu } from "../../molecules/EmbedMenu";
import { VisibilityToggle } from "../../molecules/VisibilityToggle";
import { PendingChanges } from "../../molecules/PendingChanges";
import { ScopeList } from "./ScopeList";
import { useHillEditor } from "./useHillEditor";

// Mirrors the server's stalledAfter: a scope untouched for a week (and not done).
const STALLED_AFTER_MS = 7 * 24 * 60 * 60 * 1000;

const isStalled = (s: Scope) =>
  Date.now() - new Date(s.MovedAt).getTime() > STALLED_AFTER_MS && s.Position < 100;

export function HillEditor({ slug }: { slug: string }) {
  const { query, save, unsavedMoves, openScopeId, stage, setNote, toggle, discard, saveMoves } =
    useHillEditor(slug);
  const { data, isLoading, isError, error } = query;

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
    position: unsavedMoves[s.ID]?.position ?? s.Position,
    stalled: isStalled(s) && !unsavedMoves[s.ID],
    pending: Boolean(unsavedMoves[s.ID]),
  }));
  const nextSortOrder = scopes.reduce((max, s) => Math.max(max, s.SortOrder), 0) + 1;

  return (
    <main className="mx-auto flex w-full max-w-[1700px] flex-col gap-8 px-8 py-8">
      <header className="flex items-start justify-between gap-6">
        <div className="min-w-0 flex-1">
          <Link
            href="/"
            className="font-mono text-xs uppercase tracking-widest text-sage hover:text-ink"
          >
            ← shapehill
          </Link>
          <TitleForm slug={slug} title={hill.Title} />
        </div>
        <div className="flex items-center gap-4 pt-6">
          <VisibilityToggle slug={slug} isPublic={hill.IsPublic} />
          <EmbedMenu slug={slug} title={hill.Title} />
        </div>
      </header>

      <div className="grid grid-cols-1 gap-10 xl:grid-cols-[minmax(0,1fr)_440px]">
        <div className="flex min-w-0 flex-col gap-4">
          <HillChart dots={dots} onStage={stage} selectedId={openScopeId} onSelect={toggle} />

          <PendingChanges
            scopes={scopes}
            pending={unsavedMoves}
            onNote={setNote}
            onSave={saveMoves}
            onDiscard={discard}
            saving={save.isPending}
            failed={save.isError}
          />
        </div>

        <ScopeList
          slug={slug}
          scopes={scopes}
          isStalled={isStalled}
          selected={openScopeId}
          onToggle={toggle}
          nextSortOrder={nextSortOrder}
        />
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
