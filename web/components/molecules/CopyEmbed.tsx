"use client";

import { useState } from "react";
import { Button } from "../atoms/Button";

// Where the embed image is actually served (the Go API), used to build an
// absolute URL that works when pasted into someone else's README.
const EMBED_BASE = process.env.NEXT_PUBLIC_EMBED_BASE ?? "http://localhost:8080";

export function CopyEmbed({ slug, title }: { slug: string; title: string }) {
  const markdown = `![${title}](${EMBED_BASE}/hill/${slug}.svg)`;
  const [copied, setCopied] = useState(false);

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(markdown);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      setCopied(false);
    }
  };

  return (
    <div className="flex items-center gap-3">
      <code className="min-w-0 flex-1 truncate font-mono text-xs text-ink">{markdown}</code>
      <Button variant="ghost" type="button" onClick={copy} aria-label="Copy embed markdown">
        {copied ? "Copied" : "Copy markdown"}
      </Button>
    </div>
  );
}
