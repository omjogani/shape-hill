"use client";

import { useEffect, useRef } from "react";
import { phase } from "@/lib/geometry";
import type { Scope } from "@/lib/api";
import { ScopeTimeline } from "./ScopeTimeline";

export function ScopePanel({
  index,
  scope,
  stalled,
  open,
  onToggle,
}: {
  index: number;
  scope: Scope;
  stalled: boolean;
  open: boolean;
  onToggle: () => void;
}) {
  const ref = useRef<HTMLLIElement>(null);

  useEffect(() => {
    if (open) ref.current?.scrollIntoView({ block: "nearest" });
  }, [open]);

  return (
    <li ref={ref} className="border-b border-line/60 first:border-t">
      <button
        type="button"
        onClick={onToggle}
        aria-expanded={open}
        className="flex w-full items-center gap-2.5 px-1 py-2.5 text-left hover:bg-hill/30 focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-sage"
      >
        <span
          className="flex h-5 w-5 shrink-0 items-center justify-center rounded-[3px] font-mono text-[11px] font-semibold text-paper"
          style={{ backgroundColor: stalled ? "var(--alarm)" : scope.Color }}
        >
          {index + 1}
        </span>
        <span className="min-w-0 flex-1 truncate font-display text-sm text-ink">{scope.Title}</span>
        <span className={`shrink-0 font-mono text-[11px] ${stalled ? "text-alarm" : "text-sage"}`}>
          {stalled ? "Not moving" : phase(scope.Position)}{" "}
          <span className="tabular-nums text-ink">{scope.Position}%</span>
        </span>
        <svg
          viewBox="0 0 12 12"
          aria-hidden="true"
          className={`h-3 w-3 shrink-0 text-sage transition-transform ${open ? "rotate-90" : ""}`}
        >
          <path d="M4 2 L8 6 L4 10" fill="none" stroke="currentColor" strokeWidth="1.6" />
        </svg>
      </button>

      {open && (
        <div className="pb-4 pl-8.5 pr-1 pt-0.5">
          <ScopeTimeline scopeId={scope.ID} open stalled={stalled} />
        </div>
      )}
    </li>
  );
}
