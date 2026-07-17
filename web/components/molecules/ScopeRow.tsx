import { phase } from "@/lib/geometry";
import type { Scope } from "@/lib/api";

export function ScopeRow({
  index,
  scope,
  stalled,
}: {
  index: number;
  scope: Scope;
  stalled: boolean;
}) {
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
    </li>
  );
}
