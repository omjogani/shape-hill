import { x, y } from "@/lib/geometry";
import type { ChartDot } from "./types";

export function Dot({
  dot,
  index,
  position,
  selected,
  dragging,
  readOnly = false,
  onPointerDown,
  onKeyDown,
}: {
  dot: ChartDot;
  index: number;
  position: number;
  selected: boolean;
  dragging: boolean;
  readOnly?: boolean;
  onPointerDown: (e: React.PointerEvent) => void;
  onKeyDown: (e: React.KeyboardEvent) => void;
}) {
  return (
    <g
      transform={`translate(${x(position)} ${y(position)})`}
      className={`hill-dot ${readOnly ? "cursor-pointer" : dragging ? "cursor-grabbing" : "cursor-grab"}`}
      tabIndex={0}
      role={readOnly ? "button" : "slider"}
      aria-label={readOnly ? `${dot.label} — ${position} percent` : `${dot.label} position`}
      aria-valuemin={readOnly ? undefined : 0}
      aria-valuemax={readOnly ? undefined : 100}
      aria-valuenow={readOnly ? undefined : position}
      aria-valuetext={readOnly ? undefined : `${position} percent`}
      onPointerDown={onPointerDown}
      onKeyDown={onKeyDown}
    >
      {(dot.pending || selected) && (
        <circle
          r={dot.pending ? 16 : 18}
          className="fill-none stroke-sage"
          strokeWidth="1.5"
          strokeDasharray="3 3"
        />
      )}
      <circle
        r={selected ? 13 : 11}
        strokeWidth="2.5"
        className="stroke-paper"
        style={{ fill: dot.stalled ? "var(--alarm)" : dot.color }}
      />
      <text
        textAnchor="middle"
        dominantBaseline="central"
        className="fill-paper font-mono"
        fontSize="11"
        fontWeight="600"
      >
        {index + 1}
      </text>
    </g>
  );
}
