package auth

import (
	"context"
	"fmt"

	"github.com/flyluman/scratch/internal/platform/config"
	"github.com/flyluman/scratch/internal/ports"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/jwk"
)

type JWTValidator struct {
	issuer string
}

func NewJWTValidator(cfg config.Config) *JWTValidator {
	return &JWTValidator{issuer: cfg.AuthJWKSEndpoint}
}

func (v *JWTValidator) Validate(ctx context.Context, tokenString string) (ports.Claims, error) {
	keySet, err := jwk.Fetch(ctx, v.issuer)
	if err != nil {
		return ports.Claims{}, fmt.Errorf("fetch jwks: %w", err)
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		kid, ok := t.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid missing")
		}
		key, ok := keySet.LookupKeyID(kid)
		if !ok {
			return nil, fmt.Errorf("unknown kid: %s", kid)
		}
		var raw any
		return raw, key.Raw(&raw)
	})
	if err != nil {
		return ports.Claims{}, fmt.Errorf("token parse: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return ports.Claims{}, fmt.Errorf("invalid token")
	}

	return ports.Claims{
		Subject: claims["sub"].(string),
		Issuer:  claims["iss"].(string),
	}, nil
}
