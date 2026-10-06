"use client";

import { useState } from "react";

export function CodeBlock({ code, label }: { code: string; label?: string }) {
  const [copied, setCopied] = useState(false);

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      setCopied(false);
    }
  };

  return (
    <div className="mt-4 rounded-lg border border-[var(--edge)] bg-[#0d0d0f]">
      <div className="flex items-center justify-between border-b border-[var(--edge)] py-1.5 pr-1.5 pl-4">
        <p className="font-mono text-[11px] uppercase tracking-widest text-[var(--dim)]">{label}</p>
        <button
          type="button"
          onClick={copy}
          className="rounded-md px-2 py-1 font-mono text-[11px] text-[var(--dim)] transition-colors hover:text-[var(--text)] focus-visible:outline-2 focus-visible:outline-[var(--accent)]"
        >
          {copied ? "Copied" : "Copy"}
        </button>
      </div>
      <pre className="overflow-x-auto p-4 font-mono text-xs leading-relaxed text-[var(--accent)]">
        <code>{code}</code>
      </pre>
    </div>
  );
}
