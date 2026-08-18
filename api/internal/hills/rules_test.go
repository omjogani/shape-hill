package hills

import (
	"testing"
	"time"
)

func TestStalled(t *testing.T) {
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		track   bool
		moved   time.Time
		at      int16
		stalled bool
	}{
		{"a day old is fresh", true, now.Add(-24 * time.Hour), 40, false},
		{"just under a week is fresh", true, now.Add(-StalledAfter + time.Minute), 40, false},
		{"just over a week is stalled", true, now.Add(-StalledAfter - time.Minute), 40, true},
		{"a month is stalled", true, now.Add(-30 * 24 * time.Hour), 15, true},
		{"tracking off never stalls", false, now.Add(-30 * 24 * time.Hour), 15, false},
		{"work at the summit is done, not stuck", true, now.Add(-30 * 24 * time.Hour), 100, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			scope := Scope{Position: tc.at, MovedAt: tc.moved}
			if got := scope.Stalled(Hill{TrackStalled: tc.track}, now); got != tc.stalled {
				t.Errorf("Stalled() = %v, want %v", got, tc.stalled)
			}
		})
	}
}

func TestValidPosition(t *testing.T) {
	for _, position := range []int16{0, 1, 50, 99, 100} {
		if !ValidPosition(position) {
			t.Errorf("ValidPosition(%d) = false, want true", position)
		}
	}
	for _, position := range []int16{-1, 101, 32767} {
		if ValidPosition(position) {
			t.Errorf("ValidPosition(%d) = true, want false", position)
		}
	}
}

func TestValidSlug(t *testing.T) {
	for _, slug := range []string{"ship", "ship-it", "q3-2026", "a1"} {
		if !ValidSlug(slug) {
			t.Errorf("ValidSlug(%q) = false, want true", slug)
		}
	}
	for _, slug := range []string{"", "Ship", "ship it", "ship--it", "-ship", "ship-", "ship_it"} {
		if ValidSlug(slug) {
			t.Errorf("ValidSlug(%q) = true, want false", slug)
		}
	}
}

func TestColorOrDefault(t *testing.T) {
	if got := ColorOrDefault(""); got != DefaultScopeColor {
		t.Errorf("ColorOrDefault(\"\") = %q, want %q", got, DefaultScopeColor)
	}
	if got := ColorOrDefault("#ABCDEF"); got != "#ABCDEF" {
		t.Errorf("ColorOrDefault kept %q, want it untouched", got)
	}
}

func TestOwnedBy(t *testing.T) {
	hill := Hill{OwnerID: "owner"}
	if !hill.OwnedBy("owner") {
		t.Error("the owner must own their own hill")
	}
	if hill.OwnedBy("someone-else") {
		t.Error("a hill must not be owned by anyone else")
	}
	if hill.OwnedBy("") {
		t.Error("an empty caller must never own a hill")
	}
}

func TestLastMovedOnTracksTheNewestMovement(t *testing.T) {
	hill := Hill{UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	older := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	newest := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	got := LastMovedOn(hill, []Scope{{MovedAt: older}, {MovedAt: newest}, {MovedAt: older}})
	if !got.Equal(newest) {
		t.Errorf("LastMovedOn = %v, want the newest movement %v", got, newest)
	}

	if got := LastMovedOn(hill, nil); !got.Equal(hill.UpdatedAt) {
		t.Errorf("with no scopes it should fall back to the hill itself, got %v", got)
	}
}
