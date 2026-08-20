"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import { api, ApiError } from "@/lib/api";
import { Button } from "@/components/atoms/Button";
import { TextInput } from "@/components/atoms/TextInput";
import { FieldError } from "@/components/atoms/FieldError";

export default function OnboardingPage() {
  const [username, setUsername] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const router = useRouter();
  const qc = useQueryClient();

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await api.onboard(username.trim());
      await qc.invalidateQueries({ queryKey: ["me"] });
      router.replace("/app");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Something went wrong");
      setBusy(false);
    }
  }

  return (
    <main className="flex flex-1 flex-col items-center justify-center px-8">
      <form onSubmit={onSubmit} className="flex w-full max-w-sm flex-col gap-4">
        <div>
          <p className="font-mono text-xs uppercase tracking-widest text-sage">shapehill</p>
          <h1 className="font-display text-2xl">Pick a username</h1>
          <p className="mt-1 text-sm text-sage">Lowercase letters, numbers, _ or -. This is how you appear.</p>
        </div>
        <TextInput
          autoFocus
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          placeholder="username"
          aria-label="username"
        />
        <FieldError errors={[error]} />
        <Button type="submit" disabled={busy || username.trim() === ""}>
          {busy ? "Creating…" : "Continue"}
        </Button>
      </form>
    </main>
  );
}
