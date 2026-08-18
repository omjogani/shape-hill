package server

import (
	"testing"
	"time"

	"github.com/omjogani/shape-hill/internal/hills"
)

func TestChartOfFlagsStalledOnlyWhenTracked(t *testing.T) {
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	scopes := []hills.Scope{{Title: "Refunds", Position: 15, MovedAt: now.Add(-30 * 24 * time.Hour)}}

	if !chartOf(hills.Hill{TrackStalled: true}, scopes, "", now).Dots[0].Stalled {
		t.Error("a scope untouched for a month must be flagged when tracking is on")
	}
	if chartOf(hills.Hill{TrackStalled: false}, scopes, "", now).Dots[0].Stalled {
		t.Error("tracking off must keep the chart static, however old the scope is")
	}
}
