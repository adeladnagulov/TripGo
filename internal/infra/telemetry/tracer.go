package telemetry

import (
	"context"
	"log/slog"
	"net/url"

	"github.com/adeladnagulov/TripGo/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

type TracerProvider struct {
	Provider *trace.TracerProvider
}

func NewTracer(ctx context.Context, cfg *config.Config) (*TracerProvider, error) {
	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.OtelServiceName),
		),
	)
	if err != nil {
		return nil, err
	}

	endpointHost := cfg.OtelExporterEndpoint
	if u, err := url.Parse(cfg.OtelExporterEndpoint); err == nil && u.Host != "" {
		endpointHost = u.Host
	}

	traceExp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpointHost),
		otlptracehttp.WithInsecure(),
	)

	if err != nil {
		slog.Warn("trace exporter init failed, traces disabled", "error", err)
		return &TracerProvider{}, nil
	}

	sampler := trace.ParentBased(trace.AlwaysSample())

	tp := trace.NewTracerProvider(
		trace.WithResource(res),
		trace.WithBatcher(traceExp),
		trace.WithSampler(sampler),
	)
	otel.SetTracerProvider(tp)
	return &TracerProvider{Provider: tp}, nil
}

func (tp *TracerProvider) Shutdown(ctx context.Context) {
	if tp.Provider != nil {
		_ = tp.Provider.Shutdown(ctx)
	}
}
