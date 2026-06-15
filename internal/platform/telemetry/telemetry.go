package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/flyluman/scratch/internal/platform/config"
)

type Telemetry struct {
	TracerProvider *trace.TracerProvider
	Shutdown       func(context.Context) error
}

func New(ctx context.Context, cfg config.Config) (*Telemetry, error) {
	if !cfg.TelemetryEnabled {
		return nil, nil
	}

	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.TelemetryOTLPEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create otlp exporter: %w", err)
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("scratch"),
		semconv.DeploymentEnvironment(cfg.AppEnv),
	)

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp, trace.WithBatchTimeout(5*time.Second)),
		trace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	return &Telemetry{
		TracerProvider: tp,
		Shutdown: func(ctx context.Context) error {
			return tp.Shutdown(ctx)
		},
	}, nil
}
