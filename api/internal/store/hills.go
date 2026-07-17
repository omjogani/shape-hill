package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateUser(ctx context.Context, email, username, name string) (User, error) {
	user := User{Email: email, Username: username, Name: name}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO users (email, username, name)
		VALUES ($1, $2, $3)
		RETURNING id::text
	`, email, username, name).Scan(&user.ID)
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

// DeleteUser cascades: the user's hills, their scopes, and every position go with it.
func (s *Store) DeleteUser(ctx context.Context, id string) error {
	if _, err := s.pool.Exec(ctx, `DELETE FROM users WHERE id = $1::uuid`, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

func (s *Store) CreateHill(ctx context.Context, ownerID, title, description string, isPublic bool) (Hill, error) {
	slug, err := newSlug()
	if err != nil {
		return Hill{}, err
	}

	hill := Hill{OwnerID: ownerID, Slug: slug, Title: title, Description: description, IsPublic: isPublic}
	err = s.pool.QueryRow(ctx, `
		INSERT INTO hills (owner_id, slug, title, description, is_public)
		VALUES ($1::uuid, $2, $3, $4, $5)
		RETURNING id::text, created_at, updated_at
	`, ownerID, slug, title, description, isPublic).Scan(&hill.ID, &hill.CreatedAt, &hill.UpdatedAt)
	if err != nil {
		return Hill{}, fmt.Errorf("create hill: %w", err)
	}
	return hill, nil
}

func (s *Store) HillBySlug(ctx context.Context, slug string) (Hill, error) {
	var hill Hill
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, owner_id::text, slug, title, coalesce(description, ''), is_public, created_at, updated_at
		FROM hills
		WHERE slug = $1
	`, slug).Scan(&hill.ID, &hill.OwnerID, &hill.Slug, &hill.Title, &hill.Description,
		&hill.IsPublic, &hill.CreatedAt, &hill.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Hill{}, ErrNotFound
	}
	if err != nil {
		return Hill{}, fmt.Errorf("get hill: %w", err)
	}
	return hill, nil
}

func (s *Store) UpdateHillTitle(ctx context.Context, slug, title string) (Hill, error) {
	var hill Hill
	err := s.pool.QueryRow(ctx, `
		UPDATE hills SET title = $2, updated_at = now()
		WHERE slug = $1
		RETURNING id::text, owner_id::text, slug, title, coalesce(description, ''), is_public, created_at, updated_at
	`, slug, title).Scan(&hill.ID, &hill.OwnerID, &hill.Slug, &hill.Title, &hill.Description,
		&hill.IsPublic, &hill.CreatedAt, &hill.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Hill{}, ErrNotFound
	}
	if err != nil {
		return Hill{}, fmt.Errorf("update hill title: %w", err)
	}
	return hill, nil
}

func (s *Store) CreateScope(ctx context.Context, hillID, title, description, color string, sortOrder int16) (Scope, error) {
	scope := Scope{Title: title, Description: description, Color: color, SortOrder: sortOrder}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO scopes (hill_id, title, description, color, sort_order)
		VALUES ($1::uuid, $2, $3, $4, $5)
		RETURNING id::text, created_at
	`, hillID, title, description, color, sortOrder).Scan(&scope.ID, &scope.MovedAt)
	if err != nil {
		return Scope{}, fmt.Errorf("create scope: %w", err)
	}
	return scope, nil
}

// ScopesForHill returns each live scope with its most recent position. A scope
// that has never been moved sits at 0.
func (s *Store) ScopesForHill(ctx context.Context, hillID string) ([]Scope, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT s.id::text, s.title, coalesce(s.description, ''), s.color, s.sort_order,
		       coalesce(latest.position, 0), coalesce(latest.note, ''),
		       coalesce(latest.created_at, s.created_at)
		FROM scopes s
		LEFT JOIN LATERAL (
			SELECT position, note, created_at
			FROM scope_positions
			WHERE scope_id = s.id
			ORDER BY created_at DESC
			LIMIT 1
		) latest ON true
		WHERE s.hill_id = $1::uuid AND s.archived_at IS NULL
		ORDER BY s.sort_order, s.created_at
	`, hillID)
	if err != nil {
		return nil, fmt.Errorf("list scopes: %w", err)
	}
	defer rows.Close()

	var scopes []Scope
	for rows.Next() {
		var scope Scope
		if err := rows.Scan(&scope.ID, &scope.Title, &scope.Description, &scope.Color,
			&scope.SortOrder, &scope.Position, &scope.Note, &scope.MovedAt); err != nil {
			return nil, fmt.Errorf("scan scope: %w", err)
		}
		scopes = append(scopes, scope)
	}
	return scopes, rows.Err()
}

// MoveScope appends a position. It never updates one: the rows left behind are
// the dot's trail, and the newest row is where it sits now.
func (s *Store) MoveScope(ctx context.Context, scopeID string, position int16, note, movedBy string) error {
	var mover any
	if movedBy != "" {
		mover = movedBy
	}

	tag, err := s.pool.Exec(ctx, `
		INSERT INTO scope_positions (scope_id, position, note, moved_by)
		SELECT $1::uuid, $2, $3, $4::uuid
		WHERE EXISTS (SELECT 1 FROM scopes WHERE id = $1::uuid)
	`, scopeID, position, note, mover)
	if err != nil {
		return fmt.Errorf("move scope: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// LastMovedOn is the newest movement anywhere on the hill — the ETag for its image.
func LastMovedOn(hill Hill, scopes []Scope) time.Time {
	last := hill.UpdatedAt
	for _, scope := range scopes {
		if scope.MovedAt.After(last) {
			last = scope.MovedAt
		}
	}
	return last
}
