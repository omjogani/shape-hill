import type { Hill, Scope } from "./api";

// Mirrors the server's stalledAfter: a scope untouched for a week (and not done).
const STALLED_AFTER_MS = 7 * 24 * 60 * 60 * 1000;

// A hill with TrackStalled off never flags anything: the chart stays static.
export const stalledIn = (hill: Hill) => (s: Scope) =>
  hill.TrackStalled &&
  Date.now() - new Date(s.MovedAt).getTime() > STALLED_AFTER_MS &&
  s.Position < 100;
