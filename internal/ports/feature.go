package ports

import "context"

type FeatureFlags interface {
	SoftDelete() bool
	Audit() bool
	Auth() bool
	Telemetry() bool
	ValkeyCaching() bool
}

type FeatureLoader interface {
	Load(ctx context.Context) (FeatureFlags, error)
}
