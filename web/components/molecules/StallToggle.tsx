"use client";

import { Toggle } from "../atoms/Toggle";
import { useSetTrackStalled } from "@/lib/hooks";

export function StallToggle({ slug, trackStalled }: { slug: string; trackStalled: boolean }) {
  const setTrackStalled = useSetTrackStalled(slug);

  return (
    <Toggle
      on={trackStalled}
      label={trackStalled ? "Stall alerts" : "Alerts off"}
      ariaLabel={`Not-moving alerts are ${trackStalled ? "on" : "off"} — turn them ${trackStalled ? "off" : "on"}`}
      disabled={setTrackStalled.isPending}
      onToggle={() => setTrackStalled.mutate(!trackStalled)}
    />
  );
}
