package main

import (
	"context"

	"github.com/flyluman/scratch/internal/adapters/outbound/audit"
	httpapi "github.com/flyluman/scratch/internal/adapters/inbound/http"
	item "github.com/flyluman/scratch/internal/adapters/inbound/http/item"
	"github.com/flyluman/scratch/internal/application"
	"github.com/flyluman/scratch/internal/platform/bus"
	"github.com/flyluman/scratch/internal/platform/config"
	"github.com/flyluman/scratch/internal/platform/telemetry"
	"github.com/flyluman/scratch/internal/ports"
	clock "github.com/flyluman/scratch/internal/provider/clock"
	idadapter "github.com/flyluman/scratch/internal/provider/id"
	repo "github.com/flyluman/scratch/internal/provider/repo"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AppBuilder struct {
	cfg            config.Config
	logger         ports.Logger
	tel            *telemetry.Telemetry
	pool           *pgxpool.Pool
	repos          *repo.Factory
	flags          ports.FeatureFlags
	bus            ports.EventBus
	tokenValidator ports.TokenValidator
	idGen          ports.IDGenerator
	clk            ports.Clock

	handlers []httpapi.RouteRegistrar
}

func NewAppBuilder(
	cfg config.Config,
	logger ports.Logger,
	tel *telemetry.Telemetry,
	pool *pgxpool.Pool,
	flags ports.FeatureFlags,
	auditLogger ports.AuditLogger,
	tokenValidator ports.TokenValidator,
) *AppBuilder {
	eventBus := bus.NewMemoryBus()
	eventBus.Subscribe(audit.NewEventHandler(auditLogger))

	return &AppBuilder{
		cfg:            cfg,
		logger:         logger,
		tel:            tel,
		pool:           pool,
		repos:          repo.NewFactory(pool),
		flags:          flags,
		bus:            eventBus,
		tokenValidator: tokenValidator,
		idGen:          idadapter.NewUUIDGenerator(),
		clk:            clock.NewRealClock(),
	}
}

func (b *AppBuilder) WithItem() *AppBuilder {
	mod := application.NewItemModule(b.repos.NewItemRepository(), b.idGen, b.clk, b.bus, b.flags)
	b.handlers = append(b.handlers, item.NewHandler(mod))
	return b
}

func (b *AppBuilder) Build() (*App, func(), error) {
	h := httpapi.NewHandler(b.cfg, b.handlers,
		httpapi.WithLogger(b.logger),
		httpapi.WithTelemetry(b.tel),
		httpapi.WithTokenValidator(b.tokenValidator),
	)

	cleanup := func() {
		if b.tel != nil {
			_ = b.tel.Shutdown(context.Background())
		}
		if b.pool != nil {
			b.pool.Close()
		}
		_ = b.logger.Sync()
	}

	return &App{handler: h, cfg: b.cfg, logger: b.logger, telemetry: b.tel}, cleanup, nil
}
