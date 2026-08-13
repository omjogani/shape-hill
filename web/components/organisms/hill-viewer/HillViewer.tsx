"use client";

import { useState } from "react";
import { usePublicHill } from "@/lib/hooks";
import type { Scope } from "@/lib/api";
import { HillChart, type ChartDot } from "../hill-chart/HillChart";
import { ScopeList } from "./ScopeList";

// Mirrors the server's stalledAfter: a scope untouched for a week (and not done).
const STALLED_AFTER_MS = 7 * 24 * 60 * 60 * 1000;

const isStalled = (s: Scope) =>
  Date.now() - new Date(s.MovedAt).getTime() > STALLED_AFTER_MS && s.Position < 100;

export function HillViewer({ slug }: { slug: string }) {
  const { data, isLoading, isError } = usePublicHill(slug);
  const [openScopeId, setOpenScopeId] = useState<string | null>(null);
  const toggle = (scopeId: string) => setOpenScopeId((cur) => (cur === scopeId ? null : scopeId));

  if (isLoading) return <Centered>Loading hill…</Centered>;

  if (isError) {
    return (
      <Centered>
        <p className="font-display text-lg text-alarm">Couldn&apos;t load “{slug}”.</p>
        <p className="mt-2 text-sm text-sage">
          It may not exist, or its owner hasn&apos;t made it public.
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

  return (
    <main className="mx-auto flex w-full max-w-[1700px] flex-col gap-8 px-8 py-8">
      <header className="flex items-start justify-between gap-6">
        <div className="min-w-0 flex-1">
          <p className="font-mono text-xs uppercase tracking-widest text-sage">Read only</p>
          <h1 className="font-display text-3xl font-semibold tracking-tight text-ink">
            {hill.Title}
          </h1>
        </div>
      </header>

      <div className="grid grid-cols-1 gap-10 xl:grid-cols-[minmax(0,1fr)_440px]">
        <div className="flex min-w-0 flex-col gap-4">
          <HillChart dots={dots} readOnly selectedId={openScopeId} onSelect={toggle} />
        </div>

        <ScopeList scopes={scopes} isStalled={isStalled} selected={openScopeId} onToggle={toggle} />
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
