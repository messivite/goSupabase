package auth

import "context"

type Claims struct {
	Subject   string `json:"sub"`
	Role      string `json:"role"`
	Roles     []string `json:"roles,omitempty"`
	Email     string `json:"email"`
	Audience  string `json:"aud"`
	ExpiresAt int64  `json:"exp"`
}

type ctxKey struct{}

func WithClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, ctxKey{}, c)
}

func GetClaims(ctx context.Context) *Claims {
	c, _ := ctx.Value(ctxKey{}).(*Claims)
	return c
}

// EffectiveRoles returns merged, de-duplicated roles from both `role` and `roles`.
func (c *Claims) EffectiveRoles() []string {
	if c == nil {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(c.Roles)+1)
	if c.Role != "" {
		seen[c.Role] = true
		out = append(out, c.Role)
	}
	for _, r := range c.Roles {
		if r == "" || seen[r] {
			continue
		}
		seen[r] = true
		out = append(out, r)
	}
	return out
}
