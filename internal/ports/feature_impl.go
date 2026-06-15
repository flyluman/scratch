package ports

type featureFlags struct {
	softDelete    bool
	audit         bool
	auth          bool
	telemetry     bool
	valkeyCaching bool
}

func NewFeatureFlags(softDelete, audit, auth, telemetry, valkeyCaching bool) FeatureFlags {
	return &featureFlags{
		softDelete:    softDelete,
		audit:         audit,
		auth:          auth,
		telemetry:     telemetry,
		valkeyCaching: valkeyCaching,
	}
}

func (f *featureFlags) SoftDelete() bool    { return f.softDelete }
func (f *featureFlags) Audit() bool         { return f.audit }
func (f *featureFlags) Auth() bool          { return f.auth }
func (f *featureFlags) Telemetry() bool     { return f.telemetry }
func (f *featureFlags) ValkeyCaching() bool { return f.valkeyCaching }
