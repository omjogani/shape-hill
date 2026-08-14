package server

import (
	"testing"
	"time"

	"github.com/omjogani/shape-hill/internal/store"
)

func TestChartOfFlagsStalledOnlyWhenTracked(t *testing.T) {
	scopes := []store.Scope{{Title: "Refunds", Position: 15, MovedAt: time.Now().Add(-30 * 24 * time.Hour)}}

	if !chartOf(store.Hill{TrackStalled: true}, scopes, "").Dots[0].Stalled {
		t.Error("a scope untouched for a month must be flagged when tracking is on")
	}
	if chartOf(store.Hill{TrackStalled: false}, scopes, "").Dots[0].Stalled {
		t.Error("tracking off must keep the chart static, however old the scope is")
	}
}
