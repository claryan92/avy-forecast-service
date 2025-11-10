package domain_test

import (
	"testing"

	"example.com/avalanche/internal/domain"
)

func TestNewEmail_ValidatesAndNormalizes(t *testing.T) {
	cases := []struct {
		in  string
		ok  bool
		out string
	}{
		{"User@Example.com", true, "user@example.com"},
		{" user@example.com ", true, "user@example.com"},
		{"bad-email", false, ""},
		{"", false, ""},
	}
	for _, c := range cases {
		e, err := domain.NewEmail(c.in)
		if c.ok && err != nil {
			t.Fatalf("expected ok for %q: %v", c.in, err)
		}
		if !c.ok && err == nil {
			t.Fatalf("expected error for %q", c.in)
		}
		if c.ok && e.String() != c.out {
			t.Fatalf("expected %q, got %q", c.out, e.String())
		}
	}
}
