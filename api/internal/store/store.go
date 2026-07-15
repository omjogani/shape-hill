package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() { s.pool.Close() }

type User struct {
	ID       string
	Email    string
	Username string
	Name     string
}

type Hill struct {
	ID          string
	OwnerID     string
	Slug        string
	Title       string
	Description string
	IsPublic    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Scope carries its latest position, so rendering a hill needs one query for the
// hill and one for its scopes.
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
