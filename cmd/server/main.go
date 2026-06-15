// @title           Scratch API
// @version         1.0
// @description     Production-grade Go hexagonal CRUD API for Items
// @termsOfService  http://swagger.io/terms/
// @contact.name   API Support
// @contact.email  support@scratch.dev
// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT
// @host      localhost:8080
// @BasePath  /v1
// @schemes   http https
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 JWT Bearer token authentication
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	auditadapter "github.com/flyluman/scratch/internal/adapters/outbound/audit"
	authadapter "github.com/flyluman/scratch/internal/adapters/outbound/auth"
	featurevalkey "github.com/flyluman/scratch/internal/adapters/outbound/feature"
	valkeyadapter "github.com/flyluman/scratch/internal/adapters/outbound/valkey"
	httpapi "github.com/flyluman/scratch/internal/adapters/inbound/http"
	"github.com/flyluman/scratch/internal/platform/config"
	featureenv "github.com/flyluman/scratch/internal/platform/featureenv"
	"github.com/flyluman/scratch/internal/platform/logger"
	"github.com/flyluman/scratch/internal/platform/migrations"
	"github.com/flyluman/scratch/internal/platform/telemetry"
	"github.com/flyluman/scratch/internal/ports"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	_ "github.com/flyluman/scratch/docs"
)

type App struct {
	handler   *httpapi.Handler
	cfg       config.Config
	logger    ports.Logger
	telemetry *telemetry.Telemetry
}

func (a *App) Start(ctx context.Context) error {
	e := echo.New()
	e.HideBanner = true
	a.handler.RegisterRoutes(e)

	addr := a.cfg.HTTPAddr
	if addr == "" {
		addr = ":8080"
	}

	a.logger.Info("server starting", ports.F("addr", addr))

	go func() {
		var err error
		if a.cfg.TLSEnabled {
			err = e.StartTLS(addr, a.cfg.TLSCertFile, a.cfg.TLSKeyFile)
		} else {
			err = e.Start(addr)
		}
		if err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("server failed", ports.F("error", err))
		}
	}()

	<-ctx.Done()
	a.logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if a.telemetry != nil {
		_ = a.telemetry.Shutdown(shutdownCtx)
	}

	return e.Shutdown(shutdownCtx)
}

func main() {
	mode := flag.String("mode", "server", "run mode: server or migrate")
	flag.Parse()

	switch *mode {
	case "server":
		app, cleanup, err := InitializeApp()
		if err != nil {
			log.Fatal(err)
		}
		defer cleanup()

		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		if err := app.Start(ctx); err != nil {
			log.Fatal(err)
		}

	case "migrate":
		if err := RunMigrations(); err != nil {
			log.Fatal(err)
		}

	default:
		log.Fatalf("unknown mode: %s", *mode)
	}
}

func RunMigrations() error {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	pool, err := migrations.NewPool(context.Background(), cfg.PostgresDSN)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	runner := migrations.NewRunner(pool, "migrations")
	return runner.Run(context.Background())
}

// ---------------------------------------------------------------------------
// Builder functions
// ---------------------------------------------------------------------------

func buildTelemetry(ctx context.Context, cfg config.Config) *telemetry.Telemetry {
	tel, err := telemetry.New(ctx, cfg)
	if err != nil {
		return nil
	}
	return tel
}

func buildFeatureFlags(ctx context.Context, cfg config.Config, l ports.Logger) ports.FeatureFlags {
	var loader ports.FeatureLoader

	if cfg.ValkeyAddr != "" {
		vkClient, err := valkeyadapter.NewClient(cfg)
		if err != nil {
			l.Warn("valkey connect failed, using env feature flags", ports.F("error", err))
		} else {
			fv := featurevalkey.NewLoader(vkClient.Client, cfg.AppEnv, cfg.FeatureCacheTTL)
			if err := fv.Seed(ctx); err != nil {
				l.Warn("feature flag seed failed", ports.F("error", err))
			}
			loader = fv
		}
	}

	if loader == nil {
		loader = featureenv.NewLoader()
	}

	flags, err := loader.Load(ctx)
	if err != nil {
		l.Warn("feature flags load failed, using env defaults", ports.F("error", err))
		return featureenv.LoadFlagsFromEnv()
	}
	return flags
}

func buildPostgres(ctx context.Context, cfg config.Config, l ports.Logger) *pgxpool.Pool {
	if cfg.PostgresDSN == "" {
		return nil
	}

	pool, err := migrations.NewPool(ctx, cfg.PostgresDSN)
	if err != nil {
		l.Warn("postgres connect failed, using in-memory repo", ports.F("error", err))
		return nil
	}

	runner := migrations.NewRunner(pool, "migrations")
	if err := runner.Run(ctx); err != nil {
		pool.Close()
		l.Fatal("run migrations failed", ports.F("error", err))
	}

	return pool
}

func buildAuditLogger(pool *pgxpool.Pool) ports.AuditLogger {
	if pool != nil {
		return auditadapter.NewPostgresAuditLogger(pool)
	}
	return auditadapter.NewNoOpAuditLogger()
}

func buildTokenValidator(cfg config.Config) ports.TokenValidator {
	if cfg.AuthEnabled {
		return authadapter.NewJWTValidator(cfg)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Initialization
// ---------------------------------------------------------------------------

func InitializeApp() (*App, func(), error) {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		return nil, nil, fmt.Errorf("invalid config: %w", err)
	}

	log := logger.New("scratch", cfg.AppEnv, cfg.LogLevel)
	tel := buildTelemetry(context.Background(), cfg)
	flags := buildFeatureFlags(context.Background(), cfg, log)
	pool := buildPostgres(context.Background(), cfg, log)

	auditLogger := buildAuditLogger(pool)
	tokenValidator := buildTokenValidator(cfg)

	return NewAppBuilder(cfg, log, tel, pool, flags, auditLogger, tokenValidator).
		WithItem().
		Build()
}
