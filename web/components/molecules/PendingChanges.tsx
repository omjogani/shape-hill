"use client";

import type { Scope } from "@/lib/api";
import { Button } from "../atoms/Button";
import { TextInput } from "../atoms/TextInput";

export type Staged = { position: number; note: string };

export function PendingChanges({
  scopes,
  pending,
  onNote,
  onSave,
  onDiscard,
  saving,
  failed,
}: {
  scopes: Scope[];
  pending: Record<string, Staged>;
  onNote: (scopeId: string, note: string) => void;
  onSave: () => void;
  onDiscard: () => void;
  saving: boolean;
  failed: boolean;
}) {
  const moved = scopes.filter((s) => pending[s.ID]);
  if (moved.length === 0) return null;

  return (
    <section
      aria-label="Unsaved moves"
      className="flex flex-col gap-3 rounded-xl border border-sage/40 bg-hill/20 p-4"
    >
      <p className="font-mono text-xs uppercase tracking-widest text-sage">
        Unsaved {moved.length === 1 ? "move" : "moves"}
      </p>

      <ul className="flex flex-col gap-3">
        {moved.map((s) => {
          const staged = pending[s.ID];
          return (
            <li key={s.ID} className="flex flex-col gap-1.5">
              <div className="flex items-baseline gap-2">
                <span
                  className="h-2.5 w-2.5 shrink-0 rounded-full"
                  style={{ backgroundColor: s.Color }}
                />
                <span className="font-display text-sm text-ink">{s.Title}</span>
                <span className="font-mono text-[11px] tabular-nums text-sage">
                  {s.Position}% → <span className="text-ink">{staged.position}%</span>
                </span>
              </div>
              <TextInput
                value={staged.note}
                onChange={(e) => onNote(s.ID, e.target.value)}
                placeholder="What changed? (optional note saved with this move)"
                aria-label={`Note for ${s.Title}`}
              />
            </li>
          );
        })}
      </ul>

      <div className="flex items-center gap-2">
        <Button type="button" onClick={onSave} disabled={saving}>
          {saving ? "Saving…" : `Save ${moved.length === 1 ? "move" : "moves"}`}
        </Button>
        <Button type="button" variant="ghost" onClick={onDiscard} disabled={saving}>
          Discard
        </Button>
        {failed && <span className="text-xs text-alarm">Couldn&apos;t save. Try again.</span>}
      </div>
    </section>
  );
}
