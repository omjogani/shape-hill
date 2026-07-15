package hillchart

import (
	"fmt"
	"strings"
)

func (c *canvas) drawBackground(height float64) {
	// Inset by 1 so the 2px border isn't clipped in half by the canvas edge.
	fmt.Fprintf(c, `<rect class="paper" x="1" y="1" width="%.0f" height="%.0f" rx="10"/>`, width-2, height-2)

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

	fmt.Fprintf(c, `<path class="hill" d="%s L %.0f %.0f L %.0f %.0f Z"/>`,
		curve.String(), right, baseline, left, baseline)
	fmt.Fprintf(c, `<path class="edge" d="%s"/>`, curve.String())
	fmt.Fprintf(c, `<line class="axis" x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f"/>`,
		left, baseline, right, baseline)
}

func (c *canvas) drawTitle(title string) {
	fmt.Fprintf(c, `<text class="title" x="%.0f" y="34">%s</text>`, left, escape(title))
}

func (c *canvas) drawSummit() {
	mid := x(50)
	fmt.Fprintf(c, `<line class="summit" x1="%.0f" y1="%.0f" x2="%.0f" y2="%.0f"/>`, mid, peak-8, mid, baseline)
	fmt.Fprintf(c, `<text class="label" x="%.0f" y="%.0f" text-anchor="middle">SUMMIT</text>`, mid, peak-16)
	fmt.Fprintf(c, `<text class="label" x="%.0f" y="%.0f">FIGURING IT OUT</text>`, left, baseline+22)
	fmt.Fprintf(c, `<text class="label" x="%.0f" y="%.0f" text-anchor="end">MAKING IT HAPPEN</text>`, right, baseline+22)
}

func (c *canvas) drawScope(index int, dot Dot) {
	fill := dot.Color
	class := "ring"
	if dot.Stalled {
		fill = ""
		class = "ring stalled"
	}

	fmt.Fprintf(c, `<circle class="%s" cx="%.1f" cy="%.1f" r="11"`, class, x(dot.Position), y(dot.Position))
	if fill != "" {
		fmt.Fprintf(c, ` fill="%s"`, escape(fill))
	}
	c.WriteString(`/>`)

	fmt.Fprintf(c, `<text class="num" x="%.1f" y="%.1f" text-anchor="middle" dominant-baseline="central">%d</text>`,
		x(dot.Position), y(dot.Position), index+1)
}

func (c *canvas) drawLegendEntry(index int, dot Dot) {
	top := baseline + 52 + rowStep*float64(index)

	fill := dot.Color
	class := ""
	if dot.Stalled {
		fill = ""
		class = "stalled"
	}

	fmt.Fprintf(c, `<rect class="%s" x="%.0f" y="%.0f" width="16" height="16" rx="2"`, class, left, top-12)
	if fill != "" {
		fmt.Fprintf(c, ` fill="%s"`, escape(fill))
	}
	c.WriteString(`/>`)

	fmt.Fprintf(c, `<text class="num" x="%.0f" y="%.0f" text-anchor="middle" dominant-baseline="central">%d</text>`,
		left+8, top-4, index+1)
	fmt.Fprintf(c, `<text class="name" x="%.0f" y="%.0f">%s</text>`, left+26, top, escape(dot.Label))

	status := phase(dot.Position)
	if dot.Stalled {
		status = "Not moving"
	}
	if dot.Note != "" {
		status += " — " + dot.Note
	}
	fmt.Fprintf(c, `<text class="note" x="%.0f" y="%.0f" text-anchor="end">%s</text>`, right, top, escape(status))
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
