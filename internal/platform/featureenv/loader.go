package featureenv

import (
	"context"
	"os"

	"github.com/flyluman/scratch/internal/ports"
)

type EnvLoader struct{}

func NewLoader() *EnvLoader {
	return &EnvLoader{}
}

func (l *EnvLoader) Load(ctx context.Context) (ports.FeatureFlags, error) {
	return LoadFlagsFromEnv(), nil
}

func LoadFlagsFromEnv() ports.FeatureFlags {
	return ports.NewFeatureFlags(
		os.Getenv("SCRATCH_FEATURE_SOFT_DELETE") != "false",
		os.Getenv("SCRATCH_FEATURE_AUDIT") == "true",
		os.Getenv("SCRATCH_FEATURE_AUTH") == "true",
		os.Getenv("SCRATCH_FEATURE_TELEMETRY") == "true",
		os.Getenv("SCRATCH_FEATURE_VALKEY_CACHING") != "false",
	)
}
