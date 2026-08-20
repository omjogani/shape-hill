"use client";

import { useState } from "react";
import { HillChart, type ChartDot } from "@/components/organisms/hill-chart/HillChart";
import { phase } from "@/lib/geometry";

const INITIAL: ChartDot[] = [
  { id: "a", label: "Onboarding flow", color: "#d29922", position: 14, stalled: false },
  { id: "b", label: "Import pipeline", color: "#3fb950", position: 41, stalled: false },
  { id: "c", label: "Search rewrite", color: "#bc8cff", position: 63, stalled: false },
  { id: "d", label: "Auth and billing", color: "#58a6ff", position: 88, stalled: false },
];

export function HeroChart() {
  const [dots, setDots] = useState(INITIAL);
  const [selectedId, setSelectedId] = useState<string | null>("b");

  const stage = (id: string, position: number) =>
    setDots((current) => current.map((d) => (d.id === id ? { ...d, position } : d)));

  return (
    <div className="landing-chart">
      <HillChart dots={dots} onStage={stage} selectedId={selectedId} onSelect={setSelectedId} />

      <ul className="mt-5 flex flex-wrap gap-x-6 gap-y-2 px-1">
        {dots.map((dot, i) => (
          <li key={dot.id} className="flex items-center gap-2 text-sm">
            <span
              className="grid size-5 place-items-center rounded-full font-mono text-[10px] font-semibold text-[#0d1117]"
              style={{ background: dot.color }}
            >
              {i + 1}
            </span>
            <span className="text-[var(--text)]">{dot.label}</span>
            <span className="font-mono text-xs text-[var(--dim)]">{phase(dot.position)}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}
