package account

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewAPIToken(t *testing.T) {
	raw, hint, hash := NewAPIToken()

	if !IsAPIToken(raw) {
		t.Fatalf("minted token %q must carry the %q prefix", raw, TokenPrefix)
	}
	if !strings.HasPrefix(raw, hint) || len(hint) >= len(raw)/2 {
		t.Errorf("hint %q should be a short leading slice of the token", hint)
	}
	if !bytes.Equal(hash, HashAPIToken(raw)) {
		t.Error("the stored hash must be what HashAPIToken gives for the raw token")
	}

	again, _, _ := NewAPIToken()
	if again == raw {
		t.Fatal("two minted tokens must differ")
	}
}

func TestIsAPITokenIgnoresJWTs(t *testing.T) {
	if IsAPIToken("eyJhbGciOiJFUzI1NiJ9.e30.sig") {
		t.Fatal("a JWT must go through Supabase verification, not the token table")
	}
}

func TestCleanTokenName(t *testing.T) {
	tests := []struct {
		raw   string
		want  string
		valid bool
	}{
		{"laptop", "laptop", true},
		{"  claude code  ", "claude code", true},
		{"", "", false},
		{"   ", "", false},
		{strings.Repeat("a", 64), strings.Repeat("a", 64), true},
		{strings.Repeat("a", 65), strings.Repeat("a", 65), false},
	}

	for _, tc := range tests {
		got, valid := CleanTokenName(tc.raw)
		if got != tc.want || valid != tc.valid {
			t.Errorf("CleanTokenName(%q) = (%q, %v), want (%q, %v)", tc.raw, got, valid, tc.want, tc.valid)
		}
	}
}
