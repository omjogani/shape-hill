package hillchart

type Style string

const (
	StyleDefault Style = ""       // warm "paper" theme
	StyleGitHub  Style = "github" // blends into a GitHub README, light or dark
)

var styles = map[Style]string{
	StyleDefault: styleDefault,
	StyleGitHub:  styleGitHub,
}

func styleFor(s Style) string {
	if css, ok := styles[s]; ok {
		return css
	}
	return styleDefault
}

// The stylesheet travels inside the SVG, so the image follows the reader's theme
// even when it is embedded as a plain <img> in someone else's page.
const styleDefault = `<style>
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

// styleGitHub uses GitHub's own canvas, border and text colours plus its system
// font stack, so the chart reads as a native part of a README in either theme.
const styleGitHub = `<style>
  .paper { fill: #ffffff; stroke: #d0d7de; stroke-width: 2; }
  .hill  { fill: #f6f8fa; }
  .edge  { fill: none; stroke: #1f2328; stroke-width: 2; }
  .axis  { stroke: #1f2328; stroke-width: 1.5; }
  .summit { stroke: #656d76; stroke-width: 1; stroke-dasharray: 3 4; }
  .title { font: 600 19px -apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans", Helvetica, Arial, sans-serif; fill: #1f2328; }
  .label { font: 11px ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace; fill: #656d76; letter-spacing: 1.2px; }
  .num   { font: 600 11px ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace; fill: #ffffff; }
  .name  { font: 14px -apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans", Helvetica, Arial, sans-serif; fill: #1f2328; }
  .note  { font: 11px ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace; fill: #656d76; }
  .ring  { stroke: #ffffff; stroke-width: 2.5; }
  .stalled { fill: #d1242f; }
  @media (prefers-color-scheme: dark) {
    .paper { fill: #0d1117; stroke: #30363d; }
    .hill  { fill: #161b22; }
    .edge, .axis { stroke: #e6edf3; }
    .summit { stroke: #8b949e; }
    .title, .name { fill: #e6edf3; }
    .label, .note { fill: #8b949e; }
    .ring { stroke: #0d1117; }
    .stalled { fill: #f85149; }
  }
</style>`
