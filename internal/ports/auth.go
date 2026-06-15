package ports

import "context"

type Claims struct {
	Subject string
	Issuer  string
	Email   string
	Groups  []string
}

type TokenValidator interface {
	Validate(ctx context.Context, token string) (Claims, error)
}
