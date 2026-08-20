import type { Metadata } from "next";
import Link from "next/link";
import { HeaderCta } from "@/components/landing/HeaderCta";
import { HeroChart } from "@/components/landing/HeroChart";
import { BASELINE, LEFT, RIGHT, x, y } from "@/lib/geometry";

export const metadata: Metadata = {
  title: "shapehill - hill charts for teams",
  description:
    "Where the work is, not how much is left. Draw a Shape Up hill chart and embed it in any README.",
};

const GITHUB = "https://github.com/omjogani/shape-hill";
const DEMO = "/shape-hill-readme/view";
const EMBED_SNIPPET =
  "[![Team hill](https://shape-hill.onrender.com/hill/your-slug.svg?style=github)](https://shape-hill.vercel.app/your-slug/view)";

const curve = (() => {
  let d = "";
  for (let step = 0; step <= 80; step++) {
    const p = (step * 100) / 80;
    d += `${step === 0 ? "M" : "L"} ${x(p).toFixed(1)} ${y(p).toFixed(1)} `;
  }
  return d.trim();
})();

function MiniHill({ position, color }: { position: number; color: string }) {
  return (
    <svg viewBox={`${LEFT - 20} 40 ${RIGHT - LEFT + 40} 290`} className="w-full" aria-hidden>
      <path d={`${curve} L ${RIGHT} ${BASELINE} L ${LEFT} ${BASELINE} Z`} fill="#161b22" />
      <path d={curve} fill="none" stroke="#30363d" strokeWidth="3" />
      <line x1={LEFT} y1={BASELINE} x2={RIGHT} y2={BASELINE} stroke="#30363d" strokeWidth="2.5" />
      <circle cx={x(position)} cy={y(position)} r="17" fill={color} />
    </svg>
  );
}

function Pill({ children }: { children: React.ReactNode }) {
  return (
    <span className="inline-block rounded-full border border-[var(--edge)] px-3 py-1 font-mono text-[11px] uppercase tracking-widest text-[var(--accent)]">
      {children}
    </span>
  );
}

const PHASES = [
  {
    name: "Uphill",
    sub: "Figuring it out",
    color: "#d29922",
    position: 22,
    body: "Unknowns, spikes, I think it works this way. Progress here is understanding, and it is not visible from outside.",
  },
  {
    name: "The summit",
    sub: "Nothing surprising left",
    color: "#3fb950",
    position: 50,
    body: "You can describe the rest of the work in sentences that do not contain the word probably.",
  },
  {
    name: "Downhill",
    sub: "Making it happen",
    color: "#58a6ff",
    position: 78,
    body: "Execution. Boring, predictable, mostly typing. The kind of work an estimate is actually good at.",
  },
];

export default function LandingPage() {
  return (
    <div className="landing min-h-screen">
      <header className="mx-auto flex w-full max-w-6xl items-center justify-between px-6 py-6">
        <Link href="/" className="font-mono text-sm tracking-widest">
          shapehill
        </Link>
        <nav className="flex items-center gap-2 text-sm">
          <a
            href={GITHUB}
            className="rounded-md px-3 py-1.5 text-[var(--dim)] transition-colors hover:text-[var(--text)]"
          >
            GitHub
          </a>
          <HeaderCta />
        </nav>
      </header>

      <main className="mx-auto w-full max-w-6xl px-6">
        <section className="hero-glow flex flex-col items-center pb-16 pt-20 text-center sm:pt-28">
          <h1 className="max-w-3xl text-5xl font-semibold leading-[1.05] tracking-[-0.035em] sm:text-7xl">
            Where the work is, not how much is left.
          </h1>
          <p className="mt-7 max-w-xl text-lg leading-relaxed text-[var(--dim)]">
            Hill charts for teams. Every scope is a dot: climbing while you are still figuring it
            out, descending once you are just building it.
          </p>
          <div className="mt-9 flex flex-wrap items-center justify-center gap-3">
            <Link
              href="/app"
              className="rounded-lg bg-[var(--text)] px-5 py-2.5 font-medium text-[var(--bg)] transition-opacity hover:opacity-90"
            >
              Start a hill chart
            </Link>
            <Link
              href={DEMO}
              className="rounded-lg border border-[var(--edge)] px-5 py-2.5 text-[var(--dim)] transition-colors hover:border-[var(--dim)] hover:text-[var(--text)]"
            >
              See a live one
            </Link>
          </div>
        </section>

        <section className="fade-up pb-32">
          <HeroChart />
        </section>

        <section className="fade-up border-t border-[var(--edge)] py-28">
          <div className="max-w-2xl">
            <Pill>The idea</Pill>
            <h2 className="mt-6 text-4xl font-semibold leading-[1.1] tracking-[-0.03em] sm:text-5xl">
              80% done is a lie you tell twice.
            </h2>
            <p className="mt-6 text-lg leading-relaxed text-[var(--dim)]">
              Once when you are stuck, and again a week later when you are still stuck. A percentage
              can only go up, so it goes up whether or not anything was figured out. A hill has two
              sides, and that is the whole point.
            </p>
          </div>

          <div className="mt-16 grid gap-4 sm:grid-cols-3">
            {PHASES.map((p) => (
              <div
                key={p.name}
                className="rounded-2xl border border-[var(--edge)] bg-[var(--panel)] p-6"
              >
                <MiniHill position={p.position} color={p.color} />
                <h3 className="mt-5 text-lg font-medium">{p.name}</h3>
                <p className="font-mono text-xs uppercase tracking-widest text-[var(--accent)]">
                  {p.sub}
                </p>
                <p className="mt-3 text-sm leading-relaxed text-[var(--dim)]">{p.body}</p>
              </div>
            ))}
          </div>

          <p className="mt-12 max-w-2xl text-lg leading-relaxed text-[var(--dim)]">
            Two scopes both halfway done are in completely different amounts of trouble if one is
            climbing and the other is descending.{" "}
            <span className="text-[var(--text)]">Position on the hill says which.</span>
          </p>
        </section>

        <section className="fade-up border-t border-[var(--edge)] py-28">
          <div className="max-w-2xl">
            <Pill>Embed</Pill>
            <h2 className="mt-6 text-4xl font-semibold leading-[1.1] tracking-[-0.03em] sm:text-5xl">
              One line in your README. It redraws itself.
            </h2>
            <p className="mt-6 text-lg leading-relaxed text-[var(--dim)]">
              Paste it into a README, a PR description, or a wiki page. Every time someone moves a
              dot, the image everyone is looking at changes with it.
            </p>
          </div>

          <div className="mt-14 grid gap-4 lg:grid-cols-2">
            <div className="rounded-2xl border border-[var(--edge)] bg-[var(--panel)] p-6">
              <p className="font-mono text-[11px] uppercase tracking-widest text-[var(--dim)]">
                Markdown
              </p>
              <code className="mt-3 block overflow-x-auto whitespace-pre rounded-lg bg-[#0d0d0f] p-4 font-mono text-xs leading-relaxed text-[var(--accent)]">
                {EMBED_SNIPPET}
              </code>
              <p className="mt-4 text-sm leading-relaxed text-[var(--dim)]">
                Two styles ship with it: warm paper, or one that disappears into a GitHub README.
              </p>
            </div>

            <div className="rounded-2xl border border-[var(--edge)] bg-[var(--panel)] p-6">
              <p className="font-mono text-[11px] uppercase tracking-widest text-[var(--dim)]">
                Renders as
              </p>
              <a href={DEMO} className="mt-3 block rounded-lg bg-white p-3">
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src="https://shape-hill.onrender.com/hill/shape-hill-readme.svg?style=github"
                  alt="A live shapehill chart, as embedded in this project's README"
                  className="w-full"
                  loading="lazy"
                />
              </a>
              <p className="mt-4 text-sm leading-relaxed text-[var(--dim)]">
                That is this project&apos;s own chart, live. Click it for the full view.
              </p>
            </div>
          </div>
        </section>
      </main>

      <footer className="mx-auto flex w-full max-w-6xl flex-wrap items-center justify-between gap-6 border-t border-[var(--edge)] px-6 py-10 text-sm text-[var(--dim)]">
        <p>
          Borrowed, with thanks, from{" "}
          <a
            href="https://basecamp.com/shapeup/3.4-chapter-13"
            className="text-[var(--text)] underline underline-offset-4"
          >
            Basecamp&apos;s Shape Up
          </a>
          .
        </p>
        <div className="flex items-center gap-5">
          <a href={GITHUB} className="transition-colors hover:text-[var(--text)]">
            GitHub
          </a>
          <Link href="/app" className="transition-colors hover:text-[var(--text)]">
            Open the app
          </Link>
        </div>
      </footer>
    </div>
  );
}
