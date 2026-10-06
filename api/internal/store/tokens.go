package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/omjogani/shape-hill/internal/account"
)

var _ account.TokenRepository = (*Store)(nil)

func (s *Store) CreateAPIToken(ctx context.Context, userID, name, hint string, hash []byte) (account.APIToken, error) {
	token := account.APIToken{Name: name, Hint: hint}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO api_tokens (user_id, name, hint, token_hash)
		VALUES ($1::uuid, $2, $3, $4)
		RETURNING id::text, created_at
	`, userID, name, hint, hash).Scan(&token.ID, &token.CreatedAt)
	if err != nil {
		return account.APIToken{}, fmt.Errorf("create api token: %w", err)
	}
	return token, nil
}

func (s *Store) ListAPITokens(ctx context.Context, userID string) ([]account.APIToken, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, name, hint, created_at, last_used_at
		FROM api_tokens
		WHERE user_id = $1::uuid
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list api tokens: %w", err)
	}
	defer rows.Close()

	tokens := []account.APIToken{}
	for rows.Next() {
		var token account.APIToken
		if err := rows.Scan(&token.ID, &token.Name, &token.Hint, &token.CreatedAt, &token.LastUsedAt); err != nil {
			return nil, fmt.Errorf("scan api token: %w", err)
		}
		tokens = append(tokens, token)
	}
	return tokens, rows.Err()
}

func (s *Store) DeleteAPIToken(ctx context.Context, tokenID, userID string) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM api_tokens WHERE id = $1::uuid AND user_id = $2::uuid
	`, tokenID, userID)
	if err != nil {
		return fmt.Errorf("delete api token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return account.ErrNotFound
	}
	return nil
}

func (s *Store) UserByAPIToken(ctx context.Context, hash []byte) (account.User, error) {
	var user account.User
	err := s.pool.QueryRow(ctx, `
		WITH used AS (
			UPDATE api_tokens SET last_used_at = now()
			WHERE token_hash = $1
			RETURNING user_id
		)
		SELECT u.id::text, u.email, u.username, coalesce(u.name, '')
		FROM users u JOIN used ON used.user_id = u.id
	`, hash).Scan(&user.ID, &user.Email, &user.Username, &user.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return account.User{}, account.ErrNotFound
	}
	if err != nil {
		return account.User{}, fmt.Errorf("user by api token: %w", err)
	}
	return user, nil
}
