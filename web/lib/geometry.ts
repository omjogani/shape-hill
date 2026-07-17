// Ported from api/internal/hillchart/geometry.go so the browser draws the exact
// same hill the server renders. Keep the two in sync.
export const WIDTH = 900;
export const LEFT = 60;
export const RIGHT = 840;
export const BASELINE = 300;
export const PEAK = 60;

// The interactive chart draws only the hill region (title + terrain + ground
// labels); the scope legend lives in HTML below it.
export const CHART_HEIGHT = 340;

export const clamp = (position: number) =>
  position < 0 ? 0 : position > 100 ? 100 : position;

export const x = (position: number) =>
  LEFT + (clamp(position) / 100) * (RIGHT - LEFT);

// Raised cosine: flat at both feet, flat over the summit.
export const y = (position: number) => {
  const t = clamp(position) / 100;
  return BASELINE - ((BASELINE - PEAK) * (1 - Math.cos(2 * Math.PI * t))) / 2;
};

// Inverse of x: turn a horizontal SVG coordinate back into a 0..100 position.
export const positionFromX = (svgX: number) =>
  clamp(Math.round(((svgX - LEFT) / (RIGHT - LEFT)) * 100));

export const phase = (position: number) =>
  position >= 100 ? "Done" : position < 50 ? "Uphill" : "Downhill";
