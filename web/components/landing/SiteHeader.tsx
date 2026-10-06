import Link from "next/link";
import { HeaderCta } from "./HeaderCta";

export const GITHUB = "https://github.com/omjogani/shape-hill";

export function SiteHeader() {
  return (
    <header className="mx-auto flex w-full max-w-6xl items-center justify-between px-6 py-6">
      <Link href="/" className="font-mono text-sm tracking-widest">
        shapehill
      </Link>
      <nav className="flex items-center gap-2 text-sm">
        <Link
          href="/docs/cli"
          className="rounded-md px-3 py-1.5 text-[var(--dim)] transition-colors hover:text-[var(--text)]"
        >
          CLI
        </Link>
        <a
          href={GITHUB}
          className="rounded-md px-3 py-1.5 text-[var(--dim)] transition-colors hover:text-[var(--text)]"
        >
          GitHub
        </a>
        <HeaderCta />
      </nav>
    </header>
  );
}
