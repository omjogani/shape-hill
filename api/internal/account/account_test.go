package account

import "testing"

func TestCleanUsername(t *testing.T) {
	tests := []struct {
		raw   string
		want  string
		valid bool
	}{
		{"omjogani", "omjogani", true},
		{"  OmJogani  ", "omjogani", true},
		{"om_jogani-1", "om_jogani-1", true},
		{"abc", "abc", true},
		{"ab", "ab", false},
		{"", "", false},
		{"_leading", "_leading", false},
		{"has space", "has space", false},
		{"om.jogani", "om.jogani", false},
	}

	for _, tc := range tests {
		got, valid := CleanUsername(tc.raw)
		if got != tc.want || valid != tc.valid {
			t.Errorf("CleanUsername(%q) = (%q, %v), want (%q, %v)", tc.raw, got, valid, tc.want, tc.valid)
		}
	}
}
