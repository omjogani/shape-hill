"use client";

import { useScopeSnapshots } from "@/lib/hooks";

const DAY = 24 * 60 * 60 * 1000;

function ago(iso: string) {
  const days = Math.floor((Date.now() - new Date(iso).getTime()) / DAY);
  if (days <= 0) return "today";
  if (days === 1) return "yesterday";
  return `${days}d ago`;
}

export function ScopeTimeline({
  scopeId,
  open,
  stalled,
}: {
  scopeId: string;
  open: boolean;
  stalled: boolean;
}) {
  const { data, isLoading, isError } = useScopeSnapshots(scopeId, open);

  if (isLoading) return <p className="font-mono text-xs text-sage">Loading history…</p>;
  if (isError) return <p className="font-mono text-xs text-alarm">Couldn&apos;t load history.</p>;

  const snapshots = data ?? [];
  if (snapshots.length === 0) {
    return (
      <p className="font-mono text-xs text-sage">
        No snapshots yet — move this dot to record the first.
      </p>
    );
  }

  return (
    <>
      <p className="mb-3 font-mono text-[10.5px] uppercase tracking-widest text-sage">
        Snapshots · {snapshots.length}
      </p>
      <ol className="flex flex-col gap-4 border-l border-line pl-4">
        {snapshots.map((snap, i) => {
          const previous = snapshots[i + 1];
          const delta = previous ? snap.Position - previous.Position : snap.Position;
          const marker = i === 0 ? (stalled ? "bg-alarm" : "bg-ink") : "bg-sage";

          return (
            <li key={`${snap.CreatedAt}-${i}`} className="relative">
              <span
                className={`absolute -left-[21px] top-1.5 h-2.5 w-2.5 rounded-full border-2 border-background ${marker}`}
              />
              <div className="flex flex-wrap items-baseline gap-2">
                <span className="font-mono text-xs font-semibold tabular-nums text-ink">
                  {snap.Position}%
                </span>
                <span className="font-mono text-[11px] text-sage">
                  {delta >= 0 ? "+" : ""}
                  {delta}
                </span>
                <span className="ml-auto font-mono text-[11px] text-sage">
                  {ago(snap.CreatedAt)}
                </span>
              </div>
              <p className={`mt-0.5 text-[13.5px] ${snap.Note ? "text-ink" : "italic text-sage"}`}>
                {snap.Note || "no note"}
              </p>
            </li>
          );
        })}
      </ol>
    </>
  );
}
