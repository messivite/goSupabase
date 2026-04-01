package auth

import "testing"

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
