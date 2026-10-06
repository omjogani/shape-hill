"use client";

import Link from "next/link";
import { useState, type SubmitEvent } from "react";
import { useCreateToken, useDeleteToken, useTokens } from "@/lib/hooks";
import { Button } from "../atoms/Button";
import { TextInput } from "../atoms/TextInput";

const formatDate = (iso: string) =>
  new Date(iso).toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });

export function TokenManager() {
  const { data: tokens, isLoading, isError, error } = useTokens();
  const create = useCreateToken();
  const remove = useDeleteToken();
  const [name, setName] = useState("");
  const [fresh, setFresh] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const submit = (e: SubmitEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!name.trim()) return;
    create.mutate(name.trim(), {
      onSuccess: ({ token }) => {
        setFresh(token);
        setCopied(false);
        setName("");
      },
    });
  };

  const copy = async () => {
    if (!fresh) return;
    try {
      await navigator.clipboard.writeText(fresh);
      setCopied(true);
    } catch {
      setCopied(false);
    }
  };

  const revoke = (id: string, tokenName: string) => {
    if (!window.confirm(`Revoke "${tokenName}"? Anything using it stops working.`)) return;
    remove.mutate(id);
  };

  return (
    <main className="mx-auto flex w-full max-w-3xl flex-col gap-8 px-8 py-12">
      <header>
        <Link href="/app" className="font-mono text-xs uppercase tracking-widest text-sage">
          ← shapehill
        </Link>
        <h1 className="font-display text-3xl">API tokens</h1>
        <p className="mt-2 text-sm text-sage">
          Tokens let the shapehill CLI and AI agents edit your hills as you. Treat one like a
          password.
        </p>
      </header>

      <form onSubmit={submit} className="flex items-end gap-2">
        <label className="flex flex-1 flex-col gap-1.5">
          <span className="font-mono text-[11px] uppercase tracking-widest text-sage">Name</span>
          <TextInput
            placeholder="laptop CLI"
            maxLength={64}
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
        </label>
        <Button type="submit" disabled={!name.trim() || create.isPending}>
          {create.isPending ? "Creating…" : "Create token"}
        </Button>
      </form>
      {create.isError && <p className="text-xs text-alarm">{(create.error as Error).message}</p>}

      {fresh && (
        <div className="rounded-xl border border-line bg-hill/40 p-4">
          <p className="text-sm text-ink">Copy this token now. You won&apos;t see it again.</p>
          <code className="mt-2 block overflow-x-auto whitespace-pre rounded-md border border-line bg-paper p-2.5 font-mono text-xs text-ink">
            {fresh}
          </code>
          <div className="mt-3 flex justify-end gap-2">
            <Button type="button" variant="ghost" onClick={() => setFresh(null)}>
              Done
            </Button>
            <Button type="button" onClick={copy}>
              {copied ? "Copied" : "Copy token"}
            </Button>
          </div>
        </div>
      )}

      {isLoading && <p className="text-sm text-sage">Loading tokens…</p>}
      {isError && <p className="text-sm text-alarm">{(error as Error)?.message}</p>}
      {tokens?.length === 0 && <p className="text-sm text-sage">No tokens yet.</p>}
      {remove.isError && <p className="text-xs text-alarm">{(remove.error as Error).message}</p>}

      <ul className="flex flex-col gap-3">
        {tokens?.map((token) => (
          <li
            key={token.ID}
            className="flex items-center justify-between gap-4 rounded-lg border border-sage/20 px-5 py-4"
          >
            <span className="min-w-0">
              <span className="block truncate font-display text-lg">{token.Name}</span>
              <span className="font-mono text-xs text-sage">
                {token.Hint}… · created {formatDate(token.CreatedAt)} ·{" "}
                {token.LastUsedAt ? `last used ${formatDate(token.LastUsedAt)}` : "never used"}
              </span>
            </span>
            <Button
              type="button"
              variant="ghost"
              disabled={remove.isPending}
              onClick={() => revoke(token.ID, token.Name)}
            >
              Revoke
            </Button>
          </li>
        ))}
      </ul>
    </main>
  );
}
