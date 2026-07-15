// Package hillchart renders a hill chart to SVG. It knows nothing about
// databases or HTTP: given a chart, it returns bytes.
package hillchart

import (
	"fmt"
	"math"
	"strings"
)

type Dot struct {
	Label    string
	Note     string
	Color    string
	Position int16
	Stalled  bool
}

type Chart struct {
	Title string
	Dots  []Dot
}

const (
	width    = 900.0
	left     = 60.0
	right    = 840.0
	baseline = 300.0
	peak     = 60.0
	rowStep  = 26.0
)

// x maps a 0-100 position onto the baseline.
func x(position int16) float64 {
	return left + clamp(position)/100*(right-left)
}

// y is a raised cosine: flat at both feet and flat over the summit, which is what
// makes a dot's height read as certainty rather than distance travelled.
func y(position int16) float64 {
	t := clamp(position) / 100
	return baseline - (baseline-peak)*(1-math.Cos(2*math.Pi*t))/2
}

func clamp(position int16) float64 {
	switch {
	case position < 0:
		return 0
	case position > 100:
		return 100
	default:
		return float64(position)
	}
}

func Render(chart Chart) []byte {
	height := baseline + 60 + rowStep*float64(len(chart.Dots))

	var svg strings.Builder
	fmt.Fprintf(&svg, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %.0f %.0f" width="%.0f" height="%.0f" role="img" aria-label="%s">`,
		width, height, width, height, escape(chart.Title+" — hill chart"))

	svg.WriteString(style)
	writeTerrain(&svg, height)
	writeSummit(&svg)

	fmt.Fprintf(&svg, `<text class="title" x="%.0f" y="34">%s</text>`, left, escape(chart.Title))

	for i, dot := range chart.Dots {
		writeDot(&svg, i, dot)
	}
	for i, dot := range chart.Dots {
		writeLegendRow(&svg, i, dot)
	}

	svg.WriteString(`</svg>`)
	return []byte(svg.String())
}

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
    .paper { fill: #121614; }
    .hill  { fill: #1D2320; }
    .paper { fill: #121614; stroke: #1D2320; }
    .edge, .axis { stroke: #DCDDD3; }
    .summit { stroke: #8B948C; }
    .title, .name { fill: #DCDDD3; }
    .label, .note { fill: #8B948C; }
    .num { fill: #121614; }
    .ring { stroke: #121614; }
    .stalled { fill: #D9806E; }
  }
</style>`

func writeTerrain(svg *strings.Builder, height float64) {
	// Inset by 1 so the 2px border isn't clipped in half by the canvas edge.
	fmt.Fprintf(svg, `<rect class="paper" x="1" y="1" width="%.0f" height="%.0f" rx="10"/>`, width-2, height-2)

	var curve strings.Builder
	for step := 0; step <= 240; step++ {
		position := int16(step * 100 / 240)
		point := fmt.Sprintf("%.1f %.1f", x(position), y(position))
		if step == 0 {
			curve.WriteString("M " + point)
		} else {
			curve.WriteString(" L " + point)
		}
	}

	fmt.Fprintf(svg, `<path class="hill" d="%s L %.0f %.0f L %.0f %.0f Z"/>`,
		curve.String(), right, baseline, left, baseline)
	fmt.Fprintf(svg, `<path class="edge" d="%s"/>`, curve.String())
	fmt.Fprintf(svg, `<line class="axis" x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f"/>`,
		left, baseline, right, baseline)
}

func writeSummit(svg *strings.Builder) {
	mid := x(50)
	fmt.Fprintf(svg, `<line class="summit" x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f"/>`, mid, peak-8, mid, baseline)
	fmt.Fprintf(svg, `<text class="label" x="%.0f" y="%.0f" text-anchor="middle">SUMMIT</text>`, mid, peak-16)
	fmt.Fprintf(svg, `<text class="label" x="%.0f" y="%.0f">FIGURING IT OUT</text>`, left, baseline+22)
	fmt.Fprintf(svg, `<text class="label" x="%.0f" y="%.0f" text-anchor="end">MAKING IT HAPPEN</text>`, right, baseline+22)
}

func writeDot(svg *strings.Builder, index int, dot Dot) {
	fill := dot.Color
	class := "ring"
	if dot.Stalled {
		fill = ""
		class = "ring stalled"
	}

	fmt.Fprintf(svg, `<circle class="%s" cx="%.1f" cy="%.1f" r="11"`, class, x(dot.Position), y(dot.Position))
	if fill != "" {
		fmt.Fprintf(svg, ` fill="%s"`, escape(fill))
	}
	svg.WriteString(`/>`)

	fmt.Fprintf(svg, `<text class="num" x="%.1f" y="%.1f" text-anchor="middle" dominant-baseline="central">%d</text>`,
		x(dot.Position), y(dot.Position), index+1)
}

func writeLegendRow(svg *strings.Builder, index int, dot Dot) {
	top := baseline + 52 + rowStep*float64(index)

	fill := dot.Color
	class := ""
	if dot.Stalled {
		fill = ""
		class = "stalled"
	}

	fmt.Fprintf(svg, `<rect class="%s" x="%.0f" y="%.0f" width="16" height="16" rx="2"`, class, left, top-12)
	if fill != "" {
		fmt.Fprintf(svg, ` fill="%s"`, escape(fill))
	}
	svg.WriteString(`/>`)

	fmt.Fprintf(svg, `<text class="num" x="%.0f" y="%.0f" text-anchor="middle" dominant-baseline="central">%d</text>`,
		left+8, top-4, index+1)
	fmt.Fprintf(svg, `<text class="name" x="%.0f" y="%.0f">%s</text>`, left+26, top, escape(dot.Label))

	status := phase(dot.Position)
	if dot.Stalled {
		status = "Not moving"
	}
	if dot.Note != "" {
		status += " — " + dot.Note
	}
	fmt.Fprintf(svg, `<text class="note" x="%.0f" y="%.0f" text-anchor="end">%s</text>`, right, top, escape(status))
}

func phase(position int16) string {
	if position >= 100 {
		return "Done"
	}
	if position < 50 {
		return "Uphill"
	}
	return "Downhill"
}

var escaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&quot;",
	"'", "&apos;",
)

func escape(s string) string { return escaper.Replace(s) }
