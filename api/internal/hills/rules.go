package hills

import (
	"regexp"
	"time"
)

const (
	StalledAfter      = 7 * 24 * time.Hour
	DefaultScopeColor = "#2F4C64"
	Summit            = 100
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func ValidSlug(slug string) bool { return slugPattern.MatchString(slug) }

func ValidPosition(position int16) bool { return position >= 0 && position <= Summit }

func ColorOrDefault(color string) string {
	if color == "" {
		return DefaultScopeColor
	}
	return color
}

func (h Hill) OwnedBy(userID string) bool { return h.OwnerID == userID }

func (s Scope) AtSummit() bool { return s.Position >= Summit }

func (s Scope) Stalled(h Hill, now time.Time) bool {
	return h.TrackStalled && !s.AtSummit() && now.Sub(s.MovedAt) > StalledAfter
}

func LastMovedOn(hill Hill, scopes []Scope) time.Time {
	last := hill.UpdatedAt
	for _, scope := range scopes {
		if scope.MovedAt.After(last) {
			last = scope.MovedAt
		}
	}
	return last
}
