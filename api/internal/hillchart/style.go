package hillchart

// The stylesheet travels inside the SVG, so the image follows the reader's theme
// even when it is embedded as a plain <img> in someone else's page.
const style = `<style>
  .paper { fill: #E9E7DE; stroke: #DEDBCB; stroke-width: 2; }
  .hill  { fill: #DEDBCB; }
  .edge  { fill: none; stroke: #1B211D; stroke-width: 2; }
  .axis  { stroke: #1B211D; stroke-width: 1.5; }
  .summit { stroke: #5D665F; stroke-width: 1; stroke-dasharray: 3 4; }
  .title { font: 600 19px Georgia, serif; fill: #1B211D; }
  .label { font: 11px ui-monospace, Menlo, monospace; fill: #5D665F; letter-spacing: 1.2px; }
  .num   { font: 600 11px ui-monospace, Menlo, monospace; fill: #E9E7DE; }
  .name  { font: 14px Georgia, serif; fill: #1B211D; }
  .note  { font: 11px ui-monospace, Menlo, monospace; fill: #5D665F; }
  .ring  { stroke: #E9E7DE; stroke-width: 2.5; }
  .stalled { fill: #A9483A; }
  @media (prefers-color-scheme: dark) {
    .paper { fill: #121614; stroke: #1D2320; }
    .hill  { fill: #1D2320; }
    .edge, .axis { stroke: #DCDDD3; }
    .summit { stroke: #8B948C; }
    .title, .name { fill: #DCDDD3; }
    .label, .note { fill: #8B948C; }
    .num { fill: #121614; }
    .ring { stroke: #121614; }
    .stalled { fill: #D9806E; }
  }
</style>`
