package cli

import (
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fatih/color"

	"github.com/omjogani/shape-hill/internal/hills"
)

var (
	bold    = color.New(color.Bold).SprintFunc()
	faint   = color.New(color.Faint).SprintFunc()
	accent  = color.New(color.FgCyan).SprintFunc()
	success = color.New(color.FgGreen, color.Bold).SprintFunc()
	warning = color.New(color.FgYellow).SprintFunc()
	failure = color.New(color.FgRed, color.Bold).SprintFunc()
	heading = color.New(color.Bold, color.Underline).SprintFunc()
)

var phaseColors = map[string]func(...any) string{
	"uphill":   color.New(color.FgYellow).SprintFunc(),
	"downhill": color.New(color.FgBlue).SprintFunc(),
	"done":     color.New(color.FgGreen).SprintFunc(),
	"stalled":  color.New(color.FgRed).SprintFunc(),
}

func paintPhase(s hills.Scope, stalled bool) string {
	if stalled {
		return phaseColors["stalled"](s.Phase() + " · stalled")
	}
	return phaseColors[s.Phase()](s.Phase())
}

const trackWidth = 21

// track draws ──────●───┼──────────, where ┼ is the top of the hill.
func track(s hills.Scope, stalled bool) string {
	runes := []rune(strings.Repeat("─", trackWidth))
	runes[trackWidth/2] = '┼'
	dot := int(s.Position) * (trackWidth - 1) / hills.Summit

	paint := phaseColors[s.Phase()]
	if stalled {
		paint = phaseColors["stalled"]
	}
	return faint(string(runes[:dot])) + paint("●") + faint(string(runes[dot+1:]))
}

func printOK(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, "%s %s\n", success("✓"), fmt.Sprintf(format, args...))
}

func printWarn(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, "%s %s\n", warning("!"), fmt.Sprintf(format, args...))
}

type cell struct{ plain, painted string }

func plain(s string) cell { return cell{s, s} }

func painted(s string, paint func(...any) string) cell { return cell{s, paint(s)} }

// Aligns by visible width; tabwriter miscounts ANSI colour codes.
func table(w io.Writer, headers []string, rows [][]cell) {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = utf8.RuneCountInString(h)
	}
	for _, row := range rows {
		for i, c := range row {
			widths[i] = max(widths[i], utf8.RuneCountInString(c.plain))
		}
	}

	line := func(cells []cell) {
		var b strings.Builder
		for i, c := range cells {
			b.WriteString(c.painted)
			if i < len(cells)-1 {
				b.WriteString(strings.Repeat(" ", widths[i]-utf8.RuneCountInString(c.plain)+2))
			}
		}
		fmt.Fprintln(w, b.String())
	}

	head := make([]cell, len(headers))
	for i, h := range headers {
		head[i] = painted(h, faint)
	}
	line(head)
	for _, row := range rows {
		line(row)
	}
}

func ago(t, now time.Time) string {
	d := now.Sub(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
	return t.Local().Format("2006-01-02")
}
