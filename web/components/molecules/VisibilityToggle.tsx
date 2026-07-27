"use client";

import { useSetVisibility } from "@/lib/hooks";

export function VisibilityToggle({ slug, isPublic }: { slug: string; isPublic: boolean }) {
  const setVisibility = useSetVisibility(slug);

  return (
    <button
      type="button"
      role="switch"
      aria-checked={isPublic}
      aria-label={`Hill is ${isPublic ? "public" : "private"} — make it ${isPublic ? "private" : "public"}`}
      onClick={() => setVisibility.mutate(!isPublic)}
      disabled={setVisibility.isPending}
      className="flex items-center gap-2 rounded-md px-1 py-1 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sage disabled:opacity-50"
    >
      <span
        className={`relative h-5 w-9 rounded-full transition-colors ${isPublic ? "bg-sage" : "bg-hill"}`}
      >
        <span
          className={`absolute left-0.5 top-0.5 h-4 w-4 rounded-full transition-transform ${
            isPublic ? "translate-x-4 bg-paper" : "bg-sage"
          }`}
        />
      </span>
      <span className="font-mono text-[11px] uppercase tracking-widest text-sage">
        {isPublic ? "Public" : "Private"}
      </span>
    </button>
  );
}
