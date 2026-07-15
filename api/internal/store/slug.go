package store

import (
	"crypto/rand"
	"fmt"
)

// Ambiguous characters (0/O, 1/l/I) are left out: slugs get read aloud and typed by hand.
const slugAlphabet = "23456789abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"

const slugLength = 10

func newSlug() (string, error) {
	bytes := make([]byte, slugLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate slug: %w", err)
	}
	for i, b := range bytes {
		bytes[i] = slugAlphabet[int(b)%len(slugAlphabet)]
	}
	return string(bytes), nil
}
