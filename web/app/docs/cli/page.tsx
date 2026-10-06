import type { Metadata } from "next";
import Link from "next/link";
import { CodeBlock } from "@/components/docs/CodeBlock";
import { GITHUB, SiteHeader } from "@/components/landing/SiteHeader";
import { SiteFooter } from "@/components/landing/SiteFooter";

export const metadata: Metadata = {
  title: "shapehill CLI - hill charts from the terminal",
  description:
    "Install the shapehill CLI to create hill charts, embed them in stories, and move scopes from your terminal or an AI agent.",
};

const RELEASES = `${GITHUB}/releases/latest`;

const INSTALL_SCRIPT = `curl -fsSL "${GITHUB}/releases/latest/download/shapehill_$(uname -s | tr '[:upper:]' '[:lower:]')_$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/').tar.gz" | tar -xz shapehill
sudo mv shapehill /usr/local/bin/
shapehill version`;

const GO_INSTALL = "go install github.com/omjogani/shape-hill/api/cmd/shapehill@latest";

const AUTH = `export SHAPEHILL_TOKEN=shk_...   # add to ~/.zshrc or ~/.bashrc to keep it
shapehill auth status`;

const QUICKSTART = `shapehill hill create launch-v2 --title "Launch v2" --public
shapehill scope add launch-v2 "API integration"
shapehill scope add launch-v2 "Settings UI"

shapehill scope move launch-v2 "API integration" 30 --note "comparing auth providers"
shapehill scope move launch-v2 "API integration" 55 --note "auth approach settled"
shapehill scope move launch-v2 "Settings UI" 100 --note "shipped"

shapehill hill show launch-v2`;

const SHOW_OUTPUT = `Launch v2  launch-v2 · public

HILL                   POS  PHASE     SCOPE            NOTE                   MOVED
──────────┼●─────────   55  downhill  API integration  auth approach settled  just now
──────────┼─────────●  100  done      Settings UI      shipped                just now`;

const EMBED = `shapehill hill embed launch-v2 --style github`;

const AGENT_SNIPPET = `## Hill chart

This story's progress lives on the shapehill hill \`launch-v2\`.
After finishing a piece of work, move its scope and say why:

    shapehill scope move launch-v2 "<scope>" <0-100> --note "<what changed>" --json

Below 50 while unknowns remain, 50 and up once the approach is settled,
100 when shipped. Run \`shapehill help agents\` for the rules.`;

const COMPLETION = `shapehill completion zsh > "\${fpath[1]}/_shapehill"   # zsh
shapehill completion bash > /etc/bash_completion.d/shapehill   # bash
shapehill completion fish > ~/.config/fish/completions/shapehill.fish`;

const COMMANDS = [
  {
    group: "Hills",
    rows: [
      ["hill list", "ls", "List your hills"],
      [
        "hill create <slug> --title …",
        "new",
        "Create a hill and print its embed snippet. --public, --description, --if-not-exists",
      ],
      [
        "hill show <slug>",
        "view",
        "Every scope with its place on the hill, position, phase, note and last move",
      ],
      [
        "hill update <slug>",
        "",
        "--title, --public / --private, --track-stalled / --no-track-stalled",
      ],
      [
        "hill embed <slug>",
        "",
        "Print the snippet for a story. --format markdown|html|url, --style github",
      ],
    ],
  },
  {
    group: "Scopes",
    rows: [
      [
        "scope add <hill> <title>",
        "",
        "Add a scope at position 0. --color, --description, --if-not-exists",
      ],
      [
        "scope move <hill> <scope> <0-100> --note …",
        "mv",
        "Move a scope and record a snapshot. Repeating the latest move is a no-op",
      ],
      ["scope log <hill> <scope>", "history", "Every snapshot of a scope, newest first"],
    ],
  },
  {
    group: "Setup",
    rows: [
      ["auth status", "", "Which account the token belongs to"],
      ["help agents", "", "The rules for AI agents driving the CLI"],
      ["completion <shell>", "", "Shell completion, including your hill slugs and scope titles"],
      ["version", "-v", "Print the version"],
    ],
  },
];

const EXIT_CODES = [
  ["0", "ok"],
  ["1", "failure, safe to retry"],
  ["2", "invalid input"],
  ["3", "not found"],
  ["4", "auth: missing or revoked token"],
  ["5", "conflict: already exists"],
];

const ENV = [
  ["SHAPEHILL_TOKEN", "Personal access token from /app/tokens. Required."],
  ["SHAPEHILL_API_URL", "API origin. Defaults to https://shape-hill.onrender.com"],
  [
    "SHAPEHILL_WEB_URL",
    "Web origin, used in embed links. Defaults to https://shape-hill.vercel.app",
  ],
  ["NO_COLOR", "Set to anything to turn colour off (same as --no-color)"],
];

const SECTIONS = [
  ["install", "Install"],
  ["authenticate", "Authenticate"],
  ["quickstart", "Quick start"],
  ["positions", "Positions"],
  ["commands", "Commands"],
  ["agents", "AI agents"],
  ["configuration", "Configuration"],
];

function Section({
  id,
  title,
  children,
}: {
  id: string;
  title: string;
  children: React.ReactNode;
}) {
  return (
    <section id={id} className="scroll-mt-8 border-t border-[var(--edge)] py-14">
      <h2 className="text-2xl font-semibold tracking-[-0.02em] sm:text-3xl">{title}</h2>
      {children}
    </section>
  );
}

function Prose({ children }: { children: React.ReactNode }) {
  return <p className="mt-4 max-w-2xl leading-relaxed text-[var(--dim)]">{children}</p>;
}

function Code({ children }: { children: React.ReactNode }) {
  return (
    <code className="rounded bg-[var(--panel)] px-1.5 py-0.5 font-mono text-[0.85em] text-[var(--text)]">
      {children}
    </code>
  );
}

export default function CliDocsPage() {
  return (
    <div className="landing min-h-screen">
      <SiteHeader />

      <main className="mx-auto w-full max-w-6xl px-6">
        <section className="hero-glow pb-16 pt-16 sm:pt-24">
          <p className="font-mono text-xs uppercase tracking-widest text-[var(--accent)]">
            shapehill CLI
          </p>
          <h1 className="mt-4 max-w-3xl text-4xl font-semibold leading-[1.05] tracking-[-0.035em] sm:text-6xl">
            Move the dots from your terminal.
          </h1>
          <p className="mt-6 max-w-2xl text-lg leading-relaxed text-[var(--dim)]">
            Create a hill, embed it in a story, and move its scopes as the work progresses, by hand
            or from an AI agent. Every command speaks JSON and returns exit codes a script can
            branch on.
          </p>
        </section>

        <div className="grid grid-cols-1 gap-12 lg:grid-cols-[12rem_1fr]">
          <nav aria-label="On this page" className="hidden lg:block">
            <ul className="sticky top-8 flex flex-col gap-2 border-t border-[var(--edge)] pt-14 text-sm">
              {SECTIONS.map(([id, label]) => (
                <li key={id}>
                  <a
                    href={`#${id}`}
                    className="text-[var(--dim)] transition-colors hover:text-[var(--text)]"
                  >
                    {label}
                  </a>
                </li>
              ))}
            </ul>
          </nav>

          <div className="min-w-0 pb-20">
            <Section id="install" title="Install">
              <Prose>
                On macOS or Linux, download the latest release for your machine and put it on your
                PATH:
              </Prose>
              <CodeBlock label="macOS / Linux" code={INSTALL_SCRIPT} />
              <Prose>If you have Go 1.26 or newer, install from source instead:</Prose>
              <CodeBlock label="Go" code={GO_INSTALL} />
              <Prose>
                On Windows, download <Code>shapehill_windows_amd64.zip</Code> (or <Code>arm64</Code>
                ) from the{" "}
                <a href={RELEASES} className="text-[var(--text)] underline underline-offset-4">
                  latest release
                </a>{" "}
                and put <Code>shapehill.exe</Code> somewhere on your PATH.
              </Prose>
            </Section>

            <Section id="authenticate" title="Authenticate">
              <Prose>
                The CLI acts as you through a personal access token. Create one on the{" "}
                <Link
                  href="/app/tokens"
                  className="text-[var(--text)] underline underline-offset-4"
                >
                  API tokens
                </Link>{" "}
                page (it is shown once), then export it:
              </Prose>
              <CodeBlock code={AUTH} />
              <Prose>
                Treat a token like a password. Revoke it from the same page if it leaks, and give
                each machine or agent its own so you can revoke them one at a time.
              </Prose>
            </Section>

            <Section id="quickstart" title="Quick start">
              <Prose>
                Create a public hill, add two scopes, move them, then look at where everything sits:
              </Prose>
              <CodeBlock code={QUICKSTART} />
              <CodeBlock label="Output" code={SHOW_OUTPUT} />
              <Prose>
                Print the snippet for your story or README. The image redraws itself whenever a dot
                moves:
              </Prose>
              <CodeBlock code={EMBED} />
            </Section>

            <Section id="positions" title="Positions">
              <Prose>
                A position runs from 0 to 100, and it measures certainty, not percent done.
              </Prose>
              <ul className="mt-6 grid gap-3 sm:grid-cols-3">
                {[
                  ["0–49", "Uphill", "Figuring it out. Unknowns remain.", "#d29922"],
                  ["50–99", "Downhill", "Approach settled. Executing.", "#58a6ff"],
                  ["100", "Done", "Shipped.", "#3fb950"],
                ].map(([range, name, body, color]) => (
                  <li
                    key={name}
                    className="rounded-xl border border-[var(--edge)] bg-[var(--panel)] p-5"
                  >
                    <p className="font-mono text-xs text-[var(--dim)]">{range}</p>
                    <p className="mt-1 font-medium" style={{ color }}>
                      {name}
                    </p>
                    <p className="mt-2 text-sm text-[var(--dim)]">{body}</p>
                  </li>
                ))}
              </ul>
            </Section>

            <Section id="commands" title="Commands">
              <Prose>
                Scopes can be named by their exact title (any case) or their ID. Every command takes{" "}
                <Code>--json</Code>, <Code>--no-color</Code> and <Code>--help</Code>.
              </Prose>
              {COMMANDS.map(({ group, rows }) => (
                <div key={group} className="mt-8">
                  <h3 className="font-mono text-xs uppercase tracking-widest text-[var(--accent)]">
                    {group}
                  </h3>
                  <div className="mt-3 overflow-x-auto rounded-xl border border-[var(--edge)]">
                    <table className="w-full min-w-[36rem] text-left text-sm">
                      <tbody>
                        {rows.map(([command, alias, what]) => (
                          <tr key={command} className="border-b border-[var(--edge)] last:border-0">
                            <td className="whitespace-nowrap px-4 py-3 font-mono text-xs text-[var(--text)]">
                              shapehill {command}
                            </td>
                            <td className="px-4 py-3 font-mono text-xs text-[var(--dim)]">
                              {alias}
                            </td>
                            <td className="px-4 py-3 text-[var(--dim)]">{what}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              ))}
            </Section>

            <Section id="agents" title="AI agents">
              <Prose>
                The CLI is built to be driven by agents like Claude Code. Pass <Code>--json</Code>{" "}
                for output on stdout and errors as <Code>{`{"error", "code"}`}</Code> on stderr,
                then branch on the exit code:
              </Prose>
              <ul className="mt-6 grid gap-2 sm:grid-cols-3">
                {EXIT_CODES.map(([code, meaning]) => (
                  <li
                    key={code}
                    className="flex items-baseline gap-3 rounded-lg border border-[var(--edge)] bg-[var(--panel)] px-4 py-3 text-sm"
                  >
                    <span className="font-mono text-[var(--accent)]">{code}</span>
                    <span className="text-[var(--dim)]">{meaning}</span>
                  </li>
                ))}
              </ul>
              <Prose>
                Retries are safe: <Code>hill create</Code> and <Code>scope add</Code> take{" "}
                <Code>--if-not-exists</Code>, and repeating the latest <Code>scope move</Code>{" "}
                records nothing. To have an agent keep a story&apos;s hill up to date, add this to
                the story repo&apos;s <Code>AGENTS.md</Code> or <Code>CLAUDE.md</Code>:
              </Prose>
              <CodeBlock label="AGENTS.md" code={AGENT_SNIPPET} />
            </Section>

            <Section id="configuration" title="Configuration">
              <div className="mt-6 overflow-x-auto rounded-xl border border-[var(--edge)]">
                <table className="w-full min-w-[32rem] text-left text-sm">
                  <tbody>
                    {ENV.map(([name, what]) => (
                      <tr key={name} className="border-b border-[var(--edge)] last:border-0">
                        <td className="whitespace-nowrap px-4 py-3 font-mono text-xs text-[var(--text)]">
                          {name}
                        </td>
                        <td className="px-4 py-3 text-[var(--dim)]">{what}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              <Prose>
                Tab completion knows your hill slugs and scope titles. Install it once for your
                shell:
              </Prose>
              <CodeBlock code={COMPLETION} />
            </Section>
          </div>
        </div>
      </main>

      <SiteFooter />
    </div>
  );
}
