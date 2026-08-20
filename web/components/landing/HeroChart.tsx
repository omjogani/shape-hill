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

const VERDICT = {
  Uphill: "is still being figured out. Any date you give for it today is a guess.",
  Downhill: "is over the summit. Nothing surprising left, so the rest is mostly typing.",
  Done: "is done.",
};

export function HeroChart() {
  const [dots, setDots] = useState(INITIAL);
  const [selectedId, setSelectedId] = useState<string | null>("b");

  const stage = (id: string, position: number) =>
    setDots((current) => current.map((d) => (d.id === id ? { ...d, position } : d)));

  const selected = dots.find((d) => d.id === selectedId) ?? dots[0];

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

      <p className="mx-auto mt-10 min-h-16 max-w-xl text-center text-lg leading-relaxed text-[var(--dim)]">
        <span className="font-medium text-[var(--text)]">{selected.label}</span>{" "}
        {VERDICT[phase(selected.position)]}
      </p>

      <p className="text-center font-mono text-xs uppercase tracking-widest text-[var(--dim)]">
        {phase(selected.position) === "Uphill" ? "Drag it over the summit" : "Drag it back uphill"}
      </p>
    </div>
  );
}
