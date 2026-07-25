"use client";

import { CHART_HEIGHT, WIDTH } from "@/lib/geometry";
import { HillScenery } from "./HillScenery";
import { Dot } from "./Dot";
import { useDotDrag } from "./useDotDrag";
import type { ChartDot } from "./types";

export type { ChartDot };

export function HillChart({
  dots,
  onStage,
  selectedId,
  onSelect,
}: {
  dots: ChartDot[];
  /** Records a move locally; nothing is persisted until the user saves. */
  onStage: (id: string, position: number) => void;
  selectedId: string | null;
  onSelect: (id: string) => void;
}) {
  const { svgRef, drag, start, move, end, onKey, posOf } = useDotDrag(dots, onStage, onSelect);

  const dotFor = (dot: ChartDot, i: number) => (
    <Dot
      key={dot.id}
      dot={dot}
      index={i}
      position={posOf(dot)}
      selected={dot.id === selectedId}
      dragging={drag?.id === dot.id}
      onPointerDown={(e) => start(e, dot)}
      onKeyDown={(e) => onKey(e, dot)}
    />
  );

  return (
    <svg
      ref={svgRef}
      viewBox={`0 0 ${WIDTH} ${CHART_HEIGHT}`}
      className="w-full touch-none select-none"
      onPointerMove={move}
      onPointerUp={end}
      onPointerCancel={end}
      role="group"
      aria-label="Hill chart — click a dot to open it, drag or use arrow keys to move it"
    >
      <HillScenery />

      {/* Non-selected dots, opaque. When one is selected they sit under the scrim. */}
      {dots.map((dot, i) => (dot.id === selectedId ? null : dotFor(dot, i)))}

      {selectedId !== null && (
        <>
          <rect
            x="2"
            y="2"
            width={WIDTH - 4}
            height={CHART_HEIGHT - 4}
            rx="10"
            className="fill-paper"
            opacity={0.6}
            pointerEvents="none"
          />
          {dots.map((dot, i) => (dot.id === selectedId ? dotFor(dot, i) : null))}
        </>
      )}
    </svg>
  );
}
