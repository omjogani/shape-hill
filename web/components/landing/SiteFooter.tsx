import Link from "next/link";
import { GITHUB } from "./SiteHeader";

export function SiteFooter() {
  return (
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
        <Link href="/docs/cli" className="transition-colors hover:text-[var(--text)]">
          CLI
        </Link>
        <a href={GITHUB} className="transition-colors hover:text-[var(--text)]">
          GitHub
        </a>
        <Link href="/app" className="transition-colors hover:text-[var(--text)]">
          Open the app
        </Link>
      </div>
    </footer>
  );
}
