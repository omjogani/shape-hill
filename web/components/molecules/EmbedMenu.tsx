"use client";

import { useEffect, useState, useSyncExternalStore } from "react";
import { Button } from "../atoms/Button";
import { getServerStyle, getStyle, setStyle, subscribe } from "@/lib/style-store";

const EMBED_BASE = process.env.NEXT_PUBLIC_EMBED_BASE ?? "http://localhost:8080";

const STYLES = [
  { value: "", label: "Paper (default)" },
  { value: "github", label: "GitHub" },
];

export function EmbedMenu({ slug, title }: { slug: string; title: string }) {
  const style = useSyncExternalStore(subscribe, getStyle, getServerStyle);
  const [copied, setCopied] = useState(false);
  const [open, setOpen] = useState(false);

  // The chosen style drives the whole page's palette and fonts, not just the
  // image — globals.css keys off <html data-style="…">.
  useEffect(() => {
    const root = document.documentElement;
    if (style) root.dataset.style = style;
    else delete root.dataset.style;
  }, [style]);

  const url = `${EMBED_BASE}/hill/${slug}.svg${style ? `?style=${style}` : ""}`;
  const markdown = `![${title}](${url})`;

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
    <div className="relative shrink-0">
      <Button type="button" variant="ghost" onClick={() => setOpen((o) => !o)} aria-expanded={open}>
        Embed
      </Button>

      {open && (
        <>
          {/* Click-away layer */}
          <button
            type="button"
            aria-label="Close embed menu"
            className="fixed inset-0 z-10 cursor-default"
            onClick={() => setOpen(false)}
          />
          <div className="absolute right-0 z-20 mt-2 w-[min(26rem,calc(100vw-3rem))] rounded-xl border border-line bg-paper p-4 shadow-lg">
            <label
              htmlFor="embed-style"
              className="mb-1.5 block font-mono text-[11px] uppercase tracking-widest text-sage"
            >
              Style
            </label>
            <select
              id="embed-style"
              value={style}
              onChange={(e) => setStyle(e.target.value)}
              className="w-full rounded-md border border-line bg-paper px-2.5 py-1.5 text-sm text-ink [&_option]:bg-paper [&_option]:text-ink focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-sage"
            >
              {STYLES.map((s) => (
                <option key={s.value} value={s.value}>
                  {s.label}
                </option>
              ))}
            </select>
            <p className="mt-1.5 text-xs text-sage">Changes the image and this site&apos;s look.</p>

            <p className="mb-1.5 mt-4 font-mono text-[11px] uppercase tracking-widest text-sage">
              Markdown
            </p>
            <code className="block overflow-x-auto whitespace-pre rounded-md border border-line bg-hill/40 p-2.5 font-mono text-[11px] text-ink">
              {markdown}
            </code>

            <div className="mt-3 flex justify-end">
              <Button type="button" onClick={copy}>
                {copied ? "Copied" : "Copy markdown"}
              </Button>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
