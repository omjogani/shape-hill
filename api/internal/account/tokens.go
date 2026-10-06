package account

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

const TokenPrefix = "shk_"

const maxTokenName = 64

type APIToken struct {
	ID         string
	Name       string
	Hint       string
	CreatedAt  time.Time
	LastUsedAt *time.Time
}

type TokenRepository interface {
	CreateAPIToken(ctx context.Context, userID, name, hint string, hash []byte) (APIToken, error)
	ListAPITokens(ctx context.Context, userID string) ([]APIToken, error)
	DeleteAPIToken(ctx context.Context, tokenID, userID string) error
	UserByAPIToken(ctx context.Context, hash []byte) (User, error)
}

// 32 random bytes: high enough entropy that a plain SHA-256 is safe to store.
func NewAPIToken() (raw, hint string, hash []byte) {
	secret := make([]byte, 32)
	rand.Read(secret)
	raw = TokenPrefix + hex.EncodeToString(secret)
	return raw, raw[:len(TokenPrefix)+8], HashAPIToken(raw)
}

func HashAPIToken(raw string) []byte {
	sum := sha256.Sum256([]byte(raw))
	return sum[:]
}

func IsAPIToken(raw string) bool { return strings.HasPrefix(raw, TokenPrefix) }

func CleanTokenName(raw string) (cleaned string, valid bool) {
	name := strings.TrimSpace(raw)
	return name, name != "" && len(name) <= maxTokenName
}
