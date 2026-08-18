// Package account holds who someone is and the rules for a handle.
package account

import (
	"context"
	"errors"
	"regexp"
	"strings"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrUsernameTaken = errors.New("username already taken")
	ErrEmailTaken    = errors.New("email already registered")
)

type User struct {
	ID       string
	Email    string
	Username string
	Name     string
}

type AuthUser struct {
	ID    string
	Email string
}

type VerifyToken func(ctx context.Context, token string) (AuthUser, error)

type Repository interface {
	UserByAuthID(ctx context.Context, authUserID string) (User, error)
	OnboardUser(ctx context.Context, authUserID, email, username, name string) (User, error)
}

var usernamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{2,29}$`)

func CleanUsername(raw string) (cleaned string, valid bool) {
	username := strings.ToLower(strings.TrimSpace(raw))
	return username, usernamePattern.MatchString(username)
}
