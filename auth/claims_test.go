package auth

import (
	"context"
	"testing"
)

func TestEffectiveRolesNilClaims(t *testing.T) {
	var c *Claims
	if got := c.EffectiveRoles(); got != nil {
		t.Fatalf("expected nil roles, got %v", got)
	}
}

func TestEffectiveRolesMergesAndDeduplicates(t *testing.T) {
	c := &Claims{
		Role:  "authenticated",
		Roles: []string{"admin", "authenticated", "", "editor"},
	}
	got := c.EffectiveRoles()
	want := []string{"authenticated", "admin", "editor"}
	if len(got) != len(want) {
		t.Fatalf("roles length = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("roles[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestWithClaimsAndGetClaims(t *testing.T) {
	base := context.Background()
	if got := GetClaims(base); got != nil {
		t.Fatalf("expected nil claims on empty context, got %+v", got)
	}

	c := &Claims{Subject: "u1", Role: "authenticated"}
	ctx := WithClaims(base, c)
	got := GetClaims(ctx)
	if got == nil {
		t.Fatal("expected claims in context")
	}
	if got.Subject != "u1" || got.Role != "authenticated" {
		t.Fatalf("unexpected claims: %+v", got)
	}
}
