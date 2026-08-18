// Package hills holds what a hill is and the rules that govern one.
package hills

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrSlugTaken = errors.New("slug already taken")
)

type Hill struct {
	ID           string
	OwnerID      string
	Slug         string
	Title        string
	Description  string
	IsPublic     bool
	TrackStalled bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Scope struct {
	ID          string
	Title       string
	Description string
	Color       string
	SortOrder   int16
	Position    int16
	Note        string
	MovedAt     time.Time
}

type Snapshot struct {
	Position  int16
	Note      string
	CreatedAt time.Time
}

// The ownerID arguments are here only because ownership is still enforced in SQL.
// They come out of these signatures when the rule moves into a service.
type Repository interface {
	CreateHill(ctx context.Context, ownerID, slug, title, description string, isPublic bool) (Hill, error)
	ListHillsByOwner(ctx context.Context, ownerID string) ([]Hill, error)
	HillBySlug(ctx context.Context, slug string) (Hill, error)
	UpdateHill(ctx context.Context, slug, ownerID string, title *string, isPublic, trackStalled *bool) (Hill, error)

	CreateScope(ctx context.Context, hillID, title, description, color string, sortOrder int16) (Scope, error)
	ScopesForHill(ctx context.Context, hillID string) ([]Scope, error)
	UpdateScope(ctx context.Context, scopeID, title, color, ownerID string) error
	ArchiveScope(ctx context.Context, scopeID, ownerID string) error
	MoveScope(ctx context.Context, scopeID string, position int16, note, ownerID string) error

	SnapshotsForScope(ctx context.Context, scopeID, ownerID string) ([]Snapshot, error)
	SnapshotsForPublicScope(ctx context.Context, scopeID string) ([]Snapshot, error)
}
