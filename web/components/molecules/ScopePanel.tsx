"use client";

import { useEffect, useRef, useState } from "react";
import { phase } from "@/lib/geometry";
import type { Scope } from "@/lib/api";
import { useDeleteScope } from "@/lib/hooks";
import { EditScopeForm } from "./EditScopeForm";
import { ScopeTimeline } from "./ScopeTimeline";

const action =
  "rounded px-1.5 py-0.5 font-mono text-[11px] text-sage focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-sage disabled:opacity-50";

export function ScopePanel({
  slug,
  index,
  scope,
  stalled,
  open,
  onToggle,
}: {
  slug: string;
  index: number;
  scope: Scope;
  stalled: boolean;
  open: boolean;
  onToggle: () => void;
}) {
  const ref = useRef<HTMLLIElement>(null);

  // Opened from the chart? The row may be out of view.
  useEffect(() => {
    if (open) ref.current?.scrollIntoView({ block: "nearest" });
  }, [open]);

  return (
    <li ref={ref} className="border-b border-line/60 first:border-t">
      <button
        type="button"
        onClick={onToggle}
        aria-expanded={open}
        className="flex w-full items-center gap-2.5 px-1 py-2.5 text-left hover:bg-hill/30 focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-sage"
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

      {/* Mounted only while open, so edit mode resets itself on close. */}
      {open && <ScopeBody slug={slug} scope={scope} stalled={stalled} />}
    </li>
  );
}

function ScopeBody({ slug, scope, stalled }: { slug: string; scope: Scope; stalled: boolean }) {
  const [editing, setEditing] = useState(false);
  const remove = useDeleteScope(slug);

  return (
    <div className="pb-4 pl-[34px] pr-1 pt-0.5">
      {editing ? (
        <EditScopeForm slug={slug} scope={scope} onDone={() => setEditing(false)} />
      ) : (
        <>
          <ScopeTimeline scopeId={scope.ID} open stalled={stalled} />
          <div className="mt-4 flex gap-1">
            <button
              type="button"
              onClick={() => setEditing(true)}
              className={`${action} hover:text-ink`}
            >
              Edit
            </button>
            <button
              type="button"
              onClick={() => {
                if (window.confirm(`Remove “${scope.Title}”?`)) remove.mutate(scope.ID);
              }}
              disabled={remove.isPending}
              className={`${action} hover:text-alarm`}
            >
              Remove
            </button>
          </div>
        </>
      )}
    </div>
  );
}
