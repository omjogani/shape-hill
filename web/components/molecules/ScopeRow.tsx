"use client";

import { useState } from "react";
import { phase } from "@/lib/geometry";
import type { Scope } from "@/lib/api";
import { useDeleteScope } from "@/lib/hooks";
import { EditScopeForm } from "./EditScopeForm";

const rowAction =
  "rounded px-1.5 py-0.5 font-mono text-[11px] text-sage focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-sage disabled:opacity-50";

export function ScopeRow({
  slug,
  index,
  scope,
  stalled,
}: {
  slug: string;
  index: number;
  scope: Scope;
  stalled: boolean;
}) {
  const [editing, setEditing] = useState(false);
  const remove = useDeleteScope(slug);

  if (editing) {
    return (
      <li className="py-2">
        <EditScopeForm slug={slug} scope={scope} onDone={() => setEditing(false)} />
      </li>
    );
  }

  const status = stalled ? "Not moving" : phase(scope.Position);

  return (
    <li className="flex items-center gap-3 py-2">
      <span
        className="flex h-5 w-5 shrink-0 items-center justify-center rounded-[3px] font-mono text-[11px] font-semibold text-paper"
        style={{ backgroundColor: stalled ? "var(--alarm)" : scope.Color }}
      >
        {index + 1}
      </span>
      <span className="font-serif text-sm text-ink">{scope.Title}</span>

      <span className="ml-auto text-right font-mono text-[11px] text-sage">
        <span className={stalled ? "text-alarm" : undefined}>{status}</span>
        {scope.Note ? ` — ${scope.Note}` : ""}
        <span className="ml-2 tabular-nums text-ink/70">{scope.Position}%</span>
      </span>

      <div className="flex shrink-0 gap-1">
        <button
          type="button"
          onClick={() => setEditing(true)}
          className={`${rowAction} hover:text-ink`}
          aria-label={`Edit ${scope.Title}`}
        >
          Edit
        </button>
        <button
          type="button"
          onClick={() => {
            if (window.confirm(`Remove “${scope.Title}”?`)) remove.mutate(scope.ID);
          }}
          disabled={remove.isPending}
          className={`${rowAction} hover:text-alarm`}
          aria-label={`Remove ${scope.Title}`}
        >
          Remove
        </button>
      </div>
    </li>
  );
}
