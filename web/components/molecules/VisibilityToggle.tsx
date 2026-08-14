"use client";

import { Toggle } from "../atoms/Toggle";
import { useSetVisibility } from "@/lib/hooks";

export function VisibilityToggle({ slug, isPublic }: { slug: string; isPublic: boolean }) {
  const setVisibility = useSetVisibility(slug);

  return (
    <Toggle
      on={isPublic}
      label={isPublic ? "Public" : "Private"}
      ariaLabel={`Hill is ${isPublic ? "public" : "private"} — make it ${isPublic ? "private" : "public"}`}
      disabled={setVisibility.isPending}
      onToggle={() => setVisibility.mutate(!isPublic)}
    />
  );
}
