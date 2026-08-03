"use client";

import { useRef, useState, type SubmitEvent } from "react";
import { useRouter } from "next/navigation";
import { useCreateHill } from "@/lib/hooks";
import { Button } from "../atoms/Button";
import { TextInput } from "../atoms/TextInput";

const slugify = (s: string) => {
  return s
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
};

export function NewHillDialog() {
  const router = useRouter();
  const create = useCreateHill();
  const ref = useRef<HTMLDialogElement>(null);
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [slugEdited, setSlugEdited] = useState(false);

  const open = () => {
    setName("");
    setSlug("");
    setSlugEdited(false);
    create.reset();
    ref.current?.showModal();
  };

  const onName = (value: string) => {
    setName(value);
    if (!slugEdited) setSlug(slugify(value));
  };

  const submit = (e: SubmitEvent<HTMLFormElement>) => {
    e.preventDefault();
    const finalSlug = slugify(slug);
    if (!name.trim() || !finalSlug) return;
    create.mutate(
      { title: name.trim(), slug: finalSlug },
      { onSuccess: (hill) => router.push(`/${hill.Slug}`) },
    );
  };

  return (
    <>
      <Button type="button" onClick={open}>
        New hill
      </Button>

      <dialog
        ref={ref}
        onClose={() => create.reset()}
        className="m-auto w-[min(28rem,calc(100vw-2rem))] rounded-xl border border-line bg-paper p-6 text-ink shadow-lg backdrop:bg-ink/30"
      >
        <form onSubmit={submit} className="flex flex-col gap-4">
          <h2 className="font-display text-xl">New hill chart</h2>

          <label className="flex flex-col gap-1.5">
            <span className="font-mono text-[11px] uppercase tracking-widest text-sage">Name</span>
            <TextInput
              autoFocus
              placeholder="Auth Module for ShapeFile"
              value={name}
              onChange={(e) => onName(e.target.value)}
            />
          </label>

          <label className="flex flex-col gap-1.5">
            <span className="font-mono text-[11px] uppercase tracking-widest text-sage">Slug</span>
            <TextInput
              placeholder="auth-module-for-shapefile"
              value={slug}
              onChange={(e) => {
                setSlugEdited(true);
                setSlug(e.target.value.toLowerCase().replace(/[^a-z0-9-]+/g, "-"));
              }}
            />
            <span className="font-mono text-xs text-sage">The chart lives at /{slug || "…"}</span>
          </label>

          {create.isError && (
            <p className="text-xs text-alarm">{(create.error as Error).message}</p>
          )}

          <div className="flex justify-end gap-2">
            <Button type="button" variant="ghost" onClick={() => ref.current?.close()}>
              Cancel
            </Button>
            <Button type="submit" disabled={!name.trim() || !slugify(slug) || create.isPending}>
              {create.isPending ? "Creating…" : "Create"}
            </Button>
          </div>
        </form>
      </dialog>
    </>
  );
}
