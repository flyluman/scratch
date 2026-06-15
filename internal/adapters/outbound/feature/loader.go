package feature

import (
	"context"
	"fmt"
	"time"

	"github.com/flyluman/scratch/internal/ports"
	"github.com/valkey-io/valkey-go"
)

type Loader struct {
	client   valkey.Client
	env      string
	cacheTTL time.Duration
}

func NewLoader(client valkey.Client, env string, cacheTTL time.Duration) *Loader {
	return &Loader{
		client:   client,
		env:      env,
		cacheTTL: cacheTTL,
	}
}

func (l *Loader) key() string {
	return "feature:flags:" + l.env
}

func (l *Loader) Load(ctx context.Context) (ports.FeatureFlags, error) {
	key := l.key()
	resp := l.client.DoCache(ctx, l.client.B().Hgetall().Key(key).Cache(), l.cacheTTL)
	m, err := resp.AsStrMap()
	if err != nil {
		return nil, fmt.Errorf("hgetall %s: %w", key, err)
	}

	if len(m) == 0 {
		return nil, nil
	}

	return ports.NewFeatureFlags(
		m["soft_delete"] != "false",
		m["audit"] == "true",
		m["auth"] == "true",
		m["telemetry"] == "true",
		m["valkey_caching"] != "false",
	), nil
}

func (l *Loader) Seed(ctx context.Context) error {
	key := l.key()

	exists, err := l.client.Do(ctx, l.client.B().Exists().Key(key).Build()).AsInt64()
	if err != nil {
		return fmt.Errorf("check exists: %w", err)
	}
	if exists > 0 {
		return nil
	}

	defaults := map[string]string{
		"audit":          "false",
		"soft_delete":    "true",
		"auth":           "false",
		"telemetry":      "false",
		"valkey_caching": "true",
	}
	for field, val := range defaults {
		if err := l.client.Do(ctx, l.client.B().Hsetnx().Key(key).Field(field).Value(val).Build()).Error(); err != nil {
			return fmt.Errorf("seed %s=%s: %w", field, val, err)
		}
	}
	return nil
}
