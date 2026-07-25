import { useRef, useState } from "react";
import { WIDTH, positionFromX } from "@/lib/geometry";
import type { ChartDot } from "./types";

// A pointer that travels less than this is a click (select), not a drag (move).
const DRAG_SLOP_PX = 4;

// The dot interaction state machine: pointer drag (staged move) vs. tap (select),
// plus keyboard nudge. Returns the svg ref, live drag state, and the handlers the
// chart wires onto the <svg> and each <Dot>.
export function useDotDrag(
  dots: ChartDot[],
  onStage: (id: string, position: number) => void,
  onSelect: (id: string) => void,
) {
  const svgRef = useRef<SVGSVGElement>(null);
  const [drag, setDrag] = useState<{ id: string; position: number } | null>(null);
  const startX = useRef(0);
  const moved = useRef(false);

  const toPosition = (clientX: number) => {
    const rect = svgRef.current!.getBoundingClientRect();
    return positionFromX(((clientX - rect.left) / rect.width) * WIDTH);
  };

  const start = (e: React.PointerEvent, dot: ChartDot) => {
    e.preventDefault();
    svgRef.current?.setPointerCapture(e.pointerId);
    startX.current = e.clientX;
    moved.current = false;
    setDrag({ id: dot.id, position: dot.position });
  };

  const move = (e: React.PointerEvent) => {
    if (!drag) return;
    if (Math.abs(e.clientX - startX.current) > DRAG_SLOP_PX) moved.current = true;
    setDrag({ id: drag.id, position: toPosition(e.clientX) });
  };

  const end = (e: React.PointerEvent) => {
    if (!drag) return;
    svgRef.current?.releasePointerCapture(e.pointerId);

    if (!moved.current) {
      onSelect(drag.id); // a tap, not a drag: open this scope on the right
    } else {
      const original = dots.find((d) => d.id === drag.id);
      if (original && original.position !== drag.position) onStage(drag.id, drag.position);
    }
    setDrag(null);
  };

  const onKey = (e: React.KeyboardEvent, dot: ChartDot) => {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      onSelect(dot.id);
      return;
    }
    const step = e.key === "ArrowRight" ? 2 : e.key === "ArrowLeft" ? -2 : 0;
    if (!step) return;
    e.preventDefault();
    const next = Math.max(0, Math.min(100, dot.position + step));
    if (next !== dot.position) onStage(dot.id, next);
  };

  // Live position of a dot: the in-flight drag position if it's being dragged.
  const posOf = (dot: ChartDot) => (drag?.id === dot.id ? drag.position : dot.position);

  return { svgRef, drag, start, move, end, onKey, posOf };
}
