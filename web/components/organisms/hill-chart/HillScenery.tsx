import { BASELINE, CHART_HEIGHT, LEFT, PEAK, RIGHT, WIDTH, x, y } from "@/lib/geometry";

// The hill outline, sampled once at module load (it never changes).
const curvePath = (() => {
  let d = "";
  for (let step = 0; step <= 120; step++) {
    const p = (step * 100) / 120;
    d += `${step === 0 ? "M" : "L"} ${x(p).toFixed(1)} ${y(p).toFixed(1)} `;
  }
  return d.trim();
})();

const mid = x(50);

// The static chart backdrop: frame, hill fill + outline, baseline, summit marker
// and axis labels. Nothing here reacts to props — it's pure scenery.
export function HillScenery() {
  return (
    <>
      <rect
        x="1"
        y="1"
        width={WIDTH - 2}
        height={CHART_HEIGHT - 2}
        rx="10"
        className="fill-paper stroke-line"
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
    </>
  );
}
