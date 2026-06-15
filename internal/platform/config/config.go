package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv   string
	HTTPAddr string
	LogLevel string

	PostgresDSN string

	ValkeyAddr     string
	ValkeyUsername string
	ValkeyPassword string
	ValkeyDB       int

	AuthEnabled             bool
	AuthJWKSEndpoint        string
	AuthJWKSRefreshInterval time.Duration

	RateLimitPerSecond int
	RateLimitTTL       time.Duration

	TelemetryEnabled      bool
	TelemetryOTLPEndpoint string

	TLSEnabled  bool
	TLSCertFile string
	TLSKeyFile  string

	CORSOrigins  []string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	RunMigrations bool

	FeatureCacheTTL time.Duration
}

func Load() Config {
	return Config{
		AppEnv:   getenv("SCRATCH_APP_ENV", "dev"),
		HTTPAddr: getenv("SCRATCH_HTTP_ADDR", ":8080"),
		LogLevel: getenv("SCRATCH_LOG_LEVEL", "info"),

		PostgresDSN: getenv("SCRATCH_POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/scratch?sslmode=disable"),

		ValkeyAddr:     getenv("SCRATCH_VALKEY_ADDR", "localhost:6379"),
		ValkeyUsername: getenv("SCRATCH_VALKEY_USERNAME", ""),
		ValkeyPassword: getenv("SCRATCH_VALKEY_PASSWORD", ""),
		ValkeyDB:       getenvInt("SCRATCH_VALKEY_DB", 0),

		AuthEnabled:             getenvBool("SCRATCH_AUTH_ENABLED", false),
		AuthJWKSEndpoint:        getenv("SCRATCH_AUTH_JWKS_ENDPOINT", ""),
		AuthJWKSRefreshInterval: time.Duration(getenvInt("SCRATCH_AUTH_JWKS_REFRESH_MIN", 10)) * time.Minute,

		RateLimitPerSecond: getenvInt("SCRATCH_RATE_LIMIT_PER_SECOND", 0),
		RateLimitTTL:       time.Duration(getenvInt("SCRATCH_RATE_LIMIT_TTL_SEC", 300)) * time.Second,

		TelemetryEnabled:      getenvBool("SCRATCH_TELEMETRY_ENABLED", false),
		TelemetryOTLPEndpoint: getenv("SCRATCH_TELEMETRY_OTLP_ENDPOINT", ""),

		TLSEnabled:  getenvBool("SCRATCH_TLS_ENABLED", false),
		TLSCertFile: getenv("SCRATCH_TLS_CERT_FILE", ""),
		TLSKeyFile:  getenv("SCRATCH_TLS_KEY_FILE", ""),

		CORSOrigins:  strings.Split(getenv("SCRATCH_CORS_ORIGINS", "*"), ","),
		ReadTimeout:  time.Duration(getenvInt("SCRATCH_READ_TIMEOUT_SEC", 10)) * time.Second,
		WriteTimeout: time.Duration(getenvInt("SCRATCH_WRITE_TIMEOUT_SEC", 30)) * time.Second,

		RunMigrations:    getenvBool("SCRATCH_RUN_MIGRATIONS", true),
		FeatureCacheTTL:  time.Duration(getenvInt("SCRATCH_FEATURE_CACHE_TTL_SEC", 30)) * time.Second,
	}
}

func (c Config) Validate() error {
	if c.HTTPAddr == "" {
		return fmt.Errorf("SCRATCH_HTTP_ADDR is required")
	}
	if c.PostgresDSN == "" {
		return fmt.Errorf("SCRATCH_POSTGRES_DSN is required")
	}
	if c.TLSEnabled && (c.TLSCertFile == "" || c.TLSKeyFile == "") {
		return fmt.Errorf("SCRATCH_TLS_CERT_FILE and SCRATCH_TLS_KEY_FILE required")
	}
	return nil
}

func getenv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func getenvInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	if i, err := strconv.Atoi(val); err == nil {
		return i
	}
	return fallback
}

func getenvBool(key string, fallback bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val == "true" || val == "1"
}
