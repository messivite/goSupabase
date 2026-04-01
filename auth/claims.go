package auth

import "context"

type Claims struct {
	Subject   string `json:"sub"`
	Role      string `json:"role"`
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
