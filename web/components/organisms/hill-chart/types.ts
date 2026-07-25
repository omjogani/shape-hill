export type ChartDot = {
  id: string;
  label: string;
  color: string;
  position: number;
  stalled: boolean;
  /** Moved but not yet saved — drawn with a dashed halo. */
  pending?: boolean;
};
