// Package hillchart renders a hill chart to SVG. It knows nothing about
// databases or HTTP: given a chart, it returns bytes.
package hillchart

import (
	"fmt"
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
	Style Style
}

type canvas struct {
	strings.Builder
}

func Render(chart Chart) []byte {
	height := baseline + 60 + rowStep*float64(len(chart.Dots))

	c := &canvas{}
	fmt.Fprintf(c, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %.0f %.0f" width="%.0f" height="%.0f" role="img" aria-label="%s">`,
		width, height, width, height, escape(chart.Title+" — hill chart"))

	c.WriteString(styleFor(chart.Style))
	c.drawBackground(height)
	c.drawSummit()
	c.drawTitle(chart.Title)

	for i, dot := range chart.Dots {
		c.drawScope(i, dot)
	}
	for i, dot := range chart.Dots {
		c.drawLegendEntry(i, dot)
	}

	c.WriteString(`</svg>`)
	return []byte(c.String())
}

var escaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&quot;",
	"'", "&apos;",
)

func escape(s string) string { return escaper.Replace(s) }
