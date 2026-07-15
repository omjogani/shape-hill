package hillchart

import (
	"encoding/xml"
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestRendersWellFormedXML(t *testing.T) {
	svg := Render(Chart{
		Title: "Billing v2",
		Dots: []Dot{
			{Label: "Card on file", Note: "Shipped", Color: "#2F4C64", Position: 95},
			{Label: "Admin refunds", Note: "Permissions unclear", Color: "#55704B", Position: 15, Stalled: true},
		},
	})

	decoder := xml.NewDecoder(strings.NewReader(string(svg)))
	for {
		_, err := decoder.Token()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			t.Fatalf("SVG is not well-formed XML: %v", err)
		}
	}
}

func TestFeetAndSummit(t *testing.T) {
	if got := y(0); math.Abs(got-baseline) > 0.01 {
		t.Errorf("position 0 should sit on the baseline, got y=%.2f want %.2f", got, baseline)
	}
	if got := y(100); math.Abs(got-baseline) > 0.01 {
		t.Errorf("position 100 should sit on the baseline, got y=%.2f want %.2f", got, baseline)
	}
	if got := y(50); math.Abs(got-peak) > 0.01 {
		t.Errorf("position 50 should sit at the summit, got y=%.2f want %.2f", got, peak)
	}
	if y(25) <= y(50) || y(25) >= y(0) {
		t.Errorf("position 25 should be partway up the hill, got y=%.2f", y(25))
	}
}

func TestXIncreasesWithPosition(t *testing.T) {
	previous := -1.0
	for position := int16(0); position <= 100; position += 10 {
		current := x(position)
		if current <= previous {
			t.Fatalf("x should increase with position: x(%d)=%.2f is not past %.2f", position, current, previous)
		}
		previous = current
	}
	if got := x(0); math.Abs(got-left) > 0.01 {
		t.Errorf("position 0 should start at the left edge, got %.2f want %.2f", got, left)
	}
	if got := x(100); math.Abs(got-right) > 0.01 {
		t.Errorf("position 100 should end at the right edge, got %.2f want %.2f", got, right)
	}
}

func TestPositionsOutsideRangeAreClamped(t *testing.T) {
	if x(-40) != x(0) {
		t.Error("a negative position should clamp to the left foot")
	}
	if x(9000) != x(100) {
		t.Error("a position past 100 should clamp to the right foot")
	}
}

func TestStalledDotDropsItsColor(t *testing.T) {
	stalled := string(Render(Chart{Dots: []Dot{{Label: "Refunds", Color: "#55704B", Position: 15, Stalled: true}}}))
	if strings.Contains(stalled, "#55704B") {
		t.Error("a stalled dot must not keep its own colour, or the alarm is invisible")
	}
	if !strings.Contains(stalled, "stalled") {
		t.Error("a stalled dot should carry the stalled class")
	}

	moving := string(Render(Chart{Dots: []Dot{{Label: "Refunds", Color: "#55704B", Position: 15}}}))
	if !strings.Contains(moving, "#55704B") {
		t.Error("a moving dot should keep its own colour")
	}
}

func TestLabelsAreEscaped(t *testing.T) {
	svg := string(Render(Chart{
		Title: `Tom & "Jerry"`,
		Dots:  []Dot{{Label: `<script>alert(1)</script>`, Position: 50}},
	}))
	if strings.Contains(svg, "<script>") {
		t.Fatal("scope titles must be escaped: an unescaped title is script injection into every ticket that embeds the chart")
	}
	if !strings.Contains(svg, "&amp;") {
		t.Error("ampersand in the title should be escaped")
	}
}

func TestPhaseNaming(t *testing.T) {
	cases := map[int16]string{0: "Uphill", 49: "Uphill", 50: "Downhill", 99: "Downhill", 100: "Done"}
	for position, want := range cases {
		if got := phase(position); got != want {
			t.Errorf("phase(%d) = %q, want %q", position, got, want)
		}
	}
}

// The canvas has to grow with the legend, or scope rows fall off the bottom of the
// image and land on bare page in whatever ticket embedded it.
func TestCanvasGrowsToFitEveryLegendRow(t *testing.T) {
	for _, count := range []int{0, 1, 4, 12} {
		dots := make([]Dot, count)
		for i := range dots {
			dots[i] = Dot{Label: "Scope", Position: 50}
		}
		svg := string(Render(Chart{Title: "Billing v2", Dots: dots}))

		height := baseline + 60 + rowStep*float64(count)
		background := fmt.Sprintf(`<rect class="paper" x="0" y="0" width="%.0f" height="%.0f"/>`, width, height)
		if !strings.Contains(svg, background) {
			t.Errorf("%d dots: background should span the full canvas, want %s", count, background)
		}
		if count > 0 {
			lowest := baseline + 52 + rowStep*float64(count-1)
			if lowest > height {
				t.Errorf("%d dots: last legend row at y=%.0f falls outside a canvas of %.0f", count, lowest, height)
			}
		}
	}
}

func TestEmptyChartStillRenders(t *testing.T) {
	svg := string(Render(Chart{Title: "Nothing bet yet"}))
	if !strings.Contains(svg, "</svg>") {
		t.Error("a hill with no scopes should still draw the hill")
	}
}
