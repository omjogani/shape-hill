package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/omjogani/shape-hill/internal/account"
	"github.com/omjogani/shape-hill/internal/hills"
)

func (s *Store) CreateUser(ctx context.Context, email, username, name string) (account.User, error) {
	user := account.User{Email: email, Username: username, Name: name}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO users (email, username, name)
		VALUES ($1, $2, $3)
		RETURNING id::text
	`, email, username, name).Scan(&user.ID)
	if err != nil {
		return account.User{}, fmt.Errorf("create user: %w", err)
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

func (s *Store) UserByAuthID(ctx context.Context, authUserID string) (account.User, error) {
	var user account.User
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, email, username, coalesce(name, '')
		FROM users WHERE auth_user_id = $1::uuid
	`, authUserID).Scan(&user.ID, &user.Email, &user.Username, &user.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return account.User{}, account.ErrNotFound
	}
	if err != nil {
		return account.User{}, fmt.Errorf("user by auth id: %w", err)
	}
	return user, nil
}

// OnboardUser ties the caller's Supabase identity to a local account: it adopts
// an existing row with the same email (link-by-email), otherwise creates one.
// The email must come from the verified token — it is the linking key.
func (s *Store) OnboardUser(ctx context.Context, authUserID, email, username, name string) (account.User, error) {
	var user account.User
	err := s.pool.QueryRow(ctx, `
		WITH linked AS (
			UPDATE users SET auth_user_id = $1::uuid
			WHERE lower(email) = lower($2) AND auth_user_id IS NULL
			RETURNING id, email, username, name
		),
		created AS (
			INSERT INTO users (auth_user_id, email, username, name)
			SELECT $1::uuid, $2, $3, NULLIF($4, '')
			WHERE NOT EXISTS (SELECT 1 FROM linked)
			RETURNING id, email, username, name
		)
		SELECT id::text, email, username, coalesce(name, '') FROM linked
		UNION ALL
		SELECT id::text, email, username, coalesce(name, '') FROM created
	`, authUserID, email, username, name).Scan(&user.ID, &user.Email, &user.Username, &user.Name)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "users_username_key":
			return account.User{}, account.ErrUsernameTaken
		case "users_email_key":
			return account.User{}, account.ErrEmailTaken
		}
	}
	if err != nil {
		return account.User{}, fmt.Errorf("onboard user: %w", err)
	}
	return user, nil
}

func (s *Store) CreateHill(ctx context.Context, ownerID, slug, title, description string, isPublic bool) (hills.Hill, error) {
	hill := hills.Hill{OwnerID: ownerID, Slug: slug, Title: title, Description: description, IsPublic: isPublic}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO hills (owner_id, slug, title, description, is_public)
		VALUES ($1::uuid, $2, $3, $4, $5)
		RETURNING id::text, track_stalled, created_at, updated_at
	`, ownerID, slug, title, description, isPublic).Scan(&hill.ID, &hill.TrackStalled, &hill.CreatedAt, &hill.UpdatedAt)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return hills.Hill{}, hills.ErrSlugTaken
	}
	if err != nil {
		return hills.Hill{}, fmt.Errorf("create hill: %w", err)
	}
	return hill, nil
}

func (s *Store) ListHillsByOwner(ctx context.Context, ownerID string) ([]hills.Hill, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, owner_id::text, slug, title, coalesce(description, ''), is_public, track_stalled, created_at, updated_at
		FROM hills
		WHERE owner_id = $1::uuid
		ORDER BY updated_at DESC
	`, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list hills: %w", err)
	}
	defer rows.Close()

	// Non-nil so an empty table marshals to [] rather than null.
	owned := []hills.Hill{}
	for rows.Next() {
		var hill hills.Hill
		if err := rows.Scan(&hill.ID, &hill.OwnerID, &hill.Slug, &hill.Title, &hill.Description,
			&hill.IsPublic, &hill.TrackStalled, &hill.CreatedAt, &hill.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan hill: %w", err)
		}
		owned = append(owned, hill)
	}
	return owned, rows.Err()
}

func (s *Store) HillBySlug(ctx context.Context, slug string) (hills.Hill, error) {
	var hill hills.Hill
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, owner_id::text, slug, title, coalesce(description, ''), is_public, track_stalled, created_at, updated_at
		FROM hills
		WHERE slug = $1
	`, slug).Scan(&hill.ID, &hill.OwnerID, &hill.Slug, &hill.Title, &hill.Description,
		&hill.IsPublic, &hill.TrackStalled, &hill.CreatedAt, &hill.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return hills.Hill{}, hills.ErrNotFound
	}
	if err != nil {
		return hills.Hill{}, fmt.Errorf("get hill: %w", err)
	}
	return hill, nil
}

func (s *Store) UpdateHill(ctx context.Context, slug, ownerID string, title *string, isPublic, trackStalled *bool) (hills.Hill, error) {
	var hill hills.Hill
	err := s.pool.QueryRow(ctx, `
		UPDATE hills
		SET title = coalesce($3, title),
		    is_public = coalesce($4, is_public),
		    track_stalled = coalesce($5, track_stalled),
		    updated_at = now()
		WHERE slug = $1 AND owner_id = $2::uuid
		RETURNING id::text, owner_id::text, slug, title, coalesce(description, ''), is_public, track_stalled, created_at, updated_at
	`, slug, ownerID, title, isPublic, trackStalled).Scan(&hill.ID, &hill.OwnerID, &hill.Slug, &hill.Title, &hill.Description,
		&hill.IsPublic, &hill.TrackStalled, &hill.CreatedAt, &hill.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return hills.Hill{}, hills.ErrNotFound
	}
	if err != nil {
		return hills.Hill{}, fmt.Errorf("update hill: %w", err)
	}
	return hill, nil
}

func (s *Store) CreateScope(ctx context.Context, hillID, title, description, color string, sortOrder int16) (hills.Scope, error) {
	scope := hills.Scope{Title: title, Description: description, Color: color, SortOrder: sortOrder}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO scopes (hill_id, title, description, color, sort_order)
		VALUES ($1::uuid, $2, $3, $4, $5)
		RETURNING id::text, created_at
	`, hillID, title, description, color, sortOrder).Scan(&scope.ID, &scope.MovedAt)
	if err != nil {
		return hills.Scope{}, fmt.Errorf("create scope: %w", err)
	}
	return scope, nil
}

func (s *Store) SnapshotsForScope(ctx context.Context, scopeID, ownerID string) ([]hills.Snapshot, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT sp.position, coalesce(sp.note, ''), sp.created_at
		FROM scope_positions sp
		JOIN scopes s ON s.id = sp.scope_id
		JOIN hills h ON h.id = s.hill_id
		WHERE sp.scope_id = $1::uuid AND h.owner_id = $2::uuid
		ORDER BY sp.created_at DESC
	`, scopeID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list snapshots: %w", err)
	}
	defer rows.Close()

	// Non-nil so an untouched scope marshals to [] rather than null.
	snapshots := []hills.Snapshot{}
	for rows.Next() {
		var snap hills.Snapshot
		if err := rows.Scan(&snap.Position, &snap.Note, &snap.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan snapshot: %w", err)
		}
		snapshots = append(snapshots, snap)
	}
	return snapshots, rows.Err()
}

func (s *Store) SnapshotsForPublicScope(ctx context.Context, scopeID string) ([]hills.Snapshot, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT sp.position, coalesce(sp.note, ''), sp.created_at
		FROM scope_positions sp
		JOIN scopes s ON s.id = sp.scope_id
		JOIN hills h ON h.id = s.hill_id
		WHERE sp.scope_id = $1::uuid AND h.is_public = true
		ORDER BY sp.created_at DESC
	`, scopeID)
	if err != nil {
		return nil, fmt.Errorf("list public snapshots: %w", err)
	}
	defer rows.Close()

	snapshots := []hills.Snapshot{}
	for rows.Next() {
		var snap hills.Snapshot
		if err := rows.Scan(&snap.Position, &snap.Note, &snap.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan public snapshot: %w", err)
		}
		snapshots = append(snapshots, snap)
	}
	return snapshots, rows.Err()
}

func (s *Store) UpdateScope(ctx context.Context, scopeID, title, color, ownerID string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE scopes SET title = $2, color = $3
		WHERE id = $1::uuid AND archived_at IS NULL
		  AND hill_id IN (SELECT id FROM hills WHERE owner_id = $4::uuid)
	`, scopeID, title, color, ownerID)
	if err != nil {
		return fmt.Errorf("update scope: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return hills.ErrNotFound
	}
	return nil
}

func (s *Store) ArchiveScope(ctx context.Context, scopeID, ownerID string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE scopes SET archived_at = now()
		WHERE id = $1::uuid AND archived_at IS NULL
		  AND hill_id IN (SELECT id FROM hills WHERE owner_id = $2::uuid)
	`, scopeID, ownerID)
	if err != nil {
		return fmt.Errorf("archive scope: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return hills.ErrNotFound
	}
	return nil
}

// ScopesForHill returns each live scope with its most recent position. A scope
// that has never been moved sits at 0.
func (s *Store) ScopesForHill(ctx context.Context, hillID string) ([]hills.Scope, error) {
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

	scopes := []hills.Scope{}
	for rows.Next() {
		var scope hills.Scope
		if err := rows.Scan(&scope.ID, &scope.Title, &scope.Description, &scope.Color,
			&scope.SortOrder, &scope.Position, &scope.Note, &scope.MovedAt); err != nil {
			return nil, fmt.Errorf("scan scope: %w", err)
		}
		scopes = append(scopes, scope)
	}
	return scopes, rows.Err()
}

// MoveScope appends a position for a scope on a hill the owner owns. It never
// updates a row: the rows left behind are the dot's trail, and the newest row is
// where it sits now. moved_by is the owner.
func (s *Store) MoveScope(ctx context.Context, scopeID string, position int16, note, ownerID string) error {
	tag, err := s.pool.Exec(ctx, `
		INSERT INTO scope_positions (scope_id, position, note, moved_by)
		SELECT s.id, $2, $3, $4::uuid
		FROM scopes s JOIN hills h ON h.id = s.hill_id
		WHERE s.id = $1::uuid AND h.owner_id = $4::uuid
	`, scopeID, position, note, ownerID)
	if err != nil {
		return fmt.Errorf("move scope: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return hills.ErrNotFound
	}
	return nil
}
