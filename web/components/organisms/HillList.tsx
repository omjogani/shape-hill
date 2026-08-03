"use client";

import Link from "next/link";
import { useHills } from "@/lib/hooks";

export function HillList() {
  const { data: hills, isLoading, isError, error } = useHills();

  return (
    <main className="mx-auto flex w-full max-w-3xl flex-col gap-8 px-8 py-12">
      <header>
        <p className="font-mono text-xs uppercase tracking-widest text-sage">shapehill</p>
        <h1 className="font-display text-3xl">Hill charts</h1>
      </header>

      {isLoading && <p className="text-sm text-sage">Loading charts…</p>}

      {isError && (
        <p className="text-sm text-alarm">
          {(error as Error)?.message}. Is the API running on :8080?
        </p>
      )}

      {hills?.length === 0 && <p className="text-sm text-sage">No charts yet.</p>}

      <ul className="flex flex-col gap-3">
        {hills?.map((hill) => (
          <li key={hill.ID}>
            <Link
              href={`/${hill.Slug}`}
              className="flex items-center justify-between gap-4 rounded-lg border border-sage/20 px-5 py-4 hover:border-sage/50"
            >
              <span className="min-w-0">
                <span className="block truncate font-display text-lg">{hill.Title}</span>
                <span className="font-mono text-xs text-sage">{hill.Slug}</span>
              </span>
              <span className="shrink-0 font-mono text-xs uppercase tracking-widest text-sage">
                {hill.IsPublic ? "public" : "private"}
              </span>
            </Link>
          </li>
        ))}
      </ul>
    </main>
  );
}
