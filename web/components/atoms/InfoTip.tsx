"use client";

export function InfoTip({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <span className="group relative inline-flex">
      <button
        type="button"
        aria-label={label}
        className="flex h-4 w-4 items-center justify-center rounded-full border border-line font-mono text-[10px] leading-none text-sage focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sage"
      >
        i
      </button>
      <span
        role="tooltip"
        className="pointer-events-none invisible absolute right-0 top-full z-20 mt-2 w-64 rounded-xl border border-line bg-paper p-3 text-xs leading-relaxed text-sage shadow-lg group-focus-within:visible group-hover:visible"
      >
        {children}
      </span>
    </span>
  );
}
