"use client";

import { useRef, useState } from "react";
import {
  BASELINE,
  CHART_HEIGHT,
  LEFT,
  PEAK,
  RIGHT,
  WIDTH,
  positionFromX,
  x,
  y,
} from "@/lib/geometry";

export type ChartDot = {
  id: string;
  label: string;
  color: string;
  position: number;
  stalled: boolean;
};

// The hill outline, sampled once at module load (it never changes).
const curvePath = (() => {
  let d = "";
  for (let step = 0; step <= 120; step++) {
    const p = (step * 100) / 120;
    d += `${step === 0 ? "M" : "L"} ${x(p).toFixed(1)} ${y(p).toFixed(1)} `;
  }
  return d.trim();
})();

export function HillChart({
  dots,
  onMove,
}: {
  dots: ChartDot[];
  onMove: (id: string, position: number) => void;
}) {
  const svgRef = useRef<SVGSVGElement>(null);
  const [drag, setDrag] = useState<{ id: string; position: number } | null>(null);

  const toPosition = (clientX: number) => {
    const rect = svgRef.current!.getBoundingClientRect();
    return positionFromX(((clientX - rect.left) / rect.width) * WIDTH);
  };

  const start = (e: React.PointerEvent, dot: ChartDot) => {
    e.preventDefault();
    svgRef.current?.setPointerCapture(e.pointerId);
    setDrag({ id: dot.id, position: dot.position });
  };

  const move = (e: React.PointerEvent) => {
    if (!drag) return;
    setDrag({ id: drag.id, position: toPosition(e.clientX) });
  };

  const end = (e: React.PointerEvent) => {
    if (!drag) return;
    svgRef.current?.releasePointerCapture(e.pointerId);
    const original = dots.find((d) => d.id === drag.id);
    if (original && original.position !== drag.position) onMove(drag.id, drag.position);
    setDrag(null);
  };

  // Keyboard alternative to dragging: arrow keys nudge a focused dot.
  const nudge = (e: React.KeyboardEvent, dot: ChartDot) => {
    const step = e.key === "ArrowRight" ? 2 : e.key === "ArrowLeft" ? -2 : 0;
    if (!step) return;
    e.preventDefault();
    const next = Math.max(0, Math.min(100, dot.position + step));
    if (next !== dot.position) onMove(dot.id, next);
  };

  const posOf = (dot: ChartDot) => (drag?.id === dot.id ? drag.position : dot.position);
  const mid = x(50);

  return (
    <svg
      ref={svgRef}
      viewBox={`0 0 ${WIDTH} ${CHART_HEIGHT}`}
      className="w-full touch-none select-none"
      onPointerMove={move}
      onPointerUp={end}
      onPointerCancel={end}
      role="group"
      aria-label="Hill chart — drag a dot or use arrow keys to move it"
    >
      <rect
        x="1"
        y="1"
        width={WIDTH - 2}
        height={CHART_HEIGHT - 2}
        rx="10"
        className="fill-paper stroke-hill"
        strokeWidth="2"
      />
      <path
        d={`${curvePath} L ${RIGHT} ${BASELINE} L ${LEFT} ${BASELINE} Z`}
        className="fill-hill"
      />
      <path d={curvePath} className="fill-none stroke-ink" strokeWidth="2" />
      <line
        x1={LEFT}
        y1={BASELINE}
        x2={RIGHT}
        y2={BASELINE}
        className="stroke-ink"
        strokeWidth="1.5"
      />

      <line
        x1={mid}
        y1={PEAK - 8}
        x2={mid}
        y2={BASELINE}
        className="stroke-sage"
        strokeWidth="1"
        strokeDasharray="3 4"
      />
      <text
        x={mid}
        y={PEAK - 16}
        textAnchor="middle"
        className="fill-sage font-mono"
        fontSize="11"
        letterSpacing="1.2"
      >
        SUMMIT
      </text>
      <text
        x={LEFT}
        y={BASELINE + 22}
        className="fill-sage font-mono"
        fontSize="11"
        letterSpacing="1.2"
      >
        FIGURING IT OUT
      </text>
      <text
        x={RIGHT}
        y={BASELINE + 22}
        textAnchor="end"
        className="fill-sage font-mono"
        fontSize="11"
        letterSpacing="1.2"
      >
        MAKING IT HAPPEN
      </text>

      {dots.map((dot, i) => {
        const p = posOf(dot);
        return (
          <g
            key={dot.id}
            transform={`translate(${x(p)} ${y(p)})`}
            className={`hill-dot ${drag?.id === dot.id ? "cursor-grabbing" : "cursor-grab"}`}
            tabIndex={0}
            role="slider"
            aria-label={`${dot.label} position`}
            aria-valuemin={0}
            aria-valuemax={100}
            aria-valuenow={p}
            aria-valuetext={`${p} percent`}
            onPointerDown={(e) => start(e, dot)}
            onKeyDown={(e) => nudge(e, dot)}
          >
            <circle
              r="11"
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
              {i + 1}
            </text>
          </g>
        );
      })}
    </svg>
  );
}
