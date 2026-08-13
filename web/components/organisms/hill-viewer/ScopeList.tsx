import type { Scope } from "@/lib/api";
import { ScopePanel } from "./ScopePanel";

export function ScopeList({
  scopes,
  isStalled,
  selected,
  onToggle,
}: {
  scopes: Scope[];
  isStalled: (s: Scope) => boolean;
  selected: string | null;
  onToggle: (scopeId: string) => void;
}) {
  const stalledCount = scopes.filter(isStalled).length;

  return (
    <div className="flex min-w-0 flex-col gap-3">
      <h2 className="font-mono text-xs uppercase tracking-widest text-sage">
        Scopes
        {stalledCount > 0 && <span className="ml-2 text-alarm">· {stalledCount} not moving</span>}
      </h2>

      {scopes.length === 0 ? (
        <p className="text-sm text-sage">No scopes yet.</p>
      ) : (
        <ul className="flex flex-col">
          {scopes.map((s, i) => (
            <ScopePanel
              key={s.ID}
              index={i}
              scope={s}
              stalled={isStalled(s)}
              open={selected === s.ID}
              onToggle={() => onToggle(s.ID)}
            />
          ))}
        </ul>
      )}
    </div>
  );
}
