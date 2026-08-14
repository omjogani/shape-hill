"use client";

import { InfoTip } from "../atoms/InfoTip";
import { Toggle } from "../atoms/Toggle";
import { useSetTrackStalled } from "@/lib/hooks";

export function StallToggle({ slug, trackStalled }: { slug: string; trackStalled: boolean }) {
  const setTrackStalled = useSetTrackStalled(slug);

  return (
    <div className="flex items-center gap-1.5">
      <Toggle
        on={trackStalled}
        label={trackStalled ? "Stall alerts" : "Alerts off"}
        ariaLabel={`Not-moving alerts are ${trackStalled ? "on" : "off"} - turn them ${trackStalled ? "off" : "on"}`}
        disabled={setTrackStalled.isPending}
        onToggle={() => setTrackStalled.mutate(!trackStalled)}
      />
      <InfoTip label="What are stall alerts?">
        Scopes that sit still for a week turn red, so stuck work is easy to spot.
      </InfoTip>
    </div>
  );
}
